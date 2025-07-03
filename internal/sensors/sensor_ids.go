package sensors

// MonitoredSensor represents a sensor that we (a) poll from Diplus and (b)
// may expose to downstream integrations such as MQTT / ABRP / REST.
//
// • Every entry is included in each Diplus request (see PollSensorIDs).
// • If Publish == true the raw value is allowed to leave the application –
//   currently that means it will appear in MQTT discovery/state payloads.
//   When we add other outputs (Prometheus, REST, etc.) they will consult the
//   same PublishedSensorIDs helper.
// • Entries with Publish == false stay internal – useful for building derived
//   sensors or for future features we do not want to expose yet.
//
// To add a new sensor:
//   1. Make sure it exists in sensors.AllSensors with a unique ID.
//   2. Append its ID below, choosing Publish=true/false.
//   3. No other lists need editing.

type MonitoredSensor struct {
	ID      int  // sensors.SensorDefinition.ID
	Publish bool // true → value may be published externally
}

// MonitoredSensors enumerates the subset of sensors our app currently cares
// about.  Keep this list tidy; polling *all* 100-ish sensors every 15 seconds
// would waste bandwidth and CPU on the head-unit.
var MonitoredSensors = []MonitoredSensor{
	{ID: 33, Publish: true}, // BatteryPercentage – HA battery, location attr
	{ID: 2, Publish: true},  // Speed             – HA + device_tracker state
	{ID: 3, Publish: true},  // Mileage           – HA odometer
	{ID: 53, Publish: true}, // LeftFrontTirePressure
	{ID: 54, Publish: true}, // RightFrontTirePressure
	{ID: 55, Publish: true}, // LeftRearTirePressure
	{ID: 56, Publish: true}, // RightRearTirePressure
	{ID: 10, Publish: true}, // EnginePower       – power gauge
	{ID: 26, Publish: true}, // OutsideTemperature – ambient temp
	{ID: 25, Publish: true}, // CabinTemperature  – cabin temp

	// Internal-only helpers (not published)
	{ID: 12, Publish: false}, // ChargeGunState – used for virtual charging_status

}

// PollSensorIDs returns every sensor ID we must include in the Diplus API
// template.
func PollSensorIDs() []int {
	ids := make([]int, 0, len(MonitoredSensors))
	for _, s := range MonitoredSensors {
		ids = append(ids, s.ID)
	}
	return ids
}

// PublishedSensorIDs returns only the IDs whose Publish flag is true.
func PublishedSensorIDs() []int {
	ids := make([]int, 0, len(MonitoredSensors))
	for _, s := range MonitoredSensors {
		if s.Publish {
			ids = append(ids, s.ID)
		}
	}
	return ids
}

// -----------------------------------------------------------------------------
// Integration Notes
// -----------------------------------------------------------------------------
// A Better Route Planner (ABRP) consumes the following SensorDefinition IDs via
// internal/transmission/abrp.go.  Make sure they remain present in
// MonitoredSensors – they can be Publish=false if you don’t want them in other
// outputs.
//
//   33  BatteryPercentage   (soc)
//    2  Speed               (speed / is_parked)
//    3  Mileage             (odometer)
//   10  EnginePower         (power, is_charging, is_dcfc)
//   12  ChargeGunState      (is_charging, is_dcfc)
//   15  AvgBatteryTemp      (batt_temp)
//   17  MaxBatteryVoltage   (voltage / current)
//   25  CabinTemperature    (cabin_temp)
//   26  OutsideTemperature  (ext_temp)
//   29  BatteryCapacity     (capacity, soe)
//   53-56 TirePressures LF/RF/LR/RR (tire_pressure_* – converted to kPa)
//   77  ACStatus            (hvac_power)
//   78  FanSpeedLevel       (hvac_power)
// -----------------------------------------------------------------------------
