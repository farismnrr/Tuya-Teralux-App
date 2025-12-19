package routes

import (
	"teralux_app/controllers"
	"teralux_app/utils"

	"github.com/gin-gonic/gin"
)

// SetupTuyaControlRoutes registers Tuya device control routes
func SetupTuyaControlRoutes(router gin.IRouter, controller *controllers.TuyaDeviceControlController) {
	utils.LogDebug("SetupTuyaControlRoutes initialized")
	api := router.Group("/api/tuya")
	{
		// Send commands to device (Switch/Light/etc)
		// URL: /api/tuya/devices/:id/commands/switch
		// Method: POST
		// Headers:																																															
		//    - Authorization: Bearer <token>
		// Param: id (Infrared/Gateway ID)
		// Body: {
		//    "code": "...",
		//    "value": true																																																																																																																																																																																																																																																																																																																														
		// }
		api.POST("/devices/:id/commands/switch", controller.SendCommand)

		// Send IR AC command
		// URL: /api/tuya/devices/:id/commands/ir
		// Method: POST
		// Headers:																																															
		//    - Authorization: Bearer <token>
		// Param: id (Infrared/Gateway ID)
		// Body: {																																																																			
		//    "remote_id": "...",
		//    "code": "...",
		//    "value": 1
		// }
		api.POST("/devices/:id/commands/ir", controller.SendIRACCommand)
	}
}
