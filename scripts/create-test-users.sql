-- Script SQL para crear un usuario de prueba
-- Este script crea un usuario administrador con contraseña 'admin123'
-- Insertar un usuario de prueba (la contraseña debe estar hasheada con bcrypt)
-- La contraseña 'admin123' hasheada con bcrypt (cost 10)
INSERT INTO users (username, password, role, is_active, created_at, updated_at)
VALUES (
    'admin@asam.org',
    '$2a$10$K1kCTLS6VJ9U1lhH8hfste1Z7cUB7SvQH3fFtE3AqLYJrQ3GyqIKG',  -- 'admin123' hasheado CORRECTO
    'admin',  -- Changed from 'ADMIN' to 'admin'
    true,
    NOW(),
    NOW()
) ON CONFLICT (username) DO UPDATE 
SET 
    password = EXCLUDED.password,
    role = EXCLUDED.role,
    is_active = EXCLUDED.is_active,
    updated_at = NOW();

-- También crear un usuario regular para pruebas
INSERT INTO users (username, password, role, is_active, created_at, updated_at)
VALUES (
    'user@asam.org',
    '$2a$10$K1kCTLS6VJ9U1lhH8hfste1Z7cUB7SvQH3fFtE3AqLYJrQ3GyqIKG',  -- 'admin123' hasheado CORRECTO
    'user',  -- Changed from 'USER' to 'user'
    true,
    NOW(),
    NOW()
) ON CONFLICT (username) DO UPDATE 
SET 
    password = EXCLUDED.password,
    role = EXCLUDED.role,
    is_active = EXCLUDED.is_active,
    updated_at = NOW();

-- Verificar que los usuarios se crearon
SELECT username, role, is_active FROM users WHERE username IN ('admin@asam.org', 'user@asam.org');
