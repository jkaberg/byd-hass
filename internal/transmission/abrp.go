package transmission

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/jkaberg/byd-hass/internal/sensors"
	"github.com/sirupsen/logrus"
)

// ABRP (A Better Route Planner) telemetry integration
//
// This module transmits comprehensive vehicle telemetry data to ABRP for improved
// route planning and energy consumption estimation. The telemetry includes:
//
// High Priority Parameters (most important for route planning):
//   - utc: UTC timestamp (required)
//   - soc: State of Charge percentage (required)
//   - power: Instantaneous power consumption/generation in kW
//   - speed: Vehicle speed in km/h
//   - lat/lon: GPS coordinates for location-based planning
//   - is_charging: Charging status indicator
//   - is_dcfc: DC fast charging indicator
//   - is_parked: Parking status
//
// Lower Priority Parameters (enhance accuracy):
//   - capacity: Battery capacity in kWh
//   - soe: State of Energy (absolute energy content)
//   - voltage/current: Battery electrical parameters
//   - ext_temp/batt_temp/cabin_temp: Temperature data
//   - odometer: Total mileage
//   - est_battery_range: Estimated remaining range
//   - hvac_power/hvac_setpoint: Climate control data
//   - tire_pressure_*: Tire pressure monitoring
//   - heading/elevation: Navigation enhancement data

// ABRPTransmitter transmits telemetry data to A Better Route Planner
type ABRPTransmitter struct {
	apiKey     string
	vehicleKey string
	httpClient *http.Client
	logger     *logrus.Logger
}

// ABRPTelemetry represents the telemetry data format for ABRP
type ABRPTelemetry struct {
	// High priority parameters (required)
	Utc int64   `json:"utc"` // UTC timestamp in seconds
	SOC float64 `json:"soc"` // State of charge (0-100)

	// High priority parameters (optional but important)
	Power      *float64 `json:"power,omitempty"`       // Instantaneous power in kW (positive=output, negative=charging)
	Speed      *float64 `json:"speed,omitempty"`       // Vehicle speed in km/h
	Lat        *float64 `json:"lat,omitempty"`         // Current latitude
	Lon        *float64 `json:"lon,omitempty"`         // Current longitude
	IsCharging *bool    `json:"is_charging,omitempty"` // 0=not charging, 1=charging
	IsDCFC     *bool    `json:"is_dcfc,omitempty"`     // DC fast charging indicator
	IsParked   *bool    `json:"is_parked,omitempty"`   // Vehicle gear in P or driver left car

	// Lower priority parameters
	Capacity        *float64 `json:"capacity,omitempty"`          // Estimated usable battery capacity in kWh
	SOE             *float64 `json:"soe,omitempty"`               // Present energy capacity (SoC * capacity)
	SOH             *float64 `json:"soh,omitempty"`               // State of Health (100 = no degradation)
	Heading         *float64 `json:"heading,omitempty"`           // Current heading in degrees
	Elevation       *float64 `json:"elevation,omitempty"`         // Current elevation in meters
	ExtTemp         *float64 `json:"ext_temp,omitempty"`          // Outside temperature in °C
	BattTemp        *float64 `json:"batt_temp,omitempty"`         // Battery temperature in °C
	Voltage         *float64 `json:"voltage,omitempty"`           // Battery pack voltage in V
	Current         *float64 `json:"current,omitempty"`           // Battery pack current in A
	Odometer        *float64 `json:"odometer,omitempty"`          // Current odometer reading in km
	EstBatteryRange *float64 `json:"est_battery_range,omitempty"` // Estimated remaining range in km
	HVACPower       *float64 `json:"hvac_power,omitempty"`        // HVAC power usage in kW
	HVACSetpoint    *float64 `json:"hvac_setpoint,omitempty"`     // HVAC setpoint temperature in °C
	CabinTemp       *float64 `json:"cabin_temp,omitempty"`        // Current cabin temperature in °C
	TirePressureFL  *float64 `json:"tire_pressure_fl,omitempty"`  // Front left tire pressure in kPa
	TirePressureFR  *float64 `json:"tire_pressure_fr,omitempty"`  // Front right tire pressure in kPa
	TirePressureRL  *float64 `json:"tire_pressure_rl,omitempty"`  // Rear left tire pressure in kPa
	TirePressureRR  *float64 `json:"tire_pressure_rr,omitempty"`  // Rear right tire pressure in kPa
}

// NewABRPTransmitter creates a new ABRP transmitter
func NewABRPTransmitter(apiKey, vehicleKey string, logger *logrus.Logger) *ABRPTransmitter {
	return &ABRPTransmitter{
		apiKey:     apiKey,
		vehicleKey: vehicleKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		logger: logger,
	}
}

// Transmit sends sensor data to ABRP
func (t *ABRPTransmitter) Transmit(data *sensors.SensorData) error {
	// Convert sensor data to ABRP telemetry format
	telemetry := t.buildTelemetryData(data)

	// Marshal to JSON
	payload, err := json.Marshal(telemetry)
	if err != nil {
		return fmt.Errorf("failed to marshal ABRP telemetry: %w", err)
	}

	// Build API URL
	url := fmt.Sprintf("https://api.iternio.com/1/tlm/send?token=%s&api_key=%s",
		t.vehicleKey, t.apiKey)

	// Create HTTP request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to create ABRP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "byd-hass/1.0.0")

	// Send request
	resp, err := t.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send ABRP request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ABRP API returned status %d: %s", resp.StatusCode, resp.Status)
	}

	t.logger.WithFields(logrus.Fields{
		"soc":              telemetry.SOC,
		"speed":            telemetry.Speed,
		"power":            telemetry.Power,
		"is_charging":      telemetry.IsCharging,
		"is_parked":        telemetry.IsParked,
		"lat":              telemetry.Lat,
		"lon":              telemetry.Lon,
		"ext_temp":         telemetry.ExtTemp,
		"batt_temp":        telemetry.BattTemp,
		"est_range":        telemetry.EstBatteryRange,
		"tire_pressure_fl": telemetry.TirePressureFL,
		"status_code":      resp.StatusCode,
	}).Debug("Successfully transmitted to ABRP")

	return nil
}

// IsConnected always returns true for HTTP-based transmitter
func (t *ABRPTransmitter) IsConnected() bool {
	return true
}

