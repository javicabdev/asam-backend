-- Remove email indexes
DROP INDEX IF EXISTS idx_users_email_unique;
DROP INDEX IF EXISTS idx_users_email;
