package config

import (
	"fmt"
	"strings"
	"time"
)

// Config holds all configuration options for the BYD-HASS application
type Config struct {
	// MQTT Configuration
	MQTTUrl         string `json:"mqtt_url"`         // MQTT URL (supports both WebSocket and standard MQTT)
	DiscoveryPrefix string `json:"discovery_prefix"` // Home Assistant discovery prefix

	// ABRP Configuration
	ABRPAPIKey string `json:"abrp_api_key"` // ABRP API key
	ABRPToken  string `json:"abrp_token"`   // ABRP user token

	// Device Configuration
	DeviceID string `json:"device_id"` // Unique device identifier

	// Application Configuration
	Verbose bool `json:"verbose"` // Enable verbose logging

	// API Configuration
	DiplusURL       string `json:"diplus_url"`       // Di-Plus API URL
	ExtendedPolling bool   `json:"extended_polling"` // Use extended sensor polling for more data
	APITimeout      int    `json:"api_timeout"`      // API request timeout in seconds (default: 10)

	// ABRP Configuration
	ABRPEnhanced    bool   `json:"abrp_enhanced"`     // Use enhanced ABRP telemetry data
	ABRPLocation    bool   `json:"abrp_location"`     // Include GPS location in ABRP data (if available)
	ABRPVehicleType string `json:"abrp_vehicle_type"` // ABRP vehicle type for better range estimation
}

// GetDefaultConfig returns a configuration with sensible defaults
func GetDefaultConfig() *Config {
	return &Config{
		DiscoveryPrefix: "homeassistant",
		DeviceID:        "", // Will be auto-generated
		Verbose:         false,
		DiplusURL:       "localhost:8988",

		ExtendedPolling: true,    // Enable extended polling by default
		APITimeout:      10,      // 10 second API timeout
		ABRPEnhanced:    true,    // Use enhanced ABRP data by default
		ABRPLocation:    true,    // Location ENABLED by default
		ABRPVehicleType: "byd:*", // Generic BYD vehicle type
	}
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	// Basic validation
	if c.DeviceID == "" {
		return fmt.Errorf("device ID is required")
	}

	// MQTT validation - support both WebSocket and standard MQTT protocols
	if c.MQTTUrl != "" {
		if !strings.HasPrefix(c.MQTTUrl, "ws://") &&
			!strings.HasPrefix(c.MQTTUrl, "wss://") &&
			!strings.HasPrefix(c.MQTTUrl, "mqtt://") &&
			!strings.HasPrefix(c.MQTTUrl, "mqtts://") {
			return fmt.Errorf("MQTT URL must use supported protocol (ws://, wss://, mqtt://, or mqtts://)")
		}
	}

	// ABRP validation
	if c.ABRPAPIKey != "" && c.ABRPToken == "" {
		return fmt.Errorf("ABRP token is required when API key is provided")
	}
	if c.ABRPToken != "" && c.ABRPAPIKey == "" {
		return fmt.Errorf("ABRP API key is required when token is provided")
	}

	// Set defaults for invalid values
	if c.APITimeout <= 0 {
		c.APITimeout = 10 // Set default
	}

	return nil
}

// HasMQTT returns true if MQTT is configured
func (c *Config) HasMQTT() bool {
	return c.MQTTUrl != ""
}

// HasABRP returns true if ABRP is configured
func (c *Config) HasABRP() bool {
	return c.ABRPAPIKey != "" && c.ABRPToken != ""
}

// GetAPITimeout returns the API timeout as a duration
func (c *Config) GetAPITimeout() time.Duration {
	return time.Duration(c.APITimeout) * time.Second
}
