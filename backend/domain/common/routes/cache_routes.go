package routes

import (
	"teralux_app/domain/common/controllers"

	"github.com/gin-gonic/gin"
)

// SetupCacheRoutes registers endpoints for cache management.
//
// param rg The router group to attach the cache routes to.
// param controller The controller handling cache operations.
func SetupCacheRoutes(rg *gin.RouterGroup, controller *controllers.CacheController) {
	cacheGroup := rg.Group("/api/cache")
	{
		// DELETE /api/cache/flush
		// Clears all data from the application cache (BadgerDB).
		cacheGroup.DELETE("/flush", controller.FlushCache)
	}
}