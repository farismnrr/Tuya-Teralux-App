package controllers

import (
	"fmt"
	"net/http"
	"teralux_app/dtos"
	"teralux_app/usecases"

	"github.com/gin-gonic/gin"
)

// TuyaSensorController handles sensor data requests
type TuyaSensorController struct {
	useCase *usecases.TuyaGetDeviceByIDUseCase
}

// NewTuyaSensorController creates a new TuyaSensorController
func NewTuyaSensorController(useCase *usecases.TuyaGetDeviceByIDUseCase) *TuyaSensorController {
	return &TuyaSensorController{
		useCase: useCase,
	}
}

// SensorDataResponse represents the formatted sensor data
type SensorDataResponse struct {
	Temperature       float64 `json:"temperature"`
	Humidity          int     `json:"humidity"`
	BatteryPercentage int     `json:"battery_percentage"`
	StatusText        string  `json:"status_text"`
	TempUnit          string  `json:"temp_unit"`
}

// GetSensorData handles GET /api/tuya/devices/:id/sensor endpoint
func (c *TuyaSensorController) GetSensorData(ctx *gin.Context) {
	deviceID := ctx.Param("id")
	if deviceID == "" {
		ctx.JSON(http.StatusBadRequest, dtos.StandardResponse{
			Status:  false,
			Message: "device ID is required",
			Data:    nil,
		})
		return
	}

	accessToken := ctx.MustGet("access_token").(string)

	device, err := c.useCase.GetDeviceByID(accessToken, deviceID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dtos.StandardResponse{
			Status:  false,
			Message: err.Error(),
			Data:    nil,
		})
		return
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

	response := SensorDataResponse{
		Temperature:       temperature,
		Humidity:          humidity,
		BatteryPercentage: battery,
		StatusText:        statusText,
		TempUnit:          "°C", // Defaulting as per plan
	}

	ctx.JSON(http.StatusOK, dtos.StandardResponse{
		Status:  true,
		Message: "Sensor data fetched successfully",
		Data:    response,
	})
}
