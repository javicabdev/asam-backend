-- Script para marcar el email como verificado
-- Ejecutar en la base de datos PostgreSQL

UPDATE users 
SET 
    email_verified = true,
    email_verified_at = NOW()
WHERE 
    username = 'javierfernandezc@gmail.com';

-- Verificar el cambio
SELECT 
    id, 
    username, 
    email_verified, 
    email_verified_at 
FROM users 
WHERE username = 'javierfernandezc@gmail.com';
