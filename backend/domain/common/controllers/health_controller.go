package controllers

import (
	"net/http"
	"teralux_app/domain/common/infrastructure"
	"github.com/gin-gonic/gin"
)

// HealthController handles health check requests
type HealthController struct{}

// NewHealthController creates a new HealthController instance
func NewHealthController() *HealthController {
	return &HealthController{}
}


// CheckHealth godoc
// @Summary      Health check endpoint
// @Description  Check if the application and database are healthy
// @Tags         Health
// @Produce      plain
// @Success      200  {string}  string "OK"
// @Failure      503  {string}  string "Service Unavailable"
// @Router       /health [get]
func (h *HealthController) CheckHealth(c *gin.Context) {
	// Check database connection
	if err := infrastructure.PingDB(); err != nil {
		c.String(http.StatusServiceUnavailable, "Service Unavailable")
		return
	}

	c.String(http.StatusOK, "OK")
}