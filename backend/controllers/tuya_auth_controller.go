package controllers

import (
	"net/http"
	"teralux_app/dtos"
	"teralux_app/usecases"
	"teralux_app/utils"

	"github.com/gin-gonic/gin"
)

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
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Success      200  {object}  dtos.StandardResponse{data=dtos.TuyaAuthResponseDTO}
// @Failure      500  {object}  dtos.StandardResponse
// @Security     ApiKeyAuth
// @Router       /api/tuya/auth [get]
func (c *TuyaAuthController) Authenticate(ctx *gin.Context) {
	utils.LogDebug("Authenticate request received")
	// Call use case
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

	// Return success response
	utils.LogDebug("Authentication successful")
	ctx.JSON(http.StatusOK, dtos.StandardResponse{
		Status:  true,
		Message: "Authentication successful",
		Data:    token,
	})
}
