package diplus

// This file is functionally identical to the former internal/api/diplus.go but
// has been relocated to a more appropriately named package (diplus) to avoid
// the overly generic "api" label and better follow Go package naming
// conventions.

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/jkaberg/byd-hass/internal/sensors"
	"github.com/sirupsen/logrus"
)

// Client handles communication with the local Diplus API.
// The struct and method set remains unchanged so external callers are not
// impacted by the move.

type DiplusClient struct {
	baseURL    string
	httpClient *http.Client
	logger     *logrus.Logger
}

// NewDiplusClient creates a new Diplus API client.
func NewDiplusClient(baseURL string, logger *logrus.Logger) *DiplusClient {
	return &DiplusClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		logger: logger,
	}
}

// GetSensorData fetches sensor data for the specified sensor IDs.
func (c *DiplusClient) GetSensorData(sensorIDs []int) (*sensors.SensorData, error) {
	template := c.buildAPITemplate(sensorIDs)
	if template == "" {
		return nil, fmt.Errorf("no valid sensors found for IDs: %v", sensorIDs)
	}
	c.logger.WithField("template", template).Debug("Built API template")

	responseBody, err := c.makeRequest(template)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}

	sensorData, err := sensors.ParseAPIResponse(responseBody)
	if err != nil {
		return nil, fmt.Errorf("failed to parse API response: %w", err)
	}

	if warnings := sensors.ValidateSensorData(sensorData); len(warnings) > 0 {
		for _, warning := range warnings {
			c.logger.Warn(warning)
		}
	}

	c.logger.WithField("active_sensors", len(sensors.GetNonNilFields(sensorData))).Debug("Successfully parsed sensor data")
	return sensorData, nil
}

// buildAPITemplate creates the API template string using Chinese sensor names.
func (c *DiplusClient) buildAPITemplate(sensorIDs []int) string {
	var parts []string
	for _, id := range sensorIDs {
		sensor := sensors.GetSensorByID(id)
		if sensor == nil {
			c.logger.WithField("sensor_id", id).Warn("Unknown sensor ID, skipping")
			continue
		}
		key := sensor.FieldName
		part := fmt.Sprintf("%s:{%s}", key, sensor.ChineseName)
		parts = append(parts, part)
		c.logger.WithFields(logrus.Fields{
			"sensor_id":    id,
			"chinese_name": sensor.ChineseName,
			"field_name":   sensor.FieldName,
			"key":          key,
		}).Debug("Added sensor to template")
	}

	switch len(parts) {
	case 0:
		return ""
	case 1:
		return parts[0]
	default:
		return fmt.Sprintf("%s", parts[0]) + "|" + fmt.Sprint(parts[1:])
	}
}

// makeRequest performs the HTTP request to the Diplus API.
func (c *DiplusClient) makeRequest(template string) ([]byte, error) {
	encodedTemplate := url.QueryEscape(template)
	fullURL := fmt.Sprintf("%s?text=%s", c.baseURL, encodedTemplate)
	c.logger.WithField("url", fullURL).Debug("Making API request")

	resp, err := c.httpClient.Get(fullURL)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, resp.Status)
	}

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

// GetAllSensorData fetches data for all available sensors.
func (c *DiplusClient) GetAllSensorData() (*sensors.SensorData, error) {
	return c.GetSensorData(sensors.GetAllSensorIDs())
}

// IsHealthy checks if the Diplus API is responding.
func (c *DiplusClient) IsHealthy() bool {
	testSensorIDs := []int{33} // battery percentage
	_, err := c.GetSensorData(testSensorIDs)
	return err == nil
}

// GetSensorInfo returns information about a specific sensor.
func (c *DiplusClient) GetSensorInfo(sensorID int) *sensors.SensorDefinition {
	return sensors.GetSensorByID(sensorID)
}

// GetAllSensorInfo returns all sensor definitions.
func (c *DiplusClient) GetAllSensorInfo() []sensors.SensorDefinition {
	return sensors.AllSensors
}

// SetTimeout adjusts the HTTP client timeout.
func (c *DiplusClient) SetTimeout(timeout time.Duration) { c.httpClient.Timeout = timeout }

// SetLogger swaps the logger instance.
func (c *DiplusClient) SetLogger(logger *logrus.Logger) { c.logger = logger }

// CompareAllSensors queries all sensors and compares raw vs parsed values.
func (c *DiplusClient) CompareAllSensors() error {
	c.logger.Info("Querying Diplus API for all sensors to compare raw vs parsed values…")

	sensorData, err := c.GetAllSensorData()
	if err != nil {
		return fmt.Errorf("failed to get sensor data: %w", err)
	}

	allSensorIDs := sensors.GetAllSensorIDs()
	template := c.buildAPITemplate(allSensorIDs)
	responseBody, err := c.makeRequest(template)
	if err != nil {
		return fmt.Errorf("failed to get raw API response: %w", err)
	}

	c.logger.Info("Comparing raw API values vs parsed values…")
	sensors.CompareRawVsParsed(responseBody, sensorData)
	return nil
}

// Poll polls the Diplus API for sensor data.
func (c *DiplusClient) Poll() (*sensors.SensorData, error) {
	return c.GetAllSensorData()
}
