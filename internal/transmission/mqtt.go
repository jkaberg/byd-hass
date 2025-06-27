package transmission

import (
	"encoding/json"
	"fmt"

	"github.com/jkaberg/byd-hass/internal/mqtt"
	"github.com/jkaberg/byd-hass/internal/sensors"
	"github.com/sirupsen/logrus"
)

// MQTTTransmitter transmits sensor data via MQTT
type MQTTTransmitter struct {
	client             *mqtt.Client
	deviceID           string
	discoveryPrefix    string
	logger             *logrus.Logger
	discoveryPublished bool
}

// HADiscoveryConfig represents Home Assistant MQTT discovery configuration
type HADiscoveryConfig struct {
	Name              string   `json:"name"`
	UniqueID          string   `json:"unique_id"`
	StateTopic        string   `json:"state_topic"`
	ValueTemplate     string   `json:"value_template,omitempty"`
	DeviceClass       string   `json:"device_class,omitempty"`
	UnitOfMeasurement string   `json:"unit_of_measurement,omitempty"`
	Device            HADevice `json:"device"`
	AvailabilityTopic string   `json:"availability_topic"`
	Icon              string   `json:"icon,omitempty"`
	StateClass        string   `json:"state_class,omitempty"`
}

// HADevice represents the device information for Home Assistant
type HADevice struct {
	Identifiers  []string `json:"identifiers"`
	Name         string   `json:"name"`
	Model        string   `json:"model"`
	Manufacturer string   `json:"manufacturer"`
	SWVersion    string   `json:"sw_version,omitempty"`
}

// NewMQTTTransmitter creates a new MQTT transmitter
func NewMQTTTransmitter(client *mqtt.Client, deviceID, discoveryPrefix string, logger *logrus.Logger) *MQTTTransmitter {
	return &MQTTTransmitter{
		client:          client,
		deviceID:        deviceID,
		discoveryPrefix: discoveryPrefix,
		logger:          logger,
	}
}

// publishDiscoveryConfig publishes Home Assistant discovery configuration
func (t *MQTTTransmitter) publishDiscoveryConfig() error {
	device := HADevice{
		Identifiers:  []string{fmt.Sprintf("byd_car_%s", t.deviceID)},
		Name:         "BYD Car",
		Model:        "BYD Vehicle",
		Manufacturer: "BYD",
		SWVersion:    "1.0.0",
	}

	baseTopic := fmt.Sprintf("byd_car/%s", t.deviceID)

	// Publish device_tracker discovery first
	if err := t.publishDeviceTrackerDiscovery(baseTopic, device); err != nil {
		return fmt.Errorf("failed to publish device_tracker discovery config: %w", err)
	}

	// Sensor configs with detailed information
	sensorConfigs := []struct {
		name        string
		entityID    string
		entityType  string
		deviceClass string
		unit        string
		icon        string
	}{
		{"Battery Percentage", "soc", "sensor", "battery", "%", "mdi:battery"},
		{"Speed", "speed", "sensor", "", "km/h", "mdi:speedometer"},
		{"Mileage", "mileage", "sensor", "distance", "km", "mdi:counter"},
		{"Charging Status", "charging", "binary_sensor", "battery_charging", "", "mdi:battery-charging"},
		{"Door Lock", "lock", "binary_sensor", "lock", "", "mdi:lock"},
		{"Engine Running", "engine", "binary_sensor", "running", "", "mdi:engine"},
		{"Inside Temperature", "inside_temp", "sensor", "temperature", "°C", "mdi:thermometer"},
		{"Outside Temperature", "outside_temp", "sensor", "temperature", "°C", "mdi:thermometer"},
		{"Battery Temperature", "battery_temp", "sensor", "temperature", "°C", "mdi:battery-alert"},
		{"Max Battery Temperature", "max_battery_temp", "sensor", "temperature", "°C", "mdi:battery-alert"},
		{"Min Battery Temperature", "min_battery_temp", "sensor", "temperature", "°C", "mdi:battery-alert"},
		{"Max Battery Voltage", "max_battery_voltage", "sensor", "voltage", "V", "mdi:battery-plus"},
		{"Min Battery Voltage", "min_battery_voltage", "sensor", "voltage", "V", "mdi:battery-minus"},
		{"Battery Capacity", "battery_capacity", "sensor", "energy", "kWh", "mdi:battery-outline"},
		{"Power Consumption 100km", "power_consumption_100km", "sensor", "energy", "kWh/100km", "mdi:car-electric"},
		{"Total Power Consumption", "total_power_consumption", "sensor", "energy", "kWh", "mdi:lightning-bolt"},
		{"Fuel Level", "fuel", "sensor", "", "%", "mdi:gas-station"},
	}

	// Publish sensor discovery configs
	for _, sensor := range sensorConfigs {
		config := HADiscoveryConfig{
			Name:              fmt.Sprintf("BYD Car %s", sensor.name),
			UniqueID:          fmt.Sprintf("byd_car_%s_%s", t.deviceID, sensor.entityID),
			StateTopic:        fmt.Sprintf("%s/state", baseTopic),
			ValueTemplate:     fmt.Sprintf("{{ value_json.%s | default(0) }}", sensor.entityID),
			AvailabilityTopic: fmt.Sprintf("%s/availability", baseTopic),
			Device:            device,
		}

		if sensor.deviceClass != "" {
			config.DeviceClass = sensor.deviceClass
		}
		if sensor.unit != "" {
			config.UnitOfMeasurement = sensor.unit
		}
		if sensor.icon != "" {
			config.Icon = sensor.icon
		}

		topic := fmt.Sprintf("%s/%s/byd_car_%s/%s/config",
			t.discoveryPrefix, sensor.entityType, t.deviceID, sensor.entityID)

		if err := t.publishConfigRaw(topic, config); err != nil {
			return fmt.Errorf("failed to publish %s discovery config: %w", sensor.name, err)
		}
	}

	return nil
}

