package controllers

import (
	"net/http"
	"teralux_app/dtos"
	"teralux_app/usecases"
	"teralux_app/utils"


	"github.com/gin-gonic/gin"
)

// TuyaDeviceControlController handles device control requests
type TuyaDeviceControlController struct {
	useCase *usecases.TuyaDeviceControlUseCase
}

// NewTuyaDeviceControlController creates a new TuyaDeviceControlController instance
func NewTuyaDeviceControlController(useCase *usecases.TuyaDeviceControlUseCase) *TuyaDeviceControlController {
	return &TuyaDeviceControlController{
		useCase: useCase,
	}
}

// SendCommand handles the request to send commands to a device
// @Summary      Send Command to Device
// @Description  Sends a command to a specific Tuya device
// @Tags         Device Control
// @Accept       json
// @Produce      json
// @Param        id   path      string                 true  "Device ID"
// @Param        command body      dtos.TuyaCommandDTO    true  "Command Payload"
// @Success      200  {object}  dtos.StandardResponse
// @Failure      400  {object}  dtos.StandardResponse
// @Failure      500  {object}  dtos.StandardResponse
// @Security     BearerAuth
// @Router       /api/tuya/devices/{id}/commands/switch [post]
func (ctrl *TuyaDeviceControlController) SendCommand(c *gin.Context) {
	deviceID := c.Param("id")
	accessToken := c.MustGet("access_token").(string)
	utils.LogDebug("SendCommand: received request for device %s", deviceID)

	var req dtos.TuyaCommandDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.LogError("Failed to bind command: %v", err)
		c.JSON(http.StatusBadRequest, dtos.StandardResponse{
			Status:  false,
			Message: err.Error(),
			Data:    nil,
		})
		return
	}

	commands := []dtos.TuyaCommandDTO{req}
	success, err := ctrl.useCase.SendCommand(accessToken, deviceID, commands)
	if err != nil {
		utils.LogError("SendCommand failed: %v", err)
		c.JSON(http.StatusInternalServerError, dtos.StandardResponse{
			Status:  false,
			Message: err.Error(),
			Data:    nil,
		})
		return
	}

	utils.LogDebug("SendCommand success")
	c.JSON(http.StatusOK, dtos.StandardResponse{
		Status:  true,
		Message: "Command sent successfully",
		Data:    dtos.SuccessResponseDTO{Success: success},
	})
}

// SendIRACCommand handles the request to send a command to an IR air conditioner
// @Summary      Send IR AC Command
// @Description  Sends an infrared command to an AC via a specific IR device
// @Tags         Device Control
// @Accept       json
// @Produce      json
// @Param        id   path      string                 true  "Infrared Device ID"
// @Param        command body      dtos.TuyaIRACCommandDTO true  "IR AC Command Payload"
// @Success      200  {object}  dtos.StandardResponse
// @Failure      400  {object}  dtos.StandardResponse
// @Failure      500  {object}  dtos.StandardResponse
// @Security     BearerAuth
// @Router       /api/tuya/devices/{id}/commands/ir [post]
func (ctrl *TuyaDeviceControlController) SendIRACCommand(c *gin.Context) {
	accessToken := c.MustGet("access_token").(string)

	var req dtos.TuyaIRACCommandDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.LogError("Failed to bind IR AC command: %v", err)
		c.JSON(http.StatusBadRequest, dtos.StandardResponse{
			Status:  false,
			Message: err.Error(),
			Data:    nil,
		})
		return
	}

	infraredID := c.Param("id")
	utils.LogDebug("SendIRACCommand: sending to %s, remoteID: %s, code: %s", infraredID, req.RemoteID, req.Code)

	success, err := ctrl.useCase.SendIRACCommand(accessToken, infraredID, req.RemoteID, req.Code, req.Value)
	if err != nil {
		utils.LogError("SendIRACCommand failed: %v", err)
		c.JSON(http.StatusInternalServerError, dtos.StandardResponse{
			Status:  false,
			Message: err.Error(),
			Data:    nil,
		})
		return
	}

	utils.LogDebug("SendIRACCommand success")
	c.JSON(http.StatusOK, dtos.StandardResponse{
		Status:  true,
		Message: "IR AC Command sent successfully",
		Data:    dtos.SuccessResponseDTO{Success: success},
	})
}
