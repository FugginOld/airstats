-- Create user_settings table
CREATE TABLE user_settings (
    id SERIAL PRIMARY KEY,
    setting_key VARCHAR(100) NOT NULL UNIQUE,
    setting_value TEXT NOT NULL,
    description TEXT
);

-- Create index
CREATE INDEX idx_user_settings_key ON user_settings(setting_key);

-- Insert defaults
INSERT INTO user_settings (setting_key, setting_value, description) VALUES
    ('route_table_limit', '5', 'Number of rows to display in Route Information tables');