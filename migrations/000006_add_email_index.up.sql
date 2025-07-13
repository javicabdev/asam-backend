-- Add unique index to email field for efficient lookups
-- This migration adds an index to the email field in the users table

-- First, let's check for any duplicate emails (excluding empty strings)
DO $$
BEGIN
    IF EXISTS (
        SELECT email, COUNT(*) 
        FROM users 
        WHERE email IS NOT NULL AND email != ''
        GROUP BY email 
        HAVING COUNT(*) > 1
    ) THEN
        RAISE EXCEPTION 'Duplicate emails found. Please resolve duplicates before adding unique constraint.';
    END IF;
END
$$;

-- Create a partial unique index that excludes NULL and empty string values
-- This allows multiple users to have no email, but ensures uniqueness when email is provided
CREATE UNIQUE INDEX idx_users_email_unique 
ON users (email) 
WHERE email IS NOT NULL AND email != '';

-- Add a regular index for performance on email lookups
CREATE INDEX IF NOT EXISTS idx_users_email 
ON users (email);

-- Add comment to explain the index
COMMENT ON INDEX idx_users_email_unique IS 'Ensures email uniqueness while allowing NULL or empty values';
COMMENT ON INDEX idx_users_email IS 'Improves performance for email-based searches';
