package sensors

import (
	"strings"
	"time"

	"github.com/jkaberg/byd-hass/internal/location"
)

// SensorData represents all available sensor data from the BYD car
type SensorData struct {
	// Metadata
	Timestamp time.Time `json:"timestamp"`

	// Location data from Termux
	Location *location.LocationData `json:"location,omitempty"`

	// Core Vehicle Data
	PowerStatus          *float64 `json:"power_status,omitempty"`           // 电源状态
	Speed                *float64 `json:"speed,omitempty"`                  // 车速
	Mileage              *float64 `json:"mileage,omitempty"`                // 里程
	GearPosition         *int `json:"gear_position,omitempty"`          // 档位
	EngineRPM            *float64 `json:"engine_rpm,omitempty"`             // 发动机转速
	BrakeDepth           *float64 `json:"brake_depth,omitempty"`            // 刹车深度
	AcceleratorDepth     *float64 `json:"accelerator_depth,omitempty"`      // 加速踏板深度
	FrontMotorRPM        *float64 `json:"front_motor_rpm,omitempty"`        // 前电机转速
	RearMotorRPM         *float64 `json:"rear_motor_rpm,omitempty"`         // 后电机转速
	EnginePower          *float64 `json:"engine_power,omitempty"`           // 发动机功率
	FrontMotorTorque     *float64 `json:"front_motor_torque,omitempty"`     // 前电机扭矩

	// Battery & Charging
	ChargeGunState       *float64 `json:"charge_gun_state,omitempty"`       // 充电枪插枪状态
	PowerConsumption100km *float64 `json:"power_consumption_100km,omitempty"` // 百公里电耗
	MaxBatteryTemp       *float64 `json:"max_battery_temp,omitempty"`       // 最高电池温度
	AvgBatteryTemp       *float64 `json:"avg_battery_temp,omitempty"`       // 平均电池温度
	MinBatteryTemp       *float64 `json:"min_battery_temp,omitempty"`       // 最低电池温度
	MaxBatteryVoltage    *float64 `json:"max_battery_voltage,omitempty"`    // 最高电池电压
	MinBatteryVoltage    *float64 `json:"min_battery_voltage,omitempty"`    // 最低电池电压
	BatteryCapacity      *float64 `json:"battery_capacity,omitempty"`       // 电池容量
	TotalPowerConsumption *float64 `json:"total_power_consumption,omitempty"` // 总电耗
	BatteryPercentage    *float64 `json:"battery_percentage,omitempty"`     // 电量百分比
	FuelPercentage       *float64 `json:"fuel_percentage,omitempty"`        // 油量百分比
	TotalFuelConsumption *float64 `json:"total_fuel_consumption,omitempty"` // 总燃油消耗
	ChargingStatus       *int `json:"charging_status,omitempty"`        // 充电状态
	BatteryVoltage12V    *float64 `json:"battery_voltage_12v,omitempty"`    // 蓄电池电压

	// Environment & Weather
	LastWiperTime        *float64 `json:"last_wiper_time,omitempty"`        // 上次雨刮时间
	Weather              *float64 `json:"weather,omitempty"`                // 天气
	CabinTemperature     *float64 `json:"cabin_temperature,omitempty"`      // 车内温度
	OutsideTemperature   *float64 `json:"outside_temperature,omitempty"`    // 车外温度
	DriverACTemperature  *float64 `json:"driver_ac_temperature,omitempty"`  // 主驾驶空调温度
	TemperatureUnit      *float64 `json:"temperature_unit,omitempty"`       // 温度单位
	EngineCoolantTemp    *float64 `json:"engine_coolant_temp,omitempty"`    // 发动机水温

	// Safety & Security
	DriverSeatbelt       *float64 `json:"driver_seatbelt,omitempty"`        // 主驾驶安全带状态
	RemoteLockStatus     *float64 `json:"remote_lock_status,omitempty"`     // 远程锁车状态
	PassengerSeatbeltWarn *float64 `json:"passenger_seatbelt_warn,omitempty"` // 副驾安全带警告
	Row2LeftSeatbelt     *float64 `json:"row2_left_seatbelt,omitempty"`     // 二排左安全带
	Row2RightSeatbelt    *float64 `json:"row2_right_seatbelt,omitempty"`    // 二排右安全带
	Row2CenterSeatbelt   *float64 `json:"row2_center_seatbelt,omitempty"`   // 二排中安全带

	// Steering & Control
	SteeringAngle        *float64 `json:"steering_angle,omitempty"`         // 方向盘转角
	SteeringRotationSpeed *float64 `json:"steering_rotation_speed,omitempty"` // 方向盘转速
	LaneCurvature        *float64 `json:"lane_curvature,omitempty"`         // 车道线曲率
	RightLineDistance    *float64 `json:"right_line_distance,omitempty"`    // 右侧线距离
	LeftLineDistance     *float64 `json:"left_line_distance,omitempty"`     // 左侧线距离
	CruiseSwitch         *float64 `json:"cruise_switch,omitempty"`          // 巡航开关
	DistanceToCarAhead   *float64 `json:"distance_to_car_ahead,omitempty"`  // 前车距离
	AutoParking          *float64 `json:"auto_parking,omitempty"`           // 自动驻车
	ACCCruiseStatus      *float64 `json:"acc_cruise_status,omitempty"`      // ACC巡航状态
	LaneKeepAssistStatus *float64 `json:"lane_keep_assist_status,omitempty"` // 车道保持状态

	// Radar Sensors
	RadarFrontLeft       *float64 `json:"radar_front_left,omitempty"`       // 雷达左前
	RadarFrontRight      *float64 `json:"radar_front_right,omitempty"`      // 雷达右前
	RadarRearLeft        *float64 `json:"radar_rear_left,omitempty"`        // 雷达左后
	RadarRearRight       *float64 `json:"radar_rear_right,omitempty"`       // 雷达右后
	RadarLeft            *float64 `json:"radar_left,omitempty"`             // 雷达左
	RadarFrontMidLeft    *float64 `json:"radar_front_mid_left,omitempty"`   // 雷达前左中
	RadarFrontMidRight   *float64 `json:"radar_front_mid_right,omitempty"`  // 雷达前右中
	RadarRearCenter      *float64 `json:"radar_rear_center,omitempty"`      // 雷达中后
	RearLeftProximityAlert *float64 `json:"rear_left_proximity_alert,omitempty"` // 左后接近告警
	RearRightProximityAlert *float64 `json:"rear_right_proximity_alert,omitempty"` // 右后接近告警

	// Wipers & Exterior
	FrontWiperSpeed      *float64 `json:"front_wiper_speed,omitempty"`      // 前雨刮速度
	WiperGear            *float64 `json:"wiper_gear,omitempty"`             // 雨刮档位

	// Tire Pressure
	LeftFrontTirePressure *float64 `json:"left_front_tire_pressure,omitempty"` // 左前轮气压
	RightFrontTirePressure *float64 `json:"right_front_tire_pressure,omitempty"` // 右前轮气压
	LeftRearTirePressure  *float64 `json:"left_rear_tire_pressure,omitempty"`  // 左后轮气压
	RightRearTirePressure *float64 `json:"right_rear_tire_pressure,omitempty"` // 右后轮气压

	// Turn Signals & Lights
	LeftTurnSignal       *float64 `json:"left_turn_signal,omitempty"`       // 左转向灯
	RightTurnSignal      *float64 `json:"right_turn_signal,omitempty"`      // 右转向灯
	ParkingLights        *float64 `json:"parking_lights,omitempty"`         // 小灯
	LowBeamLights        *float64 `json:"low_beam_lights,omitempty"`        // 近光灯
	HighBeamLights       *float64 `json:"high_beam_lights,omitempty"`       // 远光灯
	FrontFogLights       *float64 `json:"front_fog_lights,omitempty"`       // 前雾灯
	RearFogLights        *float64 `json:"rear_fog_lights,omitempty"`        // 后雾灯
	FootwellLights       *float64 `json:"footwell_lights,omitempty"`        // 脚照灯
	DaytimeRunningLights *float64 `json:"daytime_running_lights,omitempty"` // 日行灯
	HazardLights         *float64 `json:"hazard_lights,omitempty"`          // 双闪

	// Doors & Locks
	DriverDoorLock       *int `json:"driver_door_lock,omitempty"`       // 主驾车门锁
	DriverDoor           *int `json:"driver_door,omitempty"`            // 主驾车门
	PassengerDoor        *int `json:"passenger_door,omitempty"`         // 副驾车门
	LeftRearDoor         *int `json:"left_rear_door,omitempty"`         // 左后车门
	RightRearDoor        *int `json:"right_rear_door,omitempty"`        // 右后车门
	Hood                 *int `json:"hood,omitempty"`                   // 引擎盖
	TrunkDoor            *int `json:"trunk_door,omitempty"`             // 后备箱门
	FuelCap              *float64 `json:"fuel_cap,omitempty"`               // 油箱盖
	LeftRearDoorLock     *int `json:"left_rear_door_lock,omitempty"`    // 左后车门锁
	PassengerDoorLock    *int `json:"passenger_door_lock,omitempty"`    // 副驾车门锁
	RightRearDoorLock    *int `json:"right_rear_door_lock,omitempty"`   // 右后车门锁
	TrunkLock            *int `json:"trunk_lock,omitempty"`             // 后备箱门锁
	LeftRearChildLock    *int `json:"left_rear_child_lock,omitempty"`   // 左后儿童锁
	RightRearChildLock   *int `json:"right_rear_child_lock,omitempty"`  // 右后儿童锁

	// Windows
	DriverWindowOpenPercent    *float64 `json:"driver_window_open_percent,omitempty"`    // 主驾车窗打开百分比
	PassengerWindowOpenPercent *float64 `json:"passenger_window_open_percent,omitempty"` // 副驾车窗打开百分比
	LeftRearWindowOpenPercent  *float64 `json:"left_rear_window_open_percent,omitempty"` // 左后车窗打开百分比
	RightRearWindowOpenPercent *float64 `json:"right_rear_window_open_percent,omitempty"` // 右后车窗打开百分比
	SunroofOpenPercent         *float64 `json:"sunroof_open_percent,omitempty"`          // 天窗打开百分比
	SunshadeOpenPercent        *float64 `json:"sunshade_open_percent,omitempty"`         // 遮阳帘打开百分比

	// Vehicle Modes
	VehicleOperatingMode *float64 `json:"vehicle_operating_mode,omitempty"` // 整车工作模式
	VehicleRunningMode   *float64 `json:"vehicle_running_mode,omitempty"`   // 整车运行模式

	// Date/Time
	Month                *float64 `json:"month,omitempty"`                  // 月
	Day                  *float64 `json:"day,omitempty"`                    // 日
	Hour                 *float64 `json:"hour,omitempty"`                   // 时
	Minute               *float64 `json:"minute,omitempty"`                 // 分

	// HVAC/Climate
	ACStatus             *float64 `json:"ac_status,omitempty"`              // 空调状态
	FanSpeedLevel        *float64 `json:"fan_speed_level,omitempty"`        // 风量档位
	ACCirculationMode    *int `json:"ac_circulation_mode,omitempty"`    // 空调循环方式
	ACBlowingMode        *int `json:"ac_blowing_mode,omitempty"`        // 空调出风模式

	// Extended Features (1000+ IDs)
	SurroundViewStatus   *float64 `json:"surround_view_status,omitempty"`   // 全景状态
	UIConfigVersion      *float64 `json:"ui_config_version,omitempty"`      // 配置UI版本
	SentryModeStatus     *float64 `json:"sentry_mode_status,omitempty"`     // 哨兵状态
	PowerOffRecordingConfig *float64 `json:"power_off_recording_config,omitempty"` // 熄火录像配置开关
	PowerOffSentryAlarm  *float64 `json:"power_off_sentry_alarm,omitempty"` // 熄火哨兵报警
	WiFiStatus           *float64 `json:"wifi_status,omitempty"`            // WIFI状态
	BluetoothStatus      *float64 `json:"bluetooth_status,omitempty"`       // 蓝牙状态
	BluetoothSignalStrength *float64 `json:"bluetooth_signal_strength,omitempty"` // 蓝牙信号强度
	WirelessADBSwitch    *float64 `json:"wireless_adb_switch,omitempty"`    // 无线ADB开关

	// AI Recognition (2000+ IDs)
	AIPersonConfidence   *float64 `json:"ai_person_confidence,omitempty"`   // AI识别人可信度
	AIVehicleConfidence  *float64 `json:"ai_vehicle_confidence,omitempty"`  // AI识别车可信度
	LastSentryTriggerTime *float64 `json:"last_sentry_trigger_time,omitempty"` // 上次哨兵触发时间
	LastSentryTriggerImage *float64 `json:"last_sentry_trigger_image,omitempty"` // 上次哨兵触发画面
	LastVideoStartTime   *float64 `json:"last_video_start_time,omitempty"`  // 上次录像文件开始时间
	LastVideoEndTime     *float64 `json:"last_video_end_time,omitempty"`    // 上次录像文件结束时间
	LastVideoPath        *float64 `json:"last_video_path,omitempty"`        // 上次录像路径
}

