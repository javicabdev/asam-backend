# scripts/schedule-backups.ps1
$scriptPath = Join-Path $PSScriptRoot "backup-db.ps1"
$action = New-ScheduledTaskAction -Execute "PowerShell.exe" `
    -Argument "-NoProfile -ExecutionPolicy Bypass -File `"$scriptPath`""

$trigger = New-ScheduledTaskTrigger -Daily -At 3AM

$principal = New-ScheduledTaskPrincipal -UserId "NT AUTHORITY\SYSTEM" -LogonType ServiceAccount -RunLevel Highest

$settings = New-ScheduledTaskSettingsSet -StartWhenAvailable -DontStopOnIdleEnd -RestartCount 3 -RestartInterval (New-TimeSpan -Minutes 1)

Register-ScheduledTask -Action $action -Trigger $trigger -Principal $principal -Settings $settings `
    -TaskName "ASAMDatabaseBackup" -Description "Backup diario de la base de datos ASAM" -Force

Write-Host "Tarea programada creada exitosamente" -ForegroundColor Green