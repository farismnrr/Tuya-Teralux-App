package entities

// TuyaDevicesResponse represents the response for getting all devices from Tuya API
// TuyaDevicesResponse represents the response for getting all devices from Tuya API
type TuyaDevicesResponse struct {
	Result  []TuyaDevice `json:"result"`
	Success bool         `json:"success"`
	T       int64        `json:"t"`
	Tid     string       `json:"tid"`
	Code    int          `json:"code"`
	Msg     string       `json:"msg"`
}

// TuyaDeviceResponse represents the response for getting a single device from Tuya API
type TuyaDeviceResponse struct {
	Result  TuyaDevice `json:"result"`
	Success bool       `json:"success"`
	T       int64      `json:"t"`
	Tid     string     `json:"tid"`
	Code    int        `json:"code"`
	Msg     string     `json:"msg"`
}

// TuyaDevice represents a Tuya device
type TuyaDevice struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	RemoteName  string                 `json:"remote_name"`
	UID         string                 `json:"uid"`
	LocalKey    string                 `json:"local_key"`
	Category    string                 `json:"category"`
	ProductID   string                 `json:"product_id"`
	ProductName string                 `json:"product_name"`
	Sub         bool                   `json:"sub"`
	UUID        string                 `json:"uuid"`
	Online      bool                   `json:"online"`
	ActiveTime  int64                  `json:"active_time"`
	Icon        string                 `json:"icon"`
	IP          string                 `json:"ip"`
	TimeZone    string                 `json:"time_zone"`
	CreateTime  int64                  `json:"create_time"`
	UpdateTime  int64                  `json:"update_time"`
	Status      []TuyaDeviceStatus     `json:"status"`
	Model       string                 `json:"model"`
	CustomName  string                 `json:"custom_name"`
	AssetID     string                 `json:"asset_id"`
	OwnerID     string                 `json:"owner_id"`
	NodeID      string                 `json:"node_id"`
	GatewayID   string                 `json:"gateway_id"`
	IsShare     bool                   `json:"is_share"`
	BizType     int                    `json:"biz_type"`
	Lat         string                 `json:"lat"`
	Lon         string                 `json:"lon"`
	Functions   []TuyaDeviceFunction   `json:"functions"`
	StatusRange map[string]interface{} `json:"status_range"`
}


// TuyaDeviceStatus represents the status of a device property
type TuyaDeviceStatus struct {
	Code  string      `json:"code"`
	Value interface{} `json:"value"`
}

// TuyaDeviceFunction represents a device function
type TuyaDeviceFunction struct {
	Code   string `json:"code"`
	Type   string `json:"type"`
	Values string `json:"values"` // Changed from map to string because spec API returns JSON string
}

// TuyaBatchStatusResponse represents the response for batch device status
type TuyaBatchStatusResponse struct {
	Result  []TuyaDeviceStatusItem `json:"result"`
	Success bool                   `json:"success"`
	T       int64                  `json:"t"`
	Code    int                    `json:"code"`
	Msg     string                 `json:"msg"`
}

// TuyaDeviceStatusItem represents a single device status in the batch response
type TuyaDeviceStatusItem struct {
	ID       string `json:"id"`
	IsOnline bool   `json:"is_online"` // Tuya v2/iot-03 often uses is_online
}

// TuyaCommandRequest represents the request body for sending commands
type TuyaCommandRequest struct {
	Commands []TuyaCommand `json:"commands"`
}

// TuyaCommand represents a single command
type TuyaCommand struct {
	Code  string      `json:"code"`
	Value interface{} `json:"value"`
}

// TuyaCommandResponse represents the response after sending commands
type TuyaCommandResponse struct {
	Result  bool   `json:"result"`
	Success bool   `json:"success"`
	T       int64  `json:"t"`
	Code    int    `json:"code"`
	Msg     string `json:"msg"`
}

// TuyaDeviceSpecificationResponse represents the response for device specification
type TuyaDeviceSpecificationResponse struct {
	Result  TuyaDeviceSpecification `json:"result"`
	Success bool                    `json:"success"`
	T       int64                   `json:"t"`
	Code    int                     `json:"code"`
	Msg     string                  `json:"msg"`
}

// TuyaDeviceSpecification represents the specification result
type TuyaDeviceSpecification struct {
	Category  string               `json:"category"`
	Functions []TuyaDeviceFunction `json:"functions"`
	Status    []TuyaDeviceFunction `json:"status"`
}