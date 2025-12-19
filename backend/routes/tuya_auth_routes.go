package routes

import (
	"teralux_app/controllers"
	"teralux_app/utils"

	"github.com/gin-gonic/gin"
)

// SetupTuyaAuthRoutes registers Tuya authentication routes
func SetupTuyaAuthRoutes(router *gin.RouterGroup, controller *controllers.TuyaAuthController) {
	utils.LogDebug("SetupTuyaAuthRoutes initialized")
	api := router.Group("/api/tuya")
	{
		// Authenticate with Tuya
		// URL: /api/tuya/auth
		// Method: GET
		// Headers:
		//    X-API-KEY: <your_api_key>
		// Body: None
		// Response: {
		//    "status": true,
		//    "message": "Authentication successful",
		//    "data": {
		//      "access_token": "...",
		//      "expire_time": 7200,
		//      "refresh_token": "...",
		//      "uid": "..."
		//    }
		// }
		api.GET("/auth", controller.Authenticate)
	}
}
