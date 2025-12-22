package utils

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config holds the application's configuration parameters.
// These are loaded from environment variables or a .env file.
type Config struct {
	TuyaClientID              string
	TuyaClientSecret          string
	TuyaBaseURL               string
	TuyaUserID                string
	ApiKey                    string
	SwaggerBaseURL            string
	GetAllDevicesResponseType string
}

// AppConfig is the global configuration instance.
var AppConfig *Config

// LoadConfig initializes the AppConfig by loading variables from the environment.
// It searches for a .env file in the current and parent directories if not already set.
// It also triggers an update of the log level based on the loaded configuration.
func LoadConfig() {
	envPath := findEnvFile()
	if envPath == "" {
		log.Println("Warning: .env file not found")
	} else {
		if err := godotenv.Load(envPath); err != nil {
			log.Println("Warning: Error loading .env file")
		}
	}

	AppConfig = &Config{
		TuyaClientID:              os.Getenv("TUYA_CLIENT_ID"),
		TuyaClientSecret:          os.Getenv("TUYA_ACCESS_SECRET"),
		TuyaBaseURL:               os.Getenv("TUYA_BASE_URL"),
		TuyaUserID:                os.Getenv("TUYA_USER_ID"),
		ApiKey:                    os.Getenv("API_KEY"),
		SwaggerBaseURL:            os.Getenv("SWAGGER_BASE_URL"),
		GetAllDevicesResponseType: os.Getenv("GET_ALL_DEVICES_RESPONSE"),
	}

	UpdateLogLevel()
}

// findEnvFile searches for the .env file in the current directory and up to three parent levels.
//
// return string The path to the .env file if found, otherwise an empty string.
func findEnvFile() string {
	path := ".env"
	if _, err := os.Stat(path); err == nil {
		return path
	}

	for i := 0; i < 3; i++ {
		path = "../" + path
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return ""
}

// GetConfig returns the singleton AppConfig instance.
// If the config hasn't been loaded, it triggers LoadConfig first.
//
// return *Config The global configuration object.
func GetConfig() *Config {
	if AppConfig == nil {
		LoadConfig()
	}
	return AppConfig
}
