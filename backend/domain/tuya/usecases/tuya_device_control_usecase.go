package usecases

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"teralux_app/domain/tuya/dtos"
	"teralux_app/domain/tuya/entities"
	"teralux_app/domain/common/infrastructure/persistence"
	"teralux_app/domain/tuya/services"
	"teralux_app/domain/common/utils"
	tuya_utils "teralux_app/domain/tuya/utils"
	"time"
	"strings"
)

// TuyaDeviceControlUseCase handles the business logic for controlling Tuya devices.
// It supports both standard device control (switches, lights) and specialized IR air conditioner control.
type TuyaDeviceControlUseCase struct {
	service          *services.TuyaDeviceService
	deviceStateUC    *DeviceStateUseCase
	cache            *persistence.BadgerService
}

// NewTuyaDeviceControlUseCase initializes a new TuyaDeviceControlUseCase.
//
// param service The TuyaDeviceService used for API communication.
// param deviceStateUC The DeviceStateUseCase for saving device states.
// param cache The BadgerService for cache invalidation.
// return *TuyaDeviceControlUseCase A pointer to the initialized usecase.
func NewTuyaDeviceControlUseCase(service *services.TuyaDeviceService, deviceStateUC *DeviceStateUseCase, cache *persistence.BadgerService) *TuyaDeviceControlUseCase {
	return &TuyaDeviceControlUseCase{
		service:       service,
		deviceStateUC: deviceStateUC,
		cache:         cache,
	}
}

