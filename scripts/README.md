# Scripts Directory

This directory contains all utility scripts for the ASAM Backend project, organized by functionality.

## Directory Structure

### 📁 db/ - Database Scripts
Scripts for database operations including migrations, backups, and maintenance.

- **migrate.ps1** - Run database migrations locally
- **backup-db.ps1** - Create database backups
- **restore-db.ps1** - Restore database from backup
- **apply-email-required-changes.ps1** - Apply email requirement migration
- **rollback-email-required-changes.ps1** - Rollback email requirement changes
- **run-email-required-migration.ps1** - Run specific email migration
- **run-production-migrations.ps1/.sh** - Run migrations in production

### 📁 dev/ - Development Scripts
Scripts for development workflow and code maintenance.

- **cleanup-tech-debt-phase*.ps1/.sh** - Technical debt cleanup scripts
- **install-hooks.bat/.sh** - Install git hooks
- **regenerate-graphql.ps1/.sh** - Regenerate GraphQL code
- **format.bat** - Format Go code
- **lint.bat/.sh** - Run linters
- **test.ps1** - Run tests
- **pre-commit** - Git pre-commit hook
- **check-config.go** - Verify configuration

### 📁 docker/ - Docker Scripts
Scripts for managing Docker containers.

- **docker-up.ps1** - Start Docker containers
- **docker-down.ps1** - Stop Docker containers
- **docker-restart.ps1** - Restart Docker containers
- **logs.ps1** - View Docker logs

### 📁 ops/ - Operational Scripts
Scripts for building, deployment, and operations.

- **build.ps1** - Build the application
- **clean.ps1** - Clean build artifacts
- **run.ps1** - Run the application
- **gcp/** - Google Cloud Platform specific scripts

### 📁 verification/ - Email Verification Scripts
Scripts for managing email verification functionality.

- **check-email-verification-status.ps1** - Check user verification status
- **check-token-info.ps1** - Check verification token information
- **cleanup-verification-tokens.ps1** - Clean expired tokens
- **manually-verify-user.ps1** - Manually verify a user
- **verify-email-manual.sql** - SQL for manual verification

### 📁 user-management/ - User Management Scripts
Scripts for user administration and management.

## Usage Examples

### Running Database Migrations
```powershell
# Local development
.\scripts\db\migrate.ps1

# Production
.\scripts\db\run-production-migrations.ps1
```

### Development Workflow
```powershell
# Install git hooks
.\scripts\dev\install-hooks.bat

# Regenerate GraphQL code
.\scripts\dev\regenerate-graphql.ps1

# Run tests
.\scripts\dev\test.ps1
```

### Docker Operations
```powershell
# Start services
.\scripts\docker\docker-up.ps1

# View logs
.\scripts\docker\logs.ps1

# Restart services
.\scripts\docker\docker-restart.ps1
```

## Script Dependencies

Some scripts have dependencies on others:

1. **Database scripts** require Docker to be running
2. **Development scripts** require Go to be installed
3. **GCP scripts** require Google Cloud SDK

## Contributing

When adding new scripts:
1. Place them in the appropriate category folder
2. Add documentation to this README
3. Include usage examples if the script is complex
4. Follow the existing naming conventions
