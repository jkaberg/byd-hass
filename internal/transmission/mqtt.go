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
		client:          client,
		deviceID:        deviceID,
		discoveryPrefix: discoveryPrefix,
		logger:          logger,
	}
}

// getSensorConfigs returns all sensor configurations with proper units and scaling
func (t *MQTTTransmitter) getSensorConfigs() []SensorConfig {
	return []SensorConfig{
		// Core Vehicle Data
		{"Battery Percentage", "battery_percentage", "sensor", "battery", "%", "mdi:battery", "measurement", "", 1.0},
		{"Speed", "speed", "sensor", "speed", "km/h", "mdi:speedometer", "measurement", "", 1.0},
		{"Mileage", "mileage", "sensor", "distance", "km", "mdi:counter", "total_increasing", "", 1.0},
		{"Gear Position", "gear_position", "sensor", "", "", "mdi:car-shift-pattern", "measurement", "diagnostic", 1.0},
		{"Engine RPM", "engine_rpm", "sensor", "", "rpm", "mdi:engine", "measurement", "diagnostic", 1.0},
		{"Brake Depth", "brake_depth", "sensor", "", "%", "mdi:car-brake-alert", "measurement", "diagnostic", 1.0},
		{"Accelerator Depth", "accelerator_depth", "sensor", "", "%", "mdi:car-speed-limiter", "measurement", "diagnostic", 1.0},
		{"Front Motor RPM", "front_motor_rpm", "sensor", "", "rpm", "mdi:engine-outline", "measurement", "diagnostic", 1.0},
		{"Rear Motor RPM", "rear_motor_rpm", "sensor", "", "rpm", "mdi:engine-outline", "measurement", "diagnostic", 1.0},
		{"Engine Power", "engine_power", "sensor", "power", "kW", "mdi:engine", "measurement", "diagnostic", 1.0},
		{"Front Motor Torque", "front_motor_torque", "sensor", "", "Nm", "mdi:engine-outline", "measurement", "diagnostic", 1.0},

		// Battery & Charging
		{"Charge Gun State", "charge_gun_state", "binary_sensor", "plug", "", "mdi:power-plug", "", "", 1.0},
		{"Power Consumption 100km", "power_consumption_100km", "sensor", "energy", "kWh/100km", "mdi:car-electric", "measurement", "", 1.0},
		{"Max Battery Temperature", "max_battery_temp", "sensor", "temperature", "°C", "mdi:thermometer-high", "measurement", "", 1.0},
		{"Avg Battery Temperature", "avg_battery_temp", "sensor", "temperature", "°C", "mdi:thermometer", "measurement", "", 1.0},
		{"Min Battery Temperature", "min_battery_temp", "sensor", "temperature", "°C", "mdi:thermometer-low", "measurement", "", 1.0},
		{"Max Battery Voltage", "max_battery_voltage", "sensor", "voltage", "V", "mdi:battery-plus", "measurement", "", 1.0},
		{"Min Battery Voltage", "min_battery_voltage", "sensor", "voltage", "V", "mdi:battery-minus", "measurement", "", 1.0},
		{"Battery Capacity", "battery_capacity", "sensor", "energy", "kWh", "mdi:battery-outline", "measurement", "", 1.0},
		{"Total Power Consumption", "total_power_consumption", "sensor", "energy", "kWh", "mdi:lightning-bolt", "total_increasing", "", 1.0},
		{"Fuel Percentage", "fuel_percentage", "sensor", "", "%", "mdi:gas-station", "measurement", "", 1.0},
		{"Total Fuel Consumption", "total_fuel_consumption", "sensor", "volume", "L", "mdi:gas-station", "total_increasing", "", 1.0},
		{"Charging Status", "charging_status", "binary_sensor", "battery_charging", "", "mdi:battery-charging", "", "", 1.0},
		{"12V Battery Voltage", "battery_voltage_12v", "sensor", "voltage", "V", "mdi:car-battery", "measurement", "diagnostic", 1.0},

		// Environment & Weather
		{"Cabin Temperature", "cabin_temperature", "sensor", "temperature", "°C", "mdi:thermometer", "measurement", "", 1.0},
		{"Outside Temperature", "outside_temperature", "sensor", "temperature", "°C", "mdi:thermometer", "measurement", "", 1.0},
		{"Driver AC Temperature", "driver_ac_temperature", "sensor", "temperature", "°C", "mdi:air-conditioner", "measurement", "", 1.0},
		{"Engine Coolant Temperature", "engine_coolant_temp", "sensor", "temperature", "°C", "mdi:coolant-temperature", "measurement", "diagnostic", 1.0},

		// Steering & Control
		{"Steering Angle", "steering_angle", "sensor", "", "°", "mdi:steering", "measurement", "diagnostic", 1.0},
		{"Steering Rotation Speed", "steering_rotation_speed", "sensor", "", "°/s", "mdi:steering", "measurement", "diagnostic", 1.0},
		{"Lane Curvature", "lane_curvature", "sensor", "", "", "mdi:road-variant", "measurement", "diagnostic", 1.0},
		{"Right Line Distance", "right_line_distance", "sensor", "distance", "m", "mdi:road-variant", "measurement", "diagnostic", 1.0},
		{"Left Line Distance", "left_line_distance", "sensor", "distance", "m", "mdi:road-variant", "measurement", "diagnostic", 1.0},
		{"Distance to Car Ahead", "distance_to_car_ahead", "sensor", "distance", "m", "mdi:car-connected", "measurement", "diagnostic", 1.0},
		{"Auto Parking", "auto_parking", "binary_sensor", "", "", "mdi:car-brake-parking", "", "diagnostic", 1.0},
		{"ACC Cruise Status", "acc_cruise_status", "binary_sensor", "", "", "mdi:cruise-control", "", "diagnostic", 1.0},
		{"Lane Keep Assist", "lane_keep_assist_status", "binary_sensor", "", "", "mdi:car-cruise-control", "", "diagnostic", 1.0},

		// Radar Sensors (convert to meters)
		{"Radar Front Left", "radar_front_left", "sensor", "distance", "m", "mdi:radar", "measurement", "diagnostic", 0.01},
		{"Radar Front Right", "radar_front_right", "sensor", "distance", "m", "mdi:radar", "measurement", "diagnostic", 0.01},
		{"Radar Rear Left", "radar_rear_left", "sensor", "distance", "m", "mdi:radar", "measurement", "diagnostic", 0.01},
		{"Radar Rear Right", "radar_rear_right", "sensor", "distance", "m", "mdi:radar", "measurement", "diagnostic", 0.01},
		{"Radar Left", "radar_left", "sensor", "distance", "m", "mdi:radar", "measurement", "diagnostic", 0.01},
		{"Radar Front Mid Left", "radar_front_mid_left", "sensor", "distance", "m", "mdi:radar", "measurement", "diagnostic", 0.01},
		{"Radar Front Mid Right", "radar_front_mid_right", "sensor", "distance", "m", "mdi:radar", "measurement", "diagnostic", 0.01},
		{"Radar Rear Center", "radar_rear_center", "sensor", "distance", "m", "mdi:radar", "measurement", "diagnostic", 0.01},

		// Tire Pressure (convert from raw values to bar)
		{"Left Front Tire Pressure", "left_front_tire_pressure", "sensor", "pressure", "bar", "mdi:car-tire-alert", "measurement", "diagnostic", 0.01},
		{"Right Front Tire Pressure", "right_front_tire_pressure", "sensor", "pressure", "bar", "mdi:car-tire-alert", "measurement", "diagnostic", 0.01},
		{"Left Rear Tire Pressure", "left_rear_tire_pressure", "sensor", "pressure", "bar", "mdi:car-tire-alert", "measurement", "diagnostic", 0.01},
		{"Right Rear Tire Pressure", "right_rear_tire_pressure", "sensor", "pressure", "bar", "mdi:car-tire-alert", "measurement", "diagnostic", 0.01},

		// Turn Signals & Lights
		{"Left Turn Signal", "left_turn_signal", "binary_sensor", "", "", "mdi:arrow-left-bold", "", "diagnostic", 1.0},
		{"Right Turn Signal", "right_turn_signal", "binary_sensor", "", "", "mdi:arrow-right-bold", "", "diagnostic", 1.0},
		{"Parking Lights", "parking_lights", "binary_sensor", "", "", "mdi:car-light-dimmed", "", "diagnostic", 1.0},
		{"Low Beam Lights", "low_beam_lights", "binary_sensor", "", "", "mdi:car-light-low", "", "diagnostic", 1.0},
		{"High Beam Lights", "high_beam_lights", "binary_sensor", "", "", "mdi:car-light-high", "", "diagnostic", 1.0},
		{"Front Fog Lights", "front_fog_lights", "binary_sensor", "", "", "mdi:car-light-fog", "", "diagnostic", 1.0},
		{"Rear Fog Lights", "rear_fog_lights", "binary_sensor", "", "", "mdi:car-light-fog", "", "diagnostic", 1.0},
		{"Daytime Running Lights", "daytime_running_lights", "binary_sensor", "", "", "mdi:car-light-dimmed", "", "diagnostic", 1.0},
		{"Hazard Lights", "hazard_lights", "binary_sensor", "", "", "mdi:hazard-lights", "", "diagnostic", 1.0},

		// Doors & Locks
		{"Driver Door Lock", "driver_door_lock", "binary_sensor", "lock", "", "mdi:lock", "", "", 1.0},
		{"Driver Door", "driver_door", "binary_sensor", "door", "", "mdi:car-door", "", "", 1.0},
		{"Passenger Door", "passenger_door", "binary_sensor", "door", "", "mdi:car-door", "", "", 1.0},
		{"Left Rear Door", "left_rear_door", "binary_sensor", "door", "", "mdi:car-door", "", "", 1.0},
		{"Right Rear Door", "right_rear_door", "binary_sensor", "door", "", "mdi:car-door", "", "", 1.0},
		{"Hood", "hood", "binary_sensor", "door", "", "mdi:car", "", "", 1.0},
		{"Trunk Door", "trunk_door", "binary_sensor", "door", "", "mdi:car-back", "", "", 1.0},
		{"Remote Lock Status", "remote_lock_status", "binary_sensor", "lock", "", "mdi:lock-smart", "", "", 1.0},

		// Windows (percentage open)
		{"Driver Window", "driver_window_open_percent", "sensor", "", "%", "mdi:car-door", "measurement", "diagnostic", 1.0},
		{"Passenger Window", "passenger_window_open_percent", "sensor", "", "%", "mdi:car-door", "measurement", "diagnostic", 1.0},
		{"Left Rear Window", "left_rear_window_open_percent", "sensor", "", "%", "mdi:car-door", "measurement", "diagnostic", 1.0},
		{"Right Rear Window", "right_rear_window_open_percent", "sensor", "", "%", "mdi:car-door", "measurement", "diagnostic", 1.0},
		{"Sunroof", "sunroof_open_percent", "sensor", "", "%", "mdi:car-convertible", "measurement", "diagnostic", 1.0},

		// HVAC/Climate
		{"AC Status", "ac_status", "sensor", "", "", "mdi:air-conditioner", "measurement", "", 1.0},
		{"Fan Speed Level", "fan_speed_level", "sensor", "", "", "mdi:fan", "measurement", "", 1.0},
		{"AC Circulation Mode", "ac_circulation_mode", "sensor", "", "", "mdi:air-conditioner", "measurement", "diagnostic", 1.0},
		{"AC Blowing Mode", "ac_blowing_mode", "sensor", "", "", "mdi:air-conditioner", "measurement", "diagnostic", 1.0},

		// Safety & Security
		{"Driver Seatbelt", "driver_seatbelt", "binary_sensor", "safety", "", "mdi:seatbelt", "", "diagnostic", 1.0},
		{"Passenger Seatbelt Warning", "passenger_seatbelt_warn", "binary_sensor", "safety", "", "mdi:seatbelt", "", "diagnostic", 1.0},

		// Vehicle Modes
		{"Vehicle Operating Mode", "vehicle_operating_mode", "sensor", "", "", "mdi:car-cog", "measurement", "diagnostic", 1.0},
		{"Vehicle Running Mode", "vehicle_running_mode", "sensor", "", "", "mdi:car-cog", "measurement", "diagnostic", 1.0},
		{"Power Status", "power_status", "binary_sensor", "power", "", "mdi:power", "", "", 1.0},
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

	t.logger.WithFields(logrus.Fields{
		"device_id":    t.deviceID,
		"device_name":  device.Name,
		"manufacturer": device.Manufacturer,
	}).Info("Creating Home Assistant device configuration")

	baseTopic := fmt.Sprintf("byd_car/%s", t.deviceID)

	// Publish device_tracker discovery first
	if err := t.publishDeviceTrackerDiscovery(baseTopic, device); err != nil {
		return fmt.Errorf("failed to publish device_tracker discovery config: %w", err)
	}
	t.logger.Debug("Device tracker discovery config published")

	// Get all sensor configurations
	sensorConfigs := t.getSensorConfigs()
	t.logger.WithField("sensor_count", len(sensorConfigs)).Info("Publishing sensor discovery configurations")

	// Publish sensor discovery configs
	for _, sensor := range sensorConfigs {
		config := HADiscoveryConfig{
			Name:              sensor.Name,
			UniqueID:          fmt.Sprintf("%s_%s", t.deviceID, sensor.EntityID),
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
		}).Debug("Published sensor discovery config")
	}

	t.logger.Info("All sensor discovery configurations published successfully")
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

// applyScaling applies scaling factor to a value if needed
func applyScaling(value *float64, scaleFactor float64) interface{} {
	if value == nil {
		return nil
	}
	if scaleFactor == 1.0 {
		return *value
	}
	return *value * scaleFactor
}

// buildStatePayload creates comprehensive state payload with all sensor data
func (t *MQTTTransmitter) buildStatePayload(data *sensors.SensorData) ([]byte, error) {
	state := map[string]interface{}{
		"timestamp": data.Timestamp.Unix(),
	}

	// Core Vehicle Data
	if data.BatteryPercentage != nil {
		state["battery_percentage"] = *data.BatteryPercentage
	}
	if data.Speed != nil {
		state["speed"] = *data.Speed
	}
	if data.Mileage != nil {
		state["mileage"] = *data.Mileage
	}
	if data.GearPosition != nil {
		state["gear_position"] = *data.GearPosition
	}
	if data.EngineRPM != nil {
		state["engine_rpm"] = *data.EngineRPM
	}
	if data.BrakeDepth != nil {
		state["brake_depth"] = *data.BrakeDepth
	}
	if data.AcceleratorDepth != nil {
		state["accelerator_depth"] = *data.AcceleratorDepth
	}
	if data.FrontMotorRPM != nil {
		state["front_motor_rpm"] = *data.FrontMotorRPM
	}
	if data.RearMotorRPM != nil {
		state["rear_motor_rpm"] = *data.RearMotorRPM
	}
	if data.EnginePower != nil {
		state["engine_power"] = *data.EnginePower
	}
	if data.FrontMotorTorque != nil {
		state["front_motor_torque"] = *data.FrontMotorTorque
	}

	// Battery & Charging
	if data.ChargeGunState != nil {
		state["charge_gun_state"] = *data.ChargeGunState == 2 // 2 = connected
	}
	if data.PowerConsumption100km != nil {
		state["power_consumption_100km"] = *data.PowerConsumption100km
	}
	if data.MaxBatteryTemp != nil {
		state["max_battery_temp"] = *data.MaxBatteryTemp
	}
	if data.AvgBatteryTemp != nil {
		state["avg_battery_temp"] = *data.AvgBatteryTemp
	}
	if data.MinBatteryTemp != nil {
		state["min_battery_temp"] = *data.MinBatteryTemp
	}
	if data.MaxBatteryVoltage != nil {
		state["max_battery_voltage"] = *data.MaxBatteryVoltage
	}
	if data.MinBatteryVoltage != nil {
		state["min_battery_voltage"] = *data.MinBatteryVoltage
	}
	if data.BatteryCapacity != nil {
		state["battery_capacity"] = *data.BatteryCapacity
	}
	if data.TotalPowerConsumption != nil {
		state["total_power_consumption"] = *data.TotalPowerConsumption
	}
	if data.FuelPercentage != nil {
		state["fuel_percentage"] = *data.FuelPercentage
	}
	if data.TotalFuelConsumption != nil {
		state["total_fuel_consumption"] = *data.TotalFuelConsumption
	}
	if data.ChargingStatus != nil {
		state["charging_status"] = *data.ChargingStatus == 1 // 1 = charging
	}
	if data.BatteryVoltage12V != nil {
		state["battery_voltage_12v"] = *data.BatteryVoltage12V
	}

	// Environment & Weather
	if data.CabinTemperature != nil {
		state["cabin_temperature"] = *data.CabinTemperature
	}
	if data.OutsideTemperature != nil {
		state["outside_temperature"] = *data.OutsideTemperature
	}
	if data.DriverACTemperature != nil {
		state["driver_ac_temperature"] = *data.DriverACTemperature
	}
	if data.EngineCoolantTemp != nil {
		state["engine_coolant_temp"] = *data.EngineCoolantTemp
	}

	// Steering & Control
	if data.SteeringAngle != nil {
		state["steering_angle"] = *data.SteeringAngle
	}
	if data.SteeringRotationSpeed != nil {
		state["steering_rotation_speed"] = *data.SteeringRotationSpeed
	}
	if data.LaneCurvature != nil {
		state["lane_curvature"] = *data.LaneCurvature
	}
	if data.RightLineDistance != nil {
		state["right_line_distance"] = *data.RightLineDistance
	}
	if data.LeftLineDistance != nil {
		state["left_line_distance"] = *data.LeftLineDistance
	}
	if data.DistanceToCarAhead != nil {
		state["distance_to_car_ahead"] = *data.DistanceToCarAhead
	}
	if data.AutoParking != nil {
		state["auto_parking"] = *data.AutoParking == 1
	}
	if data.ACCCruiseStatus != nil {
		state["acc_cruise_status"] = *data.ACCCruiseStatus == 1
	}
	if data.LaneKeepAssistStatus != nil {
		state["lane_keep_assist_status"] = *data.LaneKeepAssistStatus == 1
	}

	// Radar Sensors (convert cm to meters)
	if data.RadarFrontLeft != nil {
		state["radar_front_left"] = *data.RadarFrontLeft * 0.01
	}
	if data.RadarFrontRight != nil {
		state["radar_front_right"] = *data.RadarFrontRight * 0.01
	}
	if data.RadarRearLeft != nil {
		state["radar_rear_left"] = *data.RadarRearLeft * 0.01
	}
	if data.RadarRearRight != nil {
		state["radar_rear_right"] = *data.RadarRearRight * 0.01
	}
	if data.RadarLeft != nil {
		state["radar_left"] = *data.RadarLeft * 0.01
	}
	if data.RadarFrontMidLeft != nil {
		state["radar_front_mid_left"] = *data.RadarFrontMidLeft * 0.01
	}
	if data.RadarFrontMidRight != nil {
		state["radar_front_mid_right"] = *data.RadarFrontMidRight * 0.01
	}
	if data.RadarRearCenter != nil {
		state["radar_rear_center"] = *data.RadarRearCenter * 0.01
	}

	// Tire Pressure (convert to bar)
	if data.LeftFrontTirePressure != nil {
		state["left_front_tire_pressure"] = *data.LeftFrontTirePressure * 0.01
	}
	if data.RightFrontTirePressure != nil {
		state["right_front_tire_pressure"] = *data.RightFrontTirePressure * 0.01
	}
	if data.LeftRearTirePressure != nil {
		state["left_rear_tire_pressure"] = *data.LeftRearTirePressure * 0.01
	}
	if data.RightRearTirePressure != nil {
		state["right_rear_tire_pressure"] = *data.RightRearTirePressure * 0.01
	}

	// Turn Signals & Lights
	if data.LeftTurnSignal != nil {
		state["left_turn_signal"] = *data.LeftTurnSignal == 1
	}
	if data.RightTurnSignal != nil {
		state["right_turn_signal"] = *data.RightTurnSignal == 1
	}
	if data.ParkingLights != nil {
		state["parking_lights"] = *data.ParkingLights == 1
	}
	if data.LowBeamLights != nil {
		state["low_beam_lights"] = *data.LowBeamLights == 1
	}
	if data.HighBeamLights != nil {
		state["high_beam_lights"] = *data.HighBeamLights == 1
	}
	if data.FrontFogLights != nil {
		state["front_fog_lights"] = *data.FrontFogLights == 1
	}
	if data.RearFogLights != nil {
		state["rear_fog_lights"] = *data.RearFogLights == 1
	}
	if data.DaytimeRunningLights != nil {
		state["daytime_running_lights"] = *data.DaytimeRunningLights == 1
	}
	if data.HazardLights != nil {
		state["hazard_lights"] = *data.HazardLights == 1
	}

	// Doors & Locks
	if data.DriverDoorLock != nil {
		state["driver_door_lock"] = *data.DriverDoorLock == 1
	}
	if data.DriverDoor != nil {
		state["driver_door"] = *data.DriverDoor == 1
	}
	if data.PassengerDoor != nil {
		state["passenger_door"] = *data.PassengerDoor == 1
	}
	if data.LeftRearDoor != nil {
		state["left_rear_door"] = *data.LeftRearDoor == 1
	}
	if data.RightRearDoor != nil {
		state["right_rear_door"] = *data.RightRearDoor == 1
	}
	if data.Hood != nil {
		state["hood"] = *data.Hood == 1
	}
	if data.TrunkDoor != nil {
		state["trunk_door"] = *data.TrunkDoor == 1
	}
	if data.RemoteLockStatus != nil {
		state["remote_lock_status"] = *data.RemoteLockStatus == 1
	}

	// Windows
	if data.DriverWindowOpenPercent != nil {
		state["driver_window_open_percent"] = *data.DriverWindowOpenPercent
	}
	if data.PassengerWindowOpenPercent != nil {
		state["passenger_window_open_percent"] = *data.PassengerWindowOpenPercent
	}
	if data.LeftRearWindowOpenPercent != nil {
		state["left_rear_window_open_percent"] = *data.LeftRearWindowOpenPercent
	}
	if data.RightRearWindowOpenPercent != nil {
		state["right_rear_window_open_percent"] = *data.RightRearWindowOpenPercent
	}
	if data.SunroofOpenPercent != nil {
		state["sunroof_open_percent"] = *data.SunroofOpenPercent
	}

	// HVAC/Climate
	if data.ACStatus != nil {
		state["ac_status"] = *data.ACStatus
	}
	if data.FanSpeedLevel != nil {
		state["fan_speed_level"] = *data.FanSpeedLevel
	}
	if data.ACCirculationMode != nil {
		state["ac_circulation_mode"] = *data.ACCirculationMode
	}
	if data.ACBlowingMode != nil {
		state["ac_blowing_mode"] = *data.ACBlowingMode
	}

	// Safety & Security
	if data.DriverSeatbelt != nil {
		state["driver_seatbelt"] = *data.DriverSeatbelt == 1
	}
	if data.PassengerSeatbeltWarn != nil {
		state["passenger_seatbelt_warn"] = *data.PassengerSeatbeltWarn == 1
	}

	// Vehicle Modes
	if data.VehicleOperatingMode != nil {
		state["vehicle_operating_mode"] = *data.VehicleOperatingMode
	}
	if data.VehicleRunningMode != nil {
		state["vehicle_running_mode"] = *data.VehicleRunningMode
	}
	if data.PowerStatus != nil {
		state["power_status"] = *data.PowerStatus == 1
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

	t.logger.Debug("Starting MQTT transmission...")

	// Publish discovery config and availability on the first run
	if !t.discoveryPublished {
		t.logger.Info("Publishing Home Assistant discovery configuration...")
		if err := t.publishDiscoveryConfig(); err != nil {
			t.logger.WithError(err).Error("Failed to publish Home Assistant discovery configuration")
			// We can still try to publish state, so don't return here
		} else {
			t.logger.Info("Home Assistant discovery configuration published successfully")
		}
		if err := t.publishAvailability(true); err != nil {
			t.logger.WithError(err).Warn("Failed to publish availability")
		} else {
			t.logger.Debug("Device availability published")
		}
		t.discoveryPublished = true
	}

	// Publish sensor data
	t.logger.Debug("Publishing sensor state data...")
	if err := t.publishSensorData(data); err != nil {
		return fmt.Errorf("failed to publish sensor data: %w", err)
	}
	t.logger.Debug("Sensor state data published successfully")

	// Publish location data for device_tracker
	if err := t.publishLocationData(data); err != nil {
		// Log as a warning, as location is not always critical
		t.logger.WithError(err).Warn("Failed to publish location data")
	}

	t.logger.Debug("MQTT transmission completed successfully")
	return nil
}

// publishSensorData sends the main sensor data to the state topic.
func (t *MQTTTransmitter) publishSensorData(data *sensors.SensorData) error {
	stateTopic := t.client.GetStateTopic()
	payload, err := t.buildStatePayload(data)
	if err != nil {
		return fmt.Errorf("failed to marshal sensor data: %w", err)
	}

	t.logger.WithFields(logrus.Fields{
		"topic":        stateTopic,
		"payload_size": len(payload),
	}).Debug("Publishing sensor state to MQTT topic")

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
		"latitude":     data.Location.Latitude,
		"longitude":    data.Location.Longitude,
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
		"name":                  "Location",
		"unique_id":             fmt.Sprintf("byd_car_%s_location", t.deviceID),
		"state_topic":           locationTopic,
		"json_attributes_topic": locationTopic,
		"device":                device,
		"icon":                  "mdi:car-connected",
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
