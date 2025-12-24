-- Create devices table
CREATE TABLE IF NOT EXISTS devices (
    id CHAR(36) PRIMARY KEY,
    teralux_id CHAR(36) NOT NULL,
    name VARCHAR(255) NOT NULL,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Foreign key constraint with CASCADE delete
    CONSTRAINT fk_teralux
        FOREIGN KEY (teralux_id)
        REFERENCES teralux(id)
        ON DELETE CASCADE
);

-- Create index on teralux_id
CREATE INDEX idx_devices_teralux_id ON devices(teralux_id);
