PowerShell -ExecutionPolicy Bypass -Command "cd C:\Work\babacar\asam\asam-backend; go test -v ./test/pkg/utils/... ./test/internal/domain/services/... -run '(TestUserService.*Email|TestEmail)'"
