-- migrations/000010_add_token_cleanup.up.sql

-- Crear función para limpiar tokens expirados
CREATE OR REPLACE FUNCTION cleanup_expired_tokens()
    RETURNS void AS $$
BEGIN
    DELETE FROM refresh_tokens
    WHERE expires_at < EXTRACT(EPOCH FROM NOW());
END;
$$ LANGUAGE plpgsql;