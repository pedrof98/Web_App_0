-- +goose Up
-- Update the existing users table to match our new schema
ALTER TABLE users 
ADD COLUMN IF NOT EXISTS first_name VARCHAR(50),
ADD COLUMN IF NOT EXISTS last_name VARCHAR(50),
ADD COLUMN IF NOT EXISTS is_active BOOLEAN NOT NULL DEFAULT TRUE,
ADD COLUMN IF NOT EXISTS last_login_at TIMESTAMP;

-- Update the role column to include new roles
ALTER TABLE users ALTER COLUMN role TYPE VARCHAR(20);

-- Add indexes for better performance
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);
CREATE INDEX IF NOT EXISTS idx_users_is_active ON users(is_active);

-- Insert a default admin user (password: admin123)
INSERT INTO users (email, hashed_password, role, first_name, last_name, is_active) 
VALUES (
    'admin@v2x-siem.com', 
    '$2a$10$kS4cJ0FhFKR2QgL8oV6Kp.FPiXRXDmQ6j6gVhQnFxkHrX8nJ3P6JO', -- admin123
    'admin',
    'System',
    'Administrator',
    true
) ON CONFLICT (email) DO NOTHING;

-- +goose Down
-- Remove the columns we added
ALTER TABLE users 
DROP COLUMN IF EXISTS first_name,
DROP COLUMN IF EXISTS last_name,
DROP COLUMN IF EXISTS is_active,
DROP COLUMN IF EXISTS last_login_at;

-- Remove indexes
DROP INDEX IF EXISTS idx_users_email;
DROP INDEX IF EXISTS idx_users_role;
DROP INDEX IF EXISTS idx_users_is_active;

-- Remove the admin user
DELETE FROM users WHERE email = 'admin@v2x-siem.com';