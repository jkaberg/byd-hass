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

// Default monitors – expanded version
/* var defaultMonitoredSensors = []MonitoredSensor{
	// 1‑12 ----------------------------------------------------
	{ID: 1, Publish: true},   // PowerStatus
	{ID: 2, Publish: true},   // Speed
	{ID: 3, Publish: true},   // Mileage
	{ID: 4, Publish: true},   // GearPosition
	{ID: 5, Publish: true},   // EngineRPM
	{ID: 6, Publish: true},   // BrakePedalDepth
	{ID: 7, Publish: true},   // AcceleratorPedalDepth
	{ID: 8, Publish: true},   // FrontMotorRPM
	{ID: 9, Publish: true},   // RearMotorRPM
	{ID: 10, Publish: true},  // EnginePower
	{ID: 11, Publish: true},  // FrontMotorTorque
	{ID: 12, Publish: false}, // ChargeGunState (internal‑only)

	//	{ID: 12, Publish: true}, // ChargeGunState

	// 13‑22 --------------------------------------------------- 
	{ID: 13, Publish: true}, // PowerConsumption100KM
	{ID: 14, Publish: true}, // MaxBatteryTemp
	{ID: 15, Publish: true}, // AvgBatteryTemp
	{ID: 16, Publish: true}, // MinBatteryTemp
	{ID: 17, Publish: true}, // MaxBatteryVoltage
	{ID: 18, Publish: true}, // MinBatteryVoltage
	{ID: 19, Publish: true}, // LastWiperTime
	{ID: 20, Publish: true}, // Weather
	{ID: 21, Publish: true}, // DriverSeatBeltStatus
	{ID: 22, Publish: true}, // RemoteLockStatus

	// 23‑24 --------------------------------------------------- 
	// IDs 23 and 24 are not documented in the spec – they have never been
	// present in the XML, so they are omitted here.

	// 25‑34 --------------------------------------------------- 
	{ID: 25, Publish: true}, // CabinTemperature
	{ID: 26, Publish: true}, // OutsideTemperature
	{ID: 27, Publish: true}, // DriverACTemp
	{ID: 28, Publish: true}, // TemperatureUnit
	{ID: 29, Publish: true}, // BatteryCapacity
	{ID: 30, Publish: true}, // SteeringWheelAngle
	{ID: 31, Publish: true}, // SteeringWheelSpeed
	{ID: 32, Publish: true}, // TotalPowerConsumption
	{ID: 33, Publish: true}, // BatteryPercentage
	{ID: 34, Publish: true}, // FuelPercentage

	// 35‑44 --------------------------------------------------- 
	{ID: 35, Publish: true}, // TotalFuelConsumption
	{ID: 36, Publish: true}, // LaneLineCurvature
	{ID: 37, Publish: true}, // RightLaneDistance
	{ID: 38, Publish: true}, // LeftLaneDistance
	{ID: 39, Publish: true}, // BatteryVoltage
	{ID: 40, Publish: true}, // RadarLeftFront
	{ID: 41, Publish: true}, // RadarRightFront
	{ID: 42, Publish: true}, // RadarLeftRear
	{ID: 43, Publish: true}, // RadarRightRear

	// 45‑56 --------------------------------------------------- 
	{ID: 44, Publish: true}, // RadarLeft
	{ID: 45, Publish: true}, // RadarFrontLeftCenter
	{ID: 46, Publish: true}, // RadarFrontRightCenter
	{ID: 47, Publish: true}, // RadarCenterRear
	{ID: 48, Publish: true}, // FrontWiperSpeed
	{ID: 49, Publish: true}, // WiperGear
	{ID: 50, Publish: true}, // CruiseSwitch (binary_sensor)
	{ID: 51, Publish: true}, // DistanceToVehicleAhead
	{ID: 52, Publish: true}, // ChargingStatus
	{ID: 53, Publish: true}, // LeftFrontTirePressure
	{ID: 54, Publish: true}, // RightFrontTirePressure
	{ID: 55, Publish: true}, // LeftRearTirePressure
	{ID: 56, Publish: true}, // RightRearTirePressure

	// 57‑66 ---------------------------------------------------
	{ID: 57, Publish: true}, // LeftTurnSignal (binary_sensor)
	{ID: 58, Publish: true}, // RightTurnSignal (binary_sensor)
	{ID: 59, Publish: true}, // DriverDoorLock (binary_sensor)
	// ID 60 is undocumented in the spec – it never appears in the XML.

	{ID: 61, Publish: true}, // DriverWindowOpenPercentage
	{ID: 62, Publish: true}, // PassengerWindowOpenPercentage
	{ID: 63, Publish: true}, // LeftLearWindowOpenPercentage
	{ID: 64, Publish: true}, // RightRearWindowOpenPercentage
	{ID: 65, Publish: true}, // SunroofOpenPercentage
	{ID: 66, Publish: true}, // SunshadeOpenPercentage

	// 67‑72 ---------------------------------------------------
	{ID: 67, Publish: true}, // VehicleWorkingMode
	{ID: 68, Publish: true}, // VehicleOperationMode
	{ID: 69, Publish: true}, // Month
	{ID: 70, Publish: true}, // Day
	{ID: 71, Publish: true}, // Hour
	{ID: 72, Publish: true}, // Year

	// 73‑84 ---------------------------------------------------
	{ID: 73, Publish: true}, // PassengerSeatBeltWarning (binary_sensor)
	{ID: 74, Publish: true}, // SecondRowLeftSeatBelt (binary_sensor)
	{ID: 75, Publish: true}, // SecondRowRightSeatBelt (binary_sensor)
	{ID: 76, Publish: true}, // Second Row Center Seat Belt (binary_sensor)
	{ID: 77, Publish: true}, // ACStatus
	{ID: 78, Publish: true}, // FanSpeedLevel
	{ID: 79, Publish: true}, // ACCirculationMode
	{ID: 80, Publish: true}, // ACBlowingMode
	{ID: 81, Publish: true}, // DriverDoor (binary_sensor)
	{ID: 82, Publish: true}, // PassengerDoor (binary_sensor)
	{ID: 83, Publish: true}, // LeftRearDoor (binary_sensor)
	{ID: 84, Publish: true}, // RightRearDoor (binary_sensor)

	// 85‑107 --------------------------------------------------
	{ID: 85, Publish: true}, // Hood (binary_sensor)
	{ID: 86, Publish: true}, // Trunk (binary_sensor)
	{ID: 87, Publish: true}, // FuelTankCap (binary_sensor)
	{ID: 88, Publish: true}, // AutomaticParking (binary_sensor)
	{ID: 89, Publish: true}, // ACCCruiseStatus
	{ID: 90, Publish: true}, // LeftRearApproachWarning (binary_sensor)
	{ID: 91, Publish: true}, // RightRearApproachWarning (binary_sensor)
	{ID: 92, Publish: true}, // Lane Keeping Status
	{ID: 93, Publish: true}, // LeftRearDoorLock (binary_sensor)
	{ID: 94, Publish: true}, // PassengerDoorLock (binary_sensor)
	{ID: 95, Publish: true}, // RightRearDoorLock (binary_sensor)   // note: name in XML is “上次雨刮时间”, but it represents the right rear door lock
	{ID: 96, Publish: true}, // TrunkDoorLock (binary_sensor)
	{ID: 97, Publish: true}, // LeftRearChildLock (binary_sensor)
	{ID: 98, Publish: true}, // RightRearChildLock (binary_sensor)
	{ID: 99, Publish: true}, // LowBeam (binary_sensor)
	{ID: 100, Publish: true}, // LowBeam2 (binary_sensor)
	{ID: 101, Publish: true}, // HighBeam (binary_sensor)
	// IDs 102 and 103 are undocumented – they never appear in the XML.

	{ID: 104, Publish: true}, // FrontFogLamp (binary_sensor)
	{ID: 105, Publish: true}, // RearFogLamp (binary_sensor)
	{ID: 106, Publish: true}, // Footlights (binary_sensor)
	{ID: 107, Publish: true}, // DaytimeRunningLights (binary_sensor)
	{ID: 108, Publish: true}, // EngineWaterTemperature
	{ID: 109, Publish: true}, // DoubleFlash (binary_sensor)

	// 1001‑2007 -----------------------------------------------
	{ID: 1001, Publish: true}, // PanoramaStatus (binary_sensor)
	{ID: 1002, Publish: true}, // ConfigUIVer (binary_sensor)
	{ID: 1003, Publish: true}, // SentryStatus
	{ID: 1004, Publish: true}, // RecordingConfigSwitch
	{ID: 1006, Publish: true}, // SentryAlarm (signal_strength)
	{ID: 1007, Publish: true}, // WIFIStatus
	{ID: 1008, Publish: true}, // BluetoothStatus
	{ID: 1009, Publish: true}, // BluetoothSignalStrength

	{ID: 1101, Publish: true}, // WirelessADBSwitch (binary_sensor)

	{ID: 2001, Publish: true}, // AIPersonConfidence
	{ID: 2002, Publish: true}, // AIVehicleConfidence
	{ID: 2003, Publish: true}, // LastSentryTriggerTime
	{ID: 2004, Publish: true}, // LastSentryTriggerImage
	{ID: 2005, Publish: true}, // LastVideoStartTime
	{ID: 2006, Publish: true}, // LastVideoEndTime
	{ID: 2007, Publish: true}, // LastVideoPath.
} */

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
