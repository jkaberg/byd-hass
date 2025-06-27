package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jkaberg/byd-hass/internal/api"
	"github.com/jkaberg/byd-hass/internal/cache"
	"github.com/jkaberg/byd-hass/internal/config"
	"github.com/jkaberg/byd-hass/internal/location"
	"github.com/jkaberg/byd-hass/internal/mqtt"
	"github.com/jkaberg/byd-hass/internal/sensors"
	"github.com/jkaberg/byd-hass/internal/transmission"
	"github.com/sirupsen/logrus"
)

// version is injected at build time via ldflags
var version = "dev"

// Application intervals
const (
	DiplusPollInterval   = 15 * time.Second
	ABRPTransmitInterval = 10 * time.Second
	MQTTTransmitInterval = 60 * time.Second
)

func main() {
	// Parse command line flags
	cfg := parseFlags()

	// Setup logger
	logger := setupLogger(cfg.Verbose)

	logger.WithFields(logrus.Fields{
		"version":              version,
		"device_id":            cfg.DeviceID,
		"diplus_poll_interval": DiplusPollInterval,
		"abrp_interval":        ABRPTransmitInterval,
		"mqtt_interval":        MQTTTransmitInterval,
	}).Info("Starting BYD-HASS")

	// Create application context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Initialize core components
	diplusClient := api.NewDiplusClient("http://localhost:8988/api/getDiPars", logger)
	cacheManager := cache.NewManager(logger)
	locationProvider := location.NewTermuxLocationProvider(logger)

	var mqttTransmitter *transmission.MQTTTransmitter
	var abrpTransmitter *transmission.ABRPTransmitter
	// Setup MQTT transmitter if configured
	if cfg.MQTTUrl != "" {
		mqttClient, err := mqtt.NewClient(cfg.MQTTUrl, cfg.DeviceID, logger)
		if err != nil {
			logger.WithError(err).Fatal("Failed to create MQTT client")
		}

		mqttTransmitter = transmission.NewMQTTTransmitter(
			mqttClient,
			cfg.DeviceID,
			cfg.DiscoveryPrefix,
			logger,
		)

		logger.Info("MQTT transmitter configured")
	}

	// Setup ABRP transmitter if configured
	if cfg.ABRPAPIKey != "" && cfg.ABRPVehicleKey != "" {
		abrpTransmitter = transmission.NewABRPTransmitter(
			cfg.ABRPAPIKey,
			cfg.ABRPVehicleKey,
			logger,
		)

		// Log ABRP connection status
		status := abrpTransmitter.GetConnectionStatus()
		logger.WithField("abrp_status", status).Info("ABRP transmitter configured")
	}

	if mqttTransmitter == nil && abrpTransmitter == nil {
		logger.Warn("No transmitters configured, data will only be cached")
	}

	// Create tickers for different intervals
	diplusTicker := time.NewTicker(DiplusPollInterval)
	defer diplusTicker.Stop()

	var abrpTicker *time.Ticker
	if abrpTransmitter != nil {
		abrpTicker = time.NewTicker(ABRPTransmitInterval)
		defer abrpTicker.Stop()
	}

	var mqttTicker *time.Ticker
	if mqttTransmitter != nil {
		mqttTicker = time.NewTicker(MQTTTransmitInterval)
		defer mqttTicker.Stop()
	}

	// Store latest sensor data for transmission
	var latestSensorData *sensors.SensorData

	// Flags to track if there are new changes to be sent
	var hasUnsentChangesForMQTT = true // Transmit on first run
	var hasUnsentChangesForABRP = true // Transmit on first run

	// Initial poll to populate data
	pollAndFlagChanges(diplusClient, locationProvider, cacheManager, &latestSensorData, &hasUnsentChangesForMQTT, &hasUnsentChangesForABRP, logger)

	logger.Info("BYD-HASS started successfully")

	// Main loop with multiple tickers
	for {
		select {
		case <-ctx.Done():
			logger.Info("Application context cancelled")
			return
		case <-sigChan:
			logger.Info("Received termination signal, shutting down...")
			cancel()
			return
		case <-diplusTicker.C:
			// Poll sensor data and flag changes
			pollAndFlagChanges(diplusClient, locationProvider, cacheManager, &latestSensorData, &hasUnsentChangesForMQTT, &hasUnsentChangesForABRP, logger)
		case <-func() <-chan time.Time {
			if abrpTicker != nil {
				return abrpTicker.C
			}
			return make(<-chan time.Time)
		}():
			if abrpTransmitter != nil && latestSensorData != nil && hasUnsentChangesForABRP {
				if err := transmitToABRP(abrpTransmitter, latestSensorData, logger); err != nil {
					logger.WithError(err).Error("ABRP transmission failed")
				} else {
					hasUnsentChangesForABRP = false // Clear flag on success
				}
			}
		case <-func() <-chan time.Time {
			if mqttTicker != nil {
				return mqttTicker.C
			}
			return make(<-chan time.Time)
		}():
			if mqttTransmitter != nil && latestSensorData != nil && hasUnsentChangesForMQTT {
				if err := transmitToMQTT(mqttTransmitter, latestSensorData, logger); err != nil {
					logger.WithError(err).Error("MQTT transmission failed")
				} else {
					hasUnsentChangesForMQTT = false // Clear flag on success
				}
			}
		}
	}
}

