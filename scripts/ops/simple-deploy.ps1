# Deployment simplificado para proyecto ASAM
# Un solo ambiente (production) con gestión simple de versiones

param(
    [Parameter(Position=0)]
    [ValidateSet("deploy", "status", "rollback", "logs")]
    [string]$Action = "status"
)

Write-Host "🚀 ASAM Deployment Manager" -ForegroundColor Cyan
Write-Host "=" * 50

switch ($Action) {
    "status" {
        Write-Host ""
        Write-Host "[Estado Actual]" -ForegroundColor Yellow
        
        # Obtener info del servicio
        $service = gcloud run services describe asam-backend `
            --region=europe-west1 `
            --format=json 2>$null | ConvertFrom-Json
            
        if ($service) {
            $image = $service.spec.template.spec.containers[0].image
            $tag = $image.Split(":")[-1]
            $url = $service.status.url
            
            Write-Host "  URL: $url" -ForegroundColor Green
            Write-Host "  Tag: $tag" -ForegroundColor Gray
            Write-Host "  Image: $image" -ForegroundColor Gray
            
            # Health check
            try {
                $response = Invoke-WebRequest -Uri "$url/health" -TimeoutSec 5
                if ($response.StatusCode -eq 200) {
                    Write-Host "  Health: ✅ OK" -ForegroundColor Green
                }
            } catch {
                Write-Host "  Health: ⚠️ No responde" -ForegroundColor Yellow
            }
        } else {
            Write-Host "  ❌ Servicio no encontrado" -ForegroundColor Red
        }
        
        Write-Host ""
        Write-Host "[Últimas Versiones]" -ForegroundColor Yellow
        git tag -l "v*" --sort=-version:refname | Select-Object -First 5 | ForEach-Object {
            Write-Host "  $_" -ForegroundColor Gray
        }
    }
    
    "deploy" {
        Write-Host ""
        Write-Host "[Deploy a Production]" -ForegroundColor Yellow
        Write-Host ""
        
        # Mostrar versiones disponibles
        Write-Host "Versiones disponibles:" -ForegroundColor Cyan
        $tags = git tag -l "v*" --sort=-version:refname | Select-Object -First 10
        $i = 1
        $tags | ForEach-Object {
            Write-Host "  $i. $_" -ForegroundColor Gray
            $i++
        }
        Write-Host "  L. latest (desarrollo activo)" -ForegroundColor Yellow
        Write-Host ""
        
        $selection = Read-Host "Selecciona versión (número o 'L' para latest)"
        
        if ($selection -eq "L" -or $selection -eq "l") {
            $version = "latest"
            Write-Host ""
            Write-Host "⚠️  Usando 'latest' - Solo para desarrollo" -ForegroundColor Yellow
        } else {
            $version = $tags[$selection - 1]
        }
        
        if (-not $version) {
            Write-Host "❌ Selección inválida" -ForegroundColor Red
            return
        }
        
        Write-Host ""
        Write-Host "📦 Desplegando versión: $version" -ForegroundColor Cyan
        Write-Host ""
        
        # Hacer backup antes de deploy
        Write-Host "¿Hacer backup antes del deploy? (recomendado) (s/n)" -ForegroundColor Yellow
        $backup = Read-Host
        
        if ($backup -eq "s" -or $backup -eq "S") {
            & ".\scripts\ops\backup-database.ps1"
            Write-Host ""
        }
        
        # Deploy
        Write-Host "Desplegando..." -ForegroundColor Yellow
        
        # Quitar la 'v' del tag si existe (Git usa v0.2.0, Docker usa 0.2.0)
        $dockerTag = $version
        if ($dockerTag -match "^v[0-9]") {
            $dockerTag = $dockerTag.Substring(1)
            Write-Host "  Usando tag Docker: $dockerTag (Git tag: $version)" -ForegroundColor Gray
        }
        
        gcloud run deploy asam-backend `
            --image="gcr.io/babacar-asam/asam-backend:$dockerTag" `
            --region=europe-west1 `
            --platform=managed `
            --memory=512Mi `
            --min-instances=0 `
            --max-instances=2 `
            --allow-unauthenticated `
            --port=8080 `
            --set-env-vars="ENVIRONMENT=production" `
            --set-env-vars="DB_SSL_MODE=require" `
            --set-env-vars="JWT_ACCESS_TTL=15m" `
            --set-env-vars="JWT_REFRESH_TTL=168h" `
            --set-env-vars="MAILERSEND_FROM_EMAIL=noreply@asam.org" `
            --set-env-vars="MAILERSEND_FROM_NAME=ASAM" `
            --set-env-vars="RATE_LIMIT_RPS=10" `
            --set-env-vars="RATE_LIMIT_BURST=20" `
            --set-env-vars="LOG_SLOW_QUERIES=true" `
            --set-env-vars="SLOW_QUERY_THRESHOLD=100ms" `
            --set-env-vars="LOG_SLOW_RESOLVERS=true" `
            --set-env-vars="SLOW_RESOLVER_THRESHOLD=100ms" `
            --set-env-vars="GQL_COMPLEXITY_LIMIT=1000" `
            --set-env-vars="GQL_CONCURRENT_RESOLVERS=10" `
            --set-secrets="DB_HOST=db-host:latest" `
            --set-secrets="DB_PORT=db-port:latest" `
            --set-secrets="DB_USER=db-user:latest" `
            --set-secrets="DB_PASSWORD=db-password:latest" `
            --set-secrets="DB_NAME=db-name:latest" `
            --set-secrets="JWT_ACCESS_SECRET=jwt-access-secret:latest" `
            --set-secrets="JWT_REFRESH_SECRET=jwt-refresh-secret:latest" `
            --set-secrets="ADMIN_USER=admin-user:latest" `
            --set-secrets="ADMIN_PASSWORD=admin-password:latest" `
            --set-secrets="MAILERSEND_API_KEY=mailersend-api-key:latest"
            
        if ($LASTEXITCODE -eq 0) {
            Write-Host ""
            Write-Host "✅ Deploy exitoso!" -ForegroundColor Green
            
            # Verificar health
            Start-Sleep -Seconds 5
            $url = gcloud run services describe asam-backend `
                --region=europe-west1 `
                --format="value(status.url)"
            
            Write-Host "Verificando health..." -ForegroundColor Yellow
            try {
                $response = Invoke-WebRequest -Uri "$url/health" -TimeoutSec 10
                Write-Host "✅ Aplicación respondiendo correctamente" -ForegroundColor Green
            } catch {
                Write-Host "⚠️  La aplicación puede tardar unos segundos en estar lista" -ForegroundColor Yellow
            }
            
            Write-Host ""
            Write-Host "URL: $url" -ForegroundColor Cyan
        } else {
            Write-Host "❌ Error en el deploy" -ForegroundColor Red
        }
    }
    
    "rollback" {
        Write-Host ""
        Write-Host "[Rollback]" -ForegroundColor Red
        Write-Host ""
        
        # Listar revisiones anteriores
        Write-Host "Revisiones anteriores:" -ForegroundColor Yellow
        $revisions = gcloud run revisions list `
            --service=asam-backend `
            --region=europe-west1 `
            --format="table(metadata.name,metadata.creationTimestamp)" `
            --limit=5
        
        Write-Output $revisions
        Write-Host ""
        
        $revisionName = Read-Host "Nombre de la revisión a restaurar"
        
        if ($revisionName) {
            Write-Host "Restaurando a $revisionName..." -ForegroundColor Yellow
            
            gcloud run services update-traffic asam-backend `
                --region=europe-west1 `
                --to-revisions="$revisionName=100"
                
            if ($LASTEXITCODE -eq 0) {
                Write-Host "✅ Rollback exitoso" -ForegroundColor Green
            } else {
                Write-Host "❌ Error en rollback" -ForegroundColor Red
            }
        }
    }
    
    "logs" {
        Write-Host ""
        Write-Host "[Logs en Tiempo Real]" -ForegroundColor Yellow
        Write-Host "Presiona Ctrl+C para salir" -ForegroundColor Gray
        Write-Host ""
        
        gcloud run logs tail asam-backend --region=europe-west1
    }
}

Write-Host ""
Write-Host "[Comandos Disponibles]" -ForegroundColor Cyan
Write-Host "  .\simple-deploy.ps1 status   - Ver estado actual" -ForegroundColor Gray
Write-Host "  .\simple-deploy.ps1 deploy   - Hacer deploy" -ForegroundColor Gray
Write-Host "  .\simple-deploy.ps1 rollback - Volver a versión anterior" -ForegroundColor Gray
Write-Host "  .\simple-deploy.ps1 logs     - Ver logs en tiempo real" -ForegroundColor Gray
