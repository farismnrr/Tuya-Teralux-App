package usecases

import (
	"encoding/json"
	"fmt"
	"teralux_app/domain/tuya/dtos"
	"teralux_app/domain/tuya/entities"
	"teralux_app/domain/common/infrastructure/persistence"
	"teralux_app/domain/common/utils"
	"time"
)

// DeviceStateUseCase handles business logic for device state persistence.
// It manages saving, retrieving, and cleaning up device control states in BadgerDB.
type DeviceStateUseCase struct {
	cache *persistence.BadgerService
}

// NewDeviceStateUseCase initializes a new DeviceStateUseCase.
//
// param cache The BadgerService used for persistent state storage.
// return *DeviceStateUseCase A pointer to the initialized usecase.
func NewDeviceStateUseCase(cache *persistence.BadgerService) *DeviceStateUseCase {
	return &DeviceStateUseCase{
		cache: cache,
	}
}

// SaveDeviceState saves the last control state for a device to persistent storage.
// The state is stored with key format: "device_state:{device_id}" without TTL.
// This function merges new commands with existing state to preserve all device parameters.
//
// param deviceID The unique ID of the device.
// param commands A list of commands representing the device's current state.
// return error An error if the save operation fails.
func (uc *DeviceStateUseCase) SaveDeviceState(deviceID string, commands []dtos.DeviceStateCommandDTO) error {
	// Retrieve existing state first
	existingState, err := uc.GetDeviceState(deviceID)
	if err != nil {
		utils.LogWarn("DeviceStateUseCase: Failed to retrieve existing state for merge (will create new): %v", err)
	}

	// Create a map to merge commands (code -> value)
	commandMap := make(map[string]interface{})
	
	// Add existing commands to map first
	if existingState != nil && existingState.LastCommands != nil {
		for _, cmd := range existingState.LastCommands {
			commandMap[cmd.Code] = cmd.Value
		}
		utils.LogDebug("DeviceStateUseCase: Loaded %d existing commands for device %s", len(existingState.LastCommands), deviceID)
	}
	
	// Merge/update with new commands
	for _, cmd := range commands {
		commandMap[cmd.Code] = cmd.Value
		utils.LogDebug("DeviceStateUseCase: Merging command: code=%s, value=%v", cmd.Code, cmd.Value)
	}

	// Convert map back to array
	var mergedCommands []entities.DeviceStateCommand
	for code, value := range commandMap {
		mergedCommands = append(mergedCommands, entities.DeviceStateCommand{
			Code:  code,
			Value: value,
		})
	}

	// Create state entity with merged commands
	state := entities.DeviceState{
		DeviceID:     deviceID,
		LastCommands: mergedCommands,
		UpdatedAt:    time.Now().Unix(),
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(state)
	if err != nil {
		utils.LogError("DeviceStateUseCase: Failed to marshal state for device %s: %v", deviceID, err)
		return fmt.Errorf("failed to marshal device state: %w", err)
	}

	// Save to BadgerDB with persistent key (no TTL)
	key := fmt.Sprintf("device_state:%s", deviceID)
	
	utils.LogDebug("DeviceStateUseCase: Saving merged state for device %s with %d total commands", deviceID, len(mergedCommands))
	for i, cmd := range mergedCommands {
		utils.LogDebug("  MergedCommand[%d]: code=%s, value=%v (type=%T)", i, cmd.Code, cmd.Value, cmd.Value)
	}
	utils.LogDebug("  JSON payload: %s", string(jsonData))
	
	if err := uc.cache.SetPersistent(key, jsonData); err != nil {
		utils.LogError("DeviceStateUseCase: Failed to save state for device %s: %v", deviceID, err)
		return fmt.Errorf("failed to save device state: %w", err)
	}

	utils.LogDebug("DeviceStateUseCase: Successfully saved merged state for device %s", deviceID)
	return nil
}

// GetDeviceState retrieves the last known control state for a device.
//
// param deviceID The unique ID of the device.
// return *dtos.DeviceStateDTO The device state, or nil if not found.
// return error An error if the retrieval operation fails.
func (uc *DeviceStateUseCase) GetDeviceState(deviceID string) (*dtos.DeviceStateDTO, error) {
	key := fmt.Sprintf("device_state:%s", deviceID)
	
	// Retrieve from BadgerDB
	jsonData, err := uc.cache.Get(key)
	if err != nil {
		utils.LogError("DeviceStateUseCase: Failed to get state for device %s: %v", deviceID, err)
		return nil, fmt.Errorf("failed to get device state: %w", err)
	}

	// Not found
	if jsonData == nil {
		utils.LogDebug("DeviceStateUseCase: No state found for device %s", deviceID)
		return nil, nil
	}

	// Unmarshal entity
	var state entities.DeviceState
	if err := json.Unmarshal(jsonData, &state); err != nil {
		utils.LogError("DeviceStateUseCase: Failed to unmarshal state for device %s: %v", deviceID, err)
		return nil, fmt.Errorf("failed to unmarshal device state: %w", err)
	}

	// Convert to DTO
	var commandDTOs []dtos.DeviceStateCommandDTO
	for _, cmd := range state.LastCommands {
		commandDTOs = append(commandDTOs, dtos.DeviceStateCommandDTO{
			Code:  cmd.Code,
			Value: cmd.Value,
		})
	}

	stateDTO := &dtos.DeviceStateDTO{
		DeviceID:     state.DeviceID,
		LastCommands: commandDTOs,
		UpdatedAt:    state.UpdatedAt,
	}

	utils.LogDebug("DeviceStateUseCase: Retrieved state for device %s with %d commands", deviceID, len(commandDTOs))
	utils.LogDebug("  Raw JSON: %s", string(jsonData))
	for i, cmd := range commandDTOs {
		utils.LogDebug("  RetrievedCommand[%d]: code=%s, value=%v (type=%T)", i, cmd.Code, cmd.Value, cmd.Value)
	}
	return stateDTO, nil
}

// CleanupOrphanedStates removes device states for devices that no longer exist.
// This is called after fetching the device list from Tuya API.
//
// param validDeviceIDs A list of all currently valid device IDs from Tuya.
// return error An error if the cleanup operation fails.
func (uc *DeviceStateUseCase) CleanupOrphanedStates(validDeviceIDs []string) error {
	// Get all device state keys
	allStateKeys, err := uc.cache.GetAllKeysWithPrefix("device_state:")
	if err != nil {
		utils.LogError("DeviceStateUseCase: Failed to get state keys for cleanup: %v", err)
		return fmt.Errorf("failed to get state keys: %w", err)
	}

	// Create a map of valid device IDs for fast lookup
	validIDMap := make(map[string]bool)
	for _, id := range validDeviceIDs {
		validIDMap[id] = true
	}

	// Check each state key
	deletedCount := 0
	for _, key := range allStateKeys {
		// Extract device ID from key "device_state:{device_id}"
		deviceID := key[len("device_state:"):]
		
		// If device ID is not in valid list, delete the state
		if !validIDMap[deviceID] {
			if err := uc.cache.Delete(key); err != nil {
				utils.LogWarn("DeviceStateUseCase: Failed to delete orphaned state for device %s: %v", deviceID, err)
				continue
			}
			utils.LogInfo("DeviceStateUseCase: Deleted orphaned state for device %s", deviceID)
			deletedCount++
		}
	}

	if deletedCount > 0 {
		utils.LogInfo("DeviceStateUseCase: Cleanup complete - deleted %d orphaned states", deletedCount)
	} else {
		utils.LogDebug("DeviceStateUseCase: Cleanup complete - no orphaned states found")
	}

	return nil
}