-- Script SQL para crear un usuario de prueba
-- Este script crea un usuario administrador con contraseña 'admin123'
-- IMPORTANTE: Si el login falla, ejecutar fix-passwords.ps1 para regenerar el hash

-- Insertar un usuario de prueba (la contraseña debe estar hasheada con bcrypt)
-- La contraseña 'admin123' hasheada con bcrypt (cost 10)
INSERT INTO users (username, password, role, is_active, created_at, updated_at)
VALUES (
    'admin@asam.org',
    '$2a$10$3bQXKBsekmOphw2DJQYgpOaVl5GWELfYjk0j5LzTc5FqMFKmgZALu',  -- 'admin123' hasheado
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
    '$2a$10$3bQXKBsekmOphw2DJQYgpOaVl5GWELfYjk0j5LzTc5FqMFKmgZALu',  -- 'admin123' hasheado
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
