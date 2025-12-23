package routes

import (
	"teralux_app/controllers"

	"github.com/gin-gonic/gin"
)

// SetupDeviceStateRoutes registers all device state-related routes.
//
// param router The Gin router group to register routes on.
// param controller The DeviceStateController handling the requests.
func SetupDeviceStateRoutes(router *gin.RouterGroup, controller *controllers.DeviceStateController) {
	deviceStateGroup := router.Group("/api/devices")
	{
		deviceStateGroup.POST("/:id/state", controller.SaveDeviceState)
		deviceStateGroup.GET("/:id/state", controller.GetDeviceState)
	}
}
