package transmission

import "github.com/jkaberg/byd-hass/internal/sensors"

// Transmitter defines the interface for transmitting sensor data
type Transmitter interface {
	Transmit(data *sensors.SensorData) error
	IsConnected() bool
}
