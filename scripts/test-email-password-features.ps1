param(
    [string]$coveragePackage = "github.com/javicabdev/asam-backend/test/internal/domain/services/..."
)

Write-Host "🔐 Testing Email Verification and Password Recovery..." -ForegroundColor Cyan

# Clean previous coverage files
Write-Host "`n🧹 Cleaning previous coverage files..." -ForegroundColor Yellow
Remove-Item -Path "coverage*.out" -Force -ErrorAction SilentlyContinue

# Run user service tests with email functionality
Write-Host "`n📧 Running Email Verification tests..." -ForegroundColor Yellow
go test -v ./test/internal/domain/services/... -run "TestUserService.*Email|TestUserService.*Verification" -coverprofile=coverage-email.out -coverpkg=$coveragePackage

if ($LASTEXITCODE -ne 0) {
    Write-Host "`n❌ Email Verification tests failed. Please check the output above." -ForegroundColor Red
    exit 1
}

# Run password recovery tests
Write-Host "`n🔑 Running Password Recovery tests..." -ForegroundColor Yellow
go test -v ./test/internal/domain/services/... -run "TestUserService.*Password.*Reset|TestUserService.*RequestPassword" -coverprofile=coverage-password.out -coverpkg=$coveragePackage

if ($LASTEXITCODE -ne 0) {
    Write-Host "`n❌ Password Recovery tests failed. Please check the output above." -ForegroundColor Red
    exit 1
}

# Run email utility tests
Write-Host "`n📨 Running Email Utility tests..." -ForegroundColor Yellow
go test -v ./test/pkg/utils/... -run "TestEmail|TestToken" -coverprofile=coverage-utils.out -coverpkg=$coveragePackage

if ($LASTEXITCODE -ne 0) {
    Write-Host "`n❌ Email Utility tests failed. Please check the output above." -ForegroundColor Red
    exit 1
}

# Generate coverage report
Write-Host "`n📊 Generating coverage report..." -ForegroundColor Yellow
go tool cover -html=coverage-email.out -o coverage-email.html
go tool cover -html=coverage-password.out -o coverage-password.html
go tool cover -html=coverage-utils.out -o coverage-utils.html

Write-Host "`n✅ All Email and Password Recovery tests passed!" -ForegroundColor Green
Write-Host "📊 Coverage reports generated:" -ForegroundColor Cyan
Write-Host "   - coverage-email.html" -ForegroundColor Gray
Write-Host "   - coverage-password.html" -ForegroundColor Gray
Write-Host "   - coverage-utils.html" -ForegroundColor Gray

# Show coverage summary
Write-Host "`n📊 Coverage Summary:" -ForegroundColor Cyan
go tool cover -func=coverage-email.out | Select-String -Pattern "total:"
go tool cover -func=coverage-password.out | Select-String -Pattern "total:"
go tool cover -func=coverage-utils.out | Select-String -Pattern "total:"
