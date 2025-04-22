package temporal

import (
	"log"
)

// StderrLogger implements the Temporal logger interface
// ensuring all Temporal logs go to stderr instead of stdout
type StderrLogger struct {
	logger *log.Logger
}

// Debug logs a debug message
func (l *StderrLogger) Debug(msg string, keyvals ...interface{}) {
	l.logger.Printf("[DEBUG] %s", msg)
}

// Info logs an info message
func (l *StderrLogger) Info(msg string, keyvals ...interface{}) {
	l.logger.Printf("[INFO] %s", msg)
}

// Warn logs a warning message
func (l *StderrLogger) Warn(msg string, keyvals ...interface{}) {
	l.logger.Printf("[WARN] %s", msg)
}

// Error logs an error message
func (l *StderrLogger) Error(msg string, keyvals ...interface{}) {
	l.logger.Printf("[ERROR] %s", msg)
}
