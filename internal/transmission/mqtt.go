package transmission

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/jkaberg/byd-hass/internal/mqtt"
	"github.com/jkaberg/byd-hass/internal/sensors"
	"github.com/sirupsen/logrus"
)

// MQTTTransmitter transmits sensor data via MQTT
type MQTTTransmitter struct {
	client           *mqtt.Client
	deviceID         string
	discoveryPrefix  string
	logger           *logrus.Logger
	publishedSensors map[string]bool // Tracks published discovery configs
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
	EntityCategory    string   `json:"entity_category,omitempty"`
}

// HADevice represents the device information for Home Assistant
type HADevice struct {
	Identifiers  []string `json:"identifiers"`
	Name         string   `json:"name"`
	Model        string   `json:"model"`
	Manufacturer string   `json:"manufacturer"`
	SWVersion    string   `json:"sw_version,omitempty"`
}

// SensorConfig defines the configuration for each sensor
type SensorConfig struct {
	Name        string
	EntityID    string
	EntityType  string
	DeviceClass string
	Unit        string
	Icon        string
	StateClass  string
	Category    string
	ScaleFactor float64 // For unit conversion
}

// NewMQTTTransmitter creates a new MQTT transmitter
func NewMQTTTransmitter(client *mqtt.Client, deviceID, discoveryPrefix string, logger *logrus.Logger) *MQTTTransmitter {
	return &MQTTTransmitter{
		client:           client,
		deviceID:         deviceID,
		discoveryPrefix:  discoveryPrefix,
		logger:           logger,
		publishedSensors: make(map[string]bool),
	}
}

// getSensorConfigs builds sensor discovery configurations dynamically
// from the canonical sensors.AllSensors slice. This removes the need to
// manually maintain a duplicate list every time a new sensor is added.
//
// Icons, state-classes and other Home-Assistant niceties can be added
// later via dedicated mapping tables, but we prefer to keep the core
// list lean and fully data-driven for now.
func (t *MQTTTransmitter) getSensorConfigs() []SensorConfig {
	// Build a lookup table for quick ID → definition mapping
	idSet := make(map[int]struct{}, len(MQTTSensorIDs))
	for _, id := range MQTTSensorIDs {
		idSet[id] = struct{}{}
	}

	configs := make([]SensorConfig, 0, len(idSet))

	for _, def := range sensors.AllSensors {
		if _, ok := idSet[def.ID]; !ok {
			continue // skip sensors not in the allowed MQTT list
		}
		configs = append(configs, SensorConfig{
			Name:        def.EnglishName,
			EntityID:    sensors.ToSnakeCase(def.FieldName),
			EntityType:  def.Category,          // "sensor" / "binary_sensor"
			DeviceClass: def.DeviceClass,       // may be "" if not set
			Unit:        def.UnitOfMeasurement, // may be "" if not set
			ScaleFactor: 1.0,                   // default; can be refined later
		})
	}
	return configs
}

// publishDiscoveryForSensor publishes the discovery config for a single sensor.
func (t *MQTTTransmitter) publishDiscoveryForSensor(sensor SensorConfig, device HADevice, baseTopic string) error {
	uniqueID := fmt.Sprintf("%s_%s", t.deviceID, sensor.EntityID)

	// Skip if already published
	if t.publishedSensors[uniqueID] {
		return nil
	}

	config := HADiscoveryConfig{
		Name:              sensor.Name,
		UniqueID:          uniqueID,
		StateTopic:        fmt.Sprintf("%s/state", baseTopic),
		ValueTemplate:     fmt.Sprintf("{{ value_json.%s | default(0) }}", sensor.EntityID),
		AvailabilityTopic: fmt.Sprintf("%s/availability", baseTopic),
		Device:            device,
	}

	if sensor.DeviceClass != "" {
		config.DeviceClass = sensor.DeviceClass
	}
	if sensor.Unit != "" {
		config.UnitOfMeasurement = sensor.Unit
	}
	if sensor.Icon != "" {
		config.Icon = sensor.Icon
	}
	if sensor.StateClass != "" {
		config.StateClass = sensor.StateClass
	}
	if sensor.Category != "" {
		config.EntityCategory = sensor.Category
	}

	topic := fmt.Sprintf("%s/%s/byd_car_%s/%s/config",
		t.discoveryPrefix, sensor.EntityType, t.deviceID, sensor.EntityID)

	if err := t.publishConfigRaw(topic, config); err != nil {
		return fmt.Errorf("failed to publish %s discovery config: %w", sensor.Name, err)
	}

	t.logger.WithFields(logrus.Fields{
		"sensor_name": sensor.Name,
		"entity_id":   sensor.EntityID,
		"topic":       topic,
	}).Info("Published sensor discovery config")

	// Mark as published
	t.publishedSensors[uniqueID] = true
	return nil
}

