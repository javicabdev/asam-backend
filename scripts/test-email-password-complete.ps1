param(
    [string]$coveragePackage = "github.com/javicabdev/asam-backend/..."
)

Write-Host "`n🔐 Testing Email Verification and Password Recovery..." -ForegroundColor Cyan

# Clean previous coverage files
Write-Host "`n🧹 Cleaning previous coverage files..." -ForegroundColor Yellow
Remove-Item -Path "coverage*.out" -Force -ErrorAction SilentlyContinue

# First, ensure the code compiles
Write-Host "`n🔨 Building the project..." -ForegroundColor Yellow
go build ./...

if ($LASTEXITCODE -ne 0) {
    Write-Host "`n❌ Build failed. Please check the compilation errors above." -ForegroundColor Red
    exit 1
}

# Run all user service tests
Write-Host "`n📧 Running all User Service tests..." -ForegroundColor Yellow
go test -v ./test/internal/domain/services/... -coverprofile=coverage-all.out -coverpkg=$coveragePackage

if ($LASTEXITCODE -ne 0) {
    Write-Host "`n❌ Tests failed. Please check the output above." -ForegroundColor Red
    exit 1
}

# Generate coverage report
Write-Host "`n📊 Generating coverage report..." -ForegroundColor Yellow
go tool cover -html=coverage-all.out -o coverage-all.html

Write-Host "`n✅ All tests passed!" -ForegroundColor Green
Write-Host "📊 Coverage report generated: coverage-all.html" -ForegroundColor Cyan

# Show coverage summary
Write-Host "`n📊 Coverage Summary:" -ForegroundColor Cyan
go tool cover -func=coverage-all.out | Select-String -Pattern "total:"
