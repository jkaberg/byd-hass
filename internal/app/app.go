package app

import (
	"context"
	"fmt"
	"time"

	"github.com/jkaberg/byd-hass/internal/api"
	"github.com/jkaberg/byd-hass/internal/bus"
	"github.com/jkaberg/byd-hass/internal/config"
	"github.com/jkaberg/byd-hass/internal/domain"
	"github.com/jkaberg/byd-hass/internal/location"
	"github.com/jkaberg/byd-hass/internal/sensors"
	"github.com/jkaberg/byd-hass/internal/transmission"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

// Adaptive ABRP intervals ------------------------------------------------
const (
	abrpDrivingInterval = 10 * time.Second  // default while moving / charging
	abrpIdleInterval    = 120 * time.Second // when parked & not charging
)

func computeABRPInterval(data *sensors.SensorData) time.Duration {
	if data == nil {
		return abrpDrivingInterval
	}
	// Fast cadence when speed > 0 km/h
	if data.Speed != nil && *data.Speed > 0 {
		return abrpDrivingInterval
	}
	// Fast cadence when actively charging
	if sensors.DeriveChargingStatus(data) == "charging" {
		return abrpDrivingInterval
	}
	// Otherwise we're parked / idle
	return abrpIdleInterval
}

// Run launches the hexagonal architecture and blocks until ctx is cancelled.
func Run(
	parentCtx context.Context,
	cfg *config.Config,
	diplusClient *api.DiplusClient,
	locationProvider *location.TermuxLocationProvider,
	mqttTx *transmission.MQTTTransmitter,
	abrpTx *transmission.ABRPTransmitter,
	logger *logrus.Logger,
) {
	ctx, cancel := context.WithCancel(parentCtx)
	go func() {
		<-parentCtx.Done()
		cancel()
	}()

	messageBus := bus.New()
	grp, ctx := errgroup.WithContext(ctx)

	// Collector -----------------------------------------------------------
	grp.Go(func() error {
		ticker := time.NewTicker(config.DiplusPollInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-ticker.C:
				sensorData, err := diplusClient.Poll()
				if err != nil {
					logger.WithError(err).Warn("collector: poll failed")
					continue
				}
				if cfg.ABRPLocation && locationProvider != nil {
					if loc, err := locationProvider.GetLocation(); err == nil {
						sensorData.Location = loc
					}
				}
				messageBus.Publish(sensorData)
			}
		}
	})

	// Central scheduler ----------------------------------------------------

	sub := messageBus.Subscribe()

	type txState struct {
		interval time.Duration
		lastSent time.Time
		lastSnap *sensors.SensorData
		sendFn   func(context.Context, *sensors.SensorData, *logrus.Logger) error
		name     string
	}

	var states []txState
	if mqttTx != nil {
		states = append(states, txState{
			interval: cfg.MQTTInterval,
			lastSent: time.Now().Add(-cfg.MQTTInterval),
			sendFn: func(c context.Context, s *sensors.SensorData, l *logrus.Logger) error {
				return transmitToMQTTAsync(c, mqttTx, s, l)
			},
			name: "MQTT",
		})
	}
	if abrpTx != nil {
		states = append(states, txState{
			interval: cfg.ABRPInterval,
			lastSent: time.Now().Add(-cfg.ABRPInterval),
			sendFn: func(c context.Context, s *sensors.SensorData, l *logrus.Logger) error {
				return transmitToABRPAsync(c, abrpTx, s, l)
			},
			name: "ABRP",
		})
	}

	grp.Go(func() error {
		var latest *sensors.SensorData
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case snap, ok := <-sub:
				if !ok {
					return nil
				}
				latest = snap
			case <-ticker.C:
				if latest == nil {
					continue
				}
				now := time.Now()
				for i := range states {
					st := &states[i]
					// Dynamic interval for ABRP depending on vehicle state.
					interval := st.interval
					if st.name == "ABRP" {
						interval = computeABRPInterval(latest)
					}

					if now.Sub(st.lastSent) < interval {
						continue
					}
					if !domain.Changed(st.lastSnap, latest) {
						continue
					}
					if err := st.sendFn(ctx, latest, logger); err != nil {
						logger.WithError(err).Warn(st.name + " transmit failed")
						// Ensure we retry even if no data change.
						// Reset lastSnap so Changed() will evaluate to true on the next
						// scheduler tick, and bump lastSent so we still respect the
						// configured transmission interval.
						st.lastSnap = nil
						st.lastSent = now
					} else {
						st.lastSnap = latest
						st.lastSent = now
					}
				}
			}
		}
	})

	if err := grp.Wait(); err != nil && err != context.Canceled {
		logger.WithError(err).Warn("app: background group exited")
	}
}

func transmitToABRPAsync(ctx context.Context, tx *transmission.ABRPTransmitter, data *sensors.SensorData, logger *logrus.Logger) error {
	if tx == nil || data == nil {
		return nil
	}
	// Pass the caller context down so that a global cancellation stops in-flight HTTP.
	if err := tx.TransmitWithContext(ctx, data); err != nil {
		return fmt.Errorf("ABRP transmit failed: %w", err)
	}
	return nil
}

func transmitToMQTTAsync(ctx context.Context, tx *transmission.MQTTTransmitter, data *sensors.SensorData, logger *logrus.Logger) error {
	if tx == nil || data == nil {
		return nil
	}
	_ = ctx
	if err := tx.Transmit(data); err != nil {
		return fmt.Errorf("MQTT transmit failed: %w", err)
	}
	return nil
}
