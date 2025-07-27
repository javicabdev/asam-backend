@echo off
cd /d C:\Work\babacar\asam\asam-backend

powershell -NoProfile -ExecutionPolicy Bypass -Command ^
"$content = Get-Content -Path 'test\internal\domain\services\auth_service_test.go' -Raw; ^
$content = $content -replace 'test\.MockVerificationTokenRepository', 'MockVerificationTokenRepository'; ^
$content = $content -replace 'test\.MockEmailVerificationService', 'MockEmailVerificationService'; ^
$content | Set-Content -Path 'test\internal\domain\services\auth_service_test.go' -NoNewline"

echo Replacements completed
