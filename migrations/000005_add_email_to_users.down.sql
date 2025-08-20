-- Drop unique index on email
DROP INDEX IF EXISTS uni_users_email;

-- Remove email column from users table
ALTER TABLE users DROP COLUMN IF EXISTS email;
