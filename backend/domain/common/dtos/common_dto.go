package dtos

// StandardResponse represents the standardized API response structure
type StandardResponse struct {
	Status  bool        `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// SuccessResponseDTO is a simple DTO for operations returning a success boolean
type SuccessResponseDTO struct {
	Success bool `json:"success"`
}