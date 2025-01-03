-- migrations/000008_create_users.up.sql
CREATE TABLE users (
                       id SERIAL PRIMARY KEY,
                       username VARCHAR(100) NOT NULL UNIQUE,
                       password VARCHAR(255) NOT NULL,
                       role VARCHAR(20) NOT NULL DEFAULT 'user',
                       last_login TIMESTAMP,
                       is_active BOOLEAN NOT NULL DEFAULT true,
                       refresh_token VARCHAR(255),
                       created_at TIMESTAMP NOT NULL,
                       updated_at TIMESTAMP NOT NULL,
                       deleted_at TIMESTAMP
);

-- Crear usuario admin por defecto (password: admin123)
INSERT INTO users (username, password, role, is_active, created_at, updated_at)
VALUES (
           'admin',
           '$2a$10$ZVOrvWkK0C/TGB8k9LFmA.MJoaD1bkD9DgUG5cgBv99nKvIy0qS.S',
           'admin',
           true,
           CURRENT_TIMESTAMP,
           CURRENT_TIMESTAMP
       );