package controllers

import (
	"net/http"
	"teralux_app/domain/common/dtos"
	"teralux_app/domain/common/infrastructure/persistence"
	"teralux_app/domain/common/utils"

	"github.com/gin-gonic/gin"
)

// CacheController handles cache-related operations
type CacheController struct {
	cache *persistence.BadgerService
}

// NewCacheController creates a new CacheController instance
func NewCacheController(cache *persistence.BadgerService) *CacheController {
	return &CacheController{cache: cache}
}

// FlushCache clears the entire cache
// @Summary Flush all cache
// @Description Remove all data from the cache storage
// @Tags 05. Flush
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dtos.StandardResponse
// @Failure 500 {object} dtos.StandardResponse
// @Router /api/cache/flush [delete]
func (ctrl *CacheController) FlushCache(c *gin.Context) {
	if ctrl.cache == nil {
		c.JSON(http.StatusInternalServerError, dtos.StandardResponse{
			Status:  false,
			Message: "Cache service not initialized",
			Data:    nil,
		})
		return
	}

	err := ctrl.cache.FlushAll()
	if err != nil {
		utils.LogError("Failed to flush cache: %v", err)
		c.JSON(http.StatusInternalServerError, dtos.StandardResponse{
			Status:  false,
			Message: "Failed to flush cache",
			Data:    nil,
		})
		return
	}

	c.JSON(http.StatusOK, dtos.StandardResponse{
		Status:  true,
		Message: "Cache flushed successfully",
		Data:    nil,
	})
}