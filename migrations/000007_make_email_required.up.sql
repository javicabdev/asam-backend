-- Make email column required (NOT NULL) in users table

-- First, update any NULL emails to use username as default
UPDATE users 
SET email = username 
WHERE email IS NULL;

-- Now alter the column to be NOT NULL
ALTER TABLE users ALTER COLUMN email SET NOT NULL;

-- Add comment
COMMENT ON COLUMN users.email IS 'User email address, required for notifications and password reset';
