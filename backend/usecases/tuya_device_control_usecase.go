package usecases

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"teralux_app/dtos"
	"teralux_app/entities"
	"teralux_app/services"
	"teralux_app/utils"
	"time"
	"strings"
)

// TuyaDeviceControlUseCase handles device control business logic
type TuyaDeviceControlUseCase struct {
	service *services.TuyaDeviceService
}

// NewTuyaDeviceControlUseCase creates a new TuyaDeviceControlUseCase instance
func NewTuyaDeviceControlUseCase(service *services.TuyaDeviceService) *TuyaDeviceControlUseCase {
	return &TuyaDeviceControlUseCase{
		service: service,
	}
}

// SendIRACCommand sends a command to an IR air conditioner
//
// Tuya API Documentation:
// URL: https://openapi.tuyacn.com/v2.0/infrareds/{infrared_id}/air-conditioners/{remote_id}/command
// Method: POST
//
// Headers:
//   - client_id: Your Tuya Client ID
//   - access_token: Valid Bearer Token (needed for signature, passed in header if required by specific endpoints, here used for signature)
//   - sign: HMAC-SHA256(client_id + access_token + t + stringToSign)
//   - t: timestamp (ms)
//   - sign_method: HMAC-SHA256
//
// StringToSign Format:
//   POST\n{content_hash}\n\n{url}
//
// Request Body:
//   {
//     "code": "temp",   // Command code (e.g., "power", "mode", "temp", "wind")
//     "value": 24       // Command value (e.g., 0/1 for power, 16-30 for temp)
//   }
//
// Response:
//   {
//     "success": true,
//     "result": true,
//     "code": 200,
//     "msg": "success"
//   }
func (uc *TuyaDeviceControlUseCase) SendIRACCommand(accessToken, infraredID, remoteID, code string, value int) (bool, error) {
	config := utils.GetConfig()

	// 1. Fetch Device Detais to get correct GatewayID (InfraredID)
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
	deviceStringToSign := utils.GenerateTuyaStringToSign("GET", deviceContentHash, "", deviceUrlPath)
	deviceSignature := utils.GenerateTuyaSignature(config.TuyaClientID, config.TuyaClientSecret, accessToken, deviceTimestamp, deviceStringToSign)
	
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
	} else if deviceResp.Success && deviceResp.Result.GatewayID != "" {
		utils.LogDebug("SendIRACCommand: Found GatewayID=%s for device %s. Using it as InfraredID.", deviceResp.Result.GatewayID, remoteID)
		infraredID = deviceResp.Result.GatewayID
	} else {
		utils.LogDebug("SendIRACCommand: No GatewayID found in device details. Using provided infraredID=%s", infraredID)
	}

	// 2. Send IR Command
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
	stringToSign := utils.GenerateTuyaStringToSign("POST", contentHash, "", urlPath)

	// Generate signature
	signature := utils.GenerateTuyaSignature(config.TuyaClientID, config.TuyaClientSecret, accessToken, timestamp, stringToSign)

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
		
		if resp.Code == 30100 {
			utils.LogWarn("Tuya IR API 30100 detected. Attempting fallback to Standard Device Control for device %s...", infraredID)
			
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

			fallbackUrlPath := fmt.Sprintf("/v1.0/devices/%s/commands", remoteID)
			fallbackFullURL := config.TuyaBaseURL + fallbackUrlPath

			// Generate fallback signature
			// Re-create request body for signature calculation
			fallbackReqBody := entities.TuyaCommandRequest{Commands: fallbackCommands}
			fallbackJsonBody, _ := json.Marshal(fallbackReqBody)

			// Calculate content hash
			hFallback := sha256.New()
			hFallback.Write(fallbackJsonBody)
			fallbackContentHash := hex.EncodeToString(hFallback.Sum(nil))

			// Generate string to sign
			fallbackStringToSign := utils.GenerateTuyaStringToSign("POST", fallbackContentHash, "", fallbackUrlPath)

			// Generate signature
			fallbackSignature := utils.GenerateTuyaSignature(config.TuyaClientID, config.TuyaClientSecret, accessToken, timestamp, fallbackStringToSign)

			// Prepare headers
			fallbackHeaders := map[string]string{
				"client_id":    config.TuyaClientID,
				"sign":         fallbackSignature,
				"t":            timestamp,
				"sign_method":  signMethod,
				"access_token": accessToken,
			}
			
			utils.LogDebug("Fallback Legacy Call: DeviceID=%s, URL=%s, Body=%s", remoteID, fallbackFullURL, string(fallbackJsonBody))
			fallbackResp, fallbackErr := uc.service.SendCommand(fallbackFullURL, fallbackHeaders, fallbackCommands)
			if fallbackErr != nil {
				return false, fallbackErr
			}
			
			if !fallbackResp.Success {
				utils.LogError("Fallback Legacy API Failed. Code: %d, Msg: %s", fallbackResp.Code, fallbackResp.Msg)
				return false, fmt.Errorf("tuya Legacy API failed: %s (code: %d)", fallbackResp.Msg, fallbackResp.Code)
			}
			
			return fallbackResp.Result, nil
		}

		return false, fmt.Errorf("tuya IR API failed: %s (code: %d)", resp.Msg, resp.Code)
	}

	return resp.Result, nil
}

// SendCommand sends commands to a device (standard or legacy)
//
// Tuya API Documentation (Standard v1.0 / iot-03):
// URL: https://openapi.tuyacn.com/v1.0/iot-03/devices/{device_id}/commands
// Method: POST
//
// Headers:
//   - client_id, access_token, sign, t, sign_method (Same as above)
//
// Request Body:
//   {
//     "commands": [
//       { "code": "switch_1", "value": true }
//     ]
//   }
//
// NOTE: This endpoint strictly enforces "Standard Instruction Set". If device has no standard mapping (only DPs),
// it will return specific error codes (like 2008). In that case, we fallback to the LEGACY endpoint.
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
	stringToSign := utils.GenerateTuyaStringToSign("POST", contentHash, "", urlPath)
	// log.Printf("DEBUG: Command StringToSign: %s", stringToSign)

	// Generate signature
	signature := utils.GenerateTuyaSignature(config.TuyaClientID, config.TuyaClientSecret, accessToken, timestamp, stringToSign)

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
				retryStringToSign := utils.GenerateTuyaStringToSign("POST", retryContentHash, "", retryUrlPath)
				retrySignature := utils.GenerateTuyaSignature(config.TuyaClientID, config.TuyaClientSecret, accessToken, timestamp, retryStringToSign)

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

	return resp.Result, nil
}
