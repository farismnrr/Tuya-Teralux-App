package entities

// DeviceState represents the last known control state for a device.
// This is stored persistently in BadgerDB and survives cache flushes.
type DeviceState struct {
	DeviceID     string                `json:"device_id"`
	LastCommands []DeviceStateCommand  `json:"last_commands"`
	UpdatedAt    int64                 `json:"updated_at"`
}

// DeviceStateCommand represents a single command in the device state.
type DeviceStateCommand struct {
	Code  string      `json:"code"`
	Value interface{} `json:"value"`
}
