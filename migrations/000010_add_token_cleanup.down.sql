-- migrations/000010_add_token_cleanup.down.sql
DROP FUNCTION IF EXISTS cleanup_expired_tokens();