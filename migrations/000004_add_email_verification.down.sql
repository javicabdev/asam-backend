-- Revert Migration: Add email verification and password reset support
-- Date: 2024-06-20
-- Description: Removes email verification fields and verification_tokens table

-- 1. Drop indexes
DROP INDEX IF EXISTS idx_verification_tokens_type_user;
DROP INDEX IF EXISTS idx_users_email_verified;
DROP INDEX IF EXISTS idx_verification_tokens_deleted_at;
DROP INDEX IF EXISTS idx_verification_tokens_expires_at;
DROP INDEX IF EXISTS idx_verification_tokens_token;
DROP INDEX IF EXISTS idx_verification_tokens_user_id;

-- 2. Drop trigger
DROP TRIGGER IF EXISTS update_verification_tokens_updated_at ON verification_tokens;

-- 3. Drop verification_tokens table
DROP TABLE IF EXISTS verification_tokens;

-- 4. Remove email verification fields from users table
ALTER TABLE users DROP COLUMN IF EXISTS email_verified_at;
ALTER TABLE users DROP COLUMN IF EXISTS email_verified;
