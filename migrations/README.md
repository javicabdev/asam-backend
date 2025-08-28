# Database Migrations - ASAM Backend

Clean, simple database schema for the ASAM (Asociación de Senegaleses y Amigos de Montmeló) backend.

## 🎯 Philosophy

**One Migration, Complete Schema** - Since we're in development phase with no production users, we maintain a single, comprehensive migration that includes all functionality.

## 📁 Active Migrations

- **`000001_complete_schema.up.sql`** - Complete database schema with all features
- **`000001_complete_schema.down.sql`** - Complete schema rollback

## 🗂️ Database Tables

### Core Business
- **`members`** - Association members  
- **`families`** - Family groups (Spanish field names)
- **`familiars`** - Family relatives (Spanish field names)
- **`telephones`** - Phone numbers (polymorphic)

### Financial
- **`payments`** - Payment records
- **`membership_fees`** - Monthly fee definitions  
- **`cash_flows`** - Financial movement tracking

### Authentication & Users
- **`users`** - System users with email authentication
- **`refresh_tokens`** - JWT refresh token management
- **`verification_tokens`** - Email verification & password reset

## 🚀 Usage

### Fresh Installation (Development)
```bash
# Create complete schema
go run cmd/migrate/main.go -cmd up

# Check status  
go run cmd/migrate/main.go -cmd version

# Add test data
go run cmd/seed/main.go
```

### Reset Database (Development)
```bash
# Drop everything and recreate
go run cmd/migrate/main.go -cmd drop
go run cmd/migrate/main.go -cmd up
```

### Using Helper Scripts
```bash
# Linux/macOS
./scripts/dev/fresh-database-setup.sh

# Windows PowerShell  
.\scripts\dev\fresh-database-setup.ps1
```

## 🏗️ Features Included

The single migration includes **all** functionality:

- ✅ Complete table structure with proper relationships
- ✅ All indexes for performance optimization  
- ✅ User authentication with email verification
- ✅ JWT refresh token management
- ✅ Automatic timestamp triggers (`updated_at`)
- ✅ Soft deletes with `deleted_at` columns
- ✅ Foreign key constraints with proper CASCADE/RESTRICT rules
- ✅ Comprehensive commenting for documentation

## 📋 Technical Details

- **Database**: PostgreSQL with UUID extension
- **ORM Compatibility**: Schema matches GORM models exactly
- **Field Names**: Mixed English/Spanish per existing model definitions
- **Soft Deletes**: Implemented via `deleted_at` timestamp columns
- **Auto Timestamps**: `created_at`/`updated_at` with triggers
- **Indexing**: Optimized indexes for queries and foreign keys

## 🔄 Migration History

Previous migration files (development artifacts) have been moved to `old_migrations/` for reference.

## 🌱 Production Considerations

When transitioning to production with real users, consider:
- Incremental migrations for schema changes
- Proper backup/rollback procedures  
- Migration testing in staging environment
- Zero-downtime deployment strategies

---

> **Current Status**: ✅ **Development Phase** - Single comprehensive migration approach
