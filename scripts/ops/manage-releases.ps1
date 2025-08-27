# Script para gestionar releases y tags
param(
    [Parameter(Mandatory=$false)]
    [ValidateSet("check", "create", "list")]
    [string]$Action = "check"
)

Write-Host "=== Gestión de Releases ASAM ===" -ForegroundColor Cyan
Write-Host ""

# Función para obtener la última versión
function Get-LatestVersion {
    $tags = git tag -l "v*" | Sort-Object -Descending
    if ($tags) {
        return $tags[0]
    }
    return "v0.0.0"
}

# Función para incrementar versión
function Get-NextVersion {
    param(
        [string]$CurrentVersion,
        [ValidateSet("major", "minor", "patch")]
        [string]$Type = "patch"
    )
    
    if ($CurrentVersion -match '^v(\d+)\.(\d+)\.(\d+)') {
        $major = [int]$matches[1]
        $minor = [int]$matches[2]
        $patch = [int]$matches[3]
        
        switch ($Type) {
            "major" { return "v$($major + 1).0.0" }
            "minor" { return "v$major.$($minor + 1).0" }
            "patch" { return "v$major.$minor.$($patch + 1)" }
        }
    }
    return "v1.0.0"
}

switch ($Action) {
    "check" {
        Write-Host "[Estado Actual]" -ForegroundColor Yellow
        
        # Branch actual
        $branch = git branch --show-current
        Write-Host "Branch: $branch" -ForegroundColor $(if($branch -eq "main") {"Green"} else {"Yellow"})
        
        # Última versión
        $latestVersion = Get-LatestVersion
        Write-Host "Última versión: $latestVersion" -ForegroundColor Cyan
        
        # Commits desde última versión
        if ($latestVersion -ne "v0.0.0") {
            $commitsSince = git log "$latestVersion..HEAD" --oneline 2>$null | Measure-Object -Line
            Write-Host "Commits desde $latestVersion`: $($commitsSince.Lines)" -ForegroundColor Gray
        }
        
        # Estado de staging
        Write-Host ""
        Write-Host "[Staging]" -ForegroundColor Yellow
        $stagingImage = gcloud run services describe asam-backend-staging --region=europe-west1 --format="value(spec.template.spec.containers[0].image)" 2>$null
        if ($stagingImage) {
            $stagingTag = $stagingImage.Split(":")[-1]
            Write-Host "Tag actual: $stagingTag" -ForegroundColor $(if($stagingTag -eq "latest") {"Yellow"} else {"Green"})
        } else {
            Write-Host "No desplegado" -ForegroundColor DarkGray
        }
        
        # Estado de production
        Write-Host ""
        Write-Host "[Production]" -ForegroundColor Yellow
        $prodImage = gcloud run services describe asam-backend --region=europe-west1 --format="value(spec.template.spec.containers[0].image)" 2>$null
        if ($prodImage) {
            $prodTag = $prodImage.Split(":")[-1]
            Write-Host "Tag actual: $prodTag" -ForegroundColor $(if($prodTag -eq "latest") {"Red"} else {"Green"})
            
            if ($prodTag -ne $latestVersion -and $latestVersion -ne "v0.0.0") {
                Write-Host "⚠️  Production no está en la última versión ($latestVersion)" -ForegroundColor Yellow
            }
        } else {
            Write-Host "No desplegado" -ForegroundColor DarkGray
        }
        
        Write-Host ""
        Write-Host "[Próximas Versiones Sugeridas]" -ForegroundColor Yellow
        $currentVersion = Get-LatestVersion
        Write-Host "Patch:  $(Get-NextVersion $currentVersion 'patch') (bug fixes)" -ForegroundColor Gray
        Write-Host "Minor:  $(Get-NextVersion $currentVersion 'minor') (nuevas features)" -ForegroundColor Gray
        Write-Host "Major:  $(Get-NextVersion $currentVersion 'major') (cambios breaking)" -ForegroundColor Gray
    }
    
    "create" {
        Write-Host "[Crear Nuevo Release]" -ForegroundColor Yellow
        
        # Verificar que estamos en main
        $branch = git branch --show-current
        if ($branch -ne "main") {
            Write-Host "❌ Debes estar en la rama main para crear un release" -ForegroundColor Red
            Write-Host "   Branch actual: $branch" -ForegroundColor Yellow
            exit 1
        }
        
        # Verificar que no hay cambios sin commitear
        $status = git status --porcelain
        if ($status) {
            Write-Host "❌ Hay cambios sin commitear" -ForegroundColor Red
            git status --short
            exit 1
        }
        
        # Obtener última versión y sugerir siguiente
        $currentVersion = Get-LatestVersion
        Write-Host "Versión actual: $currentVersion" -ForegroundColor Cyan
        Write-Host ""
        
        Write-Host "Selecciona tipo de release:" -ForegroundColor Yellow
        Write-Host "1. Patch $(Get-NextVersion $currentVersion 'patch') - Bug fixes" -ForegroundColor Gray
        Write-Host "2. Minor $(Get-NextVersion $currentVersion 'minor') - Nuevas features" -ForegroundColor Gray
        Write-Host "3. Major $(Get-NextVersion $currentVersion 'major') - Breaking changes" -ForegroundColor Gray
        Write-Host "4. Custom - Especificar versión manualmente" -ForegroundColor Gray
        
        $choice = Read-Host "Opción (1-4)"
        
        $newVersion = switch ($choice) {
            "1" { Get-NextVersion $currentVersion 'patch' }
            "2" { Get-NextVersion $currentVersion 'minor' }
            "3" { Get-NextVersion $currentVersion 'major' }
            "4" { 
                $custom = Read-Host "Ingresa versión (formato: v1.2.3)"
                if ($custom -notmatch '^v\d+\.\d+\.\d+') {
                    Write-Host "❌ Formato inválido. Debe ser vX.Y.Z" -ForegroundColor Red
                    exit 1
                }
                $custom
            }
            default {
                Write-Host "❌ Opción inválida" -ForegroundColor Red
                exit 1
            }
        }
        
        Write-Host ""
        Write-Host "Nueva versión: $newVersion" -ForegroundColor Green
        
        # Solicitar mensaje del release
        Write-Host ""
        Write-Host "Ingresa el mensaje del release (o Enter para generar automáticamente):" -ForegroundColor Yellow
        $message = Read-Host "Mensaje"
        if (-not $message) {
            $message = "Release $newVersion"
        }
        
        # Confirmación
        Write-Host ""
        Write-Host "Se va a crear:" -ForegroundColor Yellow
        Write-Host "  Tag: $newVersion" -ForegroundColor Cyan
        Write-Host "  Mensaje: $message" -ForegroundColor Cyan
        Write-Host ""
        $confirm = Read-Host "¿Confirmar? (s/n)"
        
        if ($confirm -eq "s") {
            Write-Host ""
            Write-Host "Creando tag..." -ForegroundColor Yellow
            git tag -a $newVersion -m $message
            
            Write-Host "Pushing tag..." -ForegroundColor Yellow
            git push origin $newVersion
            
            Write-Host ""
            Write-Host "✅ Release creado exitosamente!" -ForegroundColor Green
            Write-Host ""
            Write-Host "Próximos pasos:" -ForegroundColor Yellow
            Write-Host "1. El workflow de release construirá la imagen automáticamente" -ForegroundColor Gray
            Write-Host "2. Espera ~5 minutos para que termine" -ForegroundColor Gray
            Write-Host "3. Deploy a production con:" -ForegroundColor Gray
            Write-Host "   - Ve a GitHub Actions → Deploy to Google Cloud Run" -ForegroundColor Cyan
            Write-Host "   - Usa image_tag: $newVersion" -ForegroundColor Cyan
        } else {
            Write-Host "Cancelado" -ForegroundColor Yellow
        }
    }
    
    "list" {
        Write-Host "[Releases Disponibles]" -ForegroundColor Yellow
        
        # Tags locales
        Write-Host ""
        Write-Host "Tags Git:" -ForegroundColor Cyan
        git tag -l "v*" --sort=-version:refname | Select-Object -First 10 | ForEach-Object {
            $date = git log -1 --format=%ai $_
            Write-Host "  $_ - $date" -ForegroundColor Gray
        }
        
        # Imágenes en GCR
        Write-Host ""
        Write-Host "Imágenes en Registry:" -ForegroundColor Cyan
        gcloud container images list-tags gcr.io/babacar-asam/asam-backend `
            --filter="tags ~ '^v[0-9]'" `
            --limit=10 `
            --format="table(tags,timestamp.datetime)" `
            --sort-by="~timestamp"
    }
}

Write-Host ""
Write-Host "[Comandos Útiles]" -ForegroundColor Yellow
Write-Host "• Ver estado:        .\scripts\ops\manage-releases.ps1 check" -ForegroundColor Gray
Write-Host "• Crear release:     .\scripts\ops\manage-releases.ps1 create" -ForegroundColor Gray
Write-Host "• Listar releases:   .\scripts\ops\manage-releases.ps1 list" -ForegroundColor Gray
