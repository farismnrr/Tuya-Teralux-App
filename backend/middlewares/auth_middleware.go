package middlewares

import (
	"net/http"
	"strings"
	"teralux_app/dtos"
	"teralux_app/utils"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware handles authentication and header parsing
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		utils.LogDebug("AuthMiddleware: processing request")
		// 1. Parse Authorization Header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.LogWarn("AuthMiddleware: missing Authorization Header")
			c.JSON(http.StatusUnauthorized, dtos.StandardResponse{
				Status:  false,
				Message: "Authorization header is required",
				Data:    nil,
			})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		var accessToken string
		if len(parts) == 2 && parts[0] == "Bearer" {
			accessToken = parts[1]
		} else if len(parts) == 1 {
			accessToken = parts[0]
		} else {
			utils.LogWarn("AuthMiddleware: invalid Authorization Header format")
			c.JSON(http.StatusUnauthorized, dtos.StandardResponse{
				Status:  false,
				Message: "Invalid Authorization header format. Expected 'Bearer <token>'",
				Data:    nil,
			})
			c.Abort()
			return
		}
		c.Set("access_token", accessToken)
		utils.LogDebug("AuthMiddleware: token parsed successfully")

		// 2. Parse X-TUYA-UID Header (Optional generally, but populated if present)
		// Controllers that strictly require it should check strictness, 
		// but for now we just parse it into context if it exists.
		// However, based on previous code, GetAllDevices STRICTLY required it. 
		// Let's parse it here. If a controller needs it, it can check c.Get("tuya_uid").
		
		tuyaUID := c.GetHeader("X-TUYA-UID") 
		if tuyaUID != "" {
			c.Set("tuya_uid", tuyaUID)
		}

		c.Next()
	}
}
