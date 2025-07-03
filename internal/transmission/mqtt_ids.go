package transmission

// MQTTSensor represents a sensor we may publish via MQTT.  The Transmit flag
// allows us to keep a sensor available to the application while suppressing
// its publication.  The default (zero‐value) for bool is "false", so we must
// set Transmit: true for the sensors we actually want to expose.
type MQTTSensor struct {
	ID       int  // SensorDefinition.ID from sensors.AllSensors
	Transmit bool // Whether to publish this sensor via MQTT
}

// MQTTSensors enumerates all sensors of interest.  Set Transmit=false for any
// sensor that should be kept internal (e.g. used in virtual sensors) but not
// forwarded to MQTT directly.
var MQTTSensors = []MQTTSensor{
	{ID: 33, Transmit: true}, // BatteryPercentage
	{ID: 2, Transmit: true},  // Speed
	{ID: 3, Transmit: true},  // Mileage
	{ID: 53, Transmit: true}, // Tire pressures LF
	{ID: 54, Transmit: true}, // Tire pressures RF
	{ID: 55, Transmit: true}, // Tire pressures LR
	{ID: 56, Transmit: true}, // Tire pressures RR
	{ID: 10, Transmit: true}, // EnginePower
	{ID: 26, Transmit: true}, // OutsideTemperature
	{ID: 14, Transmit: true}, // CabinTemperature

	{ID: 12, Transmit: false}, // ChargeGunState (internal only)
}

// TransmittedSensorIDs returns the subset of sensor IDs whose Transmit flag is true.
func TransmittedSensorIDs() []int {
	ids := make([]int, 0, len(MQTTSensors))
	for _, s := range MQTTSensors {
		if s.Transmit {
			ids = append(ids, s.ID)
		}
	}
	return ids
}
