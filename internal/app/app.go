package app

import (
	"context"
	"fmt"
	"strings"
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
					if now.Sub(st.lastSent) < st.interval {
						continue
					}
					if !domain.Changed(st.lastSnap, latest) {
						continue
					}
					if err := st.sendFn(ctx, latest, logger); err != nil {
						logger.WithError(err).Warn(st.name + " transmit failed")
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
	// Transmitter has its own internal timeouts; context reserved for future.
	_ = ctx
	
	if err := tx.Transmit(data); err != nil {
		// Get detailed connection status for better error reporting
		status := tx.GetConnectionStatus()
		
		// Log with enhanced context
		logFields := logrus.Fields{
			"error":               err.Error(),
			"connected":           status["connected"],
			"consecutive_failures": status["consecutive_failures"],
		}
		
		// Add backoff information if available
		if inBackoff, ok := status["in_backoff"].(bool); ok && inBackoff {
			logFields["remaining_backoff"] = status["remaining_backoff"]
			logFields["current_backoff_delay"] = status["current_backoff_delay"]
		}
		
		// Different log levels based on error type
		if strings.Contains(err.Error(), "skipping transmission due to backoff") {
			logger.WithFields(logFields).Debug("ABRP transmission skipped due to backoff")
		} else {
			logger.WithFields(logFields).Warn("ABRP transmission failed after retries")
		}
		
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
