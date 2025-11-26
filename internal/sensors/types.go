package sensors

import (
	"time"
	"github.com/jkaberg/byd-hass/internal/location"
)

// SensorData struct to hold all possible sensor values.
// We use pointers to float64 for numeric values so we can distinguish between a missing value (nil) and a value of 0.
type SensorData struct {
	Timestamp time.Time `json:"timestamp"`

	// --- Core Vehicle Data ---
	Speed            *float64 `json:"speed,omitempty"`
	Mileage          *float64 `json:"mileage,omitempty"`
	GearPosition     *float64 `json:"gear_position,omitempty"`
	PowerStatus      *float64 `json:"power_status,omitempty"`
	SteeringAngle    *float64 `json:"steering_angle,omitempty"`
	AcceleratorDepth *float64 `json:"accelerator_depth,omitempty"`
	BrakeDepth       *float64 `json:"brake_depth,omitempty"`

	// --- Powertrain & Battery ---
	EnginePower           *float64 `json:"engine_power,omitempty"`
	EngineRPM             *float64 `json:"engine_rpm,omitempty"`
	FrontMotorRPM         *float64 `json:"front_motor_rpm,omitempty"`
	FrontMotorTorque      *float64 `json:"front_motor_torque,omitempty"`
	RearMotorRPM          *float64 `json:"rear_motor_rpm,omitempty"`
	FuelPercentage        *float64 `json:"fuel_percentage,omitempty"`
	BatteryPercentage     *float64 `json:"battery_percentage,omitempty"`
	BatteryCapacity       *float64 `json:"battery_capacity,omitempty"`
	ChargingStatus        *float64 `json:"charging_status,omitempty"`
	ChargeGunState        *float64 `json:"charge_gun_state,omitempty"`
	MaxBatteryVoltage     *float64 `json:"max_battery_voltage,omitempty"`
	MinBatteryVoltage     *float64 `json:"min_battery_voltage,omitempty"`
	TotalPowerConsumption *float64 `json:"total_power_consumption,omitempty"`
	PowerConsumption100km *float64 `json:"power_consumption_100km,omitempty"`
	BatteryVoltage12V     *float64 `json:"battery_voltage_12v,omitempty"`

	// --- Temperature Sensors ---
	AvgBatteryTemp     *float64 `json:"avg_battery_temp,omitempty"`
	MinBatteryTemp     *float64 `json:"min_battery_temp,omitempty"`
	MaxBatteryTemp     *float64 `json:"max_battery_temp,omitempty"`
	CabinTemperature   *float64 `json:"cabin_temperature,omitempty"`
	OutsideTemperature *float64 `json:"outside_temperature,omitempty"`
	TemperatureUnit    *float64 `json:"temperature_unit,omitempty"`

	// --- Doors & Locks ---
	DriverDoor         *float64 `json:"driver_door,omitempty"`
	PassengerDoor      *float64 `json:"passenger_door,omitempty"`
	LeftRearDoor       *float64 `json:"left_rear_door,omitempty"`
	RightRearDoor      *float64 `json:"right_rear_door,omitempty"`
	TrunkDoor          *float64 `json:"trunk_door,omitempty"`
	Hood               *float64 `json:"hood,omitempty"`
	DriverDoorLock     *float64 `json:"driver_door_lock,omitempty"`
	PassengerDoorLock  *float64 `json:"passenger_door_lock,omitempty"`
	LeftRearDoorLock   *float64 `json:"left_rear_door_lock,omitempty"`
	RightRearDoorLock  *float64 `json:"right_rear_door_lock,omitempty"`
	TrunkLock          *float64 `json:"trunk_lock,omitempty"`
	RemoteLockStatus   *float64 `json:"remote_lock_status,omitempty"`
	LeftRearChildLock  *float64 `json:"left_rear_child_lock,omitempty"`
	RightRearChildLock *float64 `json:"right_rear_child_lock,omitempty"`

	// --- Windows & Sunroof ---
	DriverWindowOpenPercent    *float64 `json:"driver_window_open_percent,omitempty"`
	PassengerWindowOpenPercent *float64 `json:"passenger_window_open_percent,omitempty"`
	LeftRearWindowOpenPercent  *float64 `json:"left_rear_window_open_percent,omitempty"`
	RightRearWindowOpenPercent *float64 `json:"right_rear_window_open_percent,omitempty"`
	SunroofOpenPercent         *float64 `json:"sunroof_open_percent,omitempty"`
	SunshadeOpenPercent        *float64 `json:"sunshade_open_percent,omitempty"`

	// --- Tire Pressures ---
	LeftFrontTirePressure  *float64 `json:"left_front_tire_pressure,omitempty"`
	RightFrontTirePressure *float64 `json:"right_front_tire_pressure,omitempty"`
	LeftRearTirePressure   *float64 `json:"left_rear_tire_pressure,omitempty"`
	RightRearTirePressure  *float64 `json:"right_rear_tire_pressure,omitempty"`

	// --- Lights & Wipers ---
	LowBeamLights        *float64 `json:"low_beam_lights,omitempty"`
	HighBeamLights       *float64 `json:"high_beam_lights,omitempty"`
	FrontFogLights       *float64 `json:"front_fog_lights,omitempty"`
	RearFogLights        *float64 `json:"rear_fog_lights,omitempty"`
	ParkingLights        *float64 `json:"parking_lights,omitempty"`
	DaytimeRunningLights *float64 `json:"daytime_running_lights,omitempty"`
	LeftTurnSignal       *float64 `json:"left_turn_signal,omitempty"`
	RightTurnSignal      *float64 `json:"right_turn_signal,omitempty"`
	HazardLights         *float64 `json:"hazard_lights,omitempty"`
	WiperGear            *float64 `json:"wiper_gear,omitempty"`
	FrontWiperSpeed      *float64 `json:"front_wiper_speed,omitempty"`
	LastWiperTime        *float64 `json:"last_wiper_time,omitempty"`

	// --- Climate Control (AC) ---
	ACStatus            *float64 `json:"ac_status,omitempty"`
	DriverACTemperature *float64 `json:"driver_ac_temperature,omitempty"`
	FanSpeedLevel       *float64 `json:"fan_speed_level,omitempty"`
	ACBlowingMode       *float64 `json:"ac_blowing_mode,omitempty"`
	ACCirculationMode   *float64 `json:"ac_circulation_mode,omitempty"`
	Weather             *float64 `json:"weather,omitempty"`
	FootwellLights      *float64 `json:"footwell_lights,omitempty"`

	// --- Driving Assistance & Safety ---
	ACCCruiseStatus       *float64 `json:"acc_cruise_status,omitempty"`
	LaneKeepAssistStatus  *float64 `json:"lane_keep_assist_status,omitempty"`
	DriverSeatbelt        *float64 `json:"driver_seatbelt,omitempty"`
	PassengerSeatbeltWarn *float64 `json:"passenger_seatbelt_warn,omitempty"`
	Row2LeftSeatbelt      *float64 `json:"row2_left_seatbelt,omitempty"`
	Row2RightSeatbelt     *float64 `json:"row2_right_seatbelt,omitempty"`
	Row2CenterSeatbelt    *float64 `json:"row2_center_seatbelt,omitempty"`
	DistanceToCarAhead    *float64 `json:"distance_to_car_ahead,omitempty"`
	LaneCurvature         *float64 `json:"lane_curvature,omitempty"`
	RightLineDistance     *float64 `json:"right_line_distance,omitempty"`
	LeftLineDistance      *float64 `json:"left_line_distance,omitempty"`
	CruiseSwitch          *float64 `json:"cruise_switch,omitempty"`
	AutoParking           *float64 `json:"auto_parking,omitempty"`

	// --- Radar Sensors ---
	RadarFrontLeft          *float64 `json:"radar_front_left,omitempty"`
	RadarFrontRight         *float64 `json:"radar_front_right,omitempty"`
	RadarRearLeft           *float64 `json:"radar_rear_left,omitempty"`
	RadarRearRight          *float64 `json:"radar_rear_right,omitempty"`
	RadarLeft               *float64 `json:"radar_left,omitempty"`
	RadarFrontMidLeft       *float64 `json:"radar_front_mid_left,omitempty"`
	RadarFrontMidRight      *float64 `json:"radar_front_mid_right,omitempty"`
	RadarRearCenter         *float64 `json:"radar_rear_center,omitempty"`
	RearLeftProximityAlert  *float64 `json:"rear_left_proximity_alert,omitempty"`
	RearRightProximityAlert *float64 `json:"rear_right_proximity_alert,omitempty"`

	// --- Vehicle & System ---
	VehicleOperatingMode    *float64 `json:"vehicle_operating_mode,omitempty"`
	VehicleRunningMode      *float64 `json:"vehicle_running_mode,omitempty"`
	SurroundViewStatus      *float64 `json:"surround_view_status,omitempty"`
	UIConfigVersion         *float64 `json:"ui_config_version,omitempty"`
	SentryModeStatus        *float64 `json:"sentry_mode_status,omitempty"`
	PowerOffRecordingConfig *float64 `json:"power_off_recording_config,omitempty"`
	PowerOffSentryAlarm     *float64 `json:"power_off_sentry_alarm,omitempty"`
	WiFiStatus              *float64 `json:"wifi_status,omitempty"`
	BluetoothStatus         *float64 `json:"bluetooth_status,omitempty"`
	BluetoothSignalStrength *float64 `json:"bluetooth_signal_strength,omitempty"`
	WirelessADBSwitch       *float64 `json:"wireless_adb_switch,omitempty"`
	SteeringRotationSpeed   *float64 `json:"steering_rotation_speed,omitempty"`

	// --- AI & Video ---
	AIPersonConfidence     *float64 `json:"ai_person_confidence,omitempty"`
	AIVehicleConfidence    *float64 `json:"ai_vehicle_confidence,omitempty"`
	LastSentryTriggerTime  *float64 `json:"last_sentry_trigger_time,omitempty"`
	LastSentryTriggerImage *string  `json:"last_sentry_trigger_image,omitempty"`
	LastVideoStartTime     *float64 `json:"last_video_start_time,omitempty"`
	LastVideoEndTime       *float64 `json:"last_video_end_time,omitempty"`
	LastVideoPath          *string  `json:"last_video_path,omitempty"`

	// --- Location & Time ---
	Location *location.LocationData `json:"location,omitempty"`
	Year     *float64               `json:"year,omitempty"`
	Month    *float64               `json:"month,omitempty"`
	Day      *float64               `json:"day,omitempty"`
	Hour     *float64               `json:"hour,omitempty"`
	Minute   *float64               `json:"minute,omitempty"`
}

