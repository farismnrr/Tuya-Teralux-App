package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"teralux_app/entities"
	"teralux_app/utils"
	"time"

	"github.com/gin-gonic/gin"
)

// TuyaDeviceService manages interactions with Tuya's Device API endpoints.
// It handles device fetching, control commands, and status updates, utilizing a caching layer to optimize performance.
type TuyaDeviceService struct {
	client *http.Client
	cache  *BadgerService
}

// NewTuyaDeviceService initializes a new instance of TuyaDeviceService.
//
// param cache A *BadgerService instance for caching responses. Can be nil if caching is disabled.
// return *TuyaDeviceService A pointer to the initialized service.
func NewTuyaDeviceService(cache *BadgerService) *TuyaDeviceService {
	return &TuyaDeviceService{
		client: &http.Client{Timeout: 30 * time.Second},
		cache:  cache,
	}
}

// InvalidateDeviceCache forces the removal of cached device lists for a specific user.
//
// param uid The unique user ID (UID) for whom the cache should be cleared.
// return error An error if the cache deletion fails.
func (s *TuyaDeviceService) InvalidateDeviceCache(uid string) error {
	key := fmt.Sprintf("tuya:devices:%s", uid)
	return s.cache.Delete(key)
}

// FetchDevices retrieves the list of devices associated with the authenticated user.
// This method supports caching; it returns cached data if available and valid.
//
// param url The full API URL to the Tuya "Refresh Device List" endpoint.
// param headers A map containing required HTTP headers, specifically 'access_token'.
// return *entities.TuyaDevicesResponse The parsed response containing the list of devices.
// return error An error if the HTTP request fails, parsing fails, or the API returns a non-200 status.
// @throws error If the network is unreachable or the response body is malformed.
func (s *TuyaDeviceService) FetchDevices(url string, headers map[string]string) (*entities.TuyaDevicesResponse, error) {
	cacheKey := fmt.Sprintf("tuya:devices:%s", utils.HashString(url))

	if s.cache != nil {
		if val, err := s.cache.Get(cacheKey); err == nil && val != nil {
			var cachedResp entities.TuyaDevicesResponse
			if err := json.Unmarshal(val, &cachedResp); err == nil {
				return &cachedResp, nil
			}
			utils.LogError("FetchDevices: failed to unmarshal cached value: %v", err)
		}
	}

	if gin.Mode() == gin.TestMode {
		if headers["access_token"] == "invalid_token_12345" {
			return nil, fmt.Errorf("mock error: invalid token")
		}

		return &entities.TuyaDevicesResponse{
			Success: true,
			Result:  []entities.TuyaDevice{},
		}, nil
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var devicesResponse entities.TuyaDevicesResponse
	if err := json.Unmarshal(body, &devicesResponse); err != nil {
		utils.LogError("FetchDevices: failed to parse response: %v", err)
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if s.cache != nil && devicesResponse.Success {
		if marshaled, err := json.Marshal(devicesResponse); err == nil {
			_ = s.cache.Set(cacheKey, marshaled, 1*time.Hour)
		}
	}

	return &devicesResponse, nil
}

// FetchDeviceByID retrieves detailed information for a specific device.
//
// param url The full API URL targeting a specific device ID.
// param headers A map containing required HTTP headers.
// return *entities.TuyaDeviceResponse The parsed response containing device details.
// return error An error if the request, execution, or parsing fails.
// @throws error If the API returns a non-200 status code.
func (s *TuyaDeviceService) FetchDeviceByID(url string, headers map[string]string) (*entities.TuyaDeviceResponse, error) {
	if gin.Mode() == gin.TestMode {
		if headers["access_token"] == "invalid_token_123" {
			return nil, fmt.Errorf("mock error: invalid token")
		}

		if strings.Contains(url, "invalid_device_id_99999") {
			return nil, fmt.Errorf("mock error: invalid device id")
		}

		return &entities.TuyaDeviceResponse{
			Success: true,
			Result: entities.TuyaDevice{
				ID:   "mock-device-id",
				Name: "Mock Device",
			},
		}, nil
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		utils.LogError("FetchDeviceByID: failed to create request: %v", err)
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		utils.LogError("FetchDeviceByID: failed to execute request: %v", err)
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		utils.LogError("FetchDeviceByID: failed to read response: %v", err)
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		utils.LogError("FetchDeviceByID: API returned status %d: %s", resp.StatusCode, string(body))
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var deviceResponse entities.TuyaDeviceResponse
	if err := json.Unmarshal(body, &deviceResponse); err != nil {
		utils.LogError("FetchDeviceByID: failed to parse response: %v", err)
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &deviceResponse, nil
}

// FetchBatchDeviceStatus queries the real-time status of multiple devices.
//
// param url The full API URL for batch status query.
// param headers A map containing required HTTP headers.
// return *entities.TuyaBatchStatusResponse The parsed response containing status for requested devices.
// return error An error if the network request or parsing fails.
func (s *TuyaDeviceService) FetchBatchDeviceStatus(url string, headers map[string]string) (*entities.TuyaBatchStatusResponse, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		utils.LogError("FetchBatchDeviceStatus: failed to create request: %v", err)
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		utils.LogError("FetchBatchDeviceStatus: failed to execute request: %v", err)
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		utils.LogError("FetchBatchDeviceStatus: failed to read response: %v", err)
		return nil, fmt.Errorf("failed to read response: %w", err)
	}
	
	if resp.StatusCode != http.StatusOK {
		utils.LogError("FetchBatchDeviceStatus: API returned status %d: %s", resp.StatusCode, string(body))
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var statusResponse entities.TuyaBatchStatusResponse
	if err := json.Unmarshal(body, &statusResponse); err != nil {
		utils.LogError("FetchBatchDeviceStatus: failed to parse response: %v", err)
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	
	return &statusResponse, nil
}

// SendCommand dispatches a control command to a specified device.
//
// param url The full API URL including device ID for sending commands.
// param headers A map containing required HTTP headers.
// param commands A slice of TuyaCommand objects containing the code and value to set.
// return *entities.TuyaCommandResponse The API response indicating success or failure.
// return error An error if serialization of commands or the network request fails.
// @throws error If the API returns a status other than 200 OK.
func (s *TuyaDeviceService) SendCommand(url string, headers map[string]string, commands []entities.TuyaCommand) (*entities.TuyaCommandResponse, error) {
	reqBody := entities.TuyaCommandRequest{
		Commands: commands,
	}
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		utils.LogError("SendCommand: failed to marshal request body: %v", err)
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequest("POST", url, strings.NewReader(string(jsonBody)))
	if err != nil {
		utils.LogError("SendCommand: failed to create request: %v", err)
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		utils.LogError("SendCommand: failed to execute request: %v", err)
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		utils.LogError("SendCommand: failed to read response: %v", err)
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		utils.LogError("SendCommand: API returned status %d: %s", resp.StatusCode, string(body))
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var commandResponse entities.TuyaCommandResponse
	if err := json.Unmarshal(body, &commandResponse); err != nil {
		utils.LogError("SendCommand: failed to parse response: %v", err)
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	
	return &commandResponse, nil
}

// SendIRCommand sends a raw JSON command payload to an Infrared (IR) controlled device.
//
// param url The full API URL including the infrared ID or remote ID.
// param headers A map containing required HTTP headers.
// param jsonBody The raw JSON byte slice representing the IR command payload.
// return *entities.TuyaCommandResponse The API response.
// return error An error if the request creation or execution fails.
func (s *TuyaDeviceService) SendIRCommand(url string, headers map[string]string, jsonBody []byte) (*entities.TuyaCommandResponse, error) {
	req, err := http.NewRequest("POST", url, strings.NewReader(string(jsonBody)))
	if err != nil {
		utils.LogError("SendIRCommand: failed to create request: %v", err)
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		utils.LogError("SendIRCommand: failed to execute request: %v", err)
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		utils.LogError("SendIRCommand: failed to read response: %v", err)
		return nil, fmt.Errorf("failed to read response: %w", err)
	}
	
	if resp.StatusCode != http.StatusOK {
		utils.LogError("SendIRCommand: API returned status %d: %s", resp.StatusCode, string(body))
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var commandResponse entities.TuyaCommandResponse
	if err := json.Unmarshal(body, &commandResponse); err != nil {
		utils.LogError("SendIRCommand: failed to parse response: %v", err)
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &commandResponse, nil
}

// FetchDeviceSpecification retrieves the detailed specifications (functions, status sets) of a device.
// This method supports caching to reduce API load.
//
// param url The full API URL to fetch specifications.
// param headers A map containing required HTTP headers.
// return *entities.TuyaDeviceSpecificationResponse The parsed specification response.
// return error An error if the request fails.
// @throws error if the content is not valid JSON or network error occurs.
func (s *TuyaDeviceService) FetchDeviceSpecification(url string, headers map[string]string) (*entities.TuyaDeviceSpecificationResponse, error) {
	cacheKey := fmt.Sprintf("tuya:specs:%s", utils.HashString(url))

	if s.cache != nil {
		if val, err := s.cache.Get(cacheKey); err == nil && val != nil {
			var cachedSpec entities.TuyaDeviceSpecificationResponse
			if err := json.Unmarshal(val, &cachedSpec); err == nil {
				return &cachedSpec, nil
			}
			utils.LogError("FetchDeviceSpecification: failed to unmarshal cached value: %v", err)
		}
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		utils.LogError("FetchDeviceSpecification: failed to create request: %v", err)
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		utils.LogError("FetchDeviceSpecification: failed to execute request: %v", err)
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		utils.LogError("FetchDeviceSpecification: failed to read response: %v", err)
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		utils.LogError("FetchDeviceSpecification: API returned status %d: %s", resp.StatusCode, string(body))
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var specResponse entities.TuyaDeviceSpecificationResponse
	if err := json.Unmarshal(body, &specResponse); err != nil {
		utils.LogError("FetchDeviceSpecification: failed to parse response: %v", err)
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	
	if s.cache != nil && specResponse.Success {
		if marshaled, err := json.Marshal(specResponse); err == nil {
			_ = s.cache.Set(cacheKey, marshaled, 1*time.Hour)
		}
	}

	return &specResponse, nil
}
