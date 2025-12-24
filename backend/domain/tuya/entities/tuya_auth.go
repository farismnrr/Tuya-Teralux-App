package entities

// TuyaAuthResponse represents the authentication response from Tuya API
type TuyaAuthResponse struct {
	Result  TuyaAuthResult `json:"result"`
	Success bool           `json:"success"`
	T       int64          `json:"t"`
	Tid     string         `json:"tid"`
	Code    int            `json:"code"`
	Msg     string         `json:"msg"`
}

// TuyaAuthResult contains the authentication token data
type TuyaAuthResult struct {
	AccessToken  string `json:"access_token"`
	ExpireTime   int    `json:"expire_time"`
	RefreshToken string `json:"refresh_token"`
	UID          string `json:"uid"`
}