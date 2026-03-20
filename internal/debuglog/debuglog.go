package debuglog

import (
	"log"
	"os"
)

// Logger provides conditional debug logging controlled by the DEBUG
// environment variable. When DEBUG=1, messages are printed to stderr.
type Logger struct {
	enabled bool
}

// New creates a Logger. It is enabled when the DEBUG environment variable is "1".
func New() *Logger {
	return &Logger{enabled: os.Getenv("DEBUG") == "1"}
}

// Enabled returns whether debug logging is active.
func (l *Logger) Enabled() bool {
	return l.enabled
}

// Printf logs a formatted message if debug logging is enabled.
func (l *Logger) Printf(format string, args ...any) {
	if l.enabled {
		log.Printf("[DEBUG] "+format, args...)
	}
}
