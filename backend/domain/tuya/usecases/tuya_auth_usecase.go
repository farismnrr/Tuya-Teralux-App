package usecases

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"teralux_app/domain/tuya/dtos"
	"teralux_app/domain/tuya/services"
	"teralux_app/domain/common/utils"
	tuya_utils "teralux_app/domain/tuya/utils"
	"time"
)

// TuyaAuthUseCase handles the core business logic for Tuya API authentication.
// It orchestrates signature generation, timestamp creation, and service interaction.
type TuyaAuthUseCase struct {
	service *services.TuyaAuthService
}

// NewTuyaAuthUseCase creates a new instance of TuyaAuthUseCase.
//
// param service The TuyaAuthService used to perform the actual HTTP requests.
// return *TuyaAuthUseCase A pointer to the initialized usecase.
func NewTuyaAuthUseCase(service *services.TuyaAuthService) *TuyaAuthUseCase {
	return &TuyaAuthUseCase{
		service: service,
	}
}

// Authenticate performs the full authentication flow to retrieve an access token.
// It handles signature generation (HMAC-SHA256), timestamp creation, and header preparation.
//
// Tuya API Documentation (Get Token):
// URL: https://openapi.tuyacn.com/v1.0/token?grant_type=1
// Method: GET
//
// StringToSign Format:
//   GET\n{content_hash}\n\n{url}
//   (content_hash is SHA256 of empty string for GET)
//
// return *dtos.TuyaAuthResponseDTO The data transfer object containing the access token, refresh token, and expiration time.
// return error An error if configuration is missing, signature generation fails, or the API call returns an error.
// @throws error if the API returns a non-success status code (e.g., invalid client ID).
func (uc *TuyaAuthUseCase) Authenticate() (*dtos.TuyaAuthResponseDTO, error) {
	// Get config
	config := utils.GetConfig()

	// Generate timestamp in milliseconds
	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)
	signMethod := "HMAC-SHA256"

	// Build URL path
	urlPath := "/v1.0/token?grant_type=1"
	fullURL := config.TuyaBaseURL + urlPath

	// Calculate content hash (empty for GET request)
	emptyContent := ""
	h := sha256.New()
	h.Write([]byte(emptyContent))
	contentHash := hex.EncodeToString(h.Sum(nil))

	// Generate string to sign
	stringToSign := tuya_utils.GenerateTuyaStringToSign("GET", contentHash, "", urlPath)
	
	utils.LogDebug("Authenticate: generating signature for clientId=%s", config.TuyaClientID)

	// Generate signature
	signature := tuya_utils.GenerateTuyaSignature(config.TuyaClientID, config.TuyaClientSecret, "", timestamp, stringToSign)

	// Prepare headers
	headers := map[string]string{
		"client_id":   config.TuyaClientID,
		"sign":        signature,
		"t":           timestamp,
		"sign_method": signMethod,
	}

	// Call service to fetch token
	authResponse, err := uc.service.FetchToken(fullURL, headers)
	if err != nil {
		return nil, err
	}

	// Validate response
	if !authResponse.Success {
		return nil, fmt.Errorf("tuya API authentication failed: %s (code: %d)", authResponse.Msg, authResponse.Code)
	}

	// Transform entity to DTO
	dto := &dtos.TuyaAuthResponseDTO{
		AccessToken:  authResponse.Result.AccessToken,
		ExpireTime:   authResponse.Result.ExpireTime,
		RefreshToken: authResponse.Result.RefreshToken,
		UID:          authResponse.Result.UID,
	}

	// Override UID if provided in config (for testing with specific user)
	if config.TuyaUserID != "" {
		dto.UID = config.TuyaUserID
	}

	return dto, nil
}