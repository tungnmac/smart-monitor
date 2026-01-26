// Package logger provides logging utilities
package logger

import (
	"log"
	"os"
)

// Logger wraps standard logger
type Logger struct {
	*log.Logger
}

// New creates a new logger
func New() *Logger {
	return &Logger{
		Logger: log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile),
	}
}

// Info logs info message
func (l *Logger) Info(msg string) {
	l.Println("[INFO]", msg)
}

// Error logs error message
func (l *Logger) Error(msg string) {
	l.Println("[ERROR]", msg)
}

// Debug logs debug message
func (l *Logger) Debug(msg string) {
	l.Println("[DEBUG]", msg)
}

// Fatal logs fatal message and exits
func (l *Logger) Fatal(msg string) {
	l.Println("[FATAL]", msg)
	os.Exit(1)
}
