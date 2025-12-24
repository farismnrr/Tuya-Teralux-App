package usecases

import (
	"fmt"
	"teralux_app/domain/tuya/dtos"
)

// TuyaSensorUseCase handles retrieval and interpretation of sensor data.
// It parses raw device status values (like temperature, humidity) into formatted DTOs.
type TuyaSensorUseCase struct {
	getDeviceUseCase *TuyaGetDeviceByIDUseCase
}

// NewTuyaSensorUseCase initializes a new TuyaSensorUseCase.
//
// param getDeviceUseCase The usecase dependency for fetching raw device data.
// return *TuyaSensorUseCase A pointer to the initialized usecase.
func NewTuyaSensorUseCase(getDeviceUseCase *TuyaGetDeviceByIDUseCase) *TuyaSensorUseCase {
	return &TuyaSensorUseCase{
		getDeviceUseCase: getDeviceUseCase,
	}
}

// GetSensorData retrieves, interprets, and formats sensor readings for a specific device.
// It converts raw values (often integers scaled by 10) into human-readable floats and generates descriptive status text.
//
// param accessToken The valid OAuth 2.0 access token.
// param deviceID The device ID of the sensor.
// return *dtos.SensorDataDTO The structured sensor data containing temperature, humidity, and status.
// return error An error if fetching the device data fails.
func (uc *TuyaSensorUseCase) GetSensorData(accessToken, deviceID string) (*dtos.SensorDataDTO, error) {
	device, err := uc.getDeviceUseCase.GetDeviceByID(accessToken, deviceID)
	if err != nil {
		return nil, err
	}

	var temperature float64
	var humidity int
	var battery int

	// Parse status values
	for _, status := range device.Status {
		switch status.Code {
		case "va_temperature":
			// value is likely float64 or int in JSON, often comes as float64 from generic interface{} unmarshaling
			if val, ok := status.Value.(float64); ok {
				temperature = val / 10.0
			} else if val, ok := status.Value.(int); ok { // unlikely in unmarshaled json but possible
				temperature = float64(val) / 10.0
			}
		case "va_humidity":
			if val, ok := status.Value.(float64); ok {
				humidity = int(val)
			}
		case "battery_percentage":
			if val, ok := status.Value.(float64); ok {
				battery = int(val)
			}
		}
	}

	// Determine status text
	var tempStatus string
	if temperature > 28.0 {
		tempStatus = "Temperature hot"
	} else if temperature < 18.0 {
		tempStatus = "Temperature cold"
	} else {
		tempStatus = "Temperature comfortable"
	}

	var humidStatus string
	if humidity > 60 {
		humidStatus = "Air moist"
	} else if humidity < 30 {
		humidStatus = "Air dry"
	} else {
		humidStatus = "Air comfortable"
	}

	statusText := fmt.Sprintf("%s, %s", tempStatus, humidStatus)

	response := &dtos.SensorDataDTO{
		Temperature:       temperature,
		Humidity:          humidity,
		BatteryPercentage: battery,
		StatusText:        statusText,
		TempUnit:          "°C", // Defaulting as per plan
	}

	return response, nil
}