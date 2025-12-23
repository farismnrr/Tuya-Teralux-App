package usecases

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"teralux_app/dtos"
	"teralux_app/services"
	"teralux_app/utils"
	"time"
)

// TuyaGetAllDevicesUseCase orchestrates the retrieval and aggregation of device data.
// It combines the user's device list, individual device specifications, and real-time status.
type TuyaGetAllDevicesUseCase struct {
	service       *services.TuyaDeviceService
	cache         *services.BadgerService
	deviceStateUC *DeviceStateUseCase
}

// NewTuyaGetAllDevicesUseCase initializes a new TuyaGetAllDevicesUseCase.
//
// param service The TuyaDeviceService used for API interactions.
// param cache The BadgerService used for caching device lists.
// param deviceStateUC The DeviceStateUseCase for cleaning up orphaned states.
// return *TuyaGetAllDevicesUseCase A pointer to the initialized usecase.
func NewTuyaGetAllDevicesUseCase(service *services.TuyaDeviceService, cache *services.BadgerService, deviceStateUC *DeviceStateUseCase) *TuyaGetAllDevicesUseCase {
	return &TuyaGetAllDevicesUseCase{
		service:       service,
		cache:         cache,
		deviceStateUC: deviceStateUC,
	}
}

// GetAllDevices retrieves the complete list of devices for a user, including statuses and specs.
// It performs multiple API calls: fetching the device list, fetching specifications for each, and batch-fetching real-time status.
// It also handles device categorization and grouping (e.g., grouping IR ACs under a Smart IR Hub).
//
// Tuya API Interactions:
// 1. List Devices by User: GET /v1.0/users/{uid}/devices
// 2. Get Device Specifications: GET /v1.0/iot-03/devices/{device_id}/specification
// 3. Batch Get Device Status: GET /v1.0/iot-03/devices/status
//
// param accessToken The valid OAuth 2.0 access token.
// param uid The Tuya User ID for whom to fetch devices.
// param page Page number for pagination (optional, 0 to ignore).
// param limit Items per page (optional, 0 to ignore).
// param category Category to filter by (optional, empty to ignore).
// return *dtos.TuyaDevicesResponseDTO The aggregated list of devices.
// return error An error if fetching the device list fails.
// @throws error If the API returns a failure (e.g., invalid token).
func (uc *TuyaGetAllDevicesUseCase) GetAllDevices(accessToken, uid string, page, limit int, category string) (*dtos.TuyaDevicesResponseDTO, error) {
	// Get config
	config := utils.GetConfig()

	// 1. Try Cache First
	cacheKey := fmt.Sprintf("cache:devices:%s", uid)
	var deviceDTOs []dtos.TuyaDeviceDTO

	cachedData, err := uc.cache.Get(cacheKey)
	if err == nil && cachedData != nil {
		if err := json.Unmarshal(cachedData, &deviceDTOs); err == nil {
			utils.LogDebug("GetAllDevices: Cache HIT for uid %s", uid)
		} else {
			utils.LogWarn("GetAllDevices: Cache corrupted for uid %s, fetching fresh data", uid)
			cachedData = nil // Force refresh
		}
	} else {
		utils.LogDebug("GetAllDevices: Cache MISS for uid %s (err: %v)", uid, err)
	}

	// 2. If Cache Miss, Fetch from API
	if cachedData == nil {
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
			utils.LogDebug("DEVICE DEBUG: ID=%s, Name=%s, Category=%s", dev.ID, dev.Name, dev.Category)
			for _, st := range dev.Status {
				utils.LogDebug("   STATUS: Code=%s, Value=%v (Type: %T)", st.Code, st.Value, st.Value)
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
				utils.LogDebug("   SPECIFICATION for ID=%s:", dev.ID)
				for _, fn := range specResp.Result.Functions {
					utils.LogDebug("      FUNCTION: Code=%s, Type=%s, Values=%s", fn.Code, fn.Type, fn.Values)
				}
			} else {
				utils.LogError("   FAILED to fetch spec for ID=%s: %v", dev.ID, errSpec)
			}
		}

		// Transform entities to DTOs
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
				utils.LogWarn("WARN: Failed to fetch batch status: %v", err)
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


			// Determine display name (Use RemoteName if available)
			displayName := device.Name
			if device.RemoteName != "" {
				displayName = device.RemoteName
			}

			deviceDTOs = append(deviceDTOs, dtos.TuyaDeviceDTO{
				ID:          device.ID,
				Name:        displayName,
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

		// Process devices based on response type configuration
		switch config.GetAllDevicesResponseType {
		case "0":
			deviceDTOs = uc.processResponseMode0(deviceDTOs)
		case "1":
			deviceDTOs = uc.processResponseMode1(deviceDTOs)
		case "2":
			deviceDTOs = uc.processResponseMode2(deviceDTOs)
		default:
			// Default to Mode 0
			deviceDTOs = uc.processResponseMode0(deviceDTOs)
		}

		// 3. Save to Cache
		if jsonData, err := json.Marshal(deviceDTOs); err == nil {
			uc.cache.Set(cacheKey, jsonData)
			utils.LogDebug("GetAllDevices: Saved %d devices to cache for uid %s", len(deviceDTOs), uid)
		} else {
			utils.LogError("GetAllDevices: Failed to marshal devices for cache: %v", err)
		}

		// 4. Cleanup orphaned device states
		if uc.deviceStateUC != nil {
			var allDeviceIDs []string
			for _, dev := range deviceDTOs {
				allDeviceIDs = append(allDeviceIDs, dev.ID)
				// Also include remote IDs for merged devices (Mode 2)
				if dev.RemoteID != "" {
					allDeviceIDs = append(allDeviceIDs, dev.RemoteID)
				}
				// Include collection IDs (Mode 0)
				for _, coll := range dev.Collections {
					allDeviceIDs = append(allDeviceIDs, coll.ID)
				}
			}
			if err := uc.deviceStateUC.CleanupOrphanedStates(allDeviceIDs); err != nil {
				utils.LogWarn("GetAllDevices: Failed to cleanup orphaned states: %v", err)
			}
		}
	}

	// --- NEW: Filter by Category ---
	if category != "" {
		var filteredDevices []dtos.TuyaDeviceDTO
		for _, d := range deviceDTOs {
			// Check main category
			if d.Category == category {
				filteredDevices = append(filteredDevices, d)
				continue
			}
			// Also check remote category for merged devices (Mode 2)
			if d.RemoteCategory == category {
				filteredDevices = append(filteredDevices, d)
			}
		}
		deviceDTOs = filteredDevices
	}

	// Update Total after filtering
	total := len(deviceDTOs)

	// Sort devices by Name Ascending (Alphabetical)
	sort.Slice(deviceDTOs, func(i, j int) bool {
		return deviceDTOs[i].Name < deviceDTOs[j].Name
	})

	// --- NEW: Pagination ---
	if limit > 0 {
		start := (page - 1) * limit
		if start < 0 {
			start = 0
		}
		
		if start >= len(deviceDTOs) {
			// Page out of range
			deviceDTOs = []dtos.TuyaDeviceDTO{}
		} else {
			end := start + limit
			if end > len(deviceDTOs) {
				end = len(deviceDTOs)
			}
			deviceDTOs = deviceDTOs[start:end]
		}
	}

	return &dtos.TuyaDevicesResponseDTO{
		Devices:          deviceDTOs,
		TotalDevices:     total,
		CurrentPageCount: len(deviceDTOs),
	}, nil
}

// processResponseMode0 handles nesting IR devices inside Smart IR Hubs
func (uc *TuyaGetAllDevicesUseCase) processResponseMode0(deviceDTOs []dtos.TuyaDeviceDTO) []dtos.TuyaDeviceDTO {
	var finalDevices []dtos.TuyaDeviceDTO
	var irDevices []dtos.TuyaDeviceDTO
	var smartIRIndices []int

	// 1. Separate IR AC devices and identify Smart IR hubs
	for _, d := range deviceDTOs {
		if d.Category == "infrared_ac" {
			irDevices = append(irDevices, d)
			continue
		}
		finalDevices = append(finalDevices, d)
	}

	// 2. Find Smart IR hubs in the final list
	for i, d := range finalDevices {
		if d.Category == "wnykq" {
			smartIRIndices = append(smartIRIndices, i)
		}
	}

	// 3. Assign IR devices to hubs
	// If no hubs or no IR devices, just return the combined list
	if len(smartIRIndices) == 0 || len(irDevices) == 0 {
		finalDevices = append(finalDevices, irDevices...)
		return finalDevices
	}

	// Map Hub ID and LocalKey to Index for direct access
	hubIDMap := make(map[string]int)
	hubLocalKeyMap := make(map[string]int)

	for _, idx := range smartIRIndices {
		hub := finalDevices[idx]
		hubIDMap[hub.ID] = idx
		if hub.LocalKey != "" {
			hubLocalKeyMap[hub.LocalKey] = idx
		}
	}

	var orphanIRs []dtos.TuyaDeviceDTO

	for _, ir := range irDevices {
		// Strategy 1: Match by GatewayID (Official method)
		if targetIdx, ok := hubIDMap[ir.GatewayID]; ok {
			finalDevices[targetIdx].Collections = append(finalDevices[targetIdx].Collections, ir)
			continue
		}

		// Strategy 2: Match by LocalKey (Fallback method for some devices)
		if targetIdx, ok := hubLocalKeyMap[ir.LocalKey]; ok {
			finalDevices[targetIdx].Collections = append(finalDevices[targetIdx].Collections, ir)
			continue
		}

		// Strategy 3: Orphan (No parent found)
		orphanIRs = append(orphanIRs, ir)
	}

	// Add orphans back to main list
	if len(orphanIRs) > 0 {
		finalDevices = append(finalDevices, orphanIRs...)
	}

	return finalDevices
}

// processResponseMode1 handles the flat list response (Mode 1)
func (uc *TuyaGetAllDevicesUseCase) processResponseMode1(deviceDTOs []dtos.TuyaDeviceDTO) []dtos.TuyaDeviceDTO {
	// Mode 1 is the flat list response.
	// No additional processing is needed as the devices are already in a flat list.
	return deviceDTOs
}

// processResponseMode2 handles merging IR devices with their hubs in a flat list
func (uc *TuyaGetAllDevicesUseCase) processResponseMode2(deviceDTOs []dtos.TuyaDeviceDTO) []dtos.TuyaDeviceDTO {
	// 1. Identify Hubs and Remotes
	hubMap := make(map[string]dtos.TuyaDeviceDTO)         // HubID -> HubDTO
	hubLocalKeyMap := make(map[string]dtos.TuyaDeviceDTO) // LocalKey -> HubDTO

	var irRemotes []dtos.TuyaDeviceDTO
	var otherDevices []dtos.TuyaDeviceDTO

	// First pass: Index Hubs and separate Remotes
	for _, d := range deviceDTOs {
		if d.Category == "wnykq" {
			hubMap[d.ID] = d
			if d.LocalKey != "" {
				hubLocalKeyMap[d.LocalKey] = d
			}
		}
	}

	// Second pass: Categorize into Remotes and Others
	for _, d := range deviceDTOs {
		if d.Category == "infrared_ac" {
			irRemotes = append(irRemotes, d)
			continue
		}
		// Process others
		otherDevices = append(otherDevices, d)
	}

	var finalDevices []dtos.TuyaDeviceDTO
	usedHubIDs := make(map[string]bool)

	// Process IR Remotes -> Create Merged Entries
	for _, remote := range irRemotes {
		var parentHub dtos.TuyaDeviceDTO
		found := false

		// Try to find parent hub
		if hub, ok := hubMap[remote.GatewayID]; ok {
			parentHub = hub
			found = true
		}

		if !found {
			// Check local key if not found by GatewayID
			if hub, ok := hubLocalKeyMap[remote.LocalKey]; ok {
				parentHub = hub
				found = true
			}
		}

		if !found {
			// Orphan Remote? Just add it as is
			finalDevices = append(finalDevices, remote)
			continue
		}

		mergedDevice := parentHub
		mergedDevice.RemoteID = remote.ID
		mergedDevice.Name = remote.Name // Overwrite hub name with remote name
		mergedDevice.RemoteCategory = remote.Category
		mergedDevice.RemoteProductName = remote.ProductName
		mergedDevice.Icon = remote.Icon
		mergedDevice.CreateTime = remote.CreateTime
		mergedDevice.UpdateTime = remote.UpdateTime

		finalDevices = append(finalDevices, mergedDevice)
		usedHubIDs[parentHub.ID] = true
	}

	// Add non-remote devices
	for _, d := range otherDevices {
		if d.Category == "wnykq" {
			if _, used := usedHubIDs[d.ID]; used {
				continue // Skip this hub, it's represented by its children
			}
		}
		finalDevices = append(finalDevices, d)
	}

	return finalDevices
}
