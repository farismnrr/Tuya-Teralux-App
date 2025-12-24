-- Create device_statuses table
CREATE TABLE IF NOT EXISTS device_statuses (
    id CHAR(36) PRIMARY KEY,
    device_id CHAR(36) NOT NULL,
    name VARCHAR(255),
    code VARCHAR(255) NOT NULL,
    value INT,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Foreign key constraint to devices table
    CONSTRAINT fk_device
        FOREIGN KEY (device_id)
        REFERENCES devices(id)
        ON DELETE CASCADE,
        
    -- Ensure one status entry per code per device
    CONSTRAINT unique_device_code UNIQUE (device_id, code)
);

-- Create index on device_id
CREATE INDEX idx_device_statuses_device_id ON device_statuses(device_id);

-- Create index on code
CREATE INDEX idx_device_statuses_code ON device_statuses(code);
