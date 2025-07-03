package cache

import (
	"math"
	"reflect"
	"sync"
	"time"

	"github.com/jkaberg/byd-hass/internal/sensors"
	"github.com/sirupsen/logrus"
)

// Manager keeps a single previous SensorData snapshot and answers the
// question: "has anything changed since the last time I asked?".
// It is concurrency‐safe for the simple read-then-write pattern used in
// main.go.
//
// Behaviour:
//   • First call to Changed() always returns true and stores the snapshot.
//   • Timestamp fields are ignored when comparing.
//   • The stored snapshot is replaced only when a difference is detected
//     to avoid unnecessary allocations.
//
// This trims away the old per-field map, TTL handling and reflection-heavy
// loops while preserving the same outward semantics.

type Manager struct {
	mu   sync.RWMutex
	prev *sensors.SensorData
}

// Changed compares the supplied snapshot against the previously stored one,
// ignoring the Timestamp field. If a change is detected it updates the
// stored snapshot and returns true.
func (m *Manager) Changed(cur *sensors.SensorData) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.prev == nil {
		m.prev = clone(cur)
		return true
	}

	if !equalNoTimestamp(m.prev, cur) {
		m.prev = clone(cur)
		return true
	}
	return false
}

// equalNoTimestamp does a deep equality check after zeroing the Timestamp
// field on temporaries so the comparison isn't affected by the wall-clock.
func equalNoTimestamp(a, b *sensors.SensorData) bool {
	aa, bb := *a, *b
	// Ignore Timestamp for change detection.
	aa.Timestamp = time.Time{}
	bb.Timestamp = time.Time{}

	// Location – ignore micro-jitter (<10 m & <5 °).
	if aa.Location != nil && bb.Location != nil {
		const distThreshold = 10.0   // metres
		const bearingThreshold = 5.0 // degrees

		dist := haversineMeters(aa.Location.Latitude, aa.Location.Longitude,
			bb.Location.Latitude, bb.Location.Longitude)

		bearingDiff := math.Abs(aa.Location.Bearing - bb.Location.Bearing)
		if bearingDiff > 180 {
			bearingDiff = 360 - bearingDiff // smallest angular difference
		}

		if dist < distThreshold && bearingDiff < bearingThreshold {
			// Movement below thresholds → treat as unchanged
			aa.Location = nil
			bb.Location = nil
		}
	}
	return reflect.DeepEqual(aa, bb)
}

// haversineMeters returns great-circle distance in metres between two lat/lon points.
func haversineMeters(lat1, lon1, lat2, lon2 float64) float64 {
	const r = 6371000.0 // Earth radius in metres
	dLat := toRad(lat2 - lat1)
	dLon := toRad(lon2 - lon1)
	lat1Rad := toRad(lat1)
	lat2Rad := toRad(lat2)

	a := math.Sin(dLat/2)*math.Sin(dLat/2) + math.Cos(lat1Rad)*math.Cos(lat2Rad)*math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return r * c
}

func toRad(deg float64) float64 { return deg * math.Pi / 180 }

// clone returns a shallow copy of the struct. Pointer fields are copied as
// pointers; this is fine because the values are treated as immutable after
// creation.
func clone(src *sensors.SensorData) *sensors.SensorData {
	dst := *src
	return &dst
}

// NewManager returns a ready-to-use cache manager. The logger parameter is
// kept to avoid touching call-sites; it is currently unused but allows us to
// re-introduce logging cheaply in the future.
func NewManager(logger *logrus.Logger) *Manager {
	_ = logger // suppress unused param warning
	return &Manager{}
}
