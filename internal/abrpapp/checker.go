package abrpapp

import (
	"bytes"
	"context"
	"os/exec"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// Checker verifies that the ABRP Android application is running on the device.
//
// The detection strategy relies on an ADB-over-TCP connection to localhost:5555.
// For every invocation of IsRunning() (subject to a small cache TTL) it makes
// sure the connection is established and then executes:
//
//	adb -s localhost:5555 shell "pidof com.iternio.abrpapp"
//
// If the command succeeds (exit code 0) the application is considered running.
// Any non-zero exit code or unexpected error means the app is not running.
//
// Only *errors* are logged. Positive detections (or absence thereof) are silent
// to avoid log noise, per project requirements.
//
// The cached result is kept for cacheTTL to minimise ADB round-trips.
type Checker struct {
	device   string
	cacheTTL time.Duration
	logger   *logrus.Logger

	lastChecked time.Time
	lastResult  bool

	connected          bool          // Whether adb connect has succeeded recently
	lastConnectAttempt time.Time     // Timestamp of last connect attempt
	connectTTL         time.Duration // How long a connect is considered valid
}

// NewChecker returns an initialised Checker with sane defaults.
func NewChecker(logger *logrus.Logger) *Checker {
	return &Checker{
		device:     "localhost:5555",
		cacheTTL:   5 * time.Second,
		connectTTL: 60 * time.Second,
		logger:     logger,
	}
}

// IsRunning returns true if the ABRP app process is currently running.
// It caches the result for the configured cacheTTL.
func (c *Checker) IsRunning() bool {
	if time.Since(c.lastChecked) < c.cacheTTL {
		return c.lastResult
	}

	// Ensure ADB connection, avoid spamming `adb connect`.
	if !c.connected || time.Since(c.lastConnectAttempt) > c.connectTTL {
		ctxConn, cancelConn := context.WithTimeout(context.Background(), 3*time.Second)
		err := exec.CommandContext(ctxConn, "adb", "connect", c.device).Run()
		cancelConn()
		c.lastConnectAttempt = time.Now()
		if err != nil {
			c.connected = false
			c.logger.WithError(err).Warn("ADB connect failed")
		} else {
			c.connected = true
		}
	}

	// Query for the ABRP app pid.
	ctx2, cancel2 := context.WithTimeout(context.Background(), 2*time.Second)
	var stderr bytes.Buffer
	cmd := exec.CommandContext(ctx2, "adb", "-s", c.device, "shell", "pidof", "com.iternio.abrpapp")
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	// pidof returns exit status 1 when the process is not found, which isAdd commentMore actions
	// the normal signal that the ABRP app is not running. Treat that as a
	// non-error to avoid noisy warnings in the log.
	running := err == nil && len(strings.TrimSpace(string(out))) > 0
	if err != nil && ctx2.Err() == nil {
		// Check for the common "process not found" case (exit status 1).
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			// This just means the ABRP app is not running; log at debug level.
			c.logger.Debug("ABRP app not detected (exit status 1)")
		} else {
			// Unexpected ADB or shell failure: warn and reset connection state.
			c.logger.WithError(err).WithField("stderr", strings.TrimSpace(stderr.String())).Warn("ADB pidof command failed")
			c.connected = false
		}
	}

	cancel2()

	c.updateCache(running)
	return running
}

func (c *Checker) updateCache(result bool) {
	c.lastChecked = time.Now()
	c.lastResult = result
}
