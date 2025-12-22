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

// TuyaGetDeviceByIDUseCase retrieves detailed information for a specific device.
type TuyaGetDeviceByIDUseCase struct {
	service *services.TuyaDeviceService
}

// NewTuyaGetDeviceByIDUseCase initializes a new TuyaGetDeviceByIDUseCase.
//
// param service The TuyaDeviceService used regarding API requests.
// return *TuyaGetDeviceByIDUseCase A pointer to the initialized usecase.
func NewTuyaGetDeviceByIDUseCase(service *services.TuyaDeviceService) *TuyaGetDeviceByIDUseCase {
	return &TuyaGetDeviceByIDUseCase{
		service: service,
	}
}

// GetDeviceByID fetches the details of a single device from the Tuya API.
//
// Tuya API Documentation (Get Device):
// URL: https://openapi.tuyacn.com/v1.0/devices/{device_id}
// Method: GET
//
// param accessToken The valid OAuth 2.0 access token.
// param deviceID The unique ID of the device to fetch.
// return *dtos.TuyaDeviceDTO The detailed device information object.
// return error An error if the request fails.
// @throws error If the API returns a failure response.
func (uc *TuyaGetDeviceByIDUseCase) GetDeviceByID(accessToken, deviceID string) (*dtos.TuyaDeviceDTO, error) {
	// Get config
	config := utils.GetConfig()

	// Generate timestamp in milliseconds
	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)
	signMethod := "HMAC-SHA256"

	// Build URL path - using /v1.0/devices/{device_id} endpoint
	urlPath := fmt.Sprintf("/v1.0/devices/%s", deviceID)
	fullURL := config.TuyaBaseURL + urlPath

	// Calculate content hash (empty for GET request)
	emptyContent := ""
	h := sha256.New()
	h.Write([]byte(emptyContent))
	contentHash := hex.EncodeToString(h.Sum(nil))

	// Generate string to sign
	stringToSign := utils.GenerateTuyaStringToSign("GET", contentHash, "", urlPath)

	utils.LogDebug("GetDeviceByID: generating signature for device=%s", deviceID)

	// Generate signature
	signature := utils.GenerateTuyaSignature(config.TuyaClientID, config.TuyaClientSecret, accessToken, timestamp, stringToSign)

	// Prepare headers with access token
	headers := map[string]string{
		"client_id":    config.TuyaClientID,
		"sign":         signature,
		"t":            timestamp,
		"sign_method":  signMethod,
		"access_token": accessToken,
	}

	// Call service to fetch device
	deviceResponse, err := uc.service.FetchDeviceByID(fullURL, headers)
	if err != nil {
		return nil, err
	}

	// Validate response
	if !deviceResponse.Success {
		return nil, fmt.Errorf("tuya API failed to fetch device: %s (code: %d)", deviceResponse.Msg, deviceResponse.Code)
	}

	// Transform status
	statusDTOs := make([]dtos.TuyaDeviceStatusDTO, len(deviceResponse.Result.Status))
	for i, status := range deviceResponse.Result.Status {
		statusDTOs[i] = dtos.TuyaDeviceStatusDTO{
			Code:  status.Code,
			Value: status.Value,
		}
	}

	// Transform entity to DTO
	dto := &dtos.TuyaDeviceDTO{
		ID:          deviceResponse.Result.ID,
		Name:        deviceResponse.Result.Name,
		Category:    deviceResponse.Result.Category,
		ProductName: deviceResponse.Result.ProductName,
		Online:      deviceResponse.Result.Online,
		Icon:        deviceResponse.Result.Icon,
		Status:      statusDTOs,
		CustomName:  deviceResponse.Result.CustomName,
		Model:       deviceResponse.Result.Model,
		IP:          deviceResponse.Result.IP,
		LocalKey:    deviceResponse.Result.LocalKey,
		CreateTime:  deviceResponse.Result.CreateTime,
		UpdateTime:  deviceResponse.Result.UpdateTime,
	}

	return dto, nil
}