// SendIRACCommand sends a specific command to an Infrared (IR) controlled Air Conditioner.
// It first attempts to resolve the correct gateway/infrared ID before sending the command.
// If the primary IR command fails with specific error codes (e.g., 30100), it attempts a fallback to standard device control.
//
// param accessToken The valid OAuth 2.0 access token.
// param infraredID The ID of the IR blaster device (or virtual ID).
// param remoteID The ID of the configured remote control for the AC.
// param code The command code (e.g., "temp", "mode", "power", "wind").
// param value The value for the command (e.g., 24 for temp, 1 for power on).
// return bool True if the command was executed successfully.
// return error An error if the command failed after all attempts.
// @throws error If the API returns a failure code that cannot be handled by fallback logic.
func (uc *TuyaDeviceControlUseCase) SendIRACCommand(accessToken, infraredID, remoteID, code string, value int) (bool, error) {
	config := utils.GetConfig()
	forceLegacy := false
	var gatewayID string

	// 1. Fetch Device Detais to get correct GatewayID (InfraredID) and check for Custom Instructions
	//
	// Tuya API Documentation (Get Device Specification/Details):
	// URL: /v1.0/iot-03/devices/{device_id}
	// Method: GET
	// Auth: Standard Header Signature
	// Note: For GET requests, the content-hash in StringToSign must be the SHA256 of empty string.
	deviceUrlPath := fmt.Sprintf("/v1.0/iot-03/devices/%s", remoteID)
	deviceFullURL := config.TuyaBaseURL + deviceUrlPath
	
	// Generate timestamp for device fetch
	deviceTimestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)

	// Calculate content hash for empty body (GET request)
	hEmpty := sha256.New()
	hEmpty.Write([]byte(""))
	deviceContentHash := hex.EncodeToString(hEmpty.Sum(nil))
	
	// Generate signature for device fetch
	deviceStringToSign := tuya_utils.GenerateTuyaStringToSign("GET", deviceContentHash, "", deviceUrlPath)
	deviceSignature := tuya_utils.GenerateTuyaSignature(config.TuyaClientID, config.TuyaClientSecret, accessToken, deviceTimestamp, deviceStringToSign)
	
	// Prepare headers for device fetch
	deviceHeaders := map[string]string{
		"client_id":    config.TuyaClientID,
		"sign":         deviceSignature,
		"t":            deviceTimestamp,
		"sign_method":  "HMAC-SHA256",
		"access_token": accessToken,
	}

	// Call FetchDeviceByID
	utils.LogDebug("SendIRACCommand: Fetching device details for RemoteID=%s", remoteID)
	deviceResp, err := uc.service.FetchDeviceByID(deviceFullURL, deviceHeaders)
	if err != nil {
		utils.LogError("WARNING: Failed to fetch device details for IR command: %v. Continuing with provided infraredID.", err)
	} else if deviceResp.Success {
		// Check for GatewayID
		if deviceResp.Result.GatewayID != "" {
			utils.LogDebug("SendIRACCommand: Found GatewayID=%s for device %s. Using it as InfraredID.", deviceResp.Result.GatewayID, remoteID)
			gatewayID = deviceResp.Result.GatewayID
			infraredID = gatewayID
		}
		
		// Check for Custom Instructions (PowerOn/PowerOff)
		// If these exist, we MUST use the legacy Standard Control API, as the IR API will likely fail or misbehave.
		for _, fun := range deviceResp.Result.Functions {
			if fun.Code == "PowerOn" || fun.Code == "PowerOff" {
				utils.LogDebug("SendIRACCommand: Detected custom instruction set (PowerOn/Off) for device %s. Forcing Standard Control fallback.", remoteID)
				forceLegacy = true
				break
			}
		}
	} else {
		utils.LogDebug("SendIRACCommand: No GatewayID found in device details. Using provided infraredID=%s", infraredID)
	}

	// Helper function for Legacy/Fallback Call
	sendLegacy := func() (bool, error) {
		// Map IR command to Standard DP
		var fallbackCode string
		var fallbackValue interface{}
		fallbackValue = value

		switch code {
		case "temp":
			fallbackCode = "T"
			// Value is integer 16-30, same as input
		case "power":
			if value == 1 {
				fallbackCode = "PowerOn"
				fallbackValue = "PowerOn"
			} else {
				fallbackCode = "PowerOff"
				fallbackValue = "PowerOff"
			}
		case "mode":
			fallbackCode = "M"
			// Value is integer 0-4
		case "wind":
			fallbackCode = "F"
			// Value is integer 0-3
		default:
			// Try using code as is
			fallbackCode = code
		}

		utils.LogDebug("Fallback mapping: %s -> %s, %v -> %v", code, fallbackCode, value, fallbackValue)

		// Construct Standard Command Entity (not DTO, need Entity for service)
		fallbackCommands := []entities.TuyaCommand{
			{
				Code:  fallbackCode,
				Value: fallbackValue,
			},
		}

		// Use LEGACY endpoint explicitly
		retryTimestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)
		retrySignMethod := "HMAC-SHA256"
		
		fallbackUrlPath := fmt.Sprintf("/v1.0/devices/%s/commands", remoteID)
		fallbackFullURL := config.TuyaBaseURL + fallbackUrlPath

		// Generate fallback signature
		fallbackReqBody := entities.TuyaCommandRequest{Commands: fallbackCommands}
		fallbackJsonBody, _ := json.Marshal(fallbackReqBody)

		hFallback := sha256.New()
		hFallback.Write(fallbackJsonBody)
		fallbackContentHash := hex.EncodeToString(hFallback.Sum(nil))

		fallbackStringToSign := tuya_utils.GenerateTuyaStringToSign("POST", fallbackContentHash, "", fallbackUrlPath)
		fallbackSignature := tuya_utils.GenerateTuyaSignature(config.TuyaClientID, config.TuyaClientSecret, accessToken, retryTimestamp, fallbackStringToSign)

		fallbackHeaders := map[string]string{
			"client_id":    config.TuyaClientID,
			"sign":         fallbackSignature,
			"t":            retryTimestamp,
			"sign_method":  retrySignMethod,
			"access_token": accessToken,
		}
		
		utils.LogDebug("Fallback Legacy Call: DeviceID=%s, URL=%s, Body=%s", remoteID, fallbackFullURL, string(fallbackJsonBody))
		fallbackResp, fallbackErr := uc.service.SendCommand(fallbackFullURL, fallbackHeaders, fallbackCommands)
		if fallbackErr != nil {
			return false, fallbackErr
		}
		
		if !fallbackResp.Success {
			utils.LogError("Fallback Legacy API Failed. Code: %d, Msg: %s", fallbackResp.Code, fallbackResp.Msg)
			
			// Handle code 1106 (Permission Deny) - usually means incorrect request body/parameters
			if fallbackResp.Code == 1106 {
				return false, fmt.Errorf("bad request: invalid input parameters. Please verify your request body matches the device's expected command format (code: %d)", fallbackResp.Code)
			}
			
			return false, fmt.Errorf("tuya Legacy API failed: %s (code: %d)", fallbackResp.Msg, fallbackResp.Code)
		}
		
		return fallbackResp.Result, nil
	}

	// 2. Decide Execution Path
	if forceLegacy {
		return sendLegacy()
	}

	// 3. Send IR Command (Default Path)
	// Generate timestamp
	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)
	signMethod := "HMAC-SHA256"

	// Build URL path for IR AC control
	urlPath := fmt.Sprintf("/v2.0/infrareds/%s/air-conditioners/%s/command", infraredID, remoteID)
	fullURL := config.TuyaBaseURL + urlPath

	// Create request body (single command, not array)
	reqBody := map[string]interface{}{
		"code":  code,
		"value": value,
	}
	jsonBody, _ := json.Marshal(reqBody)

	// Calculate content hash
	h := sha256.New()
	h.Write(jsonBody)
	contentHash := hex.EncodeToString(h.Sum(nil))

	// Generate string to sign
	stringToSign := tuya_utils.GenerateTuyaStringToSign("POST", contentHash, "", urlPath)

	// Generate signature
	signature := tuya_utils.GenerateTuyaSignature(config.TuyaClientID, config.TuyaClientSecret, accessToken, timestamp, stringToSign)

	// Prepare headers
	headers := map[string]string{
		"client_id":    config.TuyaClientID,
		"sign":         signature,
		"t":            timestamp,
		"sign_method":  signMethod,
		"access_token": accessToken,
	}

	// Call service
	utils.LogDebug("SendIRACCommand: InfraredID=%s, RemoteID=%s, Code=%s, Value=%d, URL=%s, Body=%s", infraredID, remoteID, code, value, fullURL, string(jsonBody))
	resp, err := uc.service.SendIRCommand(fullURL, headers, jsonBody)
	if err != nil {
		return false, err
	}

	if !resp.Success {
		utils.LogError("Tuya IR API Command Failed. Code: %d, Msg: %s", resp.Code, resp.Msg)
		
		// 30100 = Custom Gateway/Device limitation?
		// 1106 = Permission Deny (often instruction set mismatch)
		if resp.Code == 30100 || resp.Code == 1106 {
			utils.LogWarn("Tuya IR API error %d detected. Attempting fallback to Standard Device Control for device %s...", resp.Code, infraredID)
			return sendLegacy()
		}
		
		return false, fmt.Errorf("tuya IR API failed: %s (code: %d)", resp.Msg, resp.Code)
	}

	// Save state after successful command
	if uc.deviceStateUC != nil {
		stateCommands := []dtos.DeviceStateCommandDTO{
			{Code: code, Value: value},
		}
		if err := uc.deviceStateUC.SaveDeviceState(remoteID, stateCommands); err != nil {
			utils.LogWarn("Failed to save device state for %s: %v", remoteID, err)
		}
	}

	// Invalidate cache for this device
	if uc.cache != nil {
		cacheKey := fmt.Sprintf("cache:tuya_device:%s", remoteID)
		if err := uc.cache.Delete(cacheKey); err != nil {
			utils.LogWarn("Failed to invalidate cache for device %s: %v", remoteID, err)
		} else {
			utils.LogDebug("Cache invalidated for device %s", remoteID)
		}
	}

	return resp.Result, nil
}