// SensorDefinition provides metadata for a sensor.
type SensorDefinition struct {
	ID                int
	FieldName         string
	ChineseName       string
	EnglishName       string
	Category          string // "sensor", "binary_sensor", "device_tracker"
	DeviceClass       string
	UnitOfMeasurement string
	ScaleFactor       float64
}

// ----------------------------------------------------------------------------
// AllSensors
// ------------
// This table contains one entry for every public field in SensorData that we
// want to surface to higher layers (Diplus polling → MQTT discovery → Home
// Assistant).  Each row provides the metadata needed to build the Diplus query
// template, scale raw values, and publish Home-Assistant discovery messages.
//
//	ID            – Stable numerical identifier (starts at 1, never reused)
//	FieldName     – _Exact_ Go struct field in SensorData (PascalCase)
//	ChineseName   – The precise label Diplus uses in its JSON output
//	EnglishName   – Clear English label for UIs / logs
//	Category      – "sensor" or "binary_sensor" (matches HA platform)
//	DeviceClass   – Optional Home-Assistant device_class (speed, voltage, …)
//	Unit          – Unit of measurement (km/h, °C, %, …) – empty if unit-less
//	ScaleFactor   – Multiply raw value by this to obtain the real value (1 = none)
//
// Whenever you add / remove a field in SensorData **make sure** to update this
// slice accordingly; build failures will warn you if you forget.
// ----------------------------------------------------------------------------
var AllSensors = []SensorDefinition{
	{1, "PowerStatus", "电源状态", "Power Status", "sensor", "", "", 1},
	{2, "Speed", "车速", "Speed", "sensor", "speed", "km/h", 1},
	{3, "Mileage", "里程", "Mileage", "sensor", "distance", "km", 0.1},
	{4, "GearPosition", "档位", "Gear Position", "sensor", "", "", 1},
	{5, "EngineRPM", "发动机转速", "Engine RPM", "sensor", "", "rpm", 1},
	{6, "BrakePedalDepth", "刹车深度", "Brake Pedal Depth", "sensor", "", "%", 1},
	{7, "AcceleratorPedalDepth", "加速踏板深度", "Accelerator Pedal Depth", "sensor", "", "%", 1},
	{8, "FrontMotorRPM", "前电机转速", "Front Motor RPM", "sensor", "", "rpm", 1},
	{9, "RearMotorRPM", "后电机转速", "Rear Motor RPM", "sensor", "", "rpm", 1},
	{10, "EnginePower", "发动机功率", "Engine Power", "sensor", "power", "kW", 1},
	{11, "FrontMotorTorque", "前电机扭矩", "Front Motor Torque", "sensor", "", "Nm", 1},
	{12, "ChargeGunState", "充电枪插枪状态", "Charge Gun State", "binary_sensor", "", "", 1},
	{13, "PowerConsumption100KM", "百公里电耗", "Power consumption per 100 kilometers", "sensor", "", "kWh/100km", 1},
	{14, "MaxBatteryTemp", "最高电池温度", "Maximum Battery Temperature", "sensor", "temperature", "°C", 1},
	{15, "AvgBatteryTemp", "平均电池温度", "Average Battery Temperature", "sensor", "temperature", "°C", 1},
	{16, "MinBatteryTemp", "最低电池温度", "Minimum Battery Temperature", "sensor", "", "°C", 1},
	{17, "MaxBatteryVoltage", "最高电池电压", "Max Battery Voltage", "sensor", "voltage", "V", 1}, // This is the 12V battery voltage
	{18, "MinBatteryVoltage", "最低电池电压", "Minimum Battery Voltage", "sensor", "", "V", 1},
	{19, "LastWiperTime", "上次雨刮时间", "Last Wiper Time", "sensor", "timestamp", "", 1},
	{20, "Weather", "天气", "Weather", "sensor", "distance", "", 1},
	{21, "DriverSeatBeltStatus", "主驾驶安全带状态", "Driver's seat belt status", "binary_sensor", "", "", 1},
	{22, "RemoteLockStatus", "远程锁车状态", "Remote Lock Status", "binary_sensor", "lock", "", 1},
	// what is ID 23 and 24? not documeneted in the spec.
	{25, "CabinTemperature", "车内温度", "Cabin Temperature", "sensor", "", "°C", 1},
	{26, "OutsideTemperature", "车外温度", "Outside Temperature", "sensor", "temperature", "°C", 1},
	{27, "DriverACTemp", "主驾驶空调温度", "Driver AC temperature", "sensor", "", "°C", 1},
	{28, "TemperatureUnit", "温度单位", "Temperature unit", "sensor", "", "", 1},
	{29, "BatteryCapacity", "电池容量", "Battery Capacity", "sensor", "energy_storage", "kWh", 1}, // seems to be 0 all the time?
	{30, "SteeringWheelAngle", "方向盘转角", "Steering Wheel Angle", "sensor", "safety", "°", 1},
	{31, "SteeringWheelSpeed", "方向盘转速", "Steering Sheel Speed", "sensor", "safety", "°/s", 1},
	{32, "TotalPowerConsumption", "总电耗", "Total Power Consumption", "sensor", "safety", "kWh", 1},
	{33, "BatteryPercentage", "电量百分比", "Battery Percentage", "sensor", "battery", "%", 1},
	{34, "FuelPercentage", "油量百分比", "Fuel Percentage", "sensor", "battery", "%", 1},
	{35, "TotalFuelConsumption", "总燃油消耗", "Total Fuel Consumption", "sensor", "timestamp", "L", 1},
	{36, "LaneLineCurvature", "车道线曲率", "Lane Line Curvature", "sensor", "timestamp", "", 1},
	{37, "RightLaneDistance", "右侧线距离", "Right Lane Distance", "sensor", "timestamp", "", 1},
	{38, "LeftLaneDistance", "左侧线距离", "Left Lane Distance", "sensor", "timestamp", "", 1},
	{39, "BatteryVoltage", "蓄电池电压", "Battery Voltage", "sensor", "", "", 1}, // seems to be 0 all the time?
	{40, "RadarLeftFront", "雷达左前", "Radar Left Front", "sensor", "", "m", 1},
	{41, "RadarRightFront", "雷达右前", "Radar Right Front", "sensor", "", "m", 1},
	{42, "RadarLeftRear", "雷达左后", "Radar Left Rear", "sensor", "", "m", 1},
	{43, "RadarRightRear", "雷达右后", "Radar Right Rear", "sensor", "", "m", 1},
	{44, "RadarLeft", "雷达左", "Radar Left", "sensor", "", "m", 1},
	{45, "RadarFrontLeftCenter", "雷达前左中", "Radar Front Left Center", "sensor", "distance", "m", 1},
	{46, "RadarFrontRightCenter", "雷达前右中", "Radar Front Right Center", "sensor", "distance", "m", 1},
	{47, "RadarCenterRear", "雷达中后", "Radar Center Rear", "sensor", "distance", "m", 1},
	{48, "FrontWiperSpeed", "前雨刮速度", "Front Wiper Speed", "sensor", "", "", 1},
	{49, "WiperGear", "雨刮档位", "WiperGear", "sensor", "", "", 1},
	{50, "CruiseSwitch", "巡航开关", "Cruise Switch", "binary_sensor", "", "", 1},
	{51, "DistanceToVehicleAhead", "前车距离", "Distance To The Vehicle Ahead", "sensor", "distance", "m", 1},
	{52, "ChargingStatus", "充电状态", "Charging Status", "sensor", "", "", 1},
	{53, "LeftFrontTirePressure", "左前轮气压", "Left Front Tire Pressure", "sensor", "pressure", "bar", 0.01},
	{54, "RightFrontTirePressure", "右前轮气压", "Right Front Tire Pressure", "sensor", "pressure", "bar", 0.01},
	{55, "LeftRearTirePressure", "左后轮气压", "Left Rear Tire Pressure", "sensor", "pressure", "bar", 0.01},
	{56, "RightRearTirePressure", "右后轮气压", "Right Rear Tire Pressure", "sensor", "pressure", "bar", 0.01},
	{57, "LeftTurnSignal", "左转向灯", "Left Turn Signal", "binary_sensor", "light", "", 1},
	{58, "RightTurnSignal", "右转向灯", "Right Turn Signal", "binary_sensor", "light", "", 1},
	{59, "DriverDoorLock", "主驾车门锁", "Driver Door Lock", "binary_sensor", "light", "", 1},
	// what is ID 60? not documeneted in the spec.
	{61, "DriverWindowOpenPercentage", "主驾车窗打开百分比", "Driver Window Open Percentage", "sensor", "light", "%", 1},
	{62, "PassengerWindowOpenPercentage", "副驾车窗打开百分比", "Passenger Window Open Percentage", "sensor", "light", "%", 1},
	{63, "LeftLearWindowOpenPercentage", "左后车窗打开百分比", "Left Rear Window Open Percentage", "sensor", "light", "%", 1},
	{64, "RightRearWindowOpenPercentage", "右后车窗打开百分比", "Right Rear Window Open Percentage", "sensor", "light", "%", 1},
	{65, "SunroofOpenPercentage", "天窗打开百分比", "Sunroof Open Percentage", "sensor", "light", "%", 1},
	{66, "SunshadeOpenPercentage", "遮阳帘打开百分比", "SunshadeOpenPercentage", "sensor", "door", "%", 1},
	{67, "VehicleWorkingMode", "整车工作模式", "Vehicle Working Mode", "sensor", "door", "", 1},
	{68, "VehicleOperationMode", "整车运行模式", "Vehicle Operation Mode", "sensor", "door", "", 1},
	{69, "Month", "月", "Month", "sensor", "door", "", 1},
	{70, "Day", "日", "Day", "sensor", "door", "", 1},
	{71, "Hour", "时", "Hour", "sensor", "door", "", 1},
	{72, "Year", "分", "Year", "sensor", "lock", "", 1},
	{73, "PassengerSeatBeltWarning", "副驾安全带警告", "Passenger Seat Belt Warning", "binary_sensor", "lock", "", 1},
	{74, "SecondRowLeftSeatBelt", "二排左安全带", "Second Row Left Seat Belt", "binary_sensor", "lock", "", 1},
	{75, "SecondRowRightSeatBelt", "二排右安全带", "Second Row Right Seat Belt", "binary_sensor", "lock", "", 1},
	{76, "Second Row Center Seat Belt", "二排中安全带", "Second Row Center Seat Belt", "binary_sensor", "lock", "", 1},
	{77, "ACStatus", "空调状态", "AC Status", "sensor", "", "", 1},
	{78, "FanSpeedLevel", "风量档位", "Fan Speed Level", "sensor", "", "", 1},
	{79, "ACCirculationMode", "空调循环方式", "AC Circulation Mode", "sensor", "", "", 1},
	{80, "AC Outlet Mode", "空调出风模式", "AC Outlet Mode", "sensor", "", "", 1},
	{81, "DriverDoor", "主驾车门", "Driver Door", "binary_sensor", "", "", 1},
	{82, "PassengerDoor", "副驾车门", "Passenger Door", "binary_sensor", "safety", "", 1},
	{83, "LeftRearDoor", "左后车门", "Left Rear Door", "binary_sensor", "safety", "", 1},
	{84, "RightRearDoor", "右后车门", "Right Rear Door", "binary_sensor", "", "", 1},
	{85, "Hood", "引擎盖", "Hood", "binary_sensor", "power", "", 1},
	{86, "Trunk", "后备箱门", "Trunk", "binary_sensor", "", "", 1},
	{87, "FuelTankCap", "油箱盖", "Fuel Tank Cap", "binary_sensor", "", "", 1},
	{88, "AutomaticParking", "自动驻车", "Automatic Parking", "binary_sensor", "", "", 1},
	{89, "ACCCruiseStatus", "ACC巡航状态", "ACC Cruise Status", "sensor", "", "", 1},
	{90, "LeftRearApproachWarning", "左后接近告警", "Left Rear Approach Warning", "binary_sensor", "power", "", 1},
	{91, "RightRearApproachWarning", "右后接近告警", "Right Rear Approach Warning", "binary_sensor", "", "", 1},
	{92, "Lane Keeping Status", "车道保持状态", "Lane Keeping Status", "sensor", "", "", 1},
	{93, "LeftRearDoorLock", "左后车门锁", "Left Rear Door Lock", "binary_sensor", "", "", 1},
	{94, "PassengerDoorLock", "副驾车门锁", "Passenger Door Lock", "binary_sensor", "", "", 1},
	{95, "RightRearDoorLock", "上次雨刮时间", "Right Rear Door Lock", "binary_sensor", "", "", 1},
	{96, "TrunkDoorLock", "后备箱门锁", "Trunk Toor Lock", "binary_sensor", "", "", 1},
	{97, "LeftRearChildLock", "左后儿童锁", "Left Rear Child Lock", "binary_sensor", "", "", 1},
	{98, "RightRearChildLock", "右后儿童锁", "Right Rear Child Lock", "binary_sensor", "", "", 1},
	{99, "LowBeam", "小灯", "Low Beam", "binary_sensor", "", "", 1},
	{100, "LowBeam2", "近光灯", "Low Beam", "binary_sensor", "", "", 1},
	{101, "HighBeam", "远光灯", "High Beam", "binary_sensor", "lock", "", 1},
	// what is ID 102 and 103? not documeneted in the spec.
	{104, "FrontFogLamp", "前雾灯", "Front Fog Lamp", "binary_sensor", "", "", 1},
	{105, "RearFogLamp", "后雾灯", "Rear Fog Lamp", "binary_sensor", "", "", 1},
	{106, "Footlights", "脚照灯", "Footlights", "binary_sensor", "", "", 1},
	{107, "DaytimeRunningLights", "日行灯", "Daytime Running Lights", "binary_sensor", "", "", 1},
	{108, "EngineWaterTemperature", "发动机水温", "Engine Water Temperature", "sensor", "", "°C", 1},
	{109, "DoubleFlash", "双闪", "DoubleFlash", "binary_sensor", "", "", 1},

	{1001, "PanoramaStatus", "熄火录制配置", "PanoramaStatus", "binary_sensor", "", "", 1},
	{1002, "ConfigUIVer", "熄火哨兵警报", "Configuration UI Version", "binary_sensor", "", "", 1},
	{1003, "SentryStatus", "WiFi状态", "Sentry Status", "binary_sensor", "connectivity", "", 1},
	{1004, "RecordingConfigSwitch", "蓝牙状态", "Recording Configuration Switch", "binary_sensor", "connectivity", "", 1},
	{1006, "SentryAlarm", "蓝牙信号强度", "Sentry Alarm", "sensor", "signal_strength", "dBm", 1},
	{1007, "WIFIStatus", "上次哨兵触发时间", "WIFI Status", "sensor", "timestamp", "", 1},
	{1008, "BluetoothStatus", "上次哨兵触发图像", "Bluetooth Status", "sensor", "", "", 1},
	{1009, "BluetoothSignalStrength", "上次录像开始时间", "Bluetooth Signal Strength", "sensor", "timestamp", "", 1},
	{1101, "WirelessADBSwitch", "上次录像结束时间", "Wireless ADB Switch", "binary_sensor", "timestamp", "", 1},
}

// GetSensorByID returns a sensor definition by its ID
func GetSensorByID(id int) *SensorDefinition {
	for _, sensor := range AllSensors {
		if sensor.ID == id {
			return &sensor
		}
	}
	return nil
}

// GetScaleFactor returns the scaling factor for a given JSON field key (snake_case).
// If no explicit factor is defined, 1.0 is returned.
func GetScaleFactor(jsonKey string) float64 {
	factor := 1.0
	for _, s := range AllSensors {
		if ToSnakeCase(s.FieldName) == jsonKey {
			if s.ScaleFactor != 0 {
				factor = s.ScaleFactor // keep updating; last match wins
			}
		}
	}
	return factor
}
