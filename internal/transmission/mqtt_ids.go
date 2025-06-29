package transmission

// MQTTSensorIDs is the authoritative list of DiPlus sensor IDs we publish
// via MQTT (state + Home-Assistant discovery).  Edit this slice if you want
// to add or remove signals; no other code changes are required.
//
// Use the IDs from sensors.AllSensors (see internal/sensors/types.go).
// Keep the order stable for readability — it has no runtime impact.
var MQTTSensorIDs = []int{
	33,             // BatteryPercentage
	2,              // Speed
	3,              // Mileage
	52,             // ChargingStatus
	10,             // EnginePower
	26,             // OutsideTemperature
	14,             // CabinTemperature
	53, 54, 55, 56, // Tire pressures LF,RF,LR,RR
}
