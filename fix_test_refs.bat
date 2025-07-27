@echo off
cd /d C:\Work\babacar\asam\asam-backend
powershell -Command "(Get-Content -Path 'test\internal\domain\services\auth_service_test.go') -replace 'test\.MockVerificationTokenRepository', 'MockVerificationTokenRepository' | Set-Content -Path 'test\internal\domain\services\auth_service_test.go'"
powershell -Command "(Get-Content -Path 'test\internal\domain\services\auth_service_test.go') -replace 'test\.MockEmailVerificationService', 'MockEmailVerificationService' | Set-Content -Path 'test\internal\domain\services\auth_service_test.go'"
echo Replacements completed
