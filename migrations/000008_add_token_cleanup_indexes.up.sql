-- Add indexes for token cleanup optimization

-- Index for expired tokens cleanup
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_expires_at ON refresh_tokens(expires_at);

-- Index for user token count queries
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id_created_at ON refresh_tokens(user_id, created_at);

-- Index for last used tracking
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_last_used_at ON refresh_tokens(last_used_at);

-- Composite index for active sessions query (without partial index)
-- The query planner will use this index efficiently for queries filtering by user_id and expires_at
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_expires ON refresh_tokens(user_id, expires_at);
