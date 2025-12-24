package routes

import (
	"teralux_app/domain/tuya/controllers"
	"teralux_app/domain/common/utils"

	"github.com/gin-gonic/gin"
)

// SetupTuyaDeviceRoutes registers endpoints for retrieving device information.
// This includes listing all devices, fetching details for a single device, and getting sensor data.
//
// param router The Gin router interface.
// param getAllDevicesController Controller for listing all devices.
// param getDeviceByIDController Controller for fetching a single device by ID.
// param sensorController Controller for retrieving sensor status.
func SetupTuyaDeviceRoutes(
	router gin.IRouter,
	getAllDevicesController *controllers.TuyaGetAllDevicesController,
	getDeviceByIDController *controllers.TuyaGetDeviceByIDController,
	sensorController *controllers.TuyaSensorController,
) {
	utils.LogDebug("SetupTuyaDeviceRoutes initialized")
	api := router.Group("/api/tuya")
	{
		// GET /api/tuya/devices
		// Retrieves a list of all devices associated with the user account.
		api.GET("/devices", getAllDevicesController.GetAllDevices)

		// GET /api/tuya/devices/:id
		// Retrieves detailed information for a specific device identified by ID.
		api.GET("/devices/:id", getDeviceByIDController.GetDeviceByID)

		// GET /api/tuya/devices/:id/sensor
		// Retrieves formatted sensor data (temperature, humidity) for a specific device.
		api.GET("/devices/:id/sensor", sensorController.GetSensorData)
	}
}