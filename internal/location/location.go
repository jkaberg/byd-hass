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
	logger       *logrus.Logger
	mu           sync.RWMutex
	cachedData   *LocationData
	lastFetch    time.Time
	cacheTTL     time.Duration
	fetchTimeout time.Duration
	ctx          context.Context
	cancel       context.CancelFunc
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
func (p *TermuxLocationProvider) fetchFromFile() (*LocationData, error) {
	const filePath = "/storage/emulated/0/bydhass/gps"

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("cannot read gps file: %w", err)
	}

	var raw struct {
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
		Speed     float64 `json:"speed"`
		Accuracy  float64 `json:"accuracy"`
		Battery   float64 `json:"battery"`
	}

	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("invalid gps json: %w", err)
	}

	return &LocationData{
		Latitude:  raw.Latitude,
		Longitude: raw.Longitude,
		Speed:     raw.Speed,
		Accuracy:  raw.Accuracy,
		Provider:  "termux-file",
		Timestamp: time.Now(),
	}, nil
}

func (p *TermuxLocationProvider) GetLocation() (*LocationData, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.cachedData == nil {
		return nil, fmt.Errorf("no location data available yet")
	}

	return &(*p.cachedData), nil
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
	loc, err := p.fetchFromFile()
	if err != nil {
		p.logger.WithError(err).Warn("Failed reading GPS file; using default")
		p.setDefaultLocation()
		return
	}

	p.mu.Lock()
	p.cachedData = loc
	p.lastFetch = time.Now()
	p.mu.Unlock()

	p.logger.WithFields(logrus.Fields{
		"latitude":  loc.Latitude,
		"longitude": loc.Longitude,
		"speed":     loc.Speed,
		"accuracy":  loc.Accuracy,
		"provider":  loc.Provider,
	}).Debug("Loaded GPS location from file")
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

