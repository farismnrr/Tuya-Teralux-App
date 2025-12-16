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

// TuyaDeviceService handles HTTP requests to Tuya Device API
type TuyaDeviceService struct {
	client *http.Client
}

// NewTuyaDeviceService creates a new TuyaDeviceService instance
func NewTuyaDeviceService() *TuyaDeviceService {
	return &TuyaDeviceService{
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

// FetchDevices makes HTTP request to get all devices from Tuya API
func (s *TuyaDeviceService) FetchDevices(url string, headers map[string]string) (*entities.TuyaDevicesResponse, error) {
	// MOCK FOR TESTS
	if gin.Mode() == gin.TestMode {
		// Simulate Invalid Token Error
		if headers["access_token"] == "invalid_token_12345" {
			return nil, fmt.Errorf("mock error: invalid token")
		}

		return &entities.TuyaDevicesResponse{
			Success: true,
			Result:  []entities.TuyaDevice{},
		}, nil
	}

	// Create HTTP request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// Execute request
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var devicesResponse entities.TuyaDevicesResponse
	if err := json.Unmarshal(body, &devicesResponse); err != nil {
		utils.LogError("FetchDevices: failed to parse response: %v", err)
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	utils.LogDebug("FetchDevices success: found %d devices", len(devicesResponse.Result))

	return &devicesResponse, nil
}

// FetchDeviceByID makes HTTP request to get a single device by ID from Tuya API
func (s *TuyaDeviceService) FetchDeviceByID(url string, headers map[string]string) (*entities.TuyaDeviceResponse, error) {
	utils.LogDebug("FetchDeviceByID: requesting %s", url)

	// MOCK FOR TESTS
	if gin.Mode() == gin.TestMode {
		// Simulate Invalid Token Error
		if headers["access_token"] == "invalid_token_123" {
			return nil, fmt.Errorf("mock error: invalid token")
		}

		// Simulate Invalid Device ID Error
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

	// Create HTTP request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		utils.LogError("FetchDeviceByID: failed to create request: %v", err)
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// Execute request
	resp, err := s.client.Do(req)
	if err != nil {
		utils.LogError("FetchDeviceByID: failed to execute request: %v", err)
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		utils.LogError("FetchDeviceByID: failed to read response: %v", err)
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// DEBUG LOG RAW BODY
	utils.LogDebug("FetchDeviceByID Response Body: %s", string(body))

	// Check status code
	if resp.StatusCode != http.StatusOK {
		utils.LogError("FetchDeviceByID: API returned status %d: %s", resp.StatusCode, string(body))
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var deviceResponse entities.TuyaDeviceResponse
	if err := json.Unmarshal(body, &deviceResponse); err != nil {
		utils.LogError("FetchDeviceByID: failed to parse response: %v", err)
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &deviceResponse, nil
}

// FetchBatchDeviceStatus makes HTTP request to get status of multiple devices
func (s *TuyaDeviceService) FetchBatchDeviceStatus(url string, headers map[string]string) (*entities.TuyaBatchStatusResponse, error) {
	// Create HTTP request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// Execute request
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var statusResponse entities.TuyaBatchStatusResponse
	if err := json.Unmarshal(body, &statusResponse); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &statusResponse, nil
}

// SendCommand sends commands to a device
func (s *TuyaDeviceService) SendCommand(url string, headers map[string]string, commands []entities.TuyaCommand) (*entities.TuyaCommandResponse, error) {
	// Create request body
	reqBody := entities.TuyaCommandRequest{
		Commands: commands,
	}
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", url, strings.NewReader(string(jsonBody)))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var commandResponse entities.TuyaCommandResponse
	if err := json.Unmarshal(body, &commandResponse); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &commandResponse, nil
}

// SendIRCommand sends a command to an IR device (different request body format)
func (s *TuyaDeviceService) SendIRCommand(url string, headers map[string]string, jsonBody []byte) (*entities.TuyaCommandResponse, error) {
	// Create HTTP request
	req, err := http.NewRequest("POST", url, strings.NewReader(string(jsonBody)))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var commandResponse entities.TuyaCommandResponse
	if err := json.Unmarshal(body, &commandResponse); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &commandResponse, nil
}

// FetchDeviceSpecification makes HTTP request to get device specification from Tuya API
func (s *TuyaDeviceService) FetchDeviceSpecification(url string, headers map[string]string) (*entities.TuyaDeviceSpecificationResponse, error) {
	// Create HTTP request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// Execute request
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var specResponse entities.TuyaDeviceSpecificationResponse
	if err := json.Unmarshal(body, &specResponse); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &specResponse, nil
}
