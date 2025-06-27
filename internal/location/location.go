package location

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"time"

	"github.com/sirupsen/logrus"
)

// LocationData represents the data from termux-location
type LocationData struct {
	Latitude         float64   `json:"latitude"`
	Longitude        float64   `json:"longitude"`
	Altitude         float64   `json:"altitude"`
	Accuracy         float64   `json:"accuracy"`
	VerticalAccuracy float64   `json:"vertical_accuracy"`
	Bearing          float64   `json:"bearing"`
	Speed            float64   `json:"speed"`
	ElapsedMs        int64     `json:"elapsed_ms"`
	Provider         string    `json:"provider"`
	Timestamp        time.Time `json:"-"`
}

// TermuxLocationProvider fetches GPS data from Termux API
type TermuxLocationProvider struct {
	logger *logrus.Logger
}

// NewTermuxLocationProvider creates a new location provider
func NewTermuxLocationProvider(logger *logrus.Logger) *TermuxLocationProvider {
	return &TermuxLocationProvider{logger: logger}
}

// GetLocation fetches the current location from termux-location
func (p *TermuxLocationProvider) GetLocation() (*LocationData, error) {
	p.logger.Debug("Fetching location from Termux API")

	// Check if termux-location is available
	_, err := exec.LookPath("termux-location")
	if err != nil {
		return nil, fmt.Errorf("termux-location command not found in PATH: %w. Is Termux:API installed?", err)
	}

	// Execute termux-location command
	cmd := exec.Command("termux-location", "-p", "gps", "-r", "once")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to execute termux-location: %w", err)
	}

	// Parse JSON output
	var loc LocationData
	if err := json.Unmarshal(output, &loc); err != nil {
		return nil, fmt.Errorf("failed to parse termux-location output: %w", err)
	}

	// Add timestamp
	loc.Timestamp = time.Now()

	p.logger.WithFields(logrus.Fields{
		"latitude":  loc.Latitude,
		"longitude": loc.Longitude,
		"speed":     loc.Speed,
		"provider":  loc.Provider,
	}).Debug("Successfully fetched location")

	return &loc, nil
} 