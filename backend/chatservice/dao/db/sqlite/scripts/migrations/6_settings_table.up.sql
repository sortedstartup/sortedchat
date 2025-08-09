CREATE TABLE IF NOT EXISTS settings (
    name TEXT PRIMARY KEY,
    settings TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Create trigger to update the updated_at timestamp
CREATE TRIGGER IF NOT EXISTS settings_updated_at 
AFTER UPDATE ON settings 
BEGIN
    UPDATE settings SET updated_at = CURRENT_TIMESTAMP WHERE name = NEW.name;
END;