package cache

import (
	"reflect"
	"sync"
	"time"

	"github.com/jkaberg/byd-hass/internal/sensors"
	"github.com/sirupsen/logrus"
)

// Manager handles caching of sensor data to detect changes
type Manager struct {
	mutex  sync.RWMutex
	cache  map[string]CacheEntry
	logger *logrus.Logger
}

// CacheEntry represents a cached sensor value with metadata
type CacheEntry struct {
	Value     interface{} `json:"value"`
	Timestamp time.Time   `json:"timestamp"`
	TTL       time.Duration `json:"ttl"`
}

// NewManager creates a new cache manager
func NewManager(logger *logrus.Logger) *Manager {
	return &Manager{
		cache:  make(map[string]CacheEntry),
		logger: logger,
	}
}

// GetChanges processes sensor data, updates the cache, and returns a map of changed values.
func (m *Manager) GetChanges(data *sensors.SensorData) map[string]interface{} {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	changes := make(map[string]interface{})
	now := time.Now()

	// Use reflection to iterate over struct fields
	v := reflect.ValueOf(data).Elem()
	t := reflect.TypeOf(data).Elem()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)
		
		// Skip non-exported fields and timestamp
		if !field.CanInterface() || fieldType.Name == "Timestamp" {
			continue
		}

		// Get JSON tag name or use field name
		fieldName := fieldType.Tag.Get("json")
		if fieldName == "" || fieldName == "-" {
			fieldName = fieldType.Name
		}
		// Remove omitempty suffix
		if len(fieldName) > 10 && fieldName[len(fieldName)-10:] == ",omitempty" {
			fieldName = fieldName[:len(fieldName)-10]
		}

		// Skip nil pointer fields
		if field.Kind() == reflect.Ptr && field.IsNil() {
			continue
		}

		var currentValue interface{}
		if field.Kind() == reflect.Ptr {
			// Handle pointer fields by getting the underlying value
			if !field.IsNil() {
				currentValue = field.Elem().Interface()
			} else {
				currentValue = nil
			}
		} else {
			currentValue = field.Interface()
		}

		// Check if value has changed or if it's a new field
		cacheKey := fieldName
		cachedEntry, exists := m.cache[cacheKey]
		
		if !exists || !reflect.DeepEqual(cachedEntry.Value, currentValue) || m.isExpired(cachedEntry) {
			// Value changed, expired, or new
			changes[fieldName] = currentValue
			
			// Update cache
			m.cache[cacheKey] = CacheEntry{
				Value:     currentValue,
				Timestamp: now,
				TTL:       time.Hour, // Default TTL
			}

			m.logger.WithFields(logrus.Fields{
				"field":     fieldName,
				"old_value": cachedEntry.Value,
				"new_value": currentValue,
				"exists":    exists,
			}).Debug("Sensor value changed")
		}
	}

	return changes
}

// isExpired checks if a cache entry has expired
func (m *Manager) isExpired(entry CacheEntry) bool {
	if entry.TTL == 0 {
		return false
	}
	return time.Since(entry.Timestamp) > entry.TTL
}

// ClearExpired removes expired entries from the cache
func (m *Manager) ClearExpired() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	now := time.Now()
	for key, entry := range m.cache {
		if entry.TTL > 0 && now.Sub(entry.Timestamp) > entry.TTL {
			delete(m.cache, key)
			m.logger.WithField("key", key).Debug("Cleared expired cache entry")
		}
	}
}

// Size returns the number of cached entries
func (m *Manager) Size() int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return len(m.cache)
} 