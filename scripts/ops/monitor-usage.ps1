# Monitorea el uso de recursos y costos estimados
Write-Host "📊 Monitor de Uso - Cloud Run Free Tier" -ForegroundColor Cyan
Write-Host "=" * 50

# Obtener métricas del último mes
$project = "babacar-asam"
$endDate = Get-Date -Format "yyyy-MM-dd"
$startDate = (Get-Date).AddDays(-30).ToString("yyyy-MM-dd")

Write-Host ""
Write-Host "Período: $startDate a $endDate" -ForegroundColor Gray
Write-Host ""

# Verificar servicios activos
Write-Host "[Servicios Cloud Run Activos]" -ForegroundColor Yellow
$services = gcloud run services list --region=europe-west1 --format="table(SERVICE,LAST_DEPLOYED_BY,LAST_DEPLOYED_AT)"
Write-Output $services

Write-Host ""
Write-Host "[Límites Gratuitos Mensuales]" -ForegroundColor Yellow
Write-Host "  • Requests:        2,000,000 por mes" -ForegroundColor Gray
Write-Host "  • CPU:             180,000 vCPU-segundos" -ForegroundColor Gray  
Write-Host "  • Memoria:         360,000 GB-segundos" -ForegroundColor Gray
Write-Host "  • Ancho de banda:  1 GB saliente" -ForegroundColor Gray

Write-Host ""
Write-Host "[Estimación de Uso Actual]" -ForegroundColor Yellow

# Intentar obtener métricas (requiere APIs habilitadas)
$hasMetrics = gcloud services list --enabled --filter="name:monitoring.googleapis.com" --format="value(name)" 2>$null

if ($hasMetrics) {
    Write-Host "  Obteniendo métricas..." -ForegroundColor Gray
    
    # Aquí podrías agregar comandos específicos de métricas si están disponibles
    # Por ahora, mostrar estimación basada en configuración
    
    $prodInstances = gcloud run services describe asam-backend --region=europe-west1 --format="value(spec.template.metadata.annotations.'autoscaling.knative.dev/maxScale')" 2>$null
    $stagingInstances = gcloud run services describe asam-backend-staging --region=europe-west1 --format="value(spec.template.metadata.annotations.'autoscaling.knative.dev/maxScale')" 2>$null
    
    if ($prodInstances) {
        Write-Host "  • Production max instances: $prodInstances" -ForegroundColor Gray
    }
    if ($stagingInstances) {
        Write-Host "  • Staging max instances: $stagingInstances" -ForegroundColor Gray
    }
} else {
    Write-Host "  ℹ️  Métricas detalladas no disponibles" -ForegroundColor Gray
    Write-Host "  Habilita Cloud Monitoring API para ver uso real" -ForegroundColor Gray
}

Write-Host ""
Write-Host "[💡 Tips para Mantenerte en Free Tier]" -ForegroundColor Green
Write-Host "  1. Usa min-instances=0 (escala a cero sin tráfico)" -ForegroundColor Gray
Write-Host "  2. Limita max-instances (1-2 para proyectos pequeños)" -ForegroundColor Gray
Write-Host "  3. Usa memoria mínima necesaria (256MB-512MB)" -ForegroundColor Gray
Write-Host "  4. Apaga staging cuando no lo uses" -ForegroundColor Gray
Write-Host "  5. Configura alertas de billing en GCP Console" -ForegroundColor Gray

Write-Host ""
Write-Host "[🔗 Enlaces Útiles]" -ForegroundColor Cyan
Write-Host "  • Ver costos actuales:" -ForegroundColor Gray
Write-Host "    https://console.cloud.google.com/billing" -ForegroundColor Blue
Write-Host "  • Configurar alertas de presupuesto:" -ForegroundColor Gray
Write-Host "    https://console.cloud.google.com/billing/budgets" -ForegroundColor Blue
Write-Host "  • Calculadora de precios:" -ForegroundColor Gray
Write-Host "    https://cloud.google.com/products/calculator" -ForegroundColor Blue
