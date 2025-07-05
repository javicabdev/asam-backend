-- Add email column to users table
ALTER TABLE users ADD COLUMN IF NOT EXISTS email VARCHAR(255);

-- Create unique index on email
CREATE UNIQUE INDEX IF NOT EXISTS uni_users_email ON users(email) WHERE email IS NOT NULL;
