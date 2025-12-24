package dtos

// TuyaAuthResponseDTO is the authentication response for API consumers
type TuyaAuthResponseDTO struct {
	AccessToken  string `json:"access_token"`
	ExpireTime   int    `json:"expire_time"`
	RefreshToken string `json:"refresh_token"`
	UID          string `json:"uid"`
}

// ErrorResponseDTO represents error response
type ErrorResponseDTO struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}