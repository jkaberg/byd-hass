package domain

import (
	"math"
	"reflect"
	"time"

	"github.com/jkaberg/byd-hass/internal/sensors"
)

// Changed returns true if *cur* differs from *prev* beyond tolerated jitter.
// It zeroes the Timestamp field and ignores small GPS noise so that minor
// location updates don't trigger a transmit.
func Changed(prev, cur *sensors.SensorData) bool {
	if prev == nil && cur == nil {
		return false
	}
	if prev == nil || cur == nil {
		return true
	}

	p, c := *prev, *cur // copy
	p.Timestamp = time.Time{}
	c.Timestamp = time.Time{}

	// Ignore wall-clock date/time fields that naturally change every minute
	p.Year, p.Month, p.Day, p.Hour, p.Minute = nil, nil, nil, nil, nil
	c.Year, c.Month, c.Day, c.Hour, c.Minute = nil, nil, nil, nil, nil

	if p.Location != nil && c.Location != nil {
		const distThr = 10.0 // metres
		const bearThr = 5.0  // degrees
		dist := haversineMeters(p.Location.Latitude, p.Location.Longitude,
			c.Location.Latitude, c.Location.Longitude)
		bearingDiff := math.Abs(p.Location.Bearing - c.Location.Bearing)
		if bearingDiff > 180 {
			bearingDiff = 360 - bearingDiff
		}
		if dist < distThr && bearingDiff < bearThr {
			p.Location = nil
			c.Location = nil
		}
	}

	return !reflect.DeepEqual(p, c)
}

func haversineMeters(lat1, lon1, lat2, lon2 float64) float64 {
	const r = 6371000.0 // Earth radius in metres
	dLat := toRad(lat2 - lat1)
	dLon := toRad(lon2 - lon1)
	lat1Rad := toRad(lat1)
	lat2Rad := toRad(lat2)

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return r * c
}

func toRad(deg float64) float64 { return deg * math.Pi / 180 }
