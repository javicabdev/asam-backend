-- Fix existing users without email
-- This script sets the email field to the username value for users that don't have an email

UPDATE users 
SET email = username 
WHERE email IS NULL OR email = '';

-- Verify the update
SELECT id, username, email, role, is_active, email_verified 
FROM users;
