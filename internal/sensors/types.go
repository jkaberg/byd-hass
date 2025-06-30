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

// AllSensors defines the metadata for all known sensors.
var AllSensors = []SensorDefinition{
	{1, "PowerStatus", "电源状态", "Power Status", "sensor", "", "", 1},
	{2, "Speed", "车速", "Speed", "sensor", "speed", "km/h", 1},
	{3, "Mileage", "里程", "Mileage", "sensor", "distance", "km", 1},
	{4, "GearPosition", "档位", "Gear Position", "sensor", "", "", 1},
	{5, "EngineRPM", "发动机转速", "Engine RPM", "sensor", "", "rpm", 1},
	{6, "BrakeDepth", "刹车深度", "Brake Depth", "sensor", "", "%", 1},
	{7, "AcceleratorDepth", "加速踏板深度", "Accelerator Depth", "sensor", "", "%", 1},
	{8, "FrontMotorRPM", "前电机转速", "Front Motor RPM", "sensor", "", "rpm", 1},
	{9, "RearMotorRPM", "后电机转速", "Rear Motor RPM", "sensor", "", "rpm", 1},
	{10, "EnginePower", "发动机功率", "Engine Power", "sensor", "power", "kW", 1},
	{11, "FrontMotorTorque", "前电机扭矩", "Front Motor Torque", "sensor", "", "Nm", 1},
	{12, "LastWiperTime", "充电枪插枪状态", "Last Wiper Time", "sensor", "timestamp", "", 1},
	{13, "Weather", "百公里电耗", "Weather", "sensor", "", "", 1},
	{14, "CabinTemperature", "最高电池温度", "Cabin Temperature", "sensor", "temperature", "°C", 1},
	{15, "AvgBatteryTemp", "平均电池温度", "Average Battery Temperature", "sensor", "temperature", "°C", 1},
	{16, "TemperatureUnit", "最低电池温度", "Temperature Unit", "sensor", "", "", 1},
	{17, "MaxBatteryVoltage", "最高电池电压", "Max Battery Voltage", "sensor", "voltage", "V", 1},
	{18, "SteeringRotationSpeed", "最低电池电压", "Steering Rotation Speed", "sensor", "", "deg/s", 1},
	{19, "LaneCurvature", "上次雨刮时间", "Lane Curvature", "sensor", "", "", 1},
	{20, "RightLineDistance", "天气", "Right Lane Distance", "sensor", "distance", "m", 1},
	{21, "LeftLineDistance", "主驾驶安全带状态", "Left Lane Distance", "sensor", "distance", "m", 1},
	{22, "CruiseSwitch", "远程锁车状态", "Cruise Switch", "binary_sensor", "power", "", 1},
	{23, "FootwellLights", "脚部照明灯", "Footwell Lights", "binary_sensor", "light", "", 1},
	{24, "FrontWiperSpeed", "前雨刮速度", "Front Wiper Speed", "sensor", "", "", 1},
	{25, "AutoParking", "车内温度", "Auto Parking", "binary_sensor", "", "", 1},
	{26, "OutsideTemperature", "车外温度", "Outside Temperature", "sensor", "temperature", "°C", 1},
	{27, "AIVehicleConfidence", "主驾驶空调温度", "AI Vehicle Confidence", "sensor", "", "%", 1},
	{28, "AIPersonConfidence", "温度单位", "AI Person Confidence", "sensor", "", "%", 1},
	{29, "BatteryCapacity", "电池容量", "Battery Capacity", "sensor", "energy_storage", "kWh", 1},
	{30, "Row2LeftSeatbelt", "方向盘转角", "Row 2 Left Seatbelt", "binary_sensor", "safety", "", 1},
	{31, "Row2RightSeatbelt", "方向盘转速", "Row 2 Right Seatbelt", "binary_sensor", "safety", "", 1},
	{32, "Row2CenterSeatbelt", "总电耗", "Row 2 Center Seatbelt", "binary_sensor", "safety", "", 1},
	{33, "BatteryPercentage", "电量百分比", "Battery Percentage", "sensor", "battery", "%", 1},
	{34, "Year", "油量百分比", "Year", "sensor", "timestamp", "", 1},
	{35, "Month", "总燃油消耗", "Month", "sensor", "timestamp", "", 1},
	{36, "Day", "车道线曲率", "Day", "sensor", "timestamp", "", 1},
	{37, "Hour", "右侧线距离", "Hour", "sensor", "timestamp", "", 1},
	{38, "Minute", "左侧线距离", "Minute", "sensor", "timestamp", "", 1},
	{39, "VehicleOperatingMode", "蓄电池电压", "Vehicle Operating Mode", "sensor", "", "", 1},
	{40, "VehicleRunningMode", "雷达左前", "Vehicle Running Mode", "sensor", "", "", 1},
	{41, "LeftRearChildLock", "雷达右前", "Left Rear Child Lock", "binary_sensor", "lock", "", 1},
	{42, "RightRearChildLock", "雷达左后", "Right Rear Child Lock", "binary_sensor", "lock", "", 1},
	{43, "RearLeftProximityAlert", "雷达右后", "Rear Left Proximity Alert", "binary_sensor", "safety", "", 1},
	{44, "RearRightProximityAlert", "雷达左", "Rear Right Proximity Alert", "binary_sensor", "safety", "", 1},
	{45, "RadarFrontLeft", "雷达前左中", "Radar Front Left", "sensor", "distance", "m", 1},
	{46, "RadarFrontRight", "雷达前右中", "Radar Front Right", "sensor", "distance", "m", 1},
	{47, "RadarRearLeft", "雷达中后", "Radar Rear Left", "sensor", "distance", "m", 1},
	{48, "RadarRearRight", "前雨刮速度", "Radar Rear Right", "sensor", "distance", "m", 1},
	{49, "RadarLeft", "雨刮档位", "Radar Left", "sensor", "distance", "m", 1},
	{50, "RadarFrontMidLeft", "巡航开关", "Radar Front Mid Left", "sensor", "distance", "m", 1},
	{51, "RadarFrontMidRight", "前车距离", "Radar Front Mid Right", "sensor", "distance", "m", 1},
	{52, "ChargingStatus", "充电状态", "Charging Status", "sensor", "", "", 1},
	{53, "LeftFrontTirePressure", "左前轮气压", "Left Front Tire Pressure", "sensor", "pressure", "bar", 0.01},
	{54, "RightFrontTirePressure", "右前轮气压", "Right Front Tire Pressure", "sensor", "pressure", "bar", 0.01},
	{55, "LeftRearTirePressure", "左后轮气压", "Left Rear Tire Pressure", "sensor", "pressure", "bar", 0.01},
	{56, "RightRearTirePressure", "右后轮气压", "Right Rear Tire Pressure", "sensor", "pressure", "bar", 0.01},
	{57, "ParkingLights", "左转向灯", "Parking Lights", "binary_sensor", "light", "", 1},
	{58, "LowBeamLights", "右转向灯", "Low Beam Lights", "binary_sensor", "light", "", 1},
	{59, "HighBeamLights", "主驾车门锁", "High Beam Lights", "binary_sensor", "light", "", 1},
	{60, "FrontFogLights", "前雾灯", "Front Fog Lights", "binary_sensor", "light", "", 1},
	{61, "RearFogLights", "主驾车窗打开百分比", "Rear Fog Lights", "binary_sensor", "light", "", 1},
	{62, "DaytimeRunningLights", "副驾车窗打开百分比", "Daytime Running Lights", "binary_sensor", "light", "", 1},
	{63, "LeftTurnSignal", "左后车窗打开百分比", "Left Turn Signal", "binary_sensor", "light", "", 1},
	{64, "RightTurnSignal", "右后车窗打开百分比", "Right Turn Signal", "binary_sensor", "light", "", 1},
	{65, "HazardLights", "天窗打开百分比", "Hazard Lights", "binary_sensor", "light", "", 1},
	{66, "DriverDoor", "遮阳帘打开百分比", "Driver Door", "binary_sensor", "door", "", 1},
	{67, "PassengerDoor", "整车工作模式", "Passenger Door", "binary_sensor", "door", "", 1},
	{68, "LeftRearDoor", "整车运行模式", "Left Rear Door", "binary_sensor", "door", "", 1},
	{69, "RightRearDoor", "月", "Right Rear Door", "binary_sensor", "door", "", 1},
	{70, "TrunkDoor", "日", "Trunk Door", "binary_sensor", "door", "", 1},
	{71, "Hood", "时", "Hood", "binary_sensor", "door", "", 1},
	{72, "DriverDoorLock", "分", "Driver Door Lock", "binary_sensor", "lock", "", 1},
	{73, "PassengerDoorLock", "副驾安全带警告", "Passenger Door Lock", "binary_sensor", "lock", "", 1},
	{74, "LeftRearDoorLock", "二排左安全带", "Left Rear Door Lock", "binary_sensor", "lock", "", 1},
	{75, "RightRearDoorLock", "二排右安全带", "Right Rear Door Lock", "binary_sensor", "lock", "", 1},
	{76, "TrunkLock", "二排中安全带", "Trunk Lock", "binary_sensor", "lock", "", 1},
	{77, "SunroofOpenPercent", "空调状态", "Sunroof Open Percent", "sensor", "", "%", 1},
	{78, "DriverWindowOpenPercent", "风量档位", "Driver Window Open Percent", "sensor", "", "%", 1},
	{79, "PassengerWindowOpenPercent", "空调循环方式", "Passenger Window Open Percent", "sensor", "", "%", 1},
	{80, "LeftRearWindowOpenPercent", "空调出风模式", "Left Rear Window Open Percent", "sensor", "", "%", 1},
	{81, "RightRearWindowOpenPercent", "主驾车门", "Right Rear Window Open Percent", "sensor", "", "%", 1},
	{82, "DriverSeatbelt", "副驾车门", "Driver Seatbelt", "binary_sensor", "safety", "", 1},
	{83, "PassengerSeatbeltWarn", "左后车门", "Passenger Seatbelt Warn", "binary_sensor", "safety", "", 1},
	{84, "PowerStatus", "右后车门", "Power Status", "sensor", "", "", 1}, // Duplicate ID 1
	{85, "ACStatus", "引擎盖", "AC Status", "binary_sensor", "power", "", 1},
	{86, "DriverACTemperature", "后备箱门", "Driver AC Temperature", "sensor", "temperature", "°C", 1},
	{87, "ACCCruiseStatus", "油箱盖", "ACC Cruise Status", "binary_sensor", "", "", 1},
	{88, "WiperGear", "自动驻车", "Wiper Gear", "sensor", "", "", 1},
	{89, "TotalPowerConsumption", "ACC巡航状态", "Total Power Consumption", "sensor", "energy", "kWh", 1},
	{90, "WirelessADBSwitch", "左后接近告警", "Wireless ADB Switch", "binary_sensor", "power", "", 1},
	{91, "SteeringAngle", "右后接近告警", "Steering Angle", "sensor", "", "°", 1},
	{92, "AcceleratorDepth", "车道保持状态", "Accelerator Depth", "sensor", "", "%", 1},
	{93, "BrakeDepth", "左后车门锁", "Brake Depth", "sensor", "", "%", 1},
	{94, "MinBatteryVoltage", "副驾车门锁", "Min Battery Voltage", "sensor", "voltage", "V", 1},
	{95, "ChargeGunState", "右后车门锁", "Charge Gun State", "sensor", "", "", 1},
	{96, "MinBatteryTemp", "后备箱门锁", "Min Battery Temp", "sensor", "temperature", "°C", 1},
	{97, "MaxBatteryTemp", "左后儿童锁", "Max Battery Temp", "sensor", "temperature", "°C", 1},
	{98, "FanSpeedLevel", "右后儿童锁", "Fan Speed Level", "sensor", "", "", 1},
	{99, "ACBlowingMode", "小灯", "AC Blowing Mode", "sensor", "", "", 1},
	{100, "ACCirculationMode", "近光灯", "AC Circulation Mode", "sensor", "", "", 1},
	{101, "RemoteLockStatus", "远光灯", "Remote Lock Status", "binary_sensor", "lock", "", 1},
	{102, "SunshadeOpenPercent", "遮阳帘开启百分比", "Sunshade Open Percent", "sensor", "", "%", 1},
	{103, "LaneKeepAssistStatus", "车道保持辅助状态", "Lane Keep Assist Status", "binary_sensor", "safety", "", 1},
	{104, "DistanceToCarAhead", "前雾灯", "Distance To Car Ahead", "sensor", "distance", "m", 1},
	{105, "BatteryVoltage12V", "后雾灯", "12V Battery Voltage", "sensor", "voltage", "V", 1},
	{106, "PowerConsumption100km", "脚照灯", "Power Consumption 100km", "sensor", "energy", "kWh/100km", 1},
	{107, "SurroundViewStatus", "日行灯", "Surround View Status", "binary_sensor", "", "", 1},
	{108, "UIConfigVersion", "发动机水温", "UI Config Version", "sensor", "", "", 1},
	{109, "SentryModeStatus", "双闪", "Sentry Mode Status", "binary_sensor", "", "", 1},
	{110, "PowerOffRecordingConfig", "熄火录制配置", "Power Off Recording Config", "binary_sensor", "", "", 1},
	{111, "PowerOffSentryAlarm", "熄火哨兵警报", "Power Off Sentry Alarm", "binary_sensor", "", "", 1},
	{112, "WiFiStatus", "WiFi状态", "WiFi Status", "binary_sensor", "connectivity", "", 1},
	{113, "BluetoothStatus", "蓝牙状态", "Bluetooth Status", "binary_sensor", "connectivity", "", 1},
	{114, "BluetoothSignalStrength", "蓝牙信号强度", "Bluetooth Signal Strength", "sensor", "signal_strength", "dBm", 1},
	{115, "LastSentryTriggerTime", "上次哨兵触发时间", "Last Sentry Trigger Time", "sensor", "timestamp", "", 1},
	{116, "LastSentryTriggerImage", "上次哨兵触发图像", "Last Sentry Trigger Image", "sensor", "", "", 1},
	{117, "LastVideoStartTime", "上次录像开始时间", "Last Video Start Time", "sensor", "timestamp", "", 1},
	{118, "LastVideoEndTime", "上次录像结束时间", "Last Video End Time", "sensor", "timestamp", "", 1},
	{119, "LastVideoPath", "上次录像路径", "Last Video Path", "sensor", "", "", 1},
	{120, "RadarRearCenter", "后中雷达", "Radar Rear Center", "sensor", "distance", "m", 1},
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