// SensorDefinition maps sensor IDs to their Chinese names and Go struct field names
type SensorDefinition struct {
	ID          int
	ChineseName string
	FieldName   string
	Description string
}

// AllSensors contains the complete mapping of all available sensors
var AllSensors = []SensorDefinition{
	{1, "电源状态", "PowerStatus", "Power Status"},
	{2, "车速", "Speed", "Vehicle Speed"},
	{3, "里程", "Mileage", "Mileage"},
	{4, "档位", "GearPosition", "Gear Position"},
	{5, "发动机转速", "EngineRPM", "Engine RPM"},
	{6, "刹车深度", "BrakeDepth", "Brake Depth"},
	{7, "加速踏板深度", "AcceleratorDepth", "Accelerator Pedal Depth"},
	{8, "前电机转速", "FrontMotorRPM", "Front Motor RPM"},
	{9, "后电机转速", "RearMotorRPM", "Rear Motor RPM"},
	{10, "发动机功率", "EnginePower", "Engine Power"},
	{11, "前电机扭矩", "FrontMotorTorque", "Front Motor Torque"},
	{12, "充电枪插枪状态", "ChargeGunState", "Charging Plug Status"},
	{13, "百公里电耗", "PowerConsumption100km", "Power Consumption per 100km"},
	{14, "最高电池温度", "MaxBatteryTemp", "Max Battery Temperature"},
	{15, "平均电池温度", "AvgBatteryTemp", "Avg Battery Temperature"},
	{16, "最低电池温度", "MinBatteryTemp", "Min Battery Temperature"},
	{17, "最高电池电压", "MaxBatteryVoltage", "Max Battery Voltage"},
	{18, "最低电池电压", "MinBatteryVoltage", "Min Battery Voltage"},
	{19, "上次雨刮时间", "LastWiperTime", "Last Wiper Time"},
	{20, "天气", "Weather", "Weather"},
	{21, "主驾驶安全带状态", "DriverSeatbelt", "Driver Seatbelt Status"},
	{22, "远程锁车状态", "RemoteLockStatus", "Remote Lock Status"},
	{25, "车内温度", "CabinTemperature", "Cabin Temperature"},
	{26, "车外温度", "OutsideTemperature", "Outside Temperature"},
	{27, "主驾驶空调温度", "DriverACTemperature", "Driver AC Temperature"},
	{28, "温度单位", "TemperatureUnit", "Temperature Unit"},
	{29, "电池容量", "BatteryCapacity", "Battery Capacity"},
	{30, "方向盘转角", "SteeringAngle", "Steering Angle"},
	{31, "方向盘转速", "SteeringRotationSpeed", "Steering Rotation Speed"},
	{32, "总电耗", "TotalPowerConsumption", "Total Power Consumption"},
	{33, "电量百分比", "BatteryPercentage", "Battery Percentage"},
	{34, "油量百分比", "FuelPercentage", "Fuel Percentage"},
	{35, "总燃油消耗", "TotalFuelConsumption", "Total Fuel Consumption"},
	{36, "车道线曲率", "LaneCurvature", "Lane Line Curvature"},
	{37, "右侧线距离", "RightLineDistance", "Right Line Distance"},
	{38, "左侧线距离", "LeftLineDistance", "Left Line Distance"},
	{39, "蓄电池电压", "BatteryVoltage12V", "12V Battery Voltage"},
	{40, "雷达左前", "RadarFrontLeft", "Radar Front Left"},
	{41, "雷达右前", "RadarFrontRight", "Radar Front Right"},
	{42, "雷达左后", "RadarRearLeft", "Radar Rear Left"},
	{43, "雷达右后", "RadarRearRight", "Radar Rear Right"},
	{44, "雷达左", "RadarLeft", "Radar Left"},
	{45, "雷达前左中", "RadarFrontMidLeft", "Radar Front Mid-Left"},
	{46, "雷达前右中", "RadarFrontMidRight", "Radar Front Mid-Right"},
	{47, "雷达中后", "RadarRearCenter", "Radar Rear Center"},
	{48, "前雨刮速度", "FrontWiperSpeed", "Front Wiper Speed"},
	{49, "雨刮档位", "WiperGear", "Wiper Gear"},
	{50, "巡航开关", "CruiseSwitch", "Cruise Control Switch"},
	{51, "前车距离", "DistanceToCarAhead", "Distance to Car Ahead"},
	{52, "充电状态", "ChargingStatus", "Charging Status"},
	{53, "左前轮气压", "LeftFrontTirePressure", "Left Front Tire Pressure"},
	{54, "右前轮气压", "RightFrontTirePressure", "Right Front Tire Pressure"},
	{55, "左后轮气压", "LeftRearTirePressure", "Left Rear Tire Pressure"},
	{56, "右后轮气压", "RightRearTirePressure", "Right Rear Tire Pressure"},
	{57, "左转向灯", "LeftTurnSignal", "Left Turn Signal"},
	{58, "右转向灯", "RightTurnSignal", "Right Turn Signal"},
	{59, "主驾车门锁", "DriverDoorLock", "Driver Door Lock"},
	{61, "主驾车窗打开百分比", "DriverWindowOpenPercent", "Driver Window Open %"},
	{62, "副驾车窗打开百分比", "PassengerWindowOpenPercent", "Passenger Window Open %"},
	{63, "左后车窗打开百分比", "LeftRearWindowOpenPercent", "Left Rear Window Open %"},
	{64, "右后车窗打开百分比", "RightRearWindowOpenPercent", "Right Rear Window Open %"},
	{65, "天窗打开百分比", "SunroofOpenPercent", "Sunroof Open %"},
	{66, "遮阳帘打开百分比", "SunshadeOpenPercent", "Sunshade Open %"},
	{67, "整车工作模式", "VehicleOperatingMode", "Vehicle Operating Mode"},
	{68, "整车运行模式", "VehicleRunningMode", "Vehicle Running Mode"},
	{69, "月", "Month", "Month"},
	{70, "日", "Day", "Day"},
	{71, "时", "Hour", "Hour"},
	{72, "分", "Minute", "Minute"},
	{73, "副驾安全带警告", "PassengerSeatbeltWarn", "Passenger Seatbelt Warning"},
	{74, "二排左安全带", "Row2LeftSeatbelt", "2nd Row Left Seatbelt"},
	{75, "二排右安全带", "Row2RightSeatbelt", "2nd Row Right Seatbelt"},
	{76, "二排中安全带", "Row2CenterSeatbelt", "2nd Row Center Seatbelt"},
	{77, "空调状态", "ACStatus", "AC Status"},
	{78, "风量档位", "FanSpeedLevel", "Fan Speed Level"},
	{79, "空调循环方式", "ACCirculationMode", "AC Circulation Mode"},
	{80, "空调出风模式", "ACBlowingMode", "AC Blowing Mode"},
	{81, "主驾车门", "DriverDoor", "Driver Door"},
	{82, "副驾车门", "PassengerDoor", "Passenger Door"},
	{83, "左后车门", "LeftRearDoor", "Left Rear Door"},
	{84, "右后车门", "RightRearDoor", "Right Rear Door"},
	{85, "引擎盖", "Hood", "Hood"},
	{86, "后备箱门", "TrunkDoor", "Trunk Door"},
	{87, "油箱盖", "FuelCap", "Fuel Cap"},
	{88, "自动驻车", "AutoParking", "Auto Parking"},
	{89, "ACC巡航状态", "ACCCruiseStatus", "ACC Cruise Status"},
	{90, "左后接近告警", "RearLeftProximityAlert", "Rear Left Proximity Alert"},
	{91, "右后接近告警", "RearRightProximityAlert", "Rear Right Proximity Alert"},
	{92, "车道保持状态", "LaneKeepAssistStatus", "Lane Keep Assist Status"},
	{93, "左后车门锁", "LeftRearDoorLock", "Left Rear Door Lock"},
	{94, "副驾车门锁", "PassengerDoorLock", "Passenger Door Lock"},
	{95, "右后车门锁", "RightRearDoorLock", "Right Rear Door Lock"},
	{96, "后备箱门锁", "TrunkLock", "Trunk Lock"},
	{97, "左后儿童锁", "LeftRearChildLock", "Left Rear Child Lock"},
	{98, "右后儿童锁", "RightRearChildLock", "Right Rear Child Lock"},
	{99, "小灯", "ParkingLights", "Parking Lights"},
	{100, "近光灯", "LowBeamLights", "Low Beam Lights"},
	{101, "远光灯", "HighBeamLights", "High Beam Lights"},
	{104, "前雾灯", "FrontFogLights", "Front Fog Lights"},
	{105, "后雾灯", "RearFogLights", "Rear Fog Lights"},
	{106, "脚照灯", "FootwellLights", "Footwell Lights"},
	{107, "日行灯", "DaytimeRunningLights", "Daytime Running Lights"},
	{108, "发动机水温", "EngineCoolantTemp", "Engine Coolant Temperature"},
	{109, "双闪", "HazardLights", "Hazard Lights"},
	{1001, "全景状态", "SurroundViewStatus", "Surround View Status"},
	{1002, "配置UI版本", "UIConfigVersion", "UI Config Version"},
	{1003, "哨兵状态", "SentryModeStatus", "Sentry Mode Status"},
	{1004, "熄火录像配置开关", "PowerOffRecordingConfig", "Power-off Recording Config Switch"},
	{1006, "熄火哨兵报警", "PowerOffSentryAlarm", "Power-off Sentry Alarm"},
	{1007, "WIFI状态", "WiFiStatus", "Wi-Fi Status"},
	{1008, "蓝牙状态", "BluetoothStatus", "Bluetooth Status"},
	{1009, "蓝牙信号强度", "BluetoothSignalStrength", "Bluetooth Signal Strength"},
	{1101, "无线ADB开关", "WirelessADBSwitch", "Wireless ADB Switch"},
	{2001, "AI识别人可信度", "AIPersonConfidence", "AI Person Recognition Confidence"},
	{2002, "AI识别车可信度", "AIVehicleConfidence", "AI Vehicle Recognition Confidence"},
	{2003, "上次哨兵触发时间", "LastSentryTriggerTime", "Last Sentry Trigger Time"},
	{2004, "上次哨兵触发画面", "LastSentryTriggerImage", "Last Sentry Trigger Image"},
	{2005, "上次录像文件开始时间", "LastVideoStartTime", "Last Video Start Time"},
	{2006, "上次录像文件结束时间", "LastVideoEndTime", "Last Video End Time"},
	{2007, "上次录像路径", "LastVideoPath", "Last Video File Path"},
}

// GetSensorByID returns the sensor definition for a given ID
func GetSensorByID(id int) *SensorDefinition {
	for _, sensor := range AllSensors {
		if sensor.ID == id {
			return &sensor
		}
	}
	return nil
}

// GetSensorByChineseName returns the sensor definition for a given Chinese name
func GetSensorByChineseName(name string) *SensorDefinition {
	for _, sensor := range AllSensors {
		if sensor.ChineseName == name {
			return &sensor
		}
	}
	return nil
}

// BuildAPITemplate creates a template string for the API call with specified sensors
func BuildAPITemplate(sensorIDs []int) string {
	var parts []string
	
	for _, id := range sensorIDs {
		sensor := GetSensorByID(id)
		if sensor != nil {
			// Format: field_name:{chinese_name}
			fieldName := ToSnakeCase(sensor.FieldName)
			parts = append(parts, fieldName+":{"+sensor.ChineseName+"}")
		}
	}
	
	return strings.Join(parts, "|")
}

// ToSnakeCase converts CamelCase to snake_case
func ToSnakeCase(s string) string {
	var result []rune
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result = append(result, '_')
		}
		result = append(result, r)
	}
	return strings.ToLower(string(result))
} 