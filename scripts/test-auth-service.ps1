# Test Runner Script for Auth Service
Write-Host "🧪 Running Auth Service Tests..." -ForegroundColor Cyan

# Get the directory of the script
$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$projectRoot = Split-Path -Parent $scriptDir

# Change to project root
Push-Location $projectRoot

try {
    # Run tests with coverage
    Write-Host "`n📊 Running tests with coverage..." -ForegroundColor Yellow
    
    $testPath = "./test/internal/domain/services/auth_service_test.go"
    $coverProfile = "coverage_auth.out"
    
    # Run the test
    & go test -v -coverprofile=$coverProfile -covermode=atomic $testPath
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host "`n✅ Tests passed!" -ForegroundColor Green
        
        # Show coverage report
        Write-Host "`n📈 Coverage Report:" -ForegroundColor Yellow
        & go tool cover -func=$coverProfile | Select-String -Pattern "auth_service.go|total"
        
        # Generate HTML coverage report
        Write-Host "`n🌐 Generating HTML coverage report..." -ForegroundColor Yellow
        & go tool cover -html=$coverProfile -o coverage_auth.html
        Write-Host "Coverage report saved to: coverage_auth.html" -ForegroundColor Gray
        
        # Ask if user wants to open the report
        $openReport = Read-Host "`nOpen coverage report in browser? (Y/N)"
        if ($openReport -eq 'Y' -or $openReport -eq 'y') {
            Start-Process "coverage_auth.html"
        }
    } else {
        Write-Host "`n❌ Tests failed!" -ForegroundColor Red
        exit 1
    }
} finally {
    # Return to original directory
    Pop-Location
}

Write-Host "`n🎉 Done!" -ForegroundColor Green