// publishConfigRaw publishes a raw configuration object
func (t *MQTTTransmitter) publishConfigRaw(topic string, config interface{}) error {
	payload, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	return t.client.Publish(topic, payload, true)
}

// buildStatePayload creates enhanced state payload with more sensor data
func (t *MQTTTransmitter) buildStatePayload(data *sensors.SensorData) ([]byte, error) {
	state := map[string]interface{}{
		"timestamp": data.Timestamp.Unix(),
	}

	// Core vehicle data
	if data.BatteryPercentage != nil {
		state["soc"] = *data.BatteryPercentage
	}
	if data.Speed != nil {
		state["speed"] = *data.Speed
	}
	if data.Mileage != nil {
		state["mileage"] = *data.Mileage
	}

	// Charging and power
	if data.ChargingStatus != nil {
		state["charging"] = *data.ChargingStatus
	}
	if data.ChargeGunState != nil {
		state["charge_gun_connected"] = *data.ChargeGunState == 2
	}
	if data.FuelPercentage != nil {
		state["fuel"] = *data.FuelPercentage
	}

	// Battery voltage data
	if data.MaxBatteryVoltage != nil {
		state["max_battery_voltage"] = *data.MaxBatteryVoltage
	}
	if data.MinBatteryVoltage != nil {
		state["min_battery_voltage"] = *data.MinBatteryVoltage
	}

	// Battery capacity and power consumption
	if data.BatteryCapacity != nil {
		state["battery_capacity"] = *data.BatteryCapacity
	}
	if data.PowerConsumption100km != nil {
		state["power_consumption_100km"] = *data.PowerConsumption100km
	}
	if data.TotalPowerConsumption != nil {
		state["total_power_consumption"] = *data.TotalPowerConsumption
	}

	// Temperature data
	if data.CabinTemperature != nil {
		state["inside_temp"] = *data.CabinTemperature
	}
	if data.OutsideTemperature != nil {
		state["outside_temp"] = *data.OutsideTemperature
	}
	if data.AvgBatteryTemp != nil {
		state["battery_temp"] = *data.AvgBatteryTemp
	}
	if data.MaxBatteryTemp != nil {
		state["max_battery_temp"] = *data.MaxBatteryTemp
	}
	if data.MinBatteryTemp != nil {
		state["min_battery_temp"] = *data.MinBatteryTemp
	}

	// Door and lock status
	if data.RemoteLockStatus != nil {
		state["door_lock"] = *data.RemoteLockStatus
	}
	if data.DriverDoor != nil {
		state["driver_door"] = *data.DriverDoor == 1
	}
	if data.PassengerDoor != nil {
		state["passenger_door"] = *data.PassengerDoor == 1
	}

	// HVAC status
	if data.ACStatus != nil {
		state["ac_status"] = *data.ACStatus
		acModes := map[float64]string{
			0: "off",
			1: "auto",
			2: "heat",
			3: "cool",
			4: "fan",
		}
		if mode, ok := acModes[*data.ACStatus]; ok {
			state["ac_mode"] = mode
		}
	}
	if data.FanSpeedLevel != nil {
		state["fan_speed"] = *data.FanSpeedLevel
	}

	// Engine and motor data
	if data.PowerStatus != nil {
		state["engine"] = *data.PowerStatus == 1
	}
	if data.EngineRPM != nil {
		state["engine_rpm"] = *data.EngineRPM
	}

	// Safety systems
	if data.DriverSeatbelt != nil {
		state["driver_seatbelt"] = *data.DriverSeatbelt == 1
	}

	// Add availability status
	state["available"] = true

	return json.Marshal(state)
}

// Transmit sends sensor data to the MQTT broker.
// On the first call, it also publishes the Home Assistant discovery configuration.
func (t *MQTTTransmitter) Transmit(data *sensors.SensorData) error {
	if !t.client.IsConnected() {
		return fmt.Errorf("MQTT client not connected")
	}

	// Publish discovery config and availability on the first run
	if !t.discoveryPublished {
		if err := t.publishDiscoveryConfig(); err != nil {
			t.logger.WithError(err).Error("Failed to publish Home Assistant discovery configuration")
			// We can still try to publish state, so don't return here
		}
		if err := t.publishAvailability(true); err != nil {
			t.logger.WithError(err).Warn("Failed to publish availability")
		}
		t.discoveryPublished = true
	}

	// Publish sensor data
	if err := t.publishSensorData(data); err != nil {
		return fmt.Errorf("failed to publish sensor data: %w", err)
	}

	// Publish location data for device_tracker
	if err := t.publishLocationData(data); err != nil {
		// Log as a warning, as location is not always critical
		t.logger.WithError(err).Warn("Failed to publish location data")
	}

	return nil
}

// publishSensorData sends the main sensor data to the state topic.
func (t *MQTTTransmitter) publishSensorData(data *sensors.SensorData) error {
	stateTopic := t.client.GetStateTopic()
	payload, err := t.buildStatePayload(data)
	if err != nil {
		return fmt.Errorf("failed to marshal sensor data: %w", err)
	}

	return t.client.Publish(stateTopic, payload, false)
}

// publishLocationData sends GPS coordinates for the device_tracker entity.
func (t *MQTTTransmitter) publishLocationData(data *sensors.SensorData) error {
	// Only publish if location data is available
	if data.Location == nil {
		t.logger.Debug("Skipping location publish: no data")
		return nil
	}

	locationTopic := fmt.Sprintf("%s/location", t.client.GetBaseTopic())
	payload := map[string]interface{}{
		"latitude":  data.Location.Latitude,
		"longitude": data.Location.Longitude,
		"gps_accuracy": data.Location.Accuracy,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal location data: %w", err)
	}

	t.logger.WithFields(logrus.Fields{
		"topic": locationTopic,
		"lat":   data.Location.Latitude,
		"lon":   data.Location.Longitude,
	}).Debug("Publishing location data")

	return t.client.Publish(locationTopic, jsonPayload, false)
}

// publishDeviceTrackerDiscovery sends the discovery config for the device_tracker.
func (t *MQTTTransmitter) publishDeviceTrackerDiscovery(baseTopic string, device HADevice) error {
	discoveryTopic := t.client.GetDiscoveryTopic(t.discoveryPrefix, "device_tracker", "location")
	locationTopic := fmt.Sprintf("%s/location", baseTopic)

	config := map[string]interface{}{
		"name":                "BYD Car Location",
		"unique_id":           fmt.Sprintf("byd_car_%s_location", t.deviceID),
		"state_topic":         locationTopic,
		"json_attributes_topic": locationTopic,
		"device":              device,
		"icon":                "mdi:car-connected",
	}

	payload, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal device_tracker config: %w", err)
	}

	t.logger.WithField("topic", discoveryTopic).Debug("Publishing device_tracker discovery config")
	return t.client.Publish(discoveryTopic, payload, true)
}

// publishAvailability sends the online/offline status of the device.
func (t *MQTTTransmitter) publishAvailability(online bool) error {
	topic := t.client.GetAvailabilityTopic()
	status := "offline"
	if online {
		status = "online"
	}
	return t.client.Publish(topic, []byte(status), true)
}

// IsConnected checks if the underlying MQTT client is connected.
func (t *MQTTTransmitter) IsConnected() bool {
	return t.client.IsConnected()
}
