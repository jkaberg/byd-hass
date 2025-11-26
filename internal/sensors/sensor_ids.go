package sensors

import (
    "os"
    "strings"
    "strconv"
)

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
//   2. Append its ID to "BYD_HASS_SENSOR_IDS" env, choosing Publish=true/false
//      in such manner: "ID:publish" for example "33:0,34:1", this will publish
//      id 34, and read but not publish id 33, you can omit ":1" as publish is 
//      the default, so you can write use "33,34:1" with the same effect
//   3. No other lists need editing.

type MonitoredSensor struct {
	ID      int  // sensors.SensorDefinition.ID
	Publish bool // true → value may be published externally
}

// MonitoredSensors enumerates the subset of sensors our app currently cares
// about.  Keep this list tidy; polling *all* 100-ish sensors every 15 seconds
// would waste bandwidth and CPU on the head-unit.
// loadMonitoredSensorsFromEnv overrides the default MonitoredSensors

// Default monitors
var defaultMonitoredSensors = []MonitoredSensor{
	{ID: 33, Publish: true}, // BatteryPercentage
	{ID: 34, Publish: true}, // FuelPercentage
	{ID: 2, Publish: true},  // Speed
	{ID: 3, Publish: true},  // Mileage
	{ID: 53, Publish: true}, // LF tire
	{ID: 54, Publish: true}, // RF tire
	{ID: 55, Publish: true}, // LR tire
	{ID: 56, Publish: true}, // RR tire
	{ID: 10, Publish: true}, // EnginePower
	{ID: 26, Publish: true}, // OutsideTemp
	{ID: 25, Publish: true}, // CabinTemp

	// Internal-only
	{ID: 12, Publish: false},
}

// Global value initialized at startup
var MonitoredSensors = loadMonitoredSensorsFromEnv()

// ---------------------------------------------------------

func loadMonitoredSensorsFromEnv() []MonitoredSensor {
	raw := os.Getenv("BYD_HASS_SENSOR_IDS")
	if raw == "" {
		return defaultMonitoredSensors
	}

	parts := strings.Split(raw, ",")
	sensorsList := make([]MonitoredSensor, 0, len(parts))

	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}

		publish := true

		// Format supports: "33" or "12:0" or "53:1"
		idStr := p
		if strings.Contains(p, ":") {
			pieces := strings.SplitN(p, ":", 2)
			idStr = pieces[0]
			if pieces[1] == "0" {
				publish = false
			}
		}

		id, err := strconv.Atoi(idStr)
		if err != nil {
			continue
		}

		sensorsList = append(sensorsList, MonitoredSensor{
			ID:	  id,
			Publish: publish,
		})
	}

	if len(sensorsList) == 0 {
		return defaultMonitoredSensors
	}

	return sensorsList
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
