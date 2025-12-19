package usecases

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"teralux_app/dtos"
	"teralux_app/services"
	"teralux_app/utils"
	"time"
)

// TuyaAuthUseCase handles authentication business logic for Tuya
type TuyaAuthUseCase struct {
	service *services.TuyaAuthService
}

// NewTuyaAuthUseCase creates a new TuyaAuthUseCase instance
func NewTuyaAuthUseCase(service *services.TuyaAuthService) *TuyaAuthUseCase {
	return &TuyaAuthUseCase{
		service: service,
	}
}

// Authenticate retrieves access token from Tuya API
//
// Tuya API Documentation (Get Token):
// URL: https://openapi.tuyacn.com/v1.0/token?grant_type=1
// Method: GET
//
// Headers:
//   - client_id: Your Tuya Client ID
//   - sign: HMAC-SHA256(client_id + t + stringToSign)
//   - t: timestamp (ms)
//   - sign_method: HMAC-SHA256
//
// StringToSign Format:
//   GET\n{content_hash}\n\n{url}
//   (content_hash is SHA256 of empty string for GET)
//
// Response:
//   {
//     "success": true,
//     "result": {
//       "access_token": "...",
//       "refresh_token": "...",
//       "expire_time": 7200,
//       "uid": "..."
//     }
//   }
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
	stringToSign := utils.GenerateTuyaStringToSign("GET", contentHash, "", urlPath)
	
	utils.LogDebug("Authenticate: generating signature for clientId=%s", config.TuyaClientID)

	// Generate signature
	signature := utils.GenerateTuyaSignature(config.TuyaClientID, config.TuyaClientSecret, "", timestamp, stringToSign)

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