// SendCommand sends a set of commands to a standard Tuya device.
// It generates the necessary signatures and headers, then dispatches the request via the service layer.
//
// param accessToken The valid OAuth 2.0 access token.
// param deviceID The unique ID of the device to control.
// param commands A list of TuyaCommandDTOs representing the instructions.
// return bool True if the command was executed successfully.
// return error An error if the API request fails or returns an error code.
// @throws error If the command fails, including specific retry logic for legacy switch commands involving naming mismatch.
func (uc *TuyaDeviceControlUseCase) SendCommand(accessToken, deviceID string, commands []dtos.TuyaCommandDTO) (bool, error) {
	// Get config
	config := utils.GetConfig()

	// Generate timestamp
	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)
	signMethod := "HMAC-SHA256"

	// Build URL path
	urlPath := fmt.Sprintf("/v1.0/iot-03/devices/%s/commands", deviceID)
	fullURL := config.TuyaBaseURL + urlPath

	// Convert DTOs to Entities
	var entityCommands []entities.TuyaCommand
	for _, cmd := range commands {
		entityCommands = append(entityCommands, entities.TuyaCommand{
			Code:  cmd.Code,
			Value: cmd.Value,
		})
	}

	// Create request body for signature calculation
	reqBody := entities.TuyaCommandRequest{
		Commands: entityCommands,
	}
	jsonBody, _ := json.Marshal(reqBody)

	// Calculate content hash
	h := sha256.New()
	h.Write(jsonBody)
	contentHash := hex.EncodeToString(h.Sum(nil))

	// Generate string to sign
	stringToSign := tuya_utils.GenerateTuyaStringToSign("POST", contentHash, "", urlPath)
	// log.Printf("DEBUG: Command StringToSign: %s", stringToSign)

	// Generate signature
	signature := tuya_utils.GenerateTuyaSignature(config.TuyaClientID, config.TuyaClientSecret, accessToken, timestamp, stringToSign)

	// Prepare headers
	headers := map[string]string{
		"client_id":    config.TuyaClientID,
		"sign":         signature,
		"t":            timestamp,
		"sign_method":  signMethod,
		"access_token": accessToken,
	}

	// Call service
	utils.LogDebug("SendCommand: DeviceID=%s, URL=%s, Body=%s", deviceID, fullURL, string(jsonBody))
	resp, err := uc.service.SendCommand(fullURL, headers, entityCommands)
	if err != nil {
		return false, err
	}

	if !resp.Success {
		utils.LogError("Tuya API Command Failed. Code: %d, Msg: %s", resp.Code, resp.Msg)

		// Handle code 1106 (Permission Deny) - usually means incorrect request body/parameters
		if resp.Code == 1106 {
			return false, fmt.Errorf("bad request: invalid input parameters. Please verify your request body matches the device's expected command format (code: %d)", resp.Code)
		}

		// RETRY LOGIC for "switch_" mismatch (switch_1 -> switch1)
		if resp.Code == 2008 {
			var retryCommands []entities.TuyaCommand
			shouldRetry := false
			
			for _, cmd := range entityCommands {
				newCode := cmd.Code
				if strings.HasPrefix(cmd.Code, "switch_") {
					newCode = strings.Replace(cmd.Code, "_", "", 1)
					if newCode != cmd.Code {
						shouldRetry = true
					}
				}
				retryCommands = append(retryCommands, entities.TuyaCommand{Code: newCode, Value: cmd.Value})
			}

			if shouldRetry {
				utils.LogDebug("Retrying with corrected commands: %+v", retryCommands)
				
				// Use LEGACY endpoint for DP instructions (v1.0/devices/{id}/commands) instead of iot-03
				// This is crucial because iot-03 endpoint validates against Standard Instruction Set (which is empty here).
				retryUrlPath := fmt.Sprintf("/v1.0/devices/%s/commands", deviceID)
				retryFullURL := config.TuyaBaseURL + retryUrlPath

				// Re-create request body
				retryReqBody := entities.TuyaCommandRequest{Commands: retryCommands}
				retryJsonBody, _ := json.Marshal(retryReqBody)

				// Re-calculate content hash
				hRetry := sha256.New()
				hRetry.Write(retryJsonBody)
				retryContentHash := hex.EncodeToString(hRetry.Sum(nil))

				// Re-sign
				retryStringToSign := tuya_utils.GenerateTuyaStringToSign("POST", retryContentHash, "", retryUrlPath)
				retrySignature := tuya_utils.GenerateTuyaSignature(config.TuyaClientID, config.TuyaClientSecret, accessToken, timestamp, retryStringToSign)

				// Re-prepare headers
				retryHeaders := map[string]string{
					"client_id":    config.TuyaClientID,
					"sign":         retrySignature,
					"t":            timestamp,
					"sign_method":  signMethod,
					"access_token": accessToken,
				}
				
				// Retry call
				retryResp, retryErr := uc.service.SendCommand(retryFullURL, retryHeaders, retryCommands)
				if retryErr == nil && retryResp.Success {
					utils.LogInfo("Retry success with corrected commands!")
					return retryResp.Result, nil
				} else if retryErr != nil {
					utils.LogError("Retry failed: %v", retryErr)
				} else {
					utils.LogError("Retry API failed: %d %s", retryResp.Code, retryResp.Msg)
				}
			}
		}
		
		return false, fmt.Errorf("tuya API failed: %s (code: %d)", resp.Msg, resp.Code)
	}

	// Save state after successful command
	if uc.deviceStateUC != nil {
		stateCommands := make([]dtos.DeviceStateCommandDTO, len(commands))
		for i, cmd := range commands {
			stateCommands[i] = dtos.DeviceStateCommandDTO{
				Code:  cmd.Code,
				Value: cmd.Value,
			}
		}
		if err := uc.deviceStateUC.SaveDeviceState(deviceID, stateCommands); err != nil {
			utils.LogWarn("Failed to save device state for %s: %v", deviceID, err)
		}
	}

	// Invalidate cache for this device
	if uc.cache != nil {
		cacheKey := fmt.Sprintf("cache:tuya_device:%s", deviceID)
		if err := uc.cache.Delete(cacheKey); err != nil {
			utils.LogWarn("Failed to invalidate cache for device %s: %v", deviceID, err)
		} else {
			utils.LogDebug("Cache invalidated for device %s", deviceID)
		}
	}

	return resp.Result, nil
}