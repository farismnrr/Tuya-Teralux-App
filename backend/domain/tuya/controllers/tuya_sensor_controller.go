package controllers

import (
	"net/http"
	"teralux_app/domain/common/dtos"
	"teralux_app/domain/tuya/usecases"
	"teralux_app/domain/common/utils"

	"github.com/gin-gonic/gin"
)

// TuyaSensorController handles sensor data requests
type TuyaSensorController struct {
	useCase *usecases.TuyaSensorUseCase
}

// NewTuyaSensorController creates a new TuyaSensorController
func NewTuyaSensorController(useCase *usecases.TuyaSensorUseCase) *TuyaSensorController {
	return &TuyaSensorController{
		useCase: useCase,
	}
}

// GetSensorData handles GET /api/tuya/devices/:id/sensor endpoint
// @Summary      Get Sensor Data
// @Description  Retrieves sensor data (temperature, humidity, etc.) for a specific device
// @Tags         04. Device Sensor
// @Accept       json
// @Produce      json
// @Param        id   path      string                 true  "Device ID"
// @Success      200  {object}  dtos.StandardResponse{data=dtos.SensorDataDTO}
// @Failure      400  {object}  dtos.StandardResponse
// @Failure      500  {object}  dtos.StandardResponse
// @Security     BearerAuth
// @Router       /api/tuya/devices/{id}/sensor [get]
func (c *TuyaSensorController) GetSensorData(ctx *gin.Context) {
	deviceID := ctx.Param("id")
	if deviceID == "" {
		ctx.JSON(http.StatusBadRequest, dtos.StandardResponse{
			Status:  false,
			Message: "device ID is required",
			Data:    nil,
		})
		return
	}

	accessToken := ctx.MustGet("access_token").(string)
	
	utils.LogDebug("GetSensorData: requesting for device %s", deviceID)

	data, err := c.useCase.GetSensorData(accessToken, deviceID)
	if err != nil {
		utils.LogError("GetSensorData failed: %v", err)
		ctx.JSON(http.StatusInternalServerError, dtos.StandardResponse{
			Status:  false,
			Message: err.Error(),
			Data:    nil,
		})
		return
	}

	ctx.JSON(http.StatusOK, dtos.StandardResponse{
		Status:  true,
		Message: "Sensor data fetched successfully",
		Data:    data,
	})
}