package bus

import (
	"sync"

	"github.com/jkaberg/byd-hass/internal/sensors"
)

// Bus provides fan-out pub/sub semantics for *sensors.SensorData* messages.
// Each Subscribe call gets its own channel that receives every future
// publication. Past messages are not replayed. The implementation is safe for
// concurrent publishers and subscribers.
type Bus struct {
	mu          sync.RWMutex
	subscribers []chan *sensors.SensorData
}

// New creates a ready-to-use Bus.
func New() *Bus { return &Bus{} }

// Subscribe returns a read-only channel that will receive all future
// SensorData snapshots.
func (b *Bus) Subscribe() <-chan *sensors.SensorData {
	ch := make(chan *sensors.SensorData, 1) // small buffer avoids blocking
	b.mu.Lock()
	b.subscribers = append(b.subscribers, ch)
	b.mu.Unlock()
	return ch
}

// Publish delivers the snapshot to all subscribers in a best-effort, non-blocking
// way. If a subscriber's buffer is full, the subscriber is dropped to keep the
// producer quick and the overall system from stalling.
func (b *Bus) Publish(s *sensors.SensorData) {
	b.mu.RLock()
	subs := make([]chan *sensors.SensorData, len(b.subscribers))
	copy(subs, b.subscribers)
	b.mu.RUnlock()

	for _, ch := range subs {
		select {
		case ch <- s:
		default:
			// Subscriber is currently busy; skip this snapshot instead of dropping the
			// subscriber entirely. The consumer will receive the next snapshot once
			// it has processed the current one.
			continue
		}
	}
}

func (b *Bus) dropSubscriber(ch chan *sensors.SensorData) {
	b.mu.Lock()
	for i, sub := range b.subscribers {
		if sub == ch {
			// remove without preserving order
			b.subscribers[i] = b.subscribers[len(b.subscribers)-1]
			b.subscribers = b.subscribers[:len(b.subscribers)-1]
			close(ch)
			break
		}
	}
	b.mu.Unlock()
}