// publishDiscoveryConfigs ensures all available sensors have their discovery configs published.
func (t *MQTTTransmitter) publishDiscoveryConfigs(data *sensors.SensorData) error {
	device := HADevice{
		Identifiers:  []string{fmt.Sprintf("byd_car_%s", t.deviceID)},
		Name:         "BYD Car",
		Model:        "Car",
		Manufacturer: "BYD",
		SWVersion:    "1.0.0",
	}
	baseTopic := fmt.Sprintf("byd_car/%s", t.deviceID)

	// Publish device_tracker discovery first (if not already done)
	if !t.publishedSensors["device_tracker"] {
		if err := t.publishDeviceTrackerDiscovery(baseTopic, device); err != nil {
			t.logger.WithError(err).Warn("Failed to publish device_tracker discovery")
		} else {
			t.logger.Info("Device tracker discovery config published")
			t.publishedSensors["device_tracker"] = true
		}
	}

	sensorConfigs := t.getSensorConfigs()

	for _, config := range sensorConfigs {
		// Always publish Home-Assistant discovery for allowed sensors even if we don't
		// currently have a value for them. This guarantees that the full set of
		// entities defined in MQTTSensorIDs becomes available in the UI right from
		// the start. The ValueTemplate in publishDiscoveryForSensor already
		// employs a `default(0)` filter, so missing values will not break
		// rendering.
		if err := t.publishDiscoveryForSensor(config, device, baseTopic); err != nil {
			t.logger.WithError(err).WithField("sensor", config.Name).Error("Failed to publish discovery config")
			// Continue to the next sensor
		}
	}

	return nil
}

// publishConfigRaw publishes a raw configuration object
func (t *MQTTTransmitter) publishConfigRaw(topic string, config interface{}) error {
	payload, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal discovery config: %w", err)
	}

	if err := t.client.Publish(topic, payload, true); err != nil {
		return fmt.Errorf("failed to publish discovery config to %s: %w", topic, err)
	}

	return nil
}

// applyScaling applies a scaling factor to a sensor value if it's not nil
func applyScaling(value *float64, scaleFactor float64) interface{} {
	if value != nil {
		return *value * scaleFactor
	}
	return nil
}

// buildStatePayload builds the JSON payload for the state topic
func (t *MQTTTransmitter) buildStatePayload(data *sensors.SensorData) ([]byte, error) {
	state := make(map[string]interface{})
	// Pre-compute allowed entityIDs in snake_case for quick filtering
	allowed := make(map[string]struct{}, len(MQTTSensorIDs))
	for _, id := range MQTTSensorIDs {
		if def := sensors.GetSensorByID(id); def != nil {
			allowed[sensors.ToSnakeCase(def.FieldName)] = struct{}{}
		}
	}

	v := reflect.ValueOf(data).Elem()
	tOf := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)

		// Skip unexported fields or fields that are nil
		if !field.CanInterface() || (field.Kind() == reflect.Ptr && field.IsNil()) {
			continue
		}

		jsonTag := tOf.Field(i).Tag.Get("json")
		jsonKey := strings.Split(jsonTag, ",")[0]

		if jsonKey == "" || jsonKey == "-" {
			continue
		}

		if _, ok := allowed[jsonKey]; !ok {
			continue // not in MQTT allow-list
		}

		// Dereference pointer to get the actual value
		var value interface{}
		if field.Kind() == reflect.Ptr {
			value = field.Elem().Interface()
		} else {
			value = field.Interface()
		}
		state[jsonKey] = value
	}
	// Add a 'state' field for the device_tracker
	if data.Speed != nil && *data.Speed > 0 {
		state["state"] = "moving"
	} else if data.ChargingStatus != nil && *data.ChargingStatus > 0 {
		state["state"] = "charging"
	} else if data.PowerStatus != nil && *data.PowerStatus > 0 {
		state["state"] = "online"
	} else {
		state["state"] = "parked"
	}

	return json.Marshal(state)
}

