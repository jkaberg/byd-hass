package wifi

import (
	"context"
	"os/exec"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// WiFiManager handles WiFi state checking and re-enabling
type WiFiManager struct {
	logger *logrus.Logger
}

// NewWiFiManager creates a new WiFi manager instance
func NewWiFiManager(logger *logrus.Logger) *WiFiManager {
	return &WiFiManager{
		logger: logger,
	}
}

// IsWiFiEnabled checks if WiFi is currently enabled
// Returns true if WiFi is enabled, false if disabled, and an error if the check fails
func (w *WiFiManager) IsWiFiEnabled(ctx context.Context) (bool, error) {
	// Use 'settings get global wifi_on' to check WiFi status
	// Returns "1" if enabled, "0" if disabled
	// Add a timeout to prevent hanging
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "settings", "get", "global", "wifi_on")
	output, err := cmd.Output()
	if err != nil {
		return false, err
	}

	status := strings.TrimSpace(string(output))
	// Check if the output is "1" (enabled)
	return status == "1", nil
}

// EnableWiFi enables WiFi using the Android service command
func (w *WiFiManager) EnableWiFi(ctx context.Context) error {
	// Use 'svc wifi enable' to enable WiFi
	// Add a timeout to prevent hanging
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "svc", "wifi", "enable")
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

// CheckAndReenable checks if WiFi is disabled and re-enables it if needed
// Returns true if WiFi was re-enabled, false otherwise
func (w *WiFiManager) CheckAndReenable(ctx context.Context) (bool, error) {
	enabled, err := w.IsWiFiEnabled(ctx)
	if err != nil {
		return false, err
	}

	if !enabled {
		w.logger.Info("WiFi is disabled, attempting to re-enable...")
		if err := w.EnableWiFi(ctx); err != nil {
			w.logger.WithError(err).Warn("Failed to enable WiFi")
			return false, err
		}
		// Give WiFi a moment to enable
		time.Sleep(500 * time.Millisecond)
		// Verify it was enabled
		enabled, err := w.IsWiFiEnabled(ctx)
		if err != nil {
			w.logger.WithError(err).Warn("Failed to verify WiFi status after enabling")
			return true, nil // Assume it worked if we can't verify
		}
		if enabled {
			w.logger.Info("WiFi successfully re-enabled")
			return true, nil
		}
		w.logger.Warn("WiFi enable command succeeded but WiFi is still disabled")
		return false, nil
	}

	return false, nil
}

// MonitorWiFi runs a monitoring loop that periodically checks and re-enables WiFi
// This function blocks until ctx is cancelled
func (w *WiFiManager) MonitorWiFi(ctx context.Context, checkInterval time.Duration) error {
	ticker := time.NewTicker(checkInterval)
	defer ticker.Stop()

	// Do an initial check
	if _, err := w.CheckAndReenable(ctx); err != nil {
		w.logger.WithError(err).Debug("Initial WiFi check failed (non-fatal)")
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			_, err := w.CheckAndReenable(ctx)
			if err != nil {
				w.logger.WithError(err).Debug("WiFi check failed (non-fatal)")
			}
		}
	}
}
