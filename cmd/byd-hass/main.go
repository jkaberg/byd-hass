package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"sync"
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

// All global intervals / timeouts are defined in internal/config/defaults.go.

// SharedState holds application state with proper synchronization
type SharedState struct {
	mu                      sync.RWMutex
	latestSensorData        *sensors.SensorData
	hasUnsentChangesForMQTT bool
	hasUnsentChangesForABRP bool
}

func (s *SharedState) GetLatestData() *sensors.SensorData {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.latestSensorData
}

func (s *SharedState) UpdateData(data *sensors.SensorData) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.latestSensorData = data
	if data != nil {
		s.hasUnsentChangesForMQTT = true
		s.hasUnsentChangesForABRP = true
	}
}

func (s *SharedState) HasUnsentMQTTChanges() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.hasUnsentChangesForMQTT
}

func (s *SharedState) HasUnsentABRPChanges() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.hasUnsentChangesForABRP
}

func (s *SharedState) ClearMQTTFlag() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.hasUnsentChangesForMQTT = false
}

func (s *SharedState) ClearABRPFlag() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.hasUnsentChangesForABRP = false
}

func main() {
	// Parse command line flags
	cfg, debugMode := parseFlags()

	// If debug mode is enabled, run diagnostics and exit
	if debugMode {
		runDebugMode(cfg)
		os.Exit(0)
	}

	// Setup logger
	logger := setupLogger(cfg.Verbose)

	// Use a custom DNS resolver to bypass broken localhost resolvers on Termux/Android
	setupCustomDNSResolver(logger)

	logger.WithFields(logrus.Fields{
		"version":              version,
		"device_id":            cfg.DeviceID,
		"diplus_poll_interval": config.DiplusPollInterval,
		"abrp_interval":        config.ABRPTransmitInterval,
		"mqtt_interval":        config.MQTTTransmitInterval,
	}).Info("Starting BYD-HASS")

	// Create application context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Initialize core components
	diplusURL := fmt.Sprintf("http://%s/api/getDiPars", cfg.DiplusURL)
	diplusClient := api.NewDiplusClient(diplusURL, logger)
	cacheManager := cache.NewManager(logger)

	// Initialize location provider only if ABRP location uploads are enabled
	var locationProvider *location.TermuxLocationProvider
	if cfg.ABRPLocation {
		locationProvider = location.NewTermuxLocationProvider(logger)
		// Ensure location provider is properly shut down
		defer locationProvider.Stop()
	} else {
		logger.Debug("ABRP location upload disabled; skipping GPS initialization")
	}

	// Initialize shared state
	sharedState := &SharedState{
		hasUnsentChangesForMQTT: true, // Transmit on first run
		hasUnsentChangesForABRP: true, // Transmit on first run
	}

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
	if cfg.ABRPAPIKey != "" && cfg.ABRPToken != "" {
		abrpTransmitter = transmission.NewABRPTransmitter(
			cfg.ABRPAPIKey,
			cfg.ABRPToken,
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
	diplusTicker := time.NewTicker(config.DiplusPollInterval)
	defer diplusTicker.Stop()

	var abrpTicker *time.Ticker
	if abrpTransmitter != nil {
		abrpTicker = time.NewTicker(config.ABRPTransmitInterval)
		defer abrpTicker.Stop()
	}

	var mqttTicker *time.Ticker
	if mqttTransmitter != nil {
		mqttTicker = time.NewTicker(config.MQTTTransmitInterval)
		defer mqttTicker.Stop()
	}

	// Initial poll to populate data
	initialDataPollAndTransmit(ctx, diplusClient, locationProvider, cacheManager, sharedState, mqttTransmitter, abrpTransmitter, logger)

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
			// Poll sensor data and flag changes (non-blocking)
			go pollAndFlagChangesAsync(ctx, diplusClient, locationProvider, cacheManager, sharedState, logger)
		case <-func() <-chan time.Time {
			if abrpTicker != nil {
				return abrpTicker.C
			}
			return make(<-chan time.Time)
		}():
			latestData := sharedState.GetLatestData()
			if abrpTransmitter != nil && latestData != nil && sharedState.HasUnsentABRPChanges() {
				// ABRP transmission (non-blocking)
				go func(data *sensors.SensorData) {
					if err := transmitToABRPAsync(ctx, abrpTransmitter, data, logger); err != nil {
						logger.WithError(err).Error("ABRP transmission failed")
					} else {
						sharedState.ClearABRPFlag() // Clear flag on success
					}
				}(latestData)
			}
		case <-func() <-chan time.Time {
			if mqttTicker != nil {
				return mqttTicker.C
			}
			return make(<-chan time.Time)
		}():
			logger.Debug("MQTT ticker triggered")
			latestData := sharedState.GetLatestData()
			if mqttTransmitter == nil {
				logger.Debug("No MQTT transmitter configured, skipping")
			} else if latestData == nil {
				logger.Debug("No sensor data available for MQTT transmission")
			} else if !sharedState.HasUnsentMQTTChanges() {
				logger.Debug("No unsent changes for MQTT, skipping transmission")
			} else {
				// MQTT transmission (non-blocking)
				go func(data *sensors.SensorData) {
					logger.Info("Transmitting sensor data to MQTT...")
					if err := transmitToMQTTAsync(ctx, mqttTransmitter, data, logger); err != nil {
						logger.WithError(err).Error("MQTT transmission failed")
					} else {
						logger.Info("MQTT transmission successful")
						sharedState.ClearMQTTFlag() // Clear flag on success
					}
				}(latestData)
			}
		}
	}
}

