-- migrations/000009_create_refresh_tokens.up.sql
CREATE TABLE refresh_tokens (
                                uuid VARCHAR(36) PRIMARY KEY,
                                user_id INTEGER NOT NULL REFERENCES users(id),
                                expires_at BIGINT NOT NULL,
                                created_at TIMESTAMP NOT NULL
);

-- Crear índices
CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX idx_refresh_tokens_expires ON refresh_tokens(expires_at);