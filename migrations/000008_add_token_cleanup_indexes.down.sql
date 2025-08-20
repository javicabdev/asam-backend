-- Remove token cleanup optimization indexes

DROP INDEX IF EXISTS idx_refresh_tokens_expires_at;
DROP INDEX IF EXISTS idx_refresh_tokens_user_id_created_at;
DROP INDEX IF EXISTS idx_refresh_tokens_last_used_at;
DROP INDEX IF EXISTS idx_refresh_tokens_user_expires;
