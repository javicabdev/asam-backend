# Script para ver el estado de todos los ambientes
Write-Host "=== Estado de Ambientes ASAM ===" -ForegroundColor Cyan
Write-Host ""

$environments = @(
    @{Name="Production"; Service="asam-backend"},
    @{Name="Staging"; Service="asam-backend-staging"}
)

foreach ($env in $environments) {
    Write-Host "[$($env.Name)]" -ForegroundColor Yellow
    
    $service = gcloud run services describe $env.Service --region=europe-west1 --format=json 2>$null | ConvertFrom-Json
    
    if ($service) {
        $image = $service.spec.template.spec.containers[0].image
        $tag = $image.Split(":")[-1]
        $url = $service.status.url
        $lastModified = $service.metadata.annotations.'serving.knative.dev/lastModifier'
        $createdTime = $service.metadata.creationTimestamp
        
        # Determinar color basado en el tag
        $tagColor = "Green"
        if ($tag -eq "latest") {
            if ($env.Name -eq "Production") {
                $tagColor = "Red"
            } else {
                $tagColor = "Yellow"
            }
        }
        
        Write-Host "  URL: $url" -ForegroundColor Cyan
        Write-Host "  Tag: $tag" -ForegroundColor $tagColor
        Write-Host "  Image: $image" -ForegroundColor Gray
        Write-Host "  Last Modified: $lastModified" -ForegroundColor Gray
        Write-Host "  Created: $createdTime" -ForegroundColor Gray
        
        # Health check
        try {
            $health = Invoke-RestMethod -Uri "$url/health" -TimeoutSec 5 -ErrorAction SilentlyContinue
            Write-Host "  Health: ✅ OK" -ForegroundColor Green
        } catch {
            Write-Host "  Health: ⚠️  Unable to check" -ForegroundColor Yellow
        }
        
        if ($env.Name -eq "Production" -and $tag -eq "latest") {
            Write-Host "  ⚠️  WARNING: Production using 'latest' - This should be fixed!" -ForegroundColor Red
        }
    } else {
        Write-Host "  ❌ Not deployed" -ForegroundColor DarkGray
        Write-Host "  To deploy staging, use:" -ForegroundColor Gray
        Write-Host "    gcloud run deploy $($env.Service) --image=gcr.io/babacar-asam/asam-backend:latest --region=europe-west1" -ForegroundColor DarkGray
    }
    
    Write-Host ""
}

# Mostrar imágenes disponibles
Write-Host "[Available Images]" -ForegroundColor Yellow
Write-Host "Recent images in registry:" -ForegroundColor Gray
gcloud container images list-tags gcr.io/babacar-asam/asam-backend `
    --limit=10 `
    --format="table(tags,timestamp.datetime)" `
    --sort-by="~timestamp"

Write-Host ""
Write-Host "[Quick Actions]" -ForegroundColor Yellow
Write-Host "• Validate deployment:  .\scripts\ops\validate-deployment.ps1 -Environment <staging|production> -ImageTag <tag>" -ForegroundColor Cyan
Write-Host "• View logs:           gcloud run logs tail <service-name> --region=europe-west1" -ForegroundColor Cyan
Write-Host "• Deploy via GitHub:   https://github.com/[your-repo]/asam-backend/actions/workflows/cloud-run-deploy.yml" -ForegroundColor Cyan
