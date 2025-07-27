@echo off
cd /d C:\Work\babacar\asam\asam-backend
go test -c ./test/internal/domain/services/
echo Test compilation result: %ERRORLEVEL%
