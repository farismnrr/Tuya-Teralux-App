package dtos

// TuyaDeviceDTO represents a single device for API consumers
type TuyaDeviceDTO struct {
	ID                string                `json:"id"`
	RemoteID          string                `json:"remote_id,omitempty"`
	Name              string                `json:"name"`
	Category          string                `json:"category"`
	RemoteCategory    string                `json:"remote_category,omitempty"`
	ProductName       string                `json:"product_name"`
	RemoteProductName string                `json:"remote_product_name,omitempty"`
	Online            bool                  `json:"online"`
	Icon              string                `json:"icon"`
	Status            []TuyaDeviceStatusDTO `json:"status"`
	CustomName        string                `json:"custom_name,omitempty"`
	Model             string                `json:"model,omitempty"`
	IP                string                `json:"ip,omitempty"`
	LocalKey          string                `json:"local_key"`
	GatewayID         string                `json:"gateway_id"`
	CreateTime        int64                 `json:"create_time"`
	UpdateTime        int64                 `json:"update_time"`
	Collections       []TuyaDeviceDTO       `json:"collections,omitempty"`
}

// TuyaCommandDTO represents a single command
type TuyaCommandDTO struct {
	Code  string      `json:"code" binding:"required"`
	Value interface{} `json:"value" binding:"required"`
}

// TuyaCommandsRequestDTO represents the request body for sending commands
type TuyaCommandsRequestDTO struct {
	Commands []TuyaCommandDTO `json:"commands" binding:"required"`
}

// TuyaIRACCommandDTO represents a single IR AC command request
type TuyaIRACCommandDTO struct {
	RemoteID string `json:"remote_id" binding:"required"`
	Code     string `json:"code" binding:"required"`
	Value    int    `json:"value"`
}

// TuyaDeviceStatusDTO represents device status for API consumers
type TuyaDeviceStatusDTO struct {
	Code  string      `json:"code"`
	Value interface{} `json:"value"`
}

// TuyaDevicesResponseDTO represents the response for getting all devices
type TuyaDevicesResponseDTO struct {
	Devices          []TuyaDeviceDTO `json:"devices"`
	TotalDevices     int             `json:"total_devices"`
	CurrentPageCount int             `json:"current_page_count"`
}

// TuyaDeviceResponseDTO represents the response for getting a single device
type TuyaDeviceResponseDTO struct {
	Device TuyaDeviceDTO `json:"device"`
}

// DeviceStateDTO represents the device state for API consumers
type DeviceStateDTO struct {
	DeviceID     string                   `json:"device_id"`
	LastCommands []DeviceStateCommandDTO  `json:"last_commands"`
	UpdatedAt    int64                    `json:"updated_at"`
}

// DeviceStateCommandDTO represents a single command in the device state
type DeviceStateCommandDTO struct {
	Code  string      `json:"code" binding:"required"`
	Value interface{} `json:"value" binding:"required"`
}

// SaveDeviceStateRequestDTO represents the request body for saving device state
type SaveDeviceStateRequestDTO struct {
	Commands []DeviceStateCommandDTO `json:"commands" binding:"required"`
}

