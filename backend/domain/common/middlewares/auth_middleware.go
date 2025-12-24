package middlewares

import (
	"net/http"
	"strings"
	"teralux_app/domain/common/dtos"
	"teralux_app/domain/common/utils"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware processes the Authorization header to extract the Bearer token.
// It also optionally parses the "X-TUYA-UID" header and stores it in the context.
//
// @return gin.HandlerFunc The Gin middleware handler.
// @throws 401 If the Authorization header is missing or malformed.
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		utils.LogDebug("AuthMiddleware: processing request")
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
	
		tuyaUID := c.GetHeader("X-TUYA-UID") 
		if tuyaUID != "" {
			c.Set("tuya_uid", tuyaUID)
		}

		c.Next()
	}
}