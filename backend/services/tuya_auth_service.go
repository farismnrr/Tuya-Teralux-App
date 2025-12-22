package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"teralux_app/entities"
	"teralux_app/utils"
	"time"
)

// TuyaAuthService handles the OAuth 2.0 authentication flow with the Tuya Cloud API.
type TuyaAuthService struct {
	client *http.Client
}

// NewTuyaAuthService initializes a new instance of TuyaAuthService.
//
// return *TuyaAuthService The initialized authentication service with a default timeout configuration.
func NewTuyaAuthService() *TuyaAuthService {
	return &TuyaAuthService{
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

// FetchToken obtains a new access token from the Tuya API.
//
// param url The complete API endpoint URL for token retrieval (e.g., /v1.0/token?grant_type=1).
// param headers A map containing the necessary signing headers (client_id, sign, t, sign_method, nonce, etc.).
// return *entities.TuyaAuthResponse The structured response containing the access token, refresh token, and expiration time.
// return error An error if the HTTP request fails, status code is not 200, or the response body cannot be parsed.
// @throws error If the Tuya API returns a non-200 status code indicating authentication failure.
func (s *TuyaAuthService) FetchToken(url string, headers map[string]string) (*entities.TuyaAuthResponse, error) {
	utils.LogDebug("FetchToken: requesting %s", url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		utils.LogError("FetchToken: failed to create request: %v", err)
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		utils.LogError("FetchToken: failed to execute request: %v", err)
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		utils.LogError("FetchToken: failed to read response: %v", err)
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	utils.LogDebug("FetchToken Response Body: %s", string(body))
	if resp.StatusCode != http.StatusOK {
		utils.LogError("FetchToken: API returned status %d: %s", resp.StatusCode, string(body))
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var authResponse entities.TuyaAuthResponse
	if err := json.Unmarshal(body, &authResponse); err != nil {
		utils.LogError("FetchToken: failed to parse response: %v", err)
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	utils.LogDebug("FetchToken success: token received, expires in %d seconds", authResponse.Result.ExpireTime)
	return &authResponse, nil
}
