package dtos

// SensorDataDTO represents the formatted sensor data
type SensorDataDTO struct {
	Temperature       float64 `json:"temperature"`
	Humidity          int     `json:"humidity"`
	BatteryPercentage int     `json:"battery_percentage"`
	StatusText        string  `json:"status_text"`
	TempUnit          string  `json:"temp_unit"`
}
