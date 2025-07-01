package abrpapp

import (
	"context"
	"os/exec"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

type Checker struct {
	logger *logrus.Logger

	cacheTTL    time.Duration
	lastChecked time.Time
	lastResult  bool
}

// NewChecker returns a new instance with a small cache TTL.
func NewChecker(logger *logrus.Logger) *Checker {
	return &Checker{
		logger:   logger,
		cacheTTL: 5 * time.Second,
	}
}

// IsRunning executes the adb pidof check unless the cached value is still
// fresh.
func (c *Checker) IsRunning() bool {
	if time.Since(c.lastChecked) < c.cacheTTL {
		return c.lastResult
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx,
		"/system/bin/pgrep", "-x", "com.iternio.abrpapp").Output()

	running := false
	if err == nil && len(strings.TrimSpace(string(out))) > 0 {
		running = true
	} else if err != nil {
		c.logger.WithError(err).Debug("ADB pidof failed or ABRP app not running")
	}

	// cache
	c.lastChecked = time.Now()
	c.lastResult = running
	return running
}
