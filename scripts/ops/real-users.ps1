# Script para cargar usuarios reales desde archivos locales
# Los datos confidenciales se leen de C:\repos\asam\asam-backend\docsTemp\usuarios
# El archivo debe contener los usuarios, guardados de este modo:
# Formato: tipo,nombre,email
# tipo puede ser: admin o user

param(
    [Parameter(Position=0)]
    [ValidateSet("load", "check", "reset")]
    [string]$Action = "check"
)

Write-Host "🔐 Gestión de Usuarios Reales - ASAM" -ForegroundColor Cyan
Write-Host "=" * 50
Write-Host ""

$dataPath = "C:\repos\asam\asam-backend\docsTemp\usuarios"
$apiUrl = "https://asam-backend-jtpswzdxuq-ew.a.run.app/graphql"

# Función para generar contraseñas seguras
function New-SecurePassword {
    $length = 12
    $chars = "ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz23456789!@#$%"
    $password = ""
    $random = New-Object System.Random
    for ($i = 0; $i -lt $length; $i++) {
        $password += $chars[$random.Next($chars.Length)]
    }
    return $password
}

switch ($Action) {
    "check" {
        Write-Host "[Verificando Configuración]" -ForegroundColor Yellow
        
        # Verificar que existe la carpeta de datos
        if (Test-Path $dataPath) {
            Write-Host "✅ Carpeta de datos encontrada: $dataPath" -ForegroundColor Green
            
            # Listar archivos disponibles
            $files = Get-ChildItem -Path $dataPath -Filter "*.txt" 2>$null
            if ($files) {
                Write-Host ""
                Write-Host "Archivos disponibles:" -ForegroundColor Cyan
                $files | ForEach-Object {
                    Write-Host "  - $($_.Name)" -ForegroundColor Gray
                }
            } else {
                Write-Host "⚠️ No hay archivos .txt en la carpeta" -ForegroundColor Yellow
            }
        } else {
            Write-Host "❌ No existe la carpeta: $dataPath" -ForegroundColor Red
            Write-Host ""
            Write-Host "Creando estructura..." -ForegroundColor Yellow
            New-Item -ItemType Directory -Path $dataPath -Force | Out-Null
            
            # Crear archivo de ejemplo
            $exampleContent = @"
# Archivo de ejemplo para usuarios
# Formato: tipo,nombre,email
admin,babacar,mmbaye@hotmail.com
admin,javiAdmin,javierfernandezc@gmail.com
user,javiUser,javi_nov20@gmail.com
"@
            $exampleContent | Out-File -FilePath "$dataPath\usuarios.txt" -Encoding UTF8
            Write-Host "✅ Creado archivo de ejemplo: usuarios.txt" -ForegroundColor Green
        }
        
        Write-Host ""
        Write-Host "[Estado del Backend]" -ForegroundColor Yellow
        try {
            $response = Invoke-WebRequest -Uri "$apiUrl" -Method GET -TimeoutSec 5
            Write-Host "✅ Backend accesible" -ForegroundColor Green
        } catch {
            Write-Host "⚠️ No se puede acceder al backend" -ForegroundColor Yellow
        }
    }
    
    "load" {
        Write-Host "[Cargando Usuarios Reales]" -ForegroundColor Yellow
        Write-Host ""
        
        # Leer archivo de usuarios
        $usuariosFile = "$dataPath\usuarios.txt"
        
        if (-not (Test-Path $usuariosFile)) {
            Write-Host "❌ No existe el archivo: $usuariosFile" -ForegroundColor Red
            Write-Host "Ejecuta primero: .\real-users.ps1 check" -ForegroundColor Yellow
            exit 1
        }
        
        # Parsear usuarios
        $usuarios = @()
        $credentials = @()
        
        Get-Content $usuariosFile | ForEach-Object {
            if ($_ -and -not $_.StartsWith("#")) {
                $parts = $_.Split(",")
                if ($parts.Length -eq 3) {
                    $password = New-SecurePassword
                    $usuario = @{
                        Tipo = $parts[0].Trim()
                        Nombre = $parts[1].Trim()
                        Email = $parts[2].Trim()
                        Password = $password
                    }
                    $usuarios += $usuario
                    $credentials += "$($usuario.Email): $password"
                }
            }
        }
        
        if ($usuarios.Count -eq 0) {
            Write-Host "❌ No se encontraron usuarios válidos en el archivo" -ForegroundColor Red
            exit 1
        }
        
        Write-Host "Usuarios a crear:" -ForegroundColor Cyan
        $usuarios | ForEach-Object {
            Write-Host "  - $($_.Nombre) ($($_.Email)) - Rol: $($_.Tipo.ToUpper())" -ForegroundColor Gray
        }
        
        Write-Host ""
        $confirm = Read-Host "¿Crear estos usuarios? (s/n)"
        
        if ($confirm -ne "s" -and $confirm -ne "S") {
            Write-Host "Cancelado" -ForegroundColor Yellow
            return
        }
        
        Write-Host ""
        Write-Host "Creando usuarios..." -ForegroundColor Yellow
        
        # Aquí necesitaríamos hacer las llamadas GraphQL
        # Por seguridad, mostramos las mutations que se ejecutarían
        
        Write-Host ""
        Write-Host "Mutations GraphQL a ejecutar:" -ForegroundColor Cyan
        Write-Host ""
        
        $mutations = @()
        $memberMutation = ""
        
        foreach ($user in $usuarios) {
            $role = if ($user.Tipo -eq "admin") { "ADMIN" } else { "USER" }
            
            # Si es USER, primero crear el miembro
            if ($role -eq "USER") {
                $memberMutation = @"
# Primero crear el miembro para $($user.Email)
mutation CreateMemberFor_$($user.Nombre) {
  createMember(input: {
    firstName: "$($user.Nombre)"
    lastName: "TestUser"
    dni: "TEST$(Get-Random -Minimum 10000000 -Maximum 99999999)"
    email: "$($user.Email)"
    phone: "600$(Get-Random -Minimum 100000 -Maximum 999999)"
    memberNumber: "M-$(Get-Date -Format 'yyyyMMdd')-$(Get-Random -Minimum 100 -Maximum 999)"
    isActive: true
  }) {
    id
    email
  }
}
"@
                Write-Host $memberMutation -ForegroundColor Green
                Write-Host ""
            }
            
            $mutation = @"
mutation CreateUser_$($user.Nombre) {
  createUser(input: {
    email: "$($user.Email)"
    password: "$($user.Password)"
    role: $role
    firstName: "$($user.Nombre)"
    lastName: "$(if ($role -eq 'ADMIN') { 'Admin' } else { 'User' })"
  }) {
    id
    email
    role
  }
}
"@
            Write-Host $mutation -ForegroundColor Green
            Write-Host ""
        }
        
        # Guardar credenciales en archivo local seguro
        $credentialsFile = "$dataPath\credenciales_$(Get-Date -Format 'yyyyMMdd_HHmmss').txt"
        
        $credentialsContent = @"
===========================================
CREDENCIALES ASAM - $(Get-Date)
===========================================
¡IMPORTANTE! Guarda estas contraseñas de forma segura
No se pueden recuperar después

ADMINISTRADORES:
"@
        
        foreach ($user in $usuarios | Where-Object { $_.Tipo -eq "admin" }) {
            $credentialsContent += "`n$($user.Email): $($user.Password)"
        }
        
        $credentialsContent += "`n`nUSUARIOS NORMALES:"
        
        foreach ($user in $usuarios | Where-Object { $_.Tipo -eq "user" }) {
            $credentialsContent += "`n$($user.Email): $($user.Password)"
        }
        
        $credentialsContent += @"

===========================================
NOTA: Este archivo contiene información confidencial
Elimínalo después de compartir las contraseñas
===========================================
"@
        
        $credentialsContent | Out-File -FilePath $credentialsFile -Encoding UTF8
        
        Write-Host "=" * 50 -ForegroundColor Yellow
        Write-Host "📝 CREDENCIALES GENERADAS" -ForegroundColor Cyan
        Write-Host "=" * 50 -ForegroundColor Yellow
        Write-Host ""
        Write-Host "ADMINS:" -ForegroundColor Yellow
        $usuarios | Where-Object { $_.Tipo -eq "admin" } | ForEach-Object {
            Write-Host "  Email: $($_.Email)" -ForegroundColor Gray
            Write-Host "  Pass:  $($_.Password)" -ForegroundColor Green
            Write-Host ""
        }
        
        Write-Host "USERS:" -ForegroundColor Yellow
        $usuarios | Where-Object { $_.Tipo -eq "user" } | ForEach-Object {
            Write-Host "  Email: $($_.Email)" -ForegroundColor Gray
            Write-Host "  Pass:  $($_.Password)" -ForegroundColor Green
            Write-Host ""
        }
        
        Write-Host "=" * 50 -ForegroundColor Yellow
        Write-Host ""
        Write-Host "✅ Credenciales guardadas en:" -ForegroundColor Green
        Write-Host "   $credentialsFile" -ForegroundColor Cyan
        Write-Host ""
        Write-Host "⚠️ IMPORTANTE:" -ForegroundColor Yellow
        Write-Host "  1. Copia las mutations anteriores" -ForegroundColor Gray
        Write-Host "  2. Pégalas en GraphQL Playground:" -ForegroundColor Gray
        Write-Host "     $apiUrl" -ForegroundColor Cyan
        Write-Host "  3. Ejecuta cada mutation una por una" -ForegroundColor Gray
        Write-Host "  4. Para usuarios tipo USER, ejecuta PRIMERO la mutation del miembro" -ForegroundColor Red
        Write-Host "  5. Guarda las contraseñas de forma segura" -ForegroundColor Gray
        Write-Host "  6. Elimina el archivo de credenciales después" -ForegroundColor Gray
    }
    
    "reset" {
        Write-Host "[Reset de Usuarios]" -ForegroundColor Red
        Write-Host ""
        Write-Host "⚠️  Esto borrará TODOS los usuarios de prueba" -ForegroundColor Yellow
        Write-Host ""
        
        $confirm = Read-Host "¿Estás seguro? (escribe 'BORRAR' para confirmar)"
        
        if ($confirm -ne "BORRAR") {
            Write-Host "Cancelado" -ForegroundColor Yellow
            return
        }
        
        Write-Host ""
        Write-Host "Mutation para borrar usuarios de prueba:" -ForegroundColor Yellow
        Write-Host @"
        
mutation DeleteTestUsers {
  deleteUsers(where: { 
    OR: [
      { email: { contains: "@hotmail.com" } },
      { email: { contains: "@gmail.com" } }
    ]
  }) {
    count
  }
}

mutation DeleteTestMembers {
  deleteMembers(where: { 
    firstName: { contains: "Test" }
  }) {
    count
  }
}
"@ -ForegroundColor Red
        
        Write-Host ""
        Write-Host "Ejecuta estas mutations en: $apiUrl" -ForegroundColor Cyan
    }
}

Write-Host ""
Write-Host "[Comandos Disponibles]" -ForegroundColor Cyan
Write-Host "  .\real-users.ps1 check  - Verificar configuración" -ForegroundColor Gray
Write-Host "  .\real-users.ps1 load   - Cargar usuarios desde archivo" -ForegroundColor Gray
Write-Host "  .\real-users.ps1 reset  - Borrar usuarios de prueba" -ForegroundColor Gray
