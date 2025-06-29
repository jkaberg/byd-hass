package config

import "time"

// Central place for all application-wide timing constants and other defaults.
// Changing a value here immediately affects all components that import
// github.com/jkaberg/byd-hass/internal/config.

const (
	// Polling / transmission intervals
	DiplusPollInterval   = 8 * time.Second  // Poll local DiPlus API
	ABRPTransmitInterval = 10 * time.Second // Push data to ABRP (HTTP)
	MQTTTransmitInterval = 60 * time.Second // Publish data to MQTT

	// Operation time-outs (to avoid blocking goroutines)
	DiplusTimeout = 8 * time.Second // DiPlus API call
	MQTTTimeout   = 5 * time.Second // MQTT publish
	ABRPTimeout   = 8 * time.Second // ABRP HTTP call

	// Cache
	DefaultCacheTTL = time.Hour
)