// buildTelemetryData converts sensor data to ABRP telemetry format
func (t *ABRPTransmitter) buildTelemetryData(data *sensors.SensorData) ABRPTelemetry {
	telemetry := ABRPTelemetry{
		Utc: data.Timestamp.Unix(),
	}

	// High priority parameters - State of charge (required)
	if data.BatteryPercentage != nil {
		telemetry.SOC = *data.BatteryPercentage
	}

	// High priority - Speed
	if data.Speed != nil {
		telemetry.Speed = data.Speed

		// Determine parking status based on speed
		isParked := *data.Speed == 0
		telemetry.IsParked = &isParked
	}

	// High priority - Location coordinates
	if data.Location != nil {
		telemetry.Lat = &data.Location.Latitude
		telemetry.Lon = &data.Location.Longitude
		telemetry.Elevation = &data.Location.Altitude
		telemetry.Heading = &data.Location.Bearing
	}

	// High priority - Charging status and power
	if data.ChargingStatus != nil {
		isCharging := *data.ChargingStatus > 0
		telemetry.IsCharging = &isCharging

		// Estimate charging power when charging (negative values for ABRP)
		if isCharging {
			// Estimate charging power based on typical BYD charging rates
			// This is a rough estimate - actual power data would be better
			chargingPower := -22.0 // Assume 22kW AC charging as default
			if data.ChargeGunState != nil && *data.ChargeGunState == 2 {
				// Connected to charging port - could be DC fast charging
				chargingPower = -50.0 // Assume DC fast charging
				isDCFC := true
				telemetry.IsDCFC = &isDCFC
			}
			telemetry.Power = &chargingPower
		}
	}

	// High priority - Driving power consumption (positive values)
	if data.Speed != nil && *data.Speed > 0 && data.EnginePower != nil {
		// Use actual engine power if available
		telemetry.Power = data.EnginePower
	} else if data.Speed != nil && *data.Speed > 0 {
		// Estimate driving power consumption based on speed and conditions
		baseConsumption := 15.0                     // Base consumption in kW
		speedFactor := (*data.Speed / 100.0) * 10.0 // Additional consumption based on speed

		// Adjust for temperature (HVAC usage)
		tempAdjustment := 0.0
		if data.OutsideTemperature != nil {
			tempC := *data.OutsideTemperature
			if tempC < 0 || tempC > 30 {
				tempAdjustment = 5.0 // Additional power for heating/cooling
			}
		}

		estimatedPower := baseConsumption + speedFactor + tempAdjustment
		telemetry.Power = &estimatedPower
	}

	// Lower priority - Battery information
	if data.BatteryCapacity != nil {
		telemetry.Capacity = data.BatteryCapacity

		// Calculate SOE (State of Energy) = SoC * capacity
		if data.BatteryPercentage != nil {
			soe := (*data.BatteryCapacity * *data.BatteryPercentage) / 100
			telemetry.SOE = &soe
		}
	}

	// Lower priority - Battery voltage and estimated current
	if data.MaxBatteryVoltage != nil {
		telemetry.Voltage = data.MaxBatteryVoltage

		// Estimate current from power and voltage (I = P / V)
		if telemetry.Power != nil && *data.MaxBatteryVoltage > 0 {
			estimatedCurrent := (*telemetry.Power * 1000) / *data.MaxBatteryVoltage // Convert kW to W, then to A
			telemetry.Current = &estimatedCurrent
		}
	}

	// Lower priority - Temperature data
	if data.OutsideTemperature != nil {
		telemetry.ExtTemp = data.OutsideTemperature
	}

	if data.AvgBatteryTemp != nil {
		telemetry.BattTemp = data.AvgBatteryTemp
	}

	if data.CabinTemperature != nil {
		telemetry.CabinTemp = data.CabinTemperature
	}

	// Lower priority - Vehicle odometer
	if data.Mileage != nil {
		telemetry.Odometer = data.Mileage
	}

	// Lower priority - Estimated battery range
	if data.BatteryPercentage != nil && data.BatteryCapacity != nil {
		// Calculate estimated range considering current conditions
		remainingCapacity := (*data.BatteryCapacity * *data.BatteryPercentage) / 100
		efficiency := 5.0 // km/kWh baseline efficiency

		// Adjust efficiency for temperature
		if data.OutsideTemperature != nil {
			tempC := *data.OutsideTemperature
			if tempC < 0 {
				efficiency *= 0.75 // Significant reduction in cold weather
			} else if tempC < 10 {
				efficiency *= 0.85 // Moderate reduction in cool weather
			} else if tempC > 35 {
				efficiency *= 0.90 // Slight reduction in very hot weather
			}
		}

		// Adjust efficiency for speed (if driving)
		if data.Speed != nil && *data.Speed > 0 {
			if *data.Speed > 100 {
				efficiency *= 0.8 // High speed reduces efficiency
			} else if *data.Speed > 80 {
				efficiency *= 0.9 // Moderate speed impact
			}
		}

		estimatedRange := remainingCapacity * efficiency
		telemetry.EstBatteryRange = &estimatedRange
	}

	// Lower priority - HVAC information
	if data.DriverACTemperature != nil {
		telemetry.HVACSetpoint = data.DriverACTemperature
	}

	// Estimate HVAC power consumption
	if data.ACStatus != nil && *data.ACStatus > 0 {
		// Estimate HVAC power based on temperature difference and fan speed
		hvacPower := 2.0 // Base HVAC power consumption in kW

		if data.CabinTemperature != nil && data.OutsideTemperature != nil {
			tempDiff := *data.CabinTemperature - *data.OutsideTemperature
			if tempDiff < 0 {
				tempDiff = -tempDiff // Absolute difference
			}

			// More temperature difference = more power needed
			hvacPower += (tempDiff / 10.0) * 1.0 // 1kW per 10°C difference
		}

		// Adjust based on fan speed level
		if data.FanSpeedLevel != nil {
			fanMultiplier := *data.FanSpeedLevel / 3.0 // Assume max fan level is 3
			hvacPower *= fanMultiplier
		}

		telemetry.HVACPower = &hvacPower
	}

	// Lower priority - Tire pressure (convert from bar to kPa)
	if data.LeftFrontTirePressure != nil {
		// BYD sensor data is in bar (scaled by 0.01), convert to kPa
		pressureKPa := (*data.LeftFrontTirePressure * 0.01) * 100 // bar to kPa
		telemetry.TirePressureFL = &pressureKPa
	}

	if data.RightFrontTirePressure != nil {
		pressureKPa := (*data.RightFrontTirePressure * 0.01) * 100
		telemetry.TirePressureFR = &pressureKPa
	}

	if data.LeftRearTirePressure != nil {
		pressureKPa := (*data.LeftRearTirePressure * 0.01) * 100
		telemetry.TirePressureRL = &pressureKPa
	}

	if data.RightRearTirePressure != nil {
		pressureKPa := (*data.RightRearTirePressure * 0.01) * 100
		telemetry.TirePressureRR = &pressureKPa
	}

	// Lower priority - State of Health (estimate based on capacity vs new capacity)
	// This would require knowing the original battery capacity when new
	// For now, we'll skip this as we don't have historical data

	return telemetry
}

// SetTimeout configures the HTTP client timeout
func (t *ABRPTransmitter) SetTimeout(timeout time.Duration) {
	t.httpClient.Timeout = timeout
}

// GetConnectionStatus returns detailed connection status for diagnostics
func (t *ABRPTransmitter) GetConnectionStatus() map[string]interface{} {
	return map[string]interface{}{
		"connected":       t.IsConnected(),
		"api_key_set":     t.apiKey != "",
		"vehicle_key_set": t.vehicleKey != "",
		"timeout":         t.httpClient.Timeout,
	}
}
