package routes

import (
	"teralux_app/domain/tuya/controllers"
	"teralux_app/domain/common/utils"

	"github.com/gin-gonic/gin"
)

// SetupTuyaControlRoutes registers endpoints for controlling Tuya devices.
// These routes handle standard device commands (e.g., switches) and specialized IR commands.
//
// param router The Gin router interface.
// param controller The controller responsible for handling device control requests.
func SetupTuyaControlRoutes(router gin.IRouter, controller *controllers.TuyaDeviceControlController) {
	utils.LogDebug("SetupTuyaControlRoutes initialized")
	api := router.Group("/api/tuya")
	{
		// POST /api/tuya/devices/:id/commands/switch
		// Sends a standard command (e.g., toggle power) to a specific device.
		api.POST("/devices/:id/commands/switch", controller.SendCommand)

		// POST /api/tuya/devices/:id/commands/ir
		// Sends an infrared command (e.g., AC control) to an IR-enabled device.
		api.POST("/devices/:id/commands/ir", controller.SendIRACCommand)
	}
}