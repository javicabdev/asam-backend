# Test Runner Script for Complete Auth System
Write-Host "🔐 Running Complete Auth System Tests..." -ForegroundColor Cyan

# Get the directory of the script
$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$projectRoot = Split-Path -Parent $scriptDir

# Change to project root
Push-Location $projectRoot

try {
    # Clean previous coverage files
    Write-Host "`n🧹 Cleaning previous coverage files..." -ForegroundColor Yellow
    Remove-Item -Path "coverage_*.out" -ErrorAction SilentlyContinue
    Remove-Item -Path "coverage_*.html" -ErrorAction SilentlyContinue
    
    # Run tests for the auth services package
    Write-Host "`n📊 Running all authentication system tests..." -ForegroundColor Yellow
    
    # Use the package path to run all tests in the directory
    $testPath = "./test/internal/domain/services"
    $coverageFile = "coverage_auth_complete.out"
    $coveragePackage = "github.com/javicabdev/asam-backend/internal/domain/services"
    
    # Run tests with coverage for the actual service package
    & go test -v -coverprofile=$coverageFile -covermode=atomic -coverpkg=$coveragePackage $testPath
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host "`n✅ All tests passed!" -ForegroundColor Green
        
        # Show coverage report
        Write-Host "`n📈 Coverage Report:" -ForegroundColor Yellow
        & go tool cover -func=$coverageFile | Select-String -Pattern "auth_service.go|user_service.go|total:" | ForEach-Object {
            $line = $_.Line
            if ($line -match "total:") {
                Write-Host "`n$line" -ForegroundColor Cyan
            } else {
                Write-Host $line -ForegroundColor Gray
            }
        }
        
        # Generate HTML coverage report
        Write-Host "`n🌐 Generating HTML coverage report..." -ForegroundColor Yellow
        & go tool cover -html=$coverageFile -o coverage_auth_complete.html
        
        Write-Host "`nCoverage report saved to: coverage_auth_complete.html" -ForegroundColor Gray
        
        # Ask if user wants to open the report
        $openReport = Read-Host "`nOpen coverage report in browser? (Y/N)"
        if ($openReport -eq 'Y' -or $openReport -eq 'y') {
            Start-Process "coverage_auth_complete.html"
        }
        
        # Summary
        Write-Host "`n📊 Test Summary:" -ForegroundColor Cyan
        Write-Host "  ✅ Authentication System Tests: PASSED" -ForegroundColor Green
        Write-Host "`n🎉 All authentication system tests completed successfully!" -ForegroundColor Green
        
    } else {
        Write-Host "`n❌ Tests failed. Please check the output above." -ForegroundColor Red
        exit 1
    }
} finally {
    # Return to original directory
    Pop-Location
}

Write-Host "`n💡 Next steps:" -ForegroundColor Yellow
Write-Host "  1. Review coverage report to identify areas needing more tests" -ForegroundColor Gray
Write-Host "  2. Add integration tests for the complete auth flow" -ForegroundColor Gray
Write-Host "  3. Consider adding benchmarks for performance testing" -ForegroundColor Gray
Write-Host "  4. Implement tests for edge cases and error scenarios" -ForegroundColor Gray
