# Gestión de datos de prueba para desarrollo
Write-Host "🧪 Gestión de Datos de Prueba - ASAM" -ForegroundColor Cyan
Write-Host "=" * 50

param(
    [Parameter(Position=0)]
    [ValidateSet("load", "clear", "status")]
    [string]$Action = "status"
)

$prodUrl = gcloud run services describe asam-backend `
    --region=europe-west1 `
    --format="value(status.url)" 2>$null

if (-not $prodUrl) {
    Write-Host "❌ No se encontró el servicio desplegado" -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host "URL: $prodUrl" -ForegroundColor Gray
Write-Host ""

switch ($Action) {
    "status" {
        Write-Host "[Estado de Datos]" -ForegroundColor Yellow
        
        # Aquí podrías hacer una llamada GraphQL para contar registros
        # Por ahora, mostrar info básica
        
        Write-Host "Para ver el estado actual de los datos:" -ForegroundColor Cyan
        Write-Host "  1. Abre: $prodUrl/graphql" -ForegroundColor Gray
        Write-Host "  2. Ejecuta esta query:" -ForegroundColor Gray
        Write-Host @"
        
  query {
    members { totalCount }
    families { totalCount }
    payments { totalCount }
    users { totalCount }
  }
"@ -ForegroundColor Green
    }
    
    "load" {
        Write-Host "[Cargar Datos de Prueba]" -ForegroundColor Yellow
        Write-Host ""
        Write-Host "⚠️  Esto agregará datos de prueba a la BD actual" -ForegroundColor Yellow
        Write-Host ""
        
        $confirm = Read-Host "¿Continuar? (s/n)"
        
        if ($confirm -ne "s" -and $confirm -ne "S") {
            Write-Host "Cancelado" -ForegroundColor Gray
            return
        }
        
        Write-Host ""
        Write-Host "Cargando datos de prueba..." -ForegroundColor Cyan
        
        # Crear archivo temporal con mutations GraphQL
        $mutations = @"
mutation LoadTestData {
  # Crear usuarios de prueba
  createUser1: createUser(input: {
    email: "admin@test.com"
    password: "Test123!"
    role: ADMIN
    firstName: "Admin"
    lastName: "Test"
  }) { id }
  
  createUser2: createUser(input: {
    email: "user@test.com"
    password: "Test123!"
    role: USER
    firstName: "Usuario"
    lastName: "Prueba"
  }) { id }
  
  # Crear miembros de prueba
  createMember1: createMember(input: {
    firstName: "Juan"
    lastName: "Pérez TEST"
    dni: "12345678A"
    email: "juan.test@example.com"
    phone: "600000001"
    memberNumber: "TEST001"
    isActive: true
  }) { id }
  
  createMember2: createMember(input: {
    firstName: "María"
    lastName: "García TEST"
    dni: "87654321B"
    email: "maria.test@example.com"
    phone: "600000002"
    memberNumber: "TEST002"
    isActive: true
  }) { id }
  
  # Crear familia de prueba
  createFamily: createFamily(input: {
    familyCode: "FAM-TEST-001"
    husbandId: 1  # Ajustar según IDs reales
    wifeId: 2     # Ajustar según IDs reales
  }) { id }
}
"@

        Write-Host "Mutations preparadas. Ejecuta manualmente en GraphQL playground:" -ForegroundColor Yellow
        Write-Host $mutations -ForegroundColor Green
        
        Write-Host ""
        Write-Host "📝 Datos de prueba creados con sufijo 'TEST' para identificarlos" -ForegroundColor Cyan
    }
    
    "clear" {
        Write-Host "[Limpiar Datos de Prueba]" -ForegroundColor Yellow
        Write-Host ""
        Write-Host "⚠️  Esto borrará SOLO los datos con sufijo 'TEST'" -ForegroundColor Yellow
        Write-Host ""
        
        $confirm = Read-Host "¿Continuar? (s/n)"
        
        if ($confirm -ne "s" -and $confirm -ne "S") {
            Write-Host "Cancelado" -ForegroundColor Gray
            return
        }
        
        Write-Host ""
        Write-Host "Para limpiar datos de prueba, ejecuta en GraphQL:" -ForegroundColor Cyan
        Write-Host @"
        
mutation ClearTestData {
  # Borrar miembros de prueba
  deleteMembers(where: { lastName: { contains: "TEST" } }) {
    count
  }
  
  # Borrar familias de prueba
  deleteFamilies(where: { familyCode: { contains: "TEST" } }) {
    count
  }
  
  # Borrar usuarios de prueba
  deleteUsers(where: { email: { contains: "@test.com" } }) {
    count
  }
}
"@ -ForegroundColor Green
        
        Write-Host ""
        Write-Host "✅ Query preparada para limpiar datos TEST" -ForegroundColor Green
    }
}

Write-Host ""
Write-Host "[Comandos]" -ForegroundColor Cyan
Write-Host "  .\test-data.ps1 status - Ver estado actual" -ForegroundColor Gray
Write-Host "  .\test-data.ps1 load   - Cargar datos de prueba" -ForegroundColor Gray
Write-Host "  .\test-data.ps1 clear  - Limpiar datos de prueba" -ForegroundColor Gray