func parseFlags() *config.Config {
	cfg := &config.Config{}

	// Version flag
	showVersion := flag.Bool("version", false, "Show version and exit")

	flag.StringVar(&cfg.MQTTUrl, "mqtt-url",
		getEnvOrDefault("BYD_HASS_MQTT_URL", ""),
		"MQTT WebSocket URL (ws://user:pass@host:port/path)")

	flag.StringVar(&cfg.ABRPAPIKey, "abrp-api-key",
		getEnvOrDefault("BYD_HASS_ABRP_API_KEY", ""),
		"ABRP API key for telemetry")

	flag.StringVar(&cfg.ABRPVehicleKey, "abrp-vehicle-key",
		getEnvOrDefault("BYD_HASS_ABRP_VEHICLE_KEY", ""),
		"ABRP vehicle identifier key")

	flag.StringVar(&cfg.DeviceID, "device-id",
		getEnvOrDefault("BYD_HASS_DEVICE_ID", generateDeviceID()),
		"Unique device identifier")

	flag.BoolVar(&cfg.Verbose, "verbose",
		getEnvOrDefault("BYD_HASS_VERBOSE", "false") == "true",
		"Enable verbose logging")

	flag.StringVar(&cfg.DiscoveryPrefix, "discovery-prefix",
		getEnvOrDefault("BYD_HASS_DISCOVERY_PREFIX", "homeassistant"),
		"Home Assistant discovery prefix")

	flag.Parse()

	// Handle version flag
	if *showVersion {
		fmt.Printf("BYD-HASS version %s\n", version)
		os.Exit(0)
	}

	return cfg
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func generateDeviceID() string {
	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "unknown"
	}
	return fmt.Sprintf("byd_%s_%d", hostname, time.Now().Unix())
}

func setupLogger(verbose bool) *logrus.Logger {
	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: time.RFC3339,
	})

	if verbose {
		logger.SetLevel(logrus.DebugLevel)
	} else {
		logger.SetLevel(logrus.InfoLevel)
	}

	return logger
}

func pollAndFlagChanges(
	diplusClient *api.DiplusClient,
	locationProvider *location.TermuxLocationProvider,
	cacheManager *cache.Manager,
	latestData **sensors.SensorData,
	hasUnsentChangesForMQTT *bool,
	hasUnsentChangesForABRP *bool,
	logger *logrus.Logger,
) {
	if err := pollSensorData(diplusClient, locationProvider, cacheManager, latestData, logger); err != nil {
		logger.WithError(err).Error("Sensor polling failed")
		return
	}
	// If pollSensorData updated latestData, it means there were changes
	if *latestData != nil {
		*hasUnsentChangesForMQTT = true
		*hasUnsentChangesForABRP = true
	}
}

func pollSensorData(
	diplusClient *api.DiplusClient,
	locationProvider *location.TermuxLocationProvider,
	cacheManager *cache.Manager,
	latestData **sensors.SensorData,
	logger *logrus.Logger,
) error {
	// Try extended polling first, fallback to default if needed
	sensorData, err := diplusClient.GetExtendedSensorData()
	if err != nil {
		logger.WithError(err).Debug("Extended polling failed, using default polling")
		sensorData, err = diplusClient.GetDefaultSensorData()
		if err != nil {
			return fmt.Errorf("failed to poll sensor data: %w", err)
		}
	}

	// Fetch location data
	locationData, err := locationProvider.GetLocation()
	if err != nil {
		// Log as a warning instead of an error, so the app can continue without location
		logger.WithError(err).Warn("Could not fetch location data")
	} else {
		sensorData.Location = locationData
	}

	logger.WithField("sensors_active", len(sensors.GetNonNilFields(sensorData))).Debug("Polled sensor data")

	// Check for changes and cache
	changes := cacheManager.GetChanges(sensorData)
	if len(changes) == 0 {
		logger.Debug("No sensor changes detected")
		// Explicitly nil the pointer if no changes, so we don't re-transmit old data
		*latestData = nil
		return nil
	}

	logger.WithField("changed_sensors", len(changes)).Info("Sensor changes detected")

	// Update latest sensor data
	*latestData = sensorData

	return nil
}

func transmitToABRP(
	transmitter *transmission.ABRPTransmitter,
	data *sensors.SensorData,
	logger *logrus.Logger,
) error {
	if err := transmitter.Transmit(data); err != nil {
		return fmt.Errorf("ABRP transmission failed: %w", err)
	}
	logger.Debug("Successfully transmitted to ABRP")
	return nil
}

func transmitToMQTT(
	transmitter *transmission.MQTTTransmitter,
	data *sensors.SensorData,
	logger *logrus.Logger,
) error {
	if err := transmitter.Transmit(data); err != nil {
		return fmt.Errorf("MQTT transmission failed: %w", err)
	}
	logger.Debug("Successfully transmitted to MQTT")
	return nil
}
