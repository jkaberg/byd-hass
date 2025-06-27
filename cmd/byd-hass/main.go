package main

import (
	"context"
	"flag"
	"fmt"
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

// Application intervals
const (
	DiplusPollInterval   = 15 * time.Second
	ABRPTransmitInterval = 10 * time.Second
	MQTTTransmitInterval = 60 * time.Second
)

// Operation timeouts to prevent blocking
const (
	DiplusTimeout = 8 * time.Second // Timeout for Diplus API calls
	MQTTTimeout   = 5 * time.Second // Timeout for MQTT operations
	ABRPTimeout   = 8 * time.Second // Timeout for ABRP API calls
)

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

	// Ensure location provider is properly shut down
	defer locationProvider.Stop()

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

	// Initial poll to populate data (non-blocking)
	go pollAndFlagChangesAsync(ctx, diplusClient, locationProvider, cacheManager, sharedState, logger)

	// For MQTT, try to publish discovery config immediately, even without sensor data
	if mqttTransmitter != nil {
		go func() {
			logger.Info("Publishing MQTT discovery configuration...")
			// Create empty sensor data just to trigger discovery config publishing
			emptySensorData := &sensors.SensorData{
				Timestamp: time.Now(),
			}
			if err := transmitToMQTTAsync(ctx, mqttTransmitter, emptySensorData, logger); err != nil {
				logger.WithError(err).Warn("Initial MQTT discovery config publishing failed, will retry with real data")
			} else {
				logger.Info("MQTT discovery configuration published successfully")
			}
		}()
	}

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

func parseFlags() *config.Config {
	cfg := &config.Config{}

	// Version flag
	showVersion := flag.Bool("version", false, "Show version and exit")

	// Debug flag for comprehensive sensor analysis
	debug := flag.Bool("debug", false, "Run comprehensive sensor debugging and exit")

	flag.StringVar(&cfg.MQTTUrl, "mqtt-url",
		getEnvOrDefault("BYD_HASS_MQTT_URL", ""),
		"MQTT URL (supports ws://, wss://, mqtt://, mqtts://)")

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

	if *showVersion {
		fmt.Printf("byd-hass %s\n", version)
		os.Exit(0)
	}

	// Handle debug mode
	if *debug {
		runDebugMode()
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
	timeoutCtx, cancel := context.WithTimeout(ctx, DiplusTimeout)
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
	// Channel to receive sensor data
	sensorDataChan := make(chan *sensors.SensorData, 1)
	errorChan := make(chan error, 1)

	// Start sensor data fetch in goroutine
	go func() {
		defer close(sensorDataChan)
		defer close(errorChan)

		// Try extended polling first, fallback to default if needed
		sensorData, err := diplusClient.GetExtendedSensorData()
		if err != nil {
			logger.WithError(err).Debug("Extended polling failed, using default polling")
			sensorData, err = diplusClient.GetDefaultSensorData()
			if err != nil {
				errorChan <- fmt.Errorf("failed to poll sensor data: %w", err)
				return
			}
		}

		// Fetch location data (non-blocking from cache)
		locationData, err := locationProvider.GetLocation()
		if err != nil {
			// Log as debug since location is fetched in background and may not be available immediately
			logger.WithError(err).Debug("Location data not available yet")
		} else {
			sensorData.Location = locationData
			logger.WithFields(logrus.Fields{
				"lat":      locationData.Latitude,
				"lon":      locationData.Longitude,
				"accuracy": locationData.Accuracy,
			}).Debug("Location data included in sensor data")
		}

		sensorDataChan <- sensorData
	}()

	// Wait for completion or timeout
	select {
	case <-ctx.Done():
		return fmt.Errorf("sensor polling timed out after %v", DiplusTimeout)
	case err := <-errorChan:
		if err != nil {
			return err
		}
	case sensorData := <-sensorDataChan:
		logger.WithField("sensors_active", len(sensors.GetNonNilFields(sensorData))).Debug("Polled sensor data")

		// Check for changes and cache
		changes := cacheManager.GetChanges(sensorData)
		if len(changes) == 0 {
			logger.Debug("No sensor changes detected")
			// Explicitly nil the pointer if no changes, so we don't re-transmit old data
			sharedState.UpdateData(nil)
			return nil
		}

		logger.WithField("changed_sensors", len(changes)).Info("Sensor changes detected")

		// Update latest sensor data
		sharedState.UpdateData(sensorData)
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
	timeoutCtx, cancel := context.WithTimeout(ctx, ABRPTimeout)
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
		return fmt.Errorf("ABRP transmission timed out after %v", ABRPTimeout)
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
	timeoutCtx, cancel := context.WithTimeout(ctx, MQTTTimeout)
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
		return fmt.Errorf("MQTT transmission timed out after %v", MQTTTimeout)
	case err := <-errorChan:
		if err != nil {
			return err
		}
		logger.Debug("Successfully transmitted to MQTT")
		return nil
	}
}

// runDebugMode runs simple raw vs parsed value comparison
func runDebugMode() {
	logger := setupLogger(true) // Always verbose for debug mode
	logger.Info("Starting BYD-HASS Raw vs Parsed Value Comparison")

	// Create Diplus client
	diplusClient := api.NewDiplusClient("http://localhost:8988/api/getDiPars", logger)

	logger.Info("Comparing raw API values vs parsed sensor data...")

	// Run comparison
	if err := diplusClient.CompareAllSensors(); err != nil {
		logger.WithError(err).Fatal("Failed to compare sensor values")
	}

	fmt.Println("\n✅ Comparison complete!")
	fmt.Println("Review the output above to identify type mismatches and parsing failures.")
}
