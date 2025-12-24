package controllers

import (
	"net/http"
	"strconv"
	"teralux_app/domain/common/dtos"
	tuya_dtos "teralux_app/domain/tuya/dtos"
	"teralux_app/domain/tuya/usecases"
	"teralux_app/domain/common/utils"

	"github.com/gin-gonic/gin"
)

// Force import for Swagger
var _ = tuya_dtos.TuyaDevicesResponseDTO{}

// TuyaGetAllDevicesController handles get all devices requests for Tuya
type TuyaGetAllDevicesController struct {
	useCase *usecases.TuyaGetAllDevicesUseCase
}

// NewTuyaGetAllDevicesController creates a new TuyaGetAllDevicesController instance
func NewTuyaGetAllDevicesController(useCase *usecases.TuyaGetAllDevicesUseCase) *TuyaGetAllDevicesController {
	return &TuyaGetAllDevicesController{
		useCase: useCase,
	}
}

// GetAllDevices handles GET /api/tuya/devices endpoint
// @Summary      Get All Devices
// @Description  Retrieves a list of all devices. Response format depends on GET_ALL_DEVICES_RESPONSE_TYPE: 0 (Nested/Default), 1 (Flat), 2 (Merged). Sorted alphabetically by Name. For infrared_ac devices, the status array is populated with saved device state (power, temp, mode, wind) or default values if no state exists.
// @Tags         02. Devices
// @Accept       json
// @Produce      json
// @Param        page      query  int     false  "Page number"
// @Param        limit     query  int     false  "Items per page"
// @Param        category  query  string  false  "Filter by category"
// @Success      200  {object}  dtos.StandardResponse{data=tuya_dtos.TuyaDevicesResponseDTO}
// @Failure      500  {object}  dtos.StandardResponse
// @Security     BearerAuth
// @Router       /api/tuya/devices [get]
func (c *TuyaGetAllDevicesController) GetAllDevices(ctx *gin.Context) {
	accessToken := ctx.MustGet("access_token").(string)

	uid := utils.AppConfig.TuyaUserID
	if uid == "" {
		utils.LogError("TUYA_USER_ID is not set in environment")
		ctx.JSON(http.StatusInternalServerError, dtos.StandardResponse{
			Status:  false,
			Message: "Server configuration error: TUYA_USER_ID missing",
			Data:    nil,
		})
		return
	}
	utils.LogDebug("Using TUYA_USER_ID from env: '%s'", uid)

	// Parse optional query parameters
	pageStr := ctx.Query("page")
	limitStr := ctx.Query("limit")
	category := ctx.Query("category")

	page := 0
	limit := 0
	var errConv error

	if pageStr != "" {
		page, errConv = strconv.Atoi(pageStr)
		if errConv != nil {
			utils.LogWarn("Invalid page parameter: %v", errConv)
			page = 0 // Default to 0 (ignored)
		}
	}

	if limitStr != "" {
		limit, errConv = strconv.Atoi(limitStr)
		if errConv != nil {
			utils.LogWarn("Invalid limit parameter: %v", errConv)
			limit = 0 // Default to 0 (ignored)
		}
	}

	devices, err := c.useCase.GetAllDevices(accessToken, uid, page, limit, category)
	if err != nil {
		utils.LogError("Error fetching devices: %v", err)
		ctx.JSON(http.StatusInternalServerError, dtos.StandardResponse{
			Status:  false,
			Message: err.Error(),
			Data:    nil,
		})
		return
	}

	ctx.JSON(http.StatusOK, dtos.StandardResponse{
		Status:  true,
		Message: "Devices fetched successfully",
		Data:    devices,
	})
}