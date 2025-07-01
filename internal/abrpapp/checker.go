package abrpapp

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// adbPath stores the absolute path to the ADB binary shipped with Termux.
// Using an absolute path (containing a slash) means os/exec skips the PATH
// search that would otherwise trigger the faccessat2 syscall blocked by
// Android 10. We determine the path once at init time.
var adbPath string

func init() {
	prefix := os.Getenv("PREFIX")
	if prefix == "" {
		// Default Termux prefix.
		prefix = "/data/data/com.termux/files/usr"
	}
	adbPath = prefix + "/bin/adb"
}

// Checker verifies that the ABRP Android application is running on the device.
//
// The detection strategy avoids spawning external commands (which would trigger
// the faccessat2 syscall that is blocked by the Android seccomp policy on
// Android 10). Instead, it walks through /proc/*/cmdline and looks for a
// process whose command-line contains the package name "com.iternio.abrpapp".
//
// The outcome is cached for cacheTTL to minimise the overhead of scanning the
// procfs on every call.
type Checker struct {
	cacheTTL time.Duration
	logger   *logrus.Logger

	lastChecked time.Time
	lastResult  bool

	// ADB connection caching
	connected          bool
	lastConnectAttempt time.Time
	connectTTL         time.Duration // how long an adb connect is considered valid
}

// NewChecker returns an initialised Checker with sane defaults.
func NewChecker(logger *logrus.Logger) *Checker {
	return &Checker{
		cacheTTL:   5 * time.Second,
		connectTTL: 60 * time.Second,
		logger:     logger,
	}
}

// IsRunning returns true if the ABRP app process is currently running. The
// result is memoised for cacheTTL.
func (c *Checker) IsRunning() bool {
	if time.Since(c.lastChecked) < c.cacheTTL {
		return c.lastResult
	}

	// Fast path: try ADB-based check first, mirroring Termux location provider
	running, adbOK := c.checkViaADB()
	if !adbOK {
		// Fall back to /proc scan if ADB unavailable (e.g. binary missing)
		running = findProcess("com.iternio.abrpapp")
	}

	c.updateCache(running)
	return running
}

// findProcess scans /proc for a process whose cmdline contains the given
// substring. It returns as soon as it finds a match. Any errors are swallowed –
// the function simply returns false when it cannot determine the state.
func findProcess(substr string) bool {
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return false
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		// Only consider PID directories (must be all digits).
		isPID := true
		for i := 0; i < len(name); i++ {
			if name[i] < '0' || name[i] > '9' {
				isPID = false
				break
			}
		}
		if !isPID {
			continue
		}

		cmdlinePath := filepath.Join("/proc", name, "cmdline")
		data, err := os.ReadFile(cmdlinePath)
		if err != nil || len(data) == 0 {
			continue
		}
		// cmdline is NUL-separated; replace NULs with spaces for searching.
		cmdline := strings.ReplaceAll(string(data), "\x00", " ")
		if strings.Contains(cmdline, substr) {
			return true
		}
	}

	return false
}

func (c *Checker) updateCache(result bool) {
	if result != c.lastResult {
		if result {
			c.logger.Debug("ABRP app detected, telemetry can resume")
		} else {
			c.logger.Debug("ABRP app no longer detected, telemetry suspended")
		}
	}

	c.lastChecked = time.Now()
	c.lastResult = result
}

// checkViaADB tries to detect the process via `adb shell pidof …`.
// It returns (running, adbOK). adbOK == false means the method couldn't be run
// (binary missing, exec failure, etc.) and the caller should fall back to
// another strategy.
func (c *Checker) checkViaADB() (bool, bool) {
	// Ensure the binary exists; if not, no point in trying.
	if _, err := os.Stat(adbPath); err != nil {
		return false, false
	}

	// Keep the connect call rate-limited.
	if !c.connected || time.Since(c.lastConnectAttempt) > c.connectTTL {
		ctxConn, cancelConn := context.WithTimeout(context.Background(), 3*time.Second)
		err := exec.CommandContext(ctxConn, adbPath, "connect", "localhost:5555").Run()
		cancelConn()
		c.lastConnectAttempt = time.Now()
		if err != nil {
			c.connected = false
			c.logger.WithError(err).Debug("ADB connect failed – falling back to /proc scan")
		} else {
			c.connected = true
		}
	}

	// If we're not connected, bail out so the caller can try the other method.
	if !c.connected {
		return false, false
	}

	// Query for the ABRP app pid.
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	var stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, adbPath, "-s", "localhost:5555", "shell", "pidof", "com.iternio.abrpapp")
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	cancel()

	// pidof returns exit status 1 when the process is not found – treat that as
	// a clean negative result.
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return false, true // adb ran successfully but app not running
		}
		// Unexpected error – abandon ADB path for now.
		c.logger.WithError(err).WithField("stderr", strings.TrimSpace(stderr.String())).Debug("ADB pidof failed – falling back to /proc scan")
		c.connected = false
		return false, false
	}

	return len(strings.TrimSpace(string(out))) > 0, true
}
