# Script completo para aplicar los cambios de email obligatorio

Write-Host "=== Aplicando cambios para hacer el email obligatorio ===" -ForegroundColor Cyan

# 1. Actualizar los usuarios existentes con el email correcto
Write-Host "`n1. Actualizando emails de usuarios existentes..." -ForegroundColor Yellow
docker exec -i asam-postgres psql -U postgres -d asam_db << 'EOF'
-- Update existing test users to have the correct email
UPDATE users 
SET email = 'javierfernandezc@gmail.com',
    updated_at = NOW()
WHERE username IN ('admin@asam.org', 'user@asam.org')
  AND (email IS NULL OR email = '');

-- Show results
SELECT id, username, email, email_verified 
FROM users 
WHERE username IN ('admin@asam.org', 'user@asam.org');
EOF

# 2. Ejecutar la migración
Write-Host "`n2. Ejecutando migración para hacer el campo email obligatorio..." -ForegroundColor Yellow

# Crear archivo temporal con la migración
$migrationContent = @"
-- Make email column required (NOT NULL) in users table

-- First, update any NULL emails to use username as default
UPDATE users 
SET email = username 
WHERE email IS NULL;

-- Now alter the column to be NOT NULL
ALTER TABLE users ALTER COLUMN email SET NOT NULL;

-- Add comment
COMMENT ON COLUMN users.email IS 'User email address, required for notifications and password reset';
"@

$migrationContent | docker exec -i asam-postgres psql -U postgres -d asam_db

if ($LASTEXITCODE -eq 0) {
    Write-Host "`n✅ Cambios aplicados exitosamente!" -ForegroundColor Green
    Write-Host "El campo email ahora es obligatorio y los usuarios de prueba tienen el email correcto." -ForegroundColor Cyan
    
    # Mostrar el estado actual
    Write-Host "`nEstado actual de los usuarios:" -ForegroundColor Yellow
    docker exec -i asam-postgres psql -U postgres -d asam_db -c "SELECT id, username, email, email_verified FROM users ORDER BY id;"
} else {
    Write-Host "`n❌ Error al aplicar los cambios" -ForegroundColor Red
}

Write-Host "`n3. Reiniciando la aplicación..." -ForegroundColor Yellow
docker restart asam-backend-api

Write-Host "`nProceso completado. Ahora puedes intentar enviar emails de verificación." -ForegroundColor Green
