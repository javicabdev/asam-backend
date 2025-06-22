-- Migration: Add email verification and password reset support
-- Date: 2024-06-20
-- Description: Adds fields for email verification and creates verification_tokens table

-- 1. Add email verification fields to users table
ALTER TABLE users ADD COLUMN email_verified BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE users ADD COLUMN email_verified_at TIMESTAMP NULL;

-- 2. Create verification_tokens table
CREATE TABLE verification_tokens (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    token VARCHAR(64) NOT NULL UNIQUE,
    user_id BIGINT UNSIGNED NOT NULL,
    type VARCHAR(20) NOT NULL,
    email VARCHAR(100) NOT NULL,
    used BOOLEAN NOT NULL DEFAULT FALSE,
    used_at TIMESTAMP NULL,
    expires_at TIMESTAMP NOT NULL,
    
    INDEX idx_user_id (user_id),
    INDEX idx_token (token),
    INDEX idx_expires_at (expires_at),
    INDEX idx_deleted_at (deleted_at),
    
    CONSTRAINT fk_verification_tokens_user FOREIGN KEY (user_id) 
        REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 3. Add indexes for better query performance
CREATE INDEX idx_users_email_verified ON users(email_verified);
CREATE INDEX idx_verification_tokens_type_user ON verification_tokens(type, user_id);

-- For PostgreSQL, use this instead:
-- CREATE TABLE verification_tokens (
--     id BIGSERIAL PRIMARY KEY,
--     created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
--     updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
--     deleted_at TIMESTAMP NULL,
--     
--     token VARCHAR(64) NOT NULL UNIQUE,
--     user_id BIGINT NOT NULL,
--     type VARCHAR(20) NOT NULL,
--     email VARCHAR(100) NOT NULL,
--     used BOOLEAN NOT NULL DEFAULT FALSE,
--     used_at TIMESTAMP NULL,
--     expires_at TIMESTAMP NOT NULL,
--     
--     CONSTRAINT fk_verification_tokens_user FOREIGN KEY (user_id) 
--         REFERENCES users(id) ON DELETE CASCADE
-- );
-- 
-- CREATE INDEX idx_verification_tokens_user_id ON verification_tokens(user_id);
-- CREATE INDEX idx_verification_tokens_token ON verification_tokens(token);
-- CREATE INDEX idx_verification_tokens_expires_at ON verification_tokens(expires_at);
-- CREATE INDEX idx_verification_tokens_deleted_at ON verification_tokens(deleted_at);
-- CREATE INDEX idx_verification_tokens_type_user ON verification_tokens(type, user_id);
