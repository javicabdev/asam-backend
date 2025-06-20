# Quick Test Runner for Auth System
Write-Host "🚀 Quick Auth System Test Run" -ForegroundColor Cyan

# Run tests from project root
$projectRoot = Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)
Push-Location $projectRoot

try {
    # Run all auth-related tests with coverage
    Write-Host "`n🧪 Running all authentication tests..." -ForegroundColor Yellow
    
    $testPath = "./test/internal/domain/services"
    $coverProfile = "coverage_auth_quick.out"
    $coveragePackage = "github.com/javicabdev/asam-backend/internal/domain/services"
    
    # Run tests with race detection and coverage for the service package
    & go test -v -race -coverprofile=$coverProfile -covermode=atomic -coverpkg=$coveragePackage $testPath
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host "`n✅ All tests passed!" -ForegroundColor Green
        
        # Show quick coverage summary
        Write-Host "`n📊 Coverage Summary:" -ForegroundColor Yellow
        & go tool cover -func=$coverProfile | Select-String -Pattern "auth_service.go|user_service.go|total:" | ForEach-Object {
            $line = $_.Line
            if ($line -match "total:") {
                Write-Host $line -ForegroundColor Cyan
            } else {
                Write-Host $line -ForegroundColor Gray
            }
        }
        
        # Generate and open HTML report
        $htmlFile = "coverage_auth_quick.html"
        & go tool cover -html=$coverProfile -o $htmlFile
        
        Write-Host "`n📄 Full report: $htmlFile" -ForegroundColor Gray
        Write-Host "💡 Run '.\scripts\test-auth-complete.ps1' for detailed analysis" -ForegroundColor DarkGray
    } else {
        Write-Host "`n❌ Tests failed!" -ForegroundColor Red
        exit 1
    }
} finally {
    Pop-Location
}
