-- Revert Migration: Add email verification and password reset support
-- Date: 2024-06-20
-- Description: Removes email verification fields and verification_tokens table

-- 1. Drop verification_tokens table and all its dependencies
-- The CASCADE will automatically drop the trigger and all indexes
DROP TABLE IF EXISTS verification_tokens CASCADE;

-- 2. Drop remaining indexes on users table
DROP INDEX IF EXISTS idx_users_email_verified;

-- 3. Remove email verification fields from users table
ALTER TABLE users DROP COLUMN IF EXISTS email_verified_at;
ALTER TABLE users DROP COLUMN IF EXISTS email_verified;
