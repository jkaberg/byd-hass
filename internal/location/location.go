package location

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"sync"
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
	logger       *logrus.Logger
	mu           sync.RWMutex
	cachedData   *LocationData
	lastFetch    time.Time
	cacheTTL     time.Duration
	fetchTimeout time.Duration
	isRunning    bool
	ctx          context.Context
	cancel       context.CancelFunc
}

// NewTermuxLocationProvider creates a new location provider with background fetching
func NewTermuxLocationProvider(logger *logrus.Logger) *TermuxLocationProvider {
	ctx, cancel := context.WithCancel(context.Background())

	p := &TermuxLocationProvider{
		logger:       logger,
		cacheTTL:     2 * time.Minute,  // Cache location for 2 minutes
		fetchTimeout: 15 * time.Second, // 15 second timeout for location requests
		ctx:          ctx,
		cancel:       cancel,
	}

	// Start background location fetching
	go p.backgroundLocationFetcher()

	return p
}

// GetLocation returns the cached location data (non-blocking)
func (p *TermuxLocationProvider) GetLocation() (*LocationData, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.cachedData == nil {
		return nil, fmt.Errorf("no location data available yet")
	}

	// Check if cached data is still valid
	if time.Since(p.lastFetch) > p.cacheTTL {
		p.logger.Debug("Cached location data is stale but returning it anyway")
	}

	// Return a copy to avoid race conditions
	locationCopy := *p.cachedData
	return &locationCopy, nil
}

// backgroundLocationFetcher runs in a separate goroutine to fetch location data
func (p *TermuxLocationProvider) backgroundLocationFetcher() {
	p.logger.Info("Started background location fetcher")

	// Initial fetch
	p.fetchLocationData()

	// Set up periodic fetching
	ticker := time.NewTicker(90 * time.Second) // Fetch every 90 seconds
	defer ticker.Stop()

	for {
		select {
		case <-p.ctx.Done():
			p.logger.Info("Background location fetcher stopped")
			return
		case <-ticker.C:
			p.fetchLocationData()
		}
	}
}

// fetchLocationData performs the actual location fetch with timeout
func (p *TermuxLocationProvider) fetchLocationData() {
	p.logger.Debug("Fetching location from Termux API (background)")

	// Create context with timeout
	ctx, cancel := context.WithTimeout(p.ctx, p.fetchTimeout)
	defer cancel()

	// Create command with context for timeout
	cmd := exec.CommandContext(ctx, "/data/data/com.termux/files/usr/bin/termux-location", "-p", "gps", "-r", "once")

	// Run command with timeout
	output, err := cmd.Output()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			p.logger.Warn("Location fetch timed out after 15 seconds")
		} else {
			p.logger.WithError(err).Debug("termux-location not available or failed")
		}

		// If we don't have any cached data, set a default location
		p.mu.Lock()
		if p.cachedData == nil {
			p.cachedData = &LocationData{
				Latitude:  0.0,
				Longitude: 0.0,
				Timestamp: time.Now(),
				Provider:  "default",
			}
			p.lastFetch = time.Now()
		}
		p.mu.Unlock()
		return
	}

	// Parse JSON output
	var loc LocationData
	if err := json.Unmarshal(output, &loc); err != nil {
		p.logger.WithError(err).Warn("Failed to parse termux-location output")
		return
	}

	// Add timestamp
	loc.Timestamp = time.Now()

	// Update cached data
	p.mu.Lock()
	p.cachedData = &loc
	p.lastFetch = time.Now()
	p.mu.Unlock()

	p.logger.WithFields(logrus.Fields{
		"latitude":  loc.Latitude,
		"longitude": loc.Longitude,
		"speed":     loc.Speed,
		"provider":  loc.Provider,
		"accuracy":  loc.Accuracy,
	}).Debug("Successfully fetched and cached location")
}

// IsLocationAvailable returns whether location data is available
func (p *TermuxLocationProvider) IsLocationAvailable() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.cachedData != nil
}

// GetLastFetchTime returns when location was last successfully fetched
func (p *TermuxLocationProvider) GetLastFetchTime() time.Time {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.lastFetch
}

// Stop gracefully stops the background location fetcher
func (p *TermuxLocationProvider) Stop() {
	p.logger.Info("Stopping location provider...")
	p.cancel()
}

// SetCacheTTL updates the cache time-to-live duration
func (p *TermuxLocationProvider) SetCacheTTL(ttl time.Duration) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.cacheTTL = ttl
}

// SetFetchTimeout updates the timeout for location fetch operations
func (p *TermuxLocationProvider) SetFetchTimeout(timeout time.Duration) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.fetchTimeout = timeout
}
