package controllers

import (
	"net/http"
	"teralux_app/dtos"
	"teralux_app/usecases"
	"teralux_app/utils"

	"github.com/gin-gonic/gin"
)

// DeviceStateController handles HTTP requests for device state management.
type DeviceStateController struct {
	usecase *usecases.DeviceStateUseCase
}

// NewDeviceStateController initializes a new DeviceStateController.
//
// param usecase The DeviceStateUseCase for business logic.
// return *DeviceStateController A pointer to the initialized controller.
func NewDeviceStateController(usecase *usecases.DeviceStateUseCase) *DeviceStateController {
	return &DeviceStateController{
		usecase: usecase,
	}
}

// SaveDeviceState handles POST /api/devices/:id/state
// Saves the last control state for a device.
//
// @Summary Save device state
// @Description Saves the last control state for a device to persistent storage
// @Tags Device State
// @Accept json
// @Produce json
// @Param id path string true "Device ID"
// @Param request body dtos.SaveDeviceStateRequestDTO true "State commands"
// @Success 200 {object} dtos.StandardResponse "State saved successfully"
// @Failure 400 {object} dtos.StandardResponse "Invalid request"
// @Failure 500 {object} dtos.StandardResponse "Internal server error"
// @Security BearerAuth
// @Router /api/devices/{id}/state [post]
func (ctrl *DeviceStateController) SaveDeviceState(c *gin.Context) {
	deviceID := c.Param("id")
	
	var req dtos.SaveDeviceStateRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.LogError("SaveDeviceState: Invalid request body: %v", err)
		c.JSON(http.StatusBadRequest, dtos.StandardResponse{
			Status:  false,
			Message: "Invalid request body",
			Data:    nil,
		})
		return
	}

	utils.LogDebug("SaveDeviceState: Received request for device %s with %d commands", deviceID, len(req.Commands))
	for i, cmd := range req.Commands {
		utils.LogDebug("  Command[%d]: code=%s, value=%v (type=%T)", i, cmd.Code, cmd.Value, cmd.Value)
	}

	// Save state
	if err := ctrl.usecase.SaveDeviceState(deviceID, req.Commands); err != nil {
		utils.LogError("SaveDeviceState: Failed to save state for device %s: %v", deviceID, err)
		c.JSON(http.StatusInternalServerError, dtos.StandardResponse{
			Status:  false,
			Message: "Failed to save device state",
			Data:    nil,
		})
		return
	}

	c.JSON(http.StatusOK, dtos.StandardResponse{
		Status:  true,
		Message: "Device state saved successfully",
		Data:    nil,
	})
}

// GetDeviceState handles GET /api/devices/:id/state
// Retrieves the last known control state for a device.
//
// @Summary Get device state
// @Description Retrieves the last known control state for a device
// @Tags Device State
// @Accept json
// @Produce json
// @Param id path string true "Device ID"
// @Success 200 {object} dtos.StandardResponse{data=dtos.DeviceStateDTO} "State retrieved successfully"
// @Success 404 {object} dtos.StandardResponse "State not found"
// @Failure 500 {object} dtos.StandardResponse "Internal server error"
// @Security BearerAuth
// @Router /api/devices/{id}/state [get]
func (ctrl *DeviceStateController) GetDeviceState(c *gin.Context) {
	deviceID := c.Param("id")

	// Get state
	state, err := ctrl.usecase.GetDeviceState(deviceID)
	if err != nil {
		utils.LogError("GetDeviceState: Failed to get state for device %s: %v", deviceID, err)
		c.JSON(http.StatusInternalServerError, dtos.StandardResponse{
			Status:  false,
			Message: "Failed to retrieve device state",
			Data:    nil,
		})
		return
	}

	// Not found
	if state == nil {
		utils.LogDebug("GetDeviceState: No state found for device %s", deviceID)
		c.JSON(http.StatusOK, dtos.StandardResponse{
			Status:  true,
			Message: "No state found for device",
			Data:    nil,
		})
		return
	}

	utils.LogDebug("GetDeviceState: Retrieved state for device %s with %d commands", deviceID, len(state.LastCommands))
	for i, cmd := range state.LastCommands {
		utils.LogDebug("  Command[%d]: code=%s, value=%v (type=%T)", i, cmd.Code, cmd.Value, cmd.Value)
	}

	c.JSON(http.StatusOK, dtos.StandardResponse{
		Status:  true,
		Message: "Device state retrieved successfully",
		Data:    state,
	})
}
