package transmission

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
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
	token      string
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
func NewABRPTransmitter(apiKey, token string, logger *logrus.Logger) *ABRPTransmitter {
	// Rely on the global custom DNS resolver installed in main.go.
	transport := &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
	}

	return &ABRPTransmitter{
		apiKey: apiKey,
		token:  token,
		httpClient: &http.Client{
			Timeout:   10 * time.Second,
			Transport: transport,
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

	// The API expects a urlencoded form post
	form := url.Values{}
	form.Add("tlm", string(payload))

	apiURL := fmt.Sprintf("https://api.iternio.com/1/tlm/send?api_key=%s&token=%s", t.apiKey, t.token)

	// Create HTTP request
	req, err := http.NewRequest("POST", apiURL, strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create ABRP request: %w", err)
	}

	req.Header.Set("User-Agent", "byd-hass/1.0.0")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

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

	// Log the entire telemetry object for full visibility
	logPayload, err := json.Marshal(telemetry)
	if err != nil {
		t.logger.WithError(err).Warn("Failed to marshal telemetry for logging")
	}

	t.logger.WithFields(logrus.Fields{
		"status_code": resp.StatusCode,
		"payload":     string(logPayload),
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
		if data.Location.Altitude > 0 {
			telemetry.Elevation = &data.Location.Altitude
		}
		if data.Location.Bearing > 0 {
			telemetry.Heading = &data.Location.Bearing
		}
	}

	// High priority - Power from engine
	if data.EnginePower != nil {
		telemetry.Power = data.EnginePower
	}

	// High priority - Charging status and power
	if data.ChargingStatus != nil {
		isCharging := *data.ChargingStatus > 0
		telemetry.IsCharging = &isCharging

		// Determine if DCFC based on power draw
		if isCharging && telemetry.Power != nil && *telemetry.Power <= -50.0 {
			isDCFC := true
			telemetry.IsDCFC = &isDCFC
		}
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
		if telemetry.Power != nil && telemetry.Voltage != nil && *telemetry.Voltage > 0 {
			// Power is in kW, convert to W for calculation
			powerWatts := *telemetry.Power * 1000
			current := powerWatts / *telemetry.Voltage
			telemetry.Current = &current
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

	// Lower priority - Odometer
	if data.Mileage != nil {
		telemetry.Odometer = data.Mileage
	}

	// Lower priority - HVAC data
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
		// BYD sensor data now in bar; convert to kPa
		pressureKPa := *data.LeftFrontTirePressure * 100
		telemetry.TirePressureFL = &pressureKPa
	}
	if data.RightFrontTirePressure != nil {
		pressureKPa := *data.RightFrontTirePressure * 100
		telemetry.TirePressureFR = &pressureKPa
	}
	if data.LeftRearTirePressure != nil {
		pressureKPa := *data.LeftRearTirePressure * 100
		telemetry.TirePressureRL = &pressureKPa
	}
	if data.RightRearTirePressure != nil {
		pressureKPa := *data.RightRearTirePressure * 100
		telemetry.TirePressureRR = &pressureKPa
	}

	return telemetry
}

// SetTimeout configures the HTTP client timeout
func (t *ABRPTransmitter) SetTimeout(timeout time.Duration) {
	t.httpClient.Timeout = timeout
}

// GetConnectionStatus returns detailed connection status for diagnostics
func (t *ABRPTransmitter) GetConnectionStatus() map[string]interface{} {
	return map[string]interface{}{
		"connected":   t.IsConnected(),
		"api_key_set": t.apiKey != "",
		"token_set":   t.token != "",
		"timeout":     t.httpClient.Timeout,
	}
}