// Btoi converts a boolean to an integer (0 or 1)
func Btoi(b bool) int {
	if b {
		return 1
	}
	return 0
}

// Transmit sends sensor data to MQTT
func (t *MQTTTransmitter) Transmit(data *sensors.SensorData) error {
	if !t.client.IsConnected() {
		return fmt.Errorf("MQTT client not connected")
	}

	// Publish discovery config for available sensors if it hasn't been done
	if err := t.publishDiscoveryConfigs(data); err != nil {
		// Log error but don't block transmission
		t.logger.WithError(err).Error("Failed to publish Home Assistant discovery configs")
	}

	// Publish sensor data
	if err := t.publishSensorData(data); err != nil {
		return fmt.Errorf("failed to publish sensor data: %w", err)
	}

	// Publish location data if available
	if data.Location != nil {
		if err := t.publishLocationData(data); err != nil {
			// Log error but don't block other publications
			t.logger.WithError(err).Warn("Failed to publish location data")
		}
	}

	// Publish availability
	if err := t.publishAvailability(true); err != nil {
		return fmt.Errorf("failed to publish availability: %w", err)
	}

	t.logger.Debug("Data transmitted successfully")
	return nil
}

// publishSensorData publishes the main sensor data payload
func (t *MQTTTransmitter) publishSensorData(data *sensors.SensorData) error {
	payload, err := t.buildStatePayload(data)
	if err != nil {
		return fmt.Errorf("failed to build state payload: %w", err)
	}

	topic := fmt.Sprintf("byd_car/%s/state", t.deviceID)
	if err := t.client.Publish(topic, payload, true); err != nil {
		return fmt.Errorf("failed to publish sensor data to %s: %w", topic, err)
	}

	t.logger.WithFields(logrus.Fields{
		"topic":   topic,
		"payload": string(payload),
	}).Info("Published sensor data")

	return nil
}

// publishLocationData publishes location data to the device_tracker entity
func (t *MQTTTransmitter) publishLocationData(data *sensors.SensorData) error {
	if data.Location == nil {
		return nil
	}

	topic := fmt.Sprintf("byd_car/%s/location", t.deviceID)
	payload := map[string]interface{}{
		"latitude":     data.Location.Latitude,
		"longitude":    data.Location.Longitude,
		"gps_accuracy": data.Location.Accuracy,
		"battery":      data.BatteryPercentage,
		"speed":        data.Speed,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal location data: %w", err)
	}

	return t.client.Publish(topic, jsonPayload, false)
}

// publishDeviceTrackerDiscovery publishes the discovery config for the device tracker.
func (t *MQTTTransmitter) publishDeviceTrackerDiscovery(baseTopic string, device HADevice) error {
	attributesTopic := fmt.Sprintf("%s/location", baseTopic)
	config := map[string]interface{}{
		"name":                  "Location",
		"unique_id":             fmt.Sprintf("%s_location", t.deviceID),
		"json_attributes_topic": attributesTopic,
		"source_type":           "gps",
		"device":                device,
		"availability_topic":    fmt.Sprintf("%s/availability", baseTopic),
	}
	topic := fmt.Sprintf("%s/device_tracker/byd_car_%s/config", t.discoveryPrefix, t.deviceID)

	return t.publishConfigRaw(topic, config)
}

// publishAvailability publishes the availability status
func (t *MQTTTransmitter) publishAvailability(online bool) error {
	payload := "online"
	if !online {
		payload = "offline"
	}

	topic := fmt.Sprintf("byd_car/%s/availability", t.deviceID)
	if err := t.client.Publish(topic, []byte(payload), true); err != nil {
		return fmt.Errorf("failed to publish availability to %s: %w", topic, err)
	}
	return nil
}

// IsConnected checks if the MQTT client is connected
func (t *MQTTTransmitter) IsConnected() bool {
	return t.client.IsConnected()
}
