// Package logger provides structured logging capabilities for the application
package logger

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/fatih/color"
)

// LogLevel represents the severity level of a log message
type LogLevel int

const (
	// DebugLevel is for detailed debug information
	DebugLevel LogLevel = iota
	// InfoLevel is for general operational information
	InfoLevel
	// WarnLevel is for warning events that might cause issues
	WarnLevel
	// ErrorLevel is for error events that might still allow the application to continue
	ErrorLevel
	// FatalLevel is for very severe error events that will lead to application termination
	FatalLevel
)

var levelNames = map[LogLevel]string{
	DebugLevel: "DEBUG",
	InfoLevel:  "INFO",
	WarnLevel:  "WARN",
	ErrorLevel: "ERROR",
	FatalLevel: "FATAL",
}

var levelColors = map[LogLevel]func(a ...any) string{
	DebugLevel: color.New(color.FgHiBlack).SprintFunc(),
	InfoLevel:  color.New(color.FgBlue).SprintFunc(),
	WarnLevel:  color.New(color.FgYellow).SprintFunc(),
	ErrorLevel: color.New(color.FgRed).SprintFunc(),
	FatalLevel: color.New(color.FgHiRed, color.Bold).SprintFunc(),
}

// Logger is the main struct for logging functionality
type Logger struct {
	level  LogLevel
	output io.Writer
	prefix string
}

// New creates a new Logger instance with default settings
func New(prefix string) *Logger {
	return &Logger{
		level:  InfoLevel,
		output: os.Stdout,
		prefix: prefix,
	}
}

// SetLevel sets the minimum log level
func (l *Logger) SetLevel(level LogLevel) {
	l.level = level
}

// SetOutput sets the output writer
func (l *Logger) SetOutput(w io.Writer) {
	l.output = w
}

// log writes a log message with the specified level
func (l *Logger) log(level LogLevel, format string, args ...any) {
	if level < l.level {
		return
	}

	now := time.Now().Format("2006-01-02 15:04:05.000")
	levelName := levelNames[level]
	coloredLevel := levelColors[level](levelName)

	prefix := ""
	if l.prefix != "" {
		prefix = fmt.Sprintf("[%s] ", l.prefix)
	}

	message := fmt.Sprintf(format, args...)
	fmt.Fprintf(l.output, "%s %s %s%s\n", now, coloredLevel, prefix, message)

	// If this is a fatal message, exit the program
	if level == FatalLevel {
		os.Exit(1)
	}
}

// Debug logs a debug message
func (l *Logger) Debug(format string, args ...any) {
	l.log(DebugLevel, format, args...)
}

// Info logs an informational message
func (l *Logger) Info(format string, args ...any) {
	l.log(InfoLevel, format, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(format string, args ...any) {
	l.log(WarnLevel, format, args...)
}

// Error logs an error message
func (l *Logger) Error(format string, args ...any) {
	l.log(ErrorLevel, format, args...)
}

// Fatal logs a fatal message and exits the program
func (l *Logger) Fatal(format string, args ...any) {
	l.log(FatalLevel, format, args...)
	// Control should never reach here due to os.Exit in log()
}

// DefaultLogger is a shared logger instance
var DefaultLogger = New("")

// SetGlobalLevel sets the log level for the default logger
func SetGlobalLevel(level LogLevel) {
	DefaultLogger.SetLevel(level)
}

// Global convenience functions that use the default logger

// Debug logs a debug message using the default logger
func Debug(format string, args ...any) {
	DefaultLogger.Debug(format, args...)
}

// Info logs an informational message using the default logger
func Info(format string, args ...any) {
	DefaultLogger.Info(format, args...)
}

// Warn logs a warning message using the default logger
func Warn(format string, args ...any) {
	DefaultLogger.Warn(format, args...)
}

// Error logs an error message using the default logger
func Error(format string, args ...any) {
	DefaultLogger.Error(format, args...)
}

// Fatal logs a fatal message and exits the program using the default logger
func Fatal(format string, args ...any) {
	DefaultLogger.Fatal(format, args...)
}
