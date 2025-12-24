package controllers

import (
	"net/http"
	"teralux_app/domain/common/dtos"
	tuya_dtos "teralux_app/domain/tuya/dtos"
	"teralux_app/domain/tuya/usecases"
	"teralux_app/domain/common/utils"

	"github.com/gin-gonic/gin"
)

// TuyaGetDeviceByIDController handles get device by ID requests for Tuya
type TuyaGetDeviceByIDController struct {
	useCase *usecases.TuyaGetDeviceByIDUseCase
}

// NewTuyaGetDeviceByIDController creates a new TuyaGetDeviceByIDController instance
func NewTuyaGetDeviceByIDController(useCase *usecases.TuyaGetDeviceByIDUseCase) *TuyaGetDeviceByIDController {
	return &TuyaGetDeviceByIDController{
		useCase: useCase,
	}
}

// GetDeviceByID handles GET /api/tuya/devices/:id endpoint
// @Summary      Get Device by ID
// @Description  Retrieves details of a specific device by its ID. Response includes last_commands field containing the last control commands sent to the device.
// @Tags         02. Devices
// @Accept       json
// @Produce      json
// @Param        id   path      string                 true  "Device ID"
// @Success      200  {object}  dtos.StandardResponse{data=tuya_dtos.TuyaDeviceResponseDTO}
// @Failure      400  {object}  dtos.StandardResponse
// @Failure      500  {object}  dtos.StandardResponse
// @Security     BearerAuth
// @Router       /api/tuya/devices/{id} [get]
func (c *TuyaGetDeviceByIDController) GetDeviceByID(ctx *gin.Context) {
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
	utils.LogDebug("GetDeviceByID: requesting device %s", deviceID)
	device, err := c.useCase.GetDeviceByID(accessToken, deviceID)
	if err != nil {
		utils.LogError("GetDeviceByID failed: %v", err)
		ctx.JSON(http.StatusInternalServerError, dtos.StandardResponse{
			Status:  false,
			Message: err.Error(),
			Data:    nil,
		})
		return
	}

	utils.LogDebug("GetDeviceByID success")
	ctx.JSON(http.StatusOK, dtos.StandardResponse{
		Status:  true,
		Message: "Device fetched successfully",
		Data:    tuya_dtos.TuyaDeviceResponseDTO{Device: *device},
	})
}