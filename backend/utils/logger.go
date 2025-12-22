package utils

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// LogLevel constants define the severity of log messages.
const (
	LevelDebug = iota
	LevelInfo
	LevelWarn
	LevelError
)

var (
	currentLogLevel = LevelInfo // Default to INFO
	levelNames      = []string{"DEBUG", "INFO", "WARN", "ERROR"}
)

// init initializes the logger configuration on package startup.
func init() {
	UpdateLogLevel()
}

// UpdateLogLevel reads the 'LOG_LEVEL' environment variable and updates the current log level.
// Valid values: DEBUG, INFO, WARN, ERROR. Defaults to INFO if invalid or unset.
func UpdateLogLevel() {
	envLevel := os.Getenv("LOG_LEVEL")
	switch strings.ToUpper(envLevel) {
	case "DEBUG":
		currentLogLevel = LevelDebug
	case "INFO":
		currentLogLevel = LevelInfo
	case "WARN":
		currentLogLevel = LevelWarn
	case "ERROR":
		currentLogLevel = LevelError
	default:
		currentLogLevel = LevelInfo // Default if unset or invalid
	}
}

// shouldLog determines if a message with the given level should be logged.
//
// param level The severity level of the message.
// return bool True if the level is greater than or equal to the current log level.
func shouldLog(level int) bool {
	return level >= currentLogLevel
}

// logMessage formats and prints a log message to stdout.
// It includes a timestamp and the log level prefix.
//
// param level The severity level of the message.
// param format The format string (printf style).
// param v The arguments for the format string.
func logMessage(level int, format string, v ...interface{}) {
	if !shouldLog(level) {
		return
	}

	msg := fmt.Sprintf(format, v...)
	timestamp := time.Now().Format("2006/01/02 15:04:05")
	prefix := levelNames[level]
	fmt.Printf("%s %s: %s\n", timestamp, prefix, msg)
}

// LogDebug logs a message at DEBUG level.
//
// param format The format string.
// param v The arguments.
func LogDebug(format string, v ...interface{}) {
	logMessage(LevelDebug, format, v...)
}

// LogInfo logs a message at INFO level.
//
// param format The format string.
// param v The arguments.
func LogInfo(format string, v ...interface{}) {
	logMessage(LevelInfo, format, v...)
}

// LogWarn logs a message at WARN level.
//
// param format The format string.
// param v The arguments.
func LogWarn(format string, v ...interface{}) {
	logMessage(LevelWarn, format, v...)
}

// LogError logs a message at ERROR level.
//
// param format The format string.
// param v The arguments.
func LogError(format string, v ...interface{}) {
	logMessage(LevelError, format, v...)
}
