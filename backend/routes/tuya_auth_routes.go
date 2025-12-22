package routes

import (
	"teralux_app/controllers"
	"teralux_app/utils"

	"github.com/gin-gonic/gin"
)

// SetupTuyaAuthRoutes registers authentication-related endpoints for Tuya.
//
// param router The Gin router group to attach routes to.
// param controller The handler controller for authentication logic.
func SetupTuyaAuthRoutes(router *gin.RouterGroup, controller *controllers.TuyaAuthController) {
	utils.LogDebug("SetupTuyaAuthRoutes initialized")
	api := router.Group("/api/tuya")
	{
		// GET /api/tuya/auth
		// Initiates the Tuya authentication process to retrieve an access token.
		api.GET("/auth", controller.Authenticate)
	}
}
