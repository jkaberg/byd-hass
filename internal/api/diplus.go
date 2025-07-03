package api

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/jkaberg/byd-hass/internal/sensors"
	"github.com/sirupsen/logrus"
)

// DiplusClient handles communication with the local Diplus API
type DiplusClient struct {
	baseURL    string
	httpClient *http.Client
	logger     *logrus.Logger
}

// NewDiplusClient creates a new Diplus API client
func NewDiplusClient(baseURL string, logger *logrus.Logger) *DiplusClient {
	return &DiplusClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		logger: logger,
	}
}

// GetSensorData fetches sensor data for the specified sensor IDs
func (c *DiplusClient) GetSensorData(sensorIDs []int) (*sensors.SensorData, error) {
	// Build the template string with Chinese sensor names
	template := c.buildAPITemplate(sensorIDs)
	if template == "" {
		return nil, fmt.Errorf("no valid sensors found for IDs: %v", sensorIDs)
	}

	//c.logger.WithField("template", template).Debug("Built API template")

	// Make the HTTP request
	responseBody, err := c.makeRequest(template)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}

	// Parse the response
	sensorData, err := sensors.ParseAPIResponse(responseBody)
	if err != nil {
		return nil, fmt.Errorf("failed to parse API response: %w", err)
	}

	// Validate the data
	if warnings := sensors.ValidateSensorData(sensorData); len(warnings) > 0 {
		for _, warning := range warnings {
			c.logger.Warn(warning)
		}
	}

	c.logger.WithField("active_sensors", len(sensors.GetNonNilFields(sensorData))).Debug("Successfully parsed sensor data")

	return sensorData, nil
}

// buildAPITemplate creates the API template string using Chinese sensor names
func (c *DiplusClient) buildAPITemplate(sensorIDs []int) string {
	var parts []string

	for _, id := range sensorIDs {
		sensor := sensors.GetSensorByID(id)
		if sensor == nil {
			c.logger.WithField("sensor_id", id).Warn("Unknown sensor ID, skipping")
			continue
		}

		// Use the struct FieldName directly as the key (e.g. BatteryPercentage)
		// This ensures the same identifier is echoed back by Diplus, eliminating
		// any key-translation logic in the parser.
		key := sensor.FieldName

		// Create template part: key:{Chinese_name}
		part := fmt.Sprintf("%s:{%s}", key, sensor.ChineseName)
		parts = append(parts, part)

		//c.logger.WithFields(logrus.Fields{
		//	"sensor_id":    id,
		//	"chinese_name": sensor.ChineseName,
		//	"field_name":   sensor.FieldName,
		//	"key":          key,
		//}).Debug("Added sensor to template")
	}

	if len(parts) == 0 {
		return ""
	}

	template := fmt.Sprintf("%s", parts[0])
	for i := 1; i < len(parts); i++ {
		template += "|" + parts[i]
	}

	return template
}

// makeRequest makes the HTTP request to the Diplus API
func (c *DiplusClient) makeRequest(template string) ([]byte, error) {
	// URL encode the template
	encodedTemplate := url.QueryEscape(template)

	// Build the full URL
	fullURL := fmt.Sprintf("%s?text=%s", c.baseURL, encodedTemplate)

	//c.logger.WithField("url", fullURL).Debug("Making API request")

	// Make the request
	resp, err := c.httpClient.Get(fullURL)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, resp.Status)
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	c.logger.WithFields(logrus.Fields{
		"status_code":   resp.StatusCode,
		"response_size": len(body),
	}).Debug("Received API response")

	return body, nil
}

// GetAllSensorData fetches data for all available sensors
func (c *DiplusClient) GetAllSensorData() (*sensors.SensorData, error) {
	return c.GetSensorData(sensors.GetAllSensorIDs())
}

// IsHealthy checks if the Diplus API is responding
func (c *DiplusClient) IsHealthy() bool {
	// Try to fetch a minimal sensor set to test connectivity
	testSensorIDs := []int{33} // Just battery percentage
	_, err := c.GetSensorData(testSensorIDs)
	return err == nil
}

// GetSensorInfo returns information about a specific sensor
func (c *DiplusClient) GetSensorInfo(sensorID int) *sensors.SensorDefinition {
	return sensors.GetSensorByID(sensorID)
}

// GetAllSensorInfo returns information about all available sensors
func (c *DiplusClient) GetAllSensorInfo() []sensors.SensorDefinition {
	return sensors.AllSensors
}

// SetTimeout configures the HTTP client timeout
func (c *DiplusClient) SetTimeout(timeout time.Duration) {
	c.httpClient.Timeout = timeout
}

// SetLogger updates the logger instance
func (c *DiplusClient) SetLogger(logger *logrus.Logger) {
	c.logger = logger
}

// CompareAllSensors queries all sensors and compares raw vs parsed values
func (c *DiplusClient) CompareAllSensors() error {
	c.logger.Info("Querying Diplus API for all sensors to compare raw vs parsed values...")

	// Get all sensor data
	sensorData, err := c.GetAllSensorData()
	if err != nil {
		return fmt.Errorf("failed to get sensor data: %w", err)
	}

	// Also get the raw response for comparison
	allSensorIDs := sensors.GetAllSensorIDs()
	template := c.buildAPITemplate(allSensorIDs)
	responseBody, err := c.makeRequest(template)
	if err != nil {
		return fmt.Errorf("failed to get raw API response: %w", err)
	}

	c.logger.Info("Comparing raw API values vs parsed values...")
	sensors.CompareRawVsParsed(responseBody, sensorData)

	return nil
}

// Poll polls the Diplus API for sensor data
func (c *DiplusClient) Poll() (*sensors.SensorData, error) {
	c.logger.Debug("Polling Diplus API for sensor data...")
	// For now, we use a minimal set of essential sensors.
	return c.GetSensorData(sensors.PollSensorIDs())
}
