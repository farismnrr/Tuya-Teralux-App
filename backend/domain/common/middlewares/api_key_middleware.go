package middlewares

import (
	"net/http"
	"teralux_app/domain/common/dtos"
	"teralux_app/domain/common/utils"

	"github.com/gin-gonic/gin"
)

// ApiKeyMiddleware validates the presence and correctness of the X-API-KEY header.
// It ensures that only clients with the correct API key can access the protected endpoints.
//
// @return gin.HandlerFunc The Gin middleware handler.
// @throws 500 If the server API key configuration is missing.
// @throws 401 If the provided API key is invalid or missing.
func ApiKeyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-KEY")
		config := utils.GetConfig()
		validApiKey := config.ApiKey

		if validApiKey == "" {
			utils.LogError("ApiKeyMiddleware: API_KEY is not set in server config!")
			c.JSON(http.StatusInternalServerError, dtos.StandardResponse{
				Status:  false,
				Message: "Server misconfiguration: API_KEY not set",
				Data:    nil,
			})
			c.Abort()
			return
		}

		if apiKey != validApiKey {
			utils.LogWarn("ApiKeyMiddleware: Invalid API Key provided")
			c.JSON(http.StatusUnauthorized, dtos.StandardResponse{
				Status:  false,
				Message: "Invalid API Key",
				Data:    nil,
			})
			c.Abort()
			return
		}
		
		utils.LogDebug("ApiKeyMiddleware: Valid API Key")

		c.Next()
	}
}