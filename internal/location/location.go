package location

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// LocationData from the JSON file
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

type TermuxLocationProvider struct {
	logger          *logrus.Logger
	mu              sync.RWMutex
	cachedData      *LocationData
	lastFetch       time.Time
	lastFileModTime time.Time // Track file modification time to detect actual updates
	cacheTTL        time.Duration
	fetchTimeout    time.Duration
	ctx             context.Context
	cancel          context.CancelFunc
}

// Create provider with background goroutine
func NewTermuxLocationProvider(logger *logrus.Logger) *TermuxLocationProvider {
	ctx, cancel := context.WithCancel(context.Background())

	p := &TermuxLocationProvider{
		logger:       logger,
		cacheTTL:     2 * time.Minute,
		fetchTimeout: 15 * time.Second,
		ctx:          ctx,
		cancel:       cancel,
	}

	go p.backgroundLocationFetcher()
	return p
}

// Read from /storage/emulated/0/bydhass/gps
func (p *TermuxLocationProvider) fetchFromFile() (*LocationData, time.Time, error) {
	const filePath = "/storage/emulated/0/bydhass/gps"

	// Get file modification time first to detect if file actually changed
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, time.Time{}, fmt.Errorf("cannot stat gps file: %w", err)
	}
	fileModTime := fileInfo.ModTime()

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, time.Time{}, fmt.Errorf("cannot read gps file: %w", err)
	}

	var raw struct {
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
		Speed     float64 `json:"speed"`
		Accuracy  float64 `json:"accuracy"`
		Battery   float64 `json:"battery"`
		Timestamp *int64  `json:"timestamp,omitempty"` // Optional timestamp from GPS script
	}

	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, time.Time{}, fmt.Errorf("invalid gps json: %w", err)
	}

	// Use timestamp from JSON if available, otherwise use file modification time
	var timestamp time.Time
	if raw.Timestamp != nil {
		timestamp = time.Unix(*raw.Timestamp, 0)
	} else {
		// Fallback to file modification time (more accurate than time.Now())
		timestamp = fileModTime
	}

	return &LocationData{
		Latitude:  raw.Latitude,
		Longitude: raw.Longitude,
		Speed:     raw.Speed,
		Accuracy:  raw.Accuracy,
		Provider:  "termux-file",
		Timestamp: timestamp,
	}, fileModTime, nil
}

func (p *TermuxLocationProvider) GetLocation() (*LocationData, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.cachedData == nil {
		return nil, fmt.Errorf("no location data available yet")
	}

	// Enforce cache TTL - reject stale data
	if p.cacheTTL > 0 {
		age := time.Since(p.cachedData.Timestamp)
		if age > p.cacheTTL {
			return nil, fmt.Errorf("location data is stale (age: %v, TTL: %v)", age, p.cacheTTL)
		}
	}

	// Return a copy to prevent external modification
	result := *p.cachedData
	return &result, nil
}

func (p *TermuxLocationProvider) backgroundLocationFetcher() {
	p.fetchLocationData()

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-p.ctx.Done():
			return
		case <-ticker.C:
			p.fetchLocationData()
		}
	}
}

func (p *TermuxLocationProvider) fetchLocationData() {
	loc, fileModTime, err := p.fetchFromFile()
	if err != nil {
		p.logger.WithError(err).Warn("Failed reading GPS file; using default")
		p.setDefaultLocation()
		return
	}

	p.mu.Lock()
	// Only update cache if file modification time changed (file was actually updated)
	if fileModTime.After(p.lastFileModTime) || p.lastFileModTime.IsZero() {
		p.cachedData = loc
		p.lastFileModTime = fileModTime
		p.lastFetch = time.Now()
		p.mu.Unlock()

		p.logger.WithFields(logrus.Fields{
			"latitude":  loc.Latitude,
			"longitude": loc.Longitude,
			"speed":     loc.Speed,
			"accuracy":  loc.Accuracy,
			"provider":  loc.Provider,
			"timestamp": loc.Timestamp,
			"file_mod":  fileModTime,
		}).Debug("Loaded GPS location from file")
	} else {
		// File hasn't changed, no need to update cache
		p.mu.Unlock()
		p.logger.WithFields(logrus.Fields{
			"last_mod":    p.lastFileModTime,
			"current_mod": fileModTime,
		}).Debug("GPS file unchanged, skipping cache update")
	}
}

func (p *TermuxLocationProvider) setDefaultLocation() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.cachedData == nil {
		p.cachedData = &LocationData{
			Latitude:  0,
			Longitude: 0,
			Provider:  "default",
			Timestamp: time.Now(),
		}
		p.lastFetch = time.Now()
	}
}

func (p *TermuxLocationProvider) Stop() {
	p.cancel()
}

func (p *TermuxLocationProvider) IsLocationAvailable() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.cachedData != nil
}

func (p *TermuxLocationProvider) GetLastFetchTime() time.Time {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.lastFetch
}

func (p *TermuxLocationProvider) SetCacheTTL(ttl time.Duration) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.cacheTTL = ttl
}

func (p *TermuxLocationProvider) SetFetchTimeout(timeout time.Duration) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.fetchTimeout = timeout
}
