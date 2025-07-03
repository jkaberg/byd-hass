package config

import "time"

const (
	// Polling / transmission intervals
	DiplusPollInterval   = 8 * time.Second  // Poll local DiPlus API
	ABRPTransmitInterval = 10 * time.Second // Push data to ABRP (HTTP)
	MQTTTransmitInterval = 60 * time.Second // Publish data to MQTT

	// Operation time-outs (to avoid blocking goroutines)
	DiplusTimeout = 3 * time.Second // DiPlus API call
	MQTTTimeout   = 5 * time.Second // MQTT publish
	ABRPTimeout   = 4 * time.Second // ABRP HTTP call

)
