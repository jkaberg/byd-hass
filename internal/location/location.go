package location

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// SourceTypeGPS defines the source type for GPS data
const SourceTypeGPS = "gps"

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
	p.logger.Debug("Location fetcher started")

	// Initial fetch
	p.fetchLocationData()

	// Set up periodic fetching
	ticker := time.NewTicker(10 * time.Second) // Fetch every 90 seconds
	defer ticker.Stop()

	for {
		select {
		case <-p.ctx.Done():
			p.logger.Debug("Location fetcher stopped")
			return
		case <-ticker.C:
			p.fetchLocationData()
		}
	}
}

// fetchLocationData performs the actual location fetch with timeout
func (p *TermuxLocationProvider) fetchLocationData() {
	p.logger.Debug("Fetching location via dumpsys (background)")

	// Create context with timeout
	ctx, cancel := context.WithTimeout(p.ctx, p.fetchTimeout)
	defer cancel()

	// On Android, dumpsys is located in /system/bin
	cmd := exec.CommandContext(ctx, "/system/bin/dumpsys", "location")

	output, err := cmd.Output()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			p.logger.Warn("Location fetch timed out after 15 seconds")
		} else {
			p.logger.WithError(err).Debug("dumpsys location failed")
		}

		p.setDefaultLocation()
		return
	}

	// Parse dumpsys output to LocationData
	loc, err := parseDumpsysLocation(string(output))
	if err != nil {
		p.logger.WithError(err).Warn("Failed to parse dumpsys location output")
		p.setDefaultLocation()
		return
	}

	// Add timestamp
	loc.Timestamp = time.Now()

	// Update cached data
	p.mu.Lock()
	p.cachedData = loc
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

// setDefaultLocation ensures we always have some location data when none is available.
func (p *TermuxLocationProvider) setDefaultLocation() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.cachedData == nil {
		p.cachedData = &LocationData{
			Latitude:  0.0,
			Longitude: 0.0,
			Timestamp: time.Now(),
			Provider:  "default",
		}
		p.lastFetch = time.Now()
	}
}

// Pre-compiled regular expressions – compiling on every GPS read is expensive and causes
// memory churn over long-running sessions, eventually slowing down the Android UI.

var (
	nativeRe  = regexp.MustCompile(`(?s)LatitudeDegrees:\s*([-0-9\.]+).*?LongitudeDegrees:\s*([-0-9\.]+).*?altitudeMeters:\s*([-0-9\.]+).*?speedMetersPerSecond:\s*([-0-9\.]+).*?bearingDegrees:\s*([-0-9\.]+).*?horizontalAccuracyMeters:\s*([-0-9\.]+).*?verticalAccuracyMeters:\s*([-0-9\.]+)`)
	gpsRe     = regexp.MustCompile(`(?m)^\s*gps:\s*Location\[[^]]*?([\-0-9\.]+),([\-0-9\.]+)[^]]*?(?:alt=([\-0-9\.]+))?[^]]*?(?:hAcc=([\-0-9\.]+))?[^]]*?(?:vAcc=([\-0-9\.]+))?[^]]*?(?:vel=([\-0-9\.]+))?[^]]*?(?:bear(?:=|ing=)([\-0-9\.]+))?[^]]*?]`)
	networkRe = regexp.MustCompile(`(?m)^\s*network:\s*Location\[[^]]*?([\-0-9\.]+),([\-0-9\.]+)[^]]*?(?:alt=([\-0-9\.]+))?[^]]*?(?:hAcc=([\-0-9\.]+))?[^]]*?(?:vAcc=([\-0-9\.]+))?[^]]*?(?:vel=([\-0-9\.]+))?[^]]*?(?:bear(?:=|ing=)([\-0-9\.]+))?[^]]*?]`)
)

// parseDumpsysLocation extracts GPS (preferred) or Network location information from the dumpsys output.
// It supports both the "Last Known Locations" block as well as the newer "native internal state" GNSS block.
func parseDumpsysLocation(out string) (*LocationData, error) {
	// Prefer GNSS native block first – contains the richest data and is present on many modern Android versions.
	if m := nativeRe.FindStringSubmatch(out); m != nil {
		lat, _ := strconv.ParseFloat(m[1], 64)
		lon, _ := strconv.ParseFloat(m[2], 64)
		alt, _ := strconv.ParseFloat(m[3], 64)
		speed, _ := strconv.ParseFloat(m[4], 64)
		bearing, _ := strconv.ParseFloat(m[5], 64)
		hAcc, _ := strconv.ParseFloat(m[6], 64)
		vAcc, _ := strconv.ParseFloat(m[7], 64)
		return &LocationData{
			Latitude:         lat,
			Longitude:        lon,
			Altitude:         alt,
			Accuracy:         hAcc,
			VerticalAccuracy: vAcc,
			Bearing:          bearing,
			Speed:            speed,
			Provider:         "gps",
		}, nil
	}

	// Friendly loop over the two secondary regexes.
	patterns := []struct {
		name string
		re   *regexp.Regexp
	}{{"gps", gpsRe}, {"network", networkRe}}

	for _, ptn := range patterns {
		if m := ptn.re.FindStringSubmatch(out); m != nil {
			lat, _ := strconv.ParseFloat(m[1], 64)
			lon, _ := strconv.ParseFloat(m[2], 64)
			alt, _ := strconv.ParseFloat(m[3], 64)
			hAcc, _ := strconv.ParseFloat(m[4], 64)
			vAcc, _ := strconv.ParseFloat(m[5], 64)
			speed, _ := strconv.ParseFloat(m[6], 64)
			bearing, _ := strconv.ParseFloat(m[7], 64)
			return &LocationData{
				Latitude:         lat,
				Longitude:        lon,
				Altitude:         alt,
				Accuracy:         hAcc,
				VerticalAccuracy: vAcc,
				Speed:            speed,
				Bearing:          bearing,
				Provider:         ptn.name,
			}, nil
		}
	}

	// Could not parse location
	return nil, fmt.Errorf("no location information found in dumpsys output")
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
	p.logger.Debug("Stopping location provider")
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
