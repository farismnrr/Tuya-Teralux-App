package utils

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config holds all application configuration
type Config struct {
	TuyaClientID     string
	TuyaClientSecret string
	TuyaBaseURL      string
	TuyaUserID       string
}

var AppConfig *Config

// LoadConfig loads environment variables and initializes AppConfig
func LoadConfig() {
	// Try to find .env file up to 3 levels up
	envPath := findEnvFile()
	if envPath == "" {
		log.Println("Warning: .env file not found")
	} else {
		if err := godotenv.Load(envPath); err != nil {
			log.Println("Warning: Error loading .env file")
		}
	}

	// Initialize config
	AppConfig = &Config{
		TuyaClientID:     os.Getenv("TUYA_CLIENT_ID"),
		TuyaClientSecret: os.Getenv("TUYA_ACCESS_SECRET"),
		TuyaBaseURL:      os.Getenv("TUYA_BASE_URL"),
		TuyaUserID:       os.Getenv("TUYA_USER_ID"),
	}

	// Refresh log level after loading config
	UpdateLogLevel()
}

func findEnvFile() string {
	path := ".env"
	// Check current directory
	if _, err := os.Stat(path); err == nil {
		return path
	}

	// Check up to 3 levels up
	for i := 0; i < 3; i++ {
		path = "../" + path
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return ""
}

// GetConfig returns the application config
func GetConfig() *Config {
	if AppConfig == nil {
		LoadConfig()
	}
	return AppConfig
}