func parseFlags() (*config.Config, bool) {
	cfg := config.GetDefaultConfig()

	// Version flag
	showVersion := flag.Bool("version", false, "Show version and exit")

	// Debug flag for comprehensive sensor analysis
	debug := flag.Bool("debug", false, "Run comprehensive sensor debugging and exit")

	flag.StringVar(&cfg.MQTTUrl, "mqtt-url",
		getEnvOrDefault("BYD_HASS_MQTT_URL", cfg.MQTTUrl),
		"MQTT URL (supports ws://, wss://, mqtt://, mqtts://)")

	flag.StringVar(&cfg.DiplusURL, "diplus-url",
		getEnvOrDefault("BYD_HASS_DIPLUS_URL", cfg.DiplusURL),
		"Di-Plus API URL (host:port)")

	flag.StringVar(&cfg.ABRPAPIKey, "abrp-api-key",
		getEnvOrDefault("BYD_HASS_ABRP_API_KEY", cfg.ABRPAPIKey),
		"ABRP API key for telemetry")

	flag.StringVar(&cfg.ABRPToken, "abrp-token",
		getEnvOrDefault("BYD_HASS_ABRP_TOKEN", cfg.ABRPToken),
		"ABRP user token for telemetry")

	flag.StringVar(&cfg.DeviceID, "device-id",
		getEnvOrDefault("BYD_HASS_DEVICE_ID", generateDeviceID()),
		"Unique device identifier")

	flag.BoolVar(&cfg.Verbose, "verbose",
		getEnvOrDefault("BYD_HASS_VERBOSE", "false") == "true",
		"Enable verbose logging")

	flag.StringVar(&cfg.DiscoveryPrefix, "discovery-prefix",
		getEnvOrDefault("BYD_HASS_DISCOVERY_PREFIX", cfg.DiscoveryPrefix),
		"Home Assistant discovery prefix")

	// Disable GPS location provider if explicitly requested (location is ON by default).
	disableLocation := flag.Bool("disable-location",
		getEnvOrDefault("BYD_HASS_DISABLE_LOCATION", "false") == "true",
		"Disable GPS location (Termux) and omit coordinates from MQTT/ABRP telemetry")

	flag.Parse()

	if *showVersion {
		fmt.Printf("byd-hass %s\n", version)
		os.Exit(0)
	}

	// Invert disable flag to set ABRPLocation.
	cfg.ABRPLocation = !*disableLocation

	return cfg, *debug
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func generateDeviceID() string {
	// Use a consistent device ID to avoid creating new Home Assistant devices on each run
	return "byd_car"
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

func pollAndFlagChangesAsync(
	ctx context.Context,
	diplusClient *api.DiplusClient,
	locationProvider *location.TermuxLocationProvider,
	cacheManager *cache.Manager,
	sharedState *SharedState,
	logger *logrus.Logger,
) {
	// Create timeout context for this operation
	timeoutCtx, cancel := context.WithTimeout(ctx, config.DiplusTimeout)
	defer cancel()

	logger.Debug("Starting sensor polling (async)...")
	if err := pollSensorDataAsync(timeoutCtx, diplusClient, locationProvider, cacheManager, sharedState, logger); err != nil {
		logger.WithError(err).Error("Sensor polling failed")
		return
	}
	// The UpdateData call in pollSensorDataAsync already handles flagging changes
	if sharedState.GetLatestData() != nil {
		logger.Info("Sensor data updated, flagged for transmission")
	} else {
		logger.Debug("No sensor data changes detected")
	}
}

func pollSensorDataAsync(
	ctx context.Context,
	diplusClient *api.DiplusClient,
	locationProvider *location.TermuxLocationProvider,
	cacheManager *cache.Manager,
	sharedState *SharedState,
	logger *logrus.Logger,
) error {
	ctx, cancel := context.WithTimeout(ctx, config.DiplusTimeout)
	defer cancel()

	logger.Info("Polling Diplus for new sensor data...")

	// Poll sensor data using the standard poll method
	sensorData, err := diplusClient.Poll()
	if err != nil {
		return fmt.Errorf("failed to poll sensor data: %w", err)
	}

	// ---- START TEMPORARY DEBUG LOG ----
	if logger.IsLevelEnabled(logrus.DebugLevel) {
		jsonData, err := json.Marshal(sensorData)
		if err == nil {
			logger.WithField("raw_polled_data", string(jsonData)).Debug("Dumping raw polled sensor data")
		}
	}
	// ---- END TEMPORARY DEBUG LOG ----

	// Enrich with location data if provider is available
	if locationProvider != nil {
		if locationData, err := locationProvider.GetLocation(); err == nil {
			sensorData.Location = locationData
			logger.WithFields(logrus.Fields{
				"lat":       locationData.Latitude,
				"lon":       locationData.Longitude,
				"accuracy":  locationData.Accuracy,
				"timestamp": locationData.Timestamp,
			}).Info("Enriched with location data")
		} else {
			logger.WithError(err).Debug("Could not get location data")
		}
	}

	if cacheManager.Changed(sensorData) {
		logger.Info("Sensor data has changed, updating state")
		sharedState.UpdateData(sensorData)
	} else {
		logger.Info("Sensor data has not changed, no update needed")
	}

	return nil
}

func transmitToABRPAsync(
	ctx context.Context,
	transmitter *transmission.ABRPTransmitter,
	data *sensors.SensorData,
	logger *logrus.Logger,
) error {
	// Create timeout context for this operation
	timeoutCtx, cancel := context.WithTimeout(ctx, config.ABRPTimeout)
	defer cancel()

	// Channel to receive result
	errorChan := make(chan error, 1)

	// Start transmission in goroutine
	go func() {
		defer close(errorChan)
		if err := transmitter.Transmit(data); err != nil {
			errorChan <- fmt.Errorf("ABRP transmission failed: %w", err)
		}
	}()

	// Wait for completion or timeout
	select {
	case <-timeoutCtx.Done():
		return fmt.Errorf("ABRP transmission timed out after %v", config.ABRPTimeout)
	case err := <-errorChan:
		if err != nil {
			return err
		}
		logger.Debug("Successfully transmitted to ABRP")
		return nil
	}
}

func transmitToMQTTAsync(
	ctx context.Context,
	transmitter *transmission.MQTTTransmitter,
	data *sensors.SensorData,
	logger *logrus.Logger,
) error {
	// Create timeout context for this operation
	timeoutCtx, cancel := context.WithTimeout(ctx, config.MQTTTimeout)
	defer cancel()

	// Channel to receive result
	errorChan := make(chan error, 1)

	// Start transmission in goroutine
	go func() {
		defer close(errorChan)
		if err := transmitter.Transmit(data); err != nil {
			errorChan <- fmt.Errorf("MQTT transmission failed: %w", err)
		}
	}()

	// Wait for completion or timeout
	select {
	case <-timeoutCtx.Done():
		return fmt.Errorf("MQTT transmission timed out after %v", config.MQTTTimeout)
	case err := <-errorChan:
		if err != nil {
			return err
		}
		logger.Debug("Successfully transmitted to MQTT")
		return nil
	}
}

// initialDataPollAndTransmit performs the first poll and transmission synchronously at startup.
func initialDataPollAndTransmit(
	ctx context.Context,
	diplusClient *api.DiplusClient,
	locationProvider *location.TermuxLocationProvider,
	cacheManager *cache.Manager,
	sharedState *SharedState,
	mqttTransmitter *transmission.MQTTTransmitter,
	abrpTransmitter *transmission.ABRPTransmitter,
	logger *logrus.Logger,
) {
	logger.Info("Performing initial data poll...")

	// Poll sensor data using the standard poll method
	sensorData, err := diplusClient.Poll()
	if err != nil {
		logger.WithError(err).Error("Initial poll failed")
		return
	}

	// Enrich with location data if provider is available
	if locationProvider != nil {
		if locationData, err := locationProvider.GetLocation(); err == nil {
			sensorData.Location = locationData
		}
	}

	sharedState.UpdateData(sensorData)
	// Store snapshot in cache manager (always true on first call)
	_ = cacheManager.Changed(sensorData)

	if mqttTransmitter != nil {
		logger.Info("Performing initial MQTT transmission...")
		if err := transmitToMQTTAsync(ctx, mqttTransmitter, sensorData, logger); err != nil {
			logger.WithError(err).Error("Initial MQTT transmission failed")
		}
	}

	if abrpTransmitter != nil {
		logger.Info("Performing initial ABRP transmission...")
		if err := transmitToABRPAsync(ctx, abrpTransmitter, sensorData, logger); err != nil {
			logger.WithError(err).Error("Initial ABRP transmission failed")
		}
	}
}

// runDebugMode runs simple raw vs parsed value comparison
func runDebugMode(cfg *config.Config) {
	logger := setupLogger(true) // Force verbose logging for debug mode
	logger.Info("--- Running in Debug Mode ---")

	diplusURL := fmt.Sprintf("http://%s/api/getDiPars", cfg.DiplusURL)
	diplusClient := api.NewDiplusClient(diplusURL, logger)

	// In debug mode, we might need more time to query all sensors
	diplusClient.SetTimeout(30 * time.Second)

	err := diplusClient.CompareAllSensors()
	if err != nil {
		logger.WithError(err).Fatal("Debug mode failed")
	}

	logger.Info("--- Debug Mode Finished ---")
}

// setupCustomDNSResolver replaces net.DefaultResolver with one that queries
// public DNS servers directly (1.1.1.1 with 8.8.8.8 as fallback).  This
// avoids "connection refused" errors when the local DNS service is not
// running, which is common in some Termux/Android environments.
func setupCustomDNSResolver(logger *logrus.Logger) {
	net.DefaultResolver = &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{Timeout: 5 * time.Second}
			conn, err := d.DialContext(ctx, "udp", "1.1.1.1:53")
			if err != nil {
				logger.WithError(err).Warn("Primary DNS (1.1.1.1) failed, falling back to 8.8.8.8")
				return d.DialContext(ctx, "udp", "8.8.8.8:53")
			}
			return conn, nil
		},
	}
}
