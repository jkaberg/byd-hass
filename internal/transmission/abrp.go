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

// ABRPTransmitter transmits telemetry data to A Better Route Planner
type ABRPTransmitter struct {
	apiKey     string
	vehicleKey string
	httpClient *http.Client
	logger     *logrus.Logger
}

// ABRPTelemetry represents the telemetry data format for ABRP
type ABRPTelemetry struct {
	// Core telemetry (required)
	Utc int64   `json:"utc"` // Unix timestamp in seconds
	SOC float64 `json:"soc"` // State of charge (0-100)

	// Enhanced telemetry (optional)
	Speed   *float64 `json:"speed,omitempty"`   // Speed in km/h
	Power   *float64 `json:"power,omitempty"`   // Power in kW (negative for charging)
	Current *float64 `json:"current,omitempty"` // Current in A
	Voltage *float64 `json:"voltage,omitempty"` // Voltage in V

	// Status flags
	IsCharging *bool `json:"is_charging,omitempty"` // Whether the car is charging
	IsParked   *bool `json:"is_parked,omitempty"`   // Whether the car is parked
	IsDriving  *bool `json:"is_driving,omitempty"`  // Whether the car is being driven

	// Battery information
	Capacity *float64 `json:"capacity,omitempty"`  // Battery capacity in kWh
	EstRange *float64 `json:"est_range,omitempty"` // Estimated range in km

	// Environmental data
	ExtTemp   *float64 `json:"ext_temp,omitempty"`  // External temperature in °C
	CabinTemp *float64 `json:"batt_temp,omitempty"` // Battery temperature in °C

	// Vehicle status
	Odometer *float64 `json:"odometer,omitempty"` // Total mileage in km

	// Charging specifics
	ChargePilotCurrent *float64 `json:"charge_pilot_current,omitempty"` // Charge pilot current in A
	ChargerACVoltage   *float64 `json:"charger_ac_voltage,omitempty"`   // AC charging voltage
	ChargerACAmps      *float64 `json:"charger_ac_amps,omitempty"`      // AC charging current
	ChargerDCVoltage   *float64 `json:"charger_dc_voltage,omitempty"`   // DC charging voltage
	ChargerDCAmps      *float64 `json:"charger_dc_amps,omitempty"`      // DC charging current

	// Advanced data for better route planning
	Elevation *float64  `json:"elevation,omitempty"` // Current elevation in meters
	Heading   *float64  `json:"heading,omitempty"`   // Heading in degrees (0-359)
	Location  *Location `json:"location,omitempty"`  // GPS coordinates

	// HVAC energy consumption
	HeaterOn  *bool `json:"heater_on,omitempty"`  // Whether heater is on
	ACOn      *bool `json:"ac_on,omitempty"`      // Whether AC is on
	DefrostOn *bool `json:"defrost_on,omitempty"` // Whether defrost is on
}

// Location represents GPS coordinates
type Location struct {
	Lat float64 `json:"lat"` // Latitude
	Lon float64 `json:"lon"` // Longitude
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
		"soc":         telemetry.SOC,
		"speed":       telemetry.Speed,
		"is_charging": telemetry.IsCharging,
		"status_code": resp.StatusCode,
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

	// Core required data - State of charge
	if data.BatteryPercentage != nil {
		telemetry.SOC = *data.BatteryPercentage
	}

	// Speed and motion
	if data.Speed != nil {
		telemetry.Speed = data.Speed
		isParked := *data.Speed == 0
		telemetry.IsParked = &isParked
		isDriving := *data.Speed > 0
		telemetry.IsDriving = &isDriving
	}

	// Charging status and power
	if data.ChargingStatus != nil && *data.ChargingStatus > 0 {
		isCharging := true
		telemetry.IsCharging = &isCharging
		// TODO: Add more charging details when available
	}

	// Calculate power consumption/generation
	if data.Speed != nil && *data.Speed > 0 {
		// Estimate driving power consumption based on speed
		// Simple model: base consumption increases with speed
		baseConsumption := 15.0                     // Base consumption in kW
		speedFactor := (*data.Speed / 100.0) * 10.0 // Additional consumption based on speed
		estimatedPower := baseConsumption + speedFactor
		telemetry.Power = &estimatedPower
	}

	// Battery information
	if data.BatteryCapacity != nil {
		telemetry.Capacity = data.BatteryCapacity
	}

	// Calculate estimated range based on battery percentage and capacity
	if data.BatteryPercentage != nil && data.BatteryCapacity != nil {
		// More sophisticated range calculation considering current conditions
		remainingCapacity := (*data.BatteryCapacity * *data.BatteryPercentage) / 100
		efficiency := 5.0 // km/kWh - could be adjusted based on speed, temperature, etc.

		// Adjust efficiency for temperature
		if data.OutsideTemperature != nil {
			tempC := *data.OutsideTemperature
			if tempC < 0 {
				efficiency *= 0.8 // Cold weather reduces efficiency
			} else if tempC > 30 {
				efficiency *= 0.9 // Hot weather reduces efficiency slightly
			}
		}

		estimatedRange := remainingCapacity * efficiency
		telemetry.EstRange = &estimatedRange
	}

	// Environmental data
	if data.OutsideTemperature != nil {
		telemetry.ExtTemp = data.OutsideTemperature
	}

	if data.AvgBatteryTemp != nil {
		telemetry.CabinTemp = data.AvgBatteryTemp // Using battery temp as cabin temp for now
	}

	// Vehicle odometer
	if data.Mileage != nil {
		telemetry.Odometer = data.Mileage
	}

	// Battery voltage and current (estimated)
	if data.MaxBatteryVoltage != nil {
		telemetry.Voltage = data.MaxBatteryVoltage
	}

	// HVAC status
	if data.ACStatus != nil {
		acOn := *data.ACStatus > 0
		telemetry.ACOn = &acOn

		// Determine if heating or cooling based on temperature difference
		if data.CabinTemperature != nil && data.DriverACTemperature != nil {
			if *data.DriverACTemperature > *data.CabinTemperature {
				heaterOn := true
				telemetry.HeaterOn = &heaterOn
			}
		}
	}

	// Location data
	if data.Location != nil {
		telemetry.Latitude = &data.Location.Latitude
		telemetry.Longitude = &data.Location.Longitude
		telemetry.Elevation = &data.Location.Altitude
		telemetry.Heading = &data.Location.Bearing
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
		"connected":       t.IsConnected(),
		"api_key_set":     t.apiKey != "",
		"vehicle_key_set": t.vehicleKey != "",
		"timeout":         t.httpClient.Timeout,
	}
}
