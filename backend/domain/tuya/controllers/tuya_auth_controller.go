package controllers

import (
	"net/http"
	"teralux_app/domain/common/dtos"
	tuya_dtos "teralux_app/domain/tuya/dtos"
	"teralux_app/domain/tuya/usecases"
	"teralux_app/domain/common/utils"

	"github.com/gin-gonic/gin"
)

// Force import for Swagger
var _ = tuya_dtos.TuyaAuthResponseDTO{}

// TuyaAuthController handles authentication requests for Tuya
type TuyaAuthController struct {
	useCase *usecases.TuyaAuthUseCase
}

// NewTuyaAuthController creates a new TuyaAuthController instance
func NewTuyaAuthController(useCase *usecases.TuyaAuthUseCase) *TuyaAuthController {
	return &TuyaAuthController{
		useCase: useCase,
	}
}

// Authenticate handles POST /api/tuya/auth endpoint
// @Summary      Authenticate with Tuya
// @Description  Authenticates the user and retrieves a Tuya access token
// @Tags         01. Auth
// @Accept       json
// @Produce      json
// @Success      200  {object}  dtos.StandardResponse{data=tuya_dtos.TuyaAuthResponseDTO}
// @Failure      500  {object}  dtos.StandardResponse
// @Security     ApiKeyAuth
// @Router       /api/tuya/auth [get]
func (c *TuyaAuthController) Authenticate(ctx *gin.Context) {
	utils.LogDebug("Authenticate request received")
	token, err := c.useCase.Authenticate()																																																																									
	if err != nil {
		utils.LogError("Authenticate failed: %v", err)
		ctx.JSON(http.StatusInternalServerError, dtos.StandardResponse{
			Status:  false,
			Message: err.Error(),
			Data:    nil,
		})
		return
	}

	utils.LogDebug("Authentication successful")
	ctx.JSON(http.StatusOK, dtos.StandardResponse{
		Status:  true,
		Message: "Authentication successful",
		Data:    token,
	})
}