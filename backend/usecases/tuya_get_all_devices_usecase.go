																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																																										package usecases

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"strconv"
	"teralux_app/dtos"
	"teralux_app/services"
	"teralux_app/utils"
	"time"
)

// TuyaGetAllDevicesUseCase handles get all devices business logic for Tuya
type TuyaGetAllDevicesUseCase struct {
	service *services.TuyaDeviceService
}

// NewTuyaGetAllDevicesUseCase creates a new TuyaGetAllDevicesUseCase instance
func NewTuyaGetAllDevicesUseCase(service *services.TuyaDeviceService) *TuyaGetAllDevicesUseCase {
	return &TuyaGetAllDevicesUseCase{
		service: service,
	}
}

// GetAllDevices retrieves all devices from Tuya API
//
// Tuya API Interactions:
// 
// 1. List Devices by User:
//    URL: https://openapi.tuyacn.com/v1.0/users/{uid}/devices
//    Method: GET
//    Auth: Standard Token Auth
//
// 2. Get Device Specifications (for each device):
//    URL: https://openapi.tuyacn.com/v1.0/iot-03/devices/{device_id}/specification
//    Method: GET
//    Purpose: Retrieves function definitions (DP codes) and status mappings. 
//             Crucial for identifying valid command codes (e.g., switch1 vs switch_1).
//
// 3. Batch Get Device Status:
//    URL: https://openapi.tuyacn.com/v1.0/iot-03/devices/status?device_ids=id1,id2...
//    Method: GET
//    Purpose: Retrieves real-time online/offline status and current DP values for multiple devices.
//
// Response Aggregation:
//    Combines basic device info, static specifications, and dynamic real-time status into a unified DTO.
func (uc *TuyaGetAllDevicesUseCase) GetAllDevices(accessToken, uid string) (*dtos.TuyaDevicesResponseDTO, error) {
	// Get config
	config := utils.GetConfig()

	// Generate timestamp in milliseconds
	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)
	signMethod := "HMAC-SHA256"

	// Build URL path - using /v1.0/users/{uid}/devices endpoint
	urlPath := fmt.Sprintf("/v1.0/users/%s/devices", uid)
	fullURL := config.TuyaBaseURL + urlPath

	// Calculate content hash (empty for GET request)
	emptyContent := ""
	h := sha256.New()
	h.Write([]byte(emptyContent))
	contentHash := hex.EncodeToString(h.Sum(nil))

	// Generate string to sign
	stringToSign := utils.GenerateTuyaStringToSign("GET", contentHash, "", urlPath)

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

	// Call service to fetch devices
	devicesResponse, err := uc.service.FetchDevices(fullURL, headers)
	if err != nil {
		return nil, err
	}

	// Validate response
	if !devicesResponse.Success {
		return nil, fmt.Errorf("tuya API failed to fetch devices: %s (code: %d)", devicesResponse.Msg, devicesResponse.Code)
	}

	// DEBUG: Log device attributes and SPECIFICATIONS to find correct command values
	for _, dev := range devicesResponse.Result {
		log.Printf("DEVICE DEBUG: ID=%s, Name=%s, Category=%s", dev.ID, dev.Name, dev.Category)
		for _, st := range dev.Status {
			log.Printf("   STATUS: Code=%s, Value=%v (Type: %T)", st.Code, st.Value, st.Value)
		}

		// Fetch and Log Specifications
		specTimestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)
		specUrlPath := fmt.Sprintf("/v1.0/iot-03/devices/%s/specification", dev.ID)
		specFullURL := config.TuyaBaseURL + specUrlPath

		specEmptyContent := ""
		hSpec := sha256.New()
		hSpec.Write([]byte(specEmptyContent))
		specContentHash := hex.EncodeToString(hSpec.Sum(nil))

		specStringToSign := utils.GenerateTuyaStringToSign("GET", specContentHash, "", specUrlPath)
		specSignature := utils.GenerateTuyaSignature(config.TuyaClientID, config.TuyaClientSecret, accessToken, specTimestamp, specStringToSign)

		specHeaders := map[string]string{
			"client_id":    config.TuyaClientID,
			"sign":         specSignature,
			"t":            specTimestamp,
			"sign_method":  signMethod,
			"access_token": accessToken,
		}

		specResp, errSpec := uc.service.FetchDeviceSpecification(specFullURL, specHeaders)
		if errSpec == nil && specResp.Success {
			log.Printf("   SPECIFICATION for ID=%s:", dev.ID)
			for _, fn := range specResp.Result.Functions {
				log.Printf("      FUNCTION: Code=%s, Type=%s, Values=%s", fn.Code, fn.Type, fn.Values)
			}
		} else {
			log.Printf("   FAILED to fetch spec for ID=%s: %v", dev.ID, errSpec)
		}
	}

	// Transform entities to DTOs
	var deviceDTOs []dtos.TuyaDeviceDTO
	var deviceIDs []string

	// Collect IDs first
	for _, device := range devicesResponse.Result {
		deviceIDs = append(deviceIDs, device.ID)
	}

	// Fetch Real-time Status Batch
	statusMap := make(map[string]bool)
	if len(deviceIDs) > 0 {
		// New timestamp/signature for status call
		statusTimestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)
		statusURLPath := "/v1.0/iot-03/devices/status"
		statusFullURL := config.TuyaBaseURL + statusURLPath + "?device_ids=" + utils.JoinStrings(deviceIDs, ",")

		statusEmptyContent := ""
		hStatus := sha256.New()
		hStatus.Write([]byte(statusEmptyContent))
		statusContentHash := hex.EncodeToString(hStatus.Sum(nil))

		statusStringToSign := utils.GenerateTuyaStringToSign("GET", statusContentHash, "", statusURLPath)
		statusSignature := utils.GenerateTuyaSignature(config.TuyaClientID, config.TuyaClientSecret, accessToken, statusTimestamp, statusStringToSign)

		statusHeaders := map[string]string{
			"client_id":    config.TuyaClientID,
			"sign":         statusSignature,
			"t":            statusTimestamp,
			"sign_method":  signMethod,
			"access_token": accessToken,
		}

		batchStatusResponse, err := uc.service.FetchBatchDeviceStatus(statusFullURL, statusHeaders)
		if err == nil && batchStatusResponse.Success {
			for _, s := range batchStatusResponse.Result {
				statusMap[s.ID] = s.IsOnline
			}
		} else {
			log.Printf("WARN: Failed to fetch batch status: %v", err)
		}
	}

	for _, device := range devicesResponse.Result {
		// Use real-time status if available, fallback to list status
		isOnline := device.Online
		if val, ok := statusMap[device.ID]; ok {
			isOnline = val
		}

		statusDTOs := make([]dtos.TuyaDeviceStatusDTO, len(device.Status))
		for j, s := range device.Status {
			statusDTOs[j] = dtos.TuyaDeviceStatusDTO{
				Code:  s.Code,
				Value: s.Value,
			}
		}

		deviceDTOs = append(deviceDTOs, dtos.TuyaDeviceDTO{
			ID:          device.ID,
			Name:        device.Name,
			ProductName: device.ProductName,
			Category:    device.Category,
			Icon:        device.Icon,
			Online:      isOnline,
			Status:      statusDTOs,
			CustomName:  device.CustomName,
			Model:       device.Model,
			IP:          device.IP,
			LocalKey:    device.LocalKey,
			GatewayID:   device.GatewayID,
			CreateTime:  device.CreateTime,
			UpdateTime:  device.UpdateTime,
		})
	}

	return &dtos.TuyaDevicesResponseDTO{
		Devices: deviceDTOs,
		Total:   len(devicesResponse.Result),
	}, nil
}
