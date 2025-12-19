package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/jkaberg/byd-hass/internal/api"
	"github.com/jkaberg/byd-hass/internal/app"
	"github.com/jkaberg/byd-hass/internal/config"
	"github.com/jkaberg/byd-hass/internal/location"
	"github.com/jkaberg/byd-hass/internal/mqtt"
	"github.com/jkaberg/byd-hass/internal/transmission"
	"github.com/sirupsen/logrus"
)

// version is injected at build time via ldflags
var version = "dev"

func main() {
	cfg, debugMode := parseFlags()

	// Debug path ------------------------------------------------------------------
	if debugMode {
		runDebugMode(cfg)
		return
	}

	logger := setupLogger(cfg.Verbose)
	setupCustomDNSResolver(logger)

	logFields := logrus.Fields{
		"version":   version,
		"device_id": cfg.DeviceID,
		"poll":      config.DiplusPollInterval,
		"abrp_int":  cfg.ABRPInterval,
		"mqtt_int":  cfg.MQTTInterval,
	}
	if cfg.ForceUpdateInterval > 0 {
		logFields["force_update_int"] = cfg.ForceUpdateInterval
	}
	logger.WithFields(logFields).Info("Starting BYD-HASS v2")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sig
		logger.Info("Shutdown signal received")
		cancel()
	}()

	// Core clients ---------------------------------------------------------------
	diplusURL := fmt.Sprintf("http://%s/api/getDiPars", cfg.DiplusURL)
	diplusClient := api.NewDiplusClient(diplusURL, logger)

	var locProvider *location.TermuxLocationProvider
	if cfg.ABRPLocation {
		locProvider = location.NewTermuxLocationProvider(logger)
		defer locProvider.Stop()
	}

	// Transmitters ---------------------------------------------------------------
	var mqttTx *transmission.MQTTTransmitter
	if cfg.MQTTUrl != "" {
		mqttClient, err := mqtt.NewClient(cfg.MQTTUrl, cfg.DeviceID, logger)
		if err != nil {
			logger.WithError(err).Fatal("Failed to create MQTT client")
		}
		mqttTx = transmission.NewMQTTTransmitter(mqttClient, cfg.DeviceID, cfg.DiscoveryPrefix, logger)
		logger.Info("MQTT transmitter ready")
	}

	var abrpTx *transmission.ABRPTransmitter
	if cfg.ABRPAPIKey != "" && cfg.ABRPToken != "" {
		abrpTx = transmission.NewABRPTransmitter(cfg.ABRPAPIKey, cfg.ABRPToken, logger)
		logger.WithField("abrp_status", abrpTx.GetConnectionStatus()).Info("ABRP transmitter ready")
	}

	if mqttTx == nil && abrpTx == nil {
		logger.Warn("No transmitters configured; data will only be logged")
	}

	// Run application ------------------------------------------------------------
	app.Run(ctx, cfg, diplusClient, locProvider, mqttTx, abrpTx, logger)

	<-ctx.Done()
	logger.Info("BYD-HASS stopped")
}

// -----------------------------------------------------------------------------
// Helpers & Flags
// -----------------------------------------------------------------------------

func parseFlags() (*config.Config, bool) {
	cfg := config.GetDefaultConfig()

	showVersion := flag.Bool("version", false, "Show version and exit")
	debug := flag.Bool("debug", false, "Run comprehensive sensor debugging and exit")

	flag.StringVar(&cfg.MQTTUrl, "mqtt-url", getEnv("BYD_HASS_MQTT_URL", cfg.MQTTUrl), "MQTT URL")
	flag.StringVar(&cfg.DiplusURL, "diplus-url", getEnv("BYD_HASS_DIPLUS_URL", cfg.DiplusURL), "Di-Plus host:port")
	flag.StringVar(&cfg.ABRPAPIKey, "abrp-api-key", getEnv("BYD_HASS_ABRP_API_KEY", cfg.ABRPAPIKey), "ABRP API key")
	flag.StringVar(&cfg.ABRPToken, "abrp-token", getEnv("BYD_HASS_ABRP_TOKEN", cfg.ABRPToken), "ABRP user token")
	flag.StringVar(&cfg.DeviceID, "device-id", getEnv("BYD_HASS_DEVICE_ID", generateDeviceID()), "Device identifier")
	flag.BoolVar(&cfg.Verbose, "verbose", getEnv("BYD_HASS_VERBOSE", "false") == "true", "Verbose logging")
	flag.StringVar(&cfg.DiscoveryPrefix, "discovery-prefix", getEnv("BYD_HASS_DISCOVERY_PREFIX", cfg.DiscoveryPrefix), "HA discovery prefix")

	mqttIntervalStr := flag.String("mqtt-interval", getEnv("BYD_HASS_MQTT_INTERVAL", ""), "MQTT interval (e.g. 60s)")
	abrpIntervalStr := flag.String("abrp-interval", getEnv("BYD_HASS_ABRP_INTERVAL", ""), "ABRP interval (e.g. 10s)")
	forceUpdateIntervalStr := flag.String("force-update-interval", getEnv("BYD_HASS_FORCE_UPDATE_INTERVAL", ""), "Force update all sensors at this interval even if unchanged (e.g. 10m, 0 = disabled)")

	flag.Parse()

	if *showVersion {
		fmt.Printf("byd-hass %s\n", version)
		os.Exit(0)
	}

	// Duration overrides
	if *mqttIntervalStr != "" {
		if d, err := time.ParseDuration(*mqttIntervalStr); err == nil && d > 0 {
			cfg.MQTTInterval = d
		} else if v, err2 := strconv.Atoi(*mqttIntervalStr); err2 == nil && v > 0 {
			cfg.MQTTInterval = time.Duration(v) * time.Second
		}
	}
	if *abrpIntervalStr != "" {
		if d, err := time.ParseDuration(*abrpIntervalStr); err == nil && d > 0 {
			cfg.ABRPInterval = d
		} else if v, err2 := strconv.Atoi(*abrpIntervalStr); err2 == nil && v > 0 {
			cfg.ABRPInterval = time.Duration(v) * time.Second
		}
	}
	if *forceUpdateIntervalStr != "" {
		if d, err := time.ParseDuration(*forceUpdateIntervalStr); err == nil && d >= 0 {
			cfg.ForceUpdateInterval = d
		} else if v, err2 := strconv.Atoi(*forceUpdateIntervalStr); err2 == nil && v >= 0 {
			cfg.ForceUpdateInterval = time.Duration(v) * time.Second
		}
	}

	return cfg, *debug
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func generateDeviceID() string { return "byd_car" }

func setupLogger(verbose bool) *logrus.Logger {
	l := logrus.New()
	l.SetFormatter(&logrus.TextFormatter{FullTimestamp: true, TimestampFormat: time.RFC3339})
	if verbose {
		l.SetLevel(logrus.DebugLevel)
	} else {
		l.SetLevel(logrus.InfoLevel)
	}
	return l
}

func setupCustomDNSResolver(logger *logrus.Logger) {
	net.DefaultResolver = &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{Timeout: time.Second}
			return d.DialContext(ctx, network, "1.1.1.1:53")
		},
	}
	logger.Debug("Custom DNS resolver installed (1.1.1.1)")
}

func runDebugMode(cfg *config.Config) {
	logger := setupLogger(true)
	diplusURL := fmt.Sprintf("http://%s/api/getDiPars", cfg.DiplusURL)
	client := api.NewDiplusClient(diplusURL, logger)
	if err := client.CompareAllSensors(); err != nil {
		logger.WithError(err).Fatal("Debug mode failed")
	}
	os.Exit(0)
}
