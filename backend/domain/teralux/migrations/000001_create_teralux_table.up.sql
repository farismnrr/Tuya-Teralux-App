-- Create teralux table
CREATE TABLE IF NOT EXISTS teralux (
    id CHAR(36) PRIMARY KEY,
    mac_address VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create index on mac_address
CREATE INDEX idx_teralux_mac_address ON teralux(mac_address);
