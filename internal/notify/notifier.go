package notify

import (
	"context"
	"os"
	"os/exec"
	"time"

	"github.com/sirupsen/logrus"
)

// termuxNotificationPath holds the absolute path to the termux-notification
// binary. Using an absolute path avoids the PATH lookup that would otherwise
// trigger the faccessat2 syscall, which is blocked by Android\'s seccomp
// policy on older versions (e.g. Android 10). The default hard-coded prefix
// points to the canonical Termux installation directory but can be overridden
// at runtime via the PREFIX environment variable.
var termuxNotificationPath string

func init() {
	prefix := os.Getenv("PREFIX")
	if prefix == "" {
		prefix = "/data/data/com.termux/files/usr"
	}
	termuxNotificationPath = prefix + "/bin/termux-notification"
}

// TermuxNotifier sends Android notifications via the `termux-notification` CLI that ships with Termux.
//
// Notifications are updated in-place by always re-using the same notification ID. This
// avoids piling up duplicate messages while still giving the user real-time feedback
// when the status changes (e.g. MQTT re-connects, ABRP fails, etc.).
//
// The implementation purposefully ignores errors from the command – if Termux is not
// available (e.g. the program is executed outside of Android) it silently degrades.
// Only unexpected failures are logged at *debug* level to avoid log noise in normal
// operation.
//
// A very small execution timeout is used to ensure the main telemetry loop will never
// be blocked by notification issues.
type TermuxNotifier struct {
	id     string
	logger *logrus.Logger
}

// NewTermuxNotifier instantiates a notifier that re-uses a constant notification ID so
// the same notification is updated instead of new ones being created on every update.
func NewTermuxNotifier(logger *logrus.Logger) *TermuxNotifier {
	return &TermuxNotifier{ // ID 1337 chosen arbitrarily but consistently.
		id:     "1337",
		logger: logger,
	}
}

// Notify posts (or updates) a notification with the supplied title and message body.
func (n *TermuxNotifier) Notify(title, content string) {
	// Fail fast if the title is empty – guards against accidental misuse.
	if title == "" {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1500*time.Millisecond)
	defer cancel()

	// Build argument slice. The flags accepted by termux-notification are documented
	// here: https://wiki.termux.com/wiki/Termux-notification
	args := []string{
		"--id", n.id,
		"-t", title,
		"-c", content,
		"--priority", "low",
		"--ongoing",
	}

	if err := exec.CommandContext(ctx, termuxNotificationPath, args...).Run(); err != nil {
		// Only log in debug mode to honour the "minimal logging" project guideline.
		n.logger.WithError(err).Debug("termux-notification execution failed")
	}
}
