package utils

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// Log levels
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

func init() {
	UpdateLogLevel()
}

// UpdateLogLevel re-reads the LOG_LEVEL from environment variables
func UpdateLogLevel() {
	// Read LOG_LEVEL from environment
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

// shouldLog checks if the message should be logged based on current level
func shouldLog(level int) bool {
	return level >= currentLogLevel
}

// logMessage formats and prints the log message
func logMessage(level int, format string, v ...interface{}) {
	if !shouldLog(level) {
		return
	}

	msg := fmt.Sprintf(format, v...)
	timestamp := time.Now().Format("2006/01/02 15:04:05")
	prefix := levelNames[level]

	// Use standard log package or fmt
	// Using fmt to control specific format requested by user (resembling the provided log snippet)
	// format: yyyy/mm/dd HH:MM:SS LEVEL: Message
	fmt.Printf("%s %s: %s\n", timestamp, prefix, msg)
}

// LogDebug logs a debug message
func LogDebug(format string, v ...interface{}) {
	logMessage(LevelDebug, format, v...)
}

// LogInfo logs an info message
func LogInfo(format string, v ...interface{}) {
	logMessage(LevelInfo, format, v...)
}

// LogWarn logs a warning message
func LogWarn(format string, v ...interface{}) {
	logMessage(LevelWarn, format, v...)
}

// LogError logs an error message
func LogError(format string, v ...interface{}) {
	logMessage(LevelError, format, v...)
}
