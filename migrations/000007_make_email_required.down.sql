-- Revert: Make email column optional (NULL) in users table

-- Remove the NOT NULL constraint
ALTER TABLE users ALTER COLUMN email DROP NOT NULL;

-- Remove comment
COMMENT ON COLUMN users.email IS NULL;
