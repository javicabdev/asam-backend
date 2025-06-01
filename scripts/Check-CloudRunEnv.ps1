# PowerShell script to check current environment variables in Cloud Run

param(
    [string]$ServiceName = "asam-backend",
    [string]$Region = "europe-west1"
)

Write-Host "Checking environment variables for $ServiceName..." -ForegroundColor Yellow
Write-Host "=================================================" -ForegroundColor Yellow

# Get the service configuration
$serviceConfig = gcloud run services describe $ServiceName `
    --region=$Region `
    --format="json" | ConvertFrom-Json

# Extract environment variables
$envVars = $serviceConfig.spec.template.spec.containers[0].env

if ($envVars) {
    foreach ($var in $envVars) {
        $name = $var.name
        $value = $var.value
        
        # Mask sensitive values
        if ($name -match "(PASSWORD|SECRET|KEY)") {
            if ([string]::IsNullOrEmpty($value)) {
                Write-Host "$name : <not set>" -ForegroundColor Red
            } else {
                Write-Host "$name : <set - masked>" -ForegroundColor Green
            }
        } else {
            Write-Host "$name : $value"
        }
    }
} else {
    Write-Host "No environment variables found!" -ForegroundColor Red
}

Write-Host "`n=================================================" -ForegroundColor Yellow
Write-Host "`nRequired variables status:" -ForegroundColor Cyan

# Check for required variables
$requiredVars = @("ENVIRONMENT", "ADMIN_USER", "ADMIN_PASSWORD", "JWT_ACCESS_SECRET", "JWT_REFRESH_SECRET")
$missingVars = @()

foreach ($reqVar in $requiredVars) {
    $found = $false
    if ($envVars) {
        foreach ($var in $envVars) {
            if ($var.name -eq $reqVar) {
                $found = $true
                break
            }
        }
    }
    
    if (-not $found) {
        $missingVars += $reqVar
        Write-Host "❌ $reqVar is MISSING" -ForegroundColor Red
    } else {
        Write-Host "✅ $reqVar is set" -ForegroundColor Green
    }
}

if ($missingVars.Count -gt 0) {
    Write-Host "`nAction needed:" -ForegroundColor Yellow
    Write-Host "Run .\scripts\Set-CloudRunEnv.ps1 to configure missing variables" -ForegroundColor Cyan
}
