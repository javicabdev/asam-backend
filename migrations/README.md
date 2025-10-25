# Database Migrations - ASAM Backend

Clean, simple database schema for the ASAM (Asociación de Senegaleses y Amigos de Montmeló) backend.

## 🎯 Philosophy

**One Migration, Complete Schema** - Since we're in development phase with no production users, we maintain a single, comprehensive migration that includes all functionality.

## 📁 Active Migrations

- **`000001_complete_schema.up.sql`** - Complete database schema with all features (v1.1 - Annual fee system)
- **`000001_complete_schema.down.sql`** - Complete schema rollback

## 📦 Version History

### v1.1 (Current) - Annual Membership Fee System
- ✅ **BREAKING CHANGE**: `membership_fees` table migrated from monthly to annual system
- ✅ Field `month` **removed** from `membership_fees`
- ✅ Field `year` now has **UNIQUE constraint** (one fee per year)
- ✅ Due date always set to **December 31st** of the year
- ✅ Simplified fee management and payment tracking

**Migration from v1.0:**
- If you have an existing database with monthly fees, see `POST_IMPLEMENTATION.md` for migration instructions
- For fresh installations, simply run the current schema

### v1.0 - Initial Monthly System
- ⚠️ **DEPRECATED**: Monthly fee system with `year` + `month` fields
- See `001_convert_to_annual_fees.sql.OBSOLETE` for reference

## 🗂️ Database Tables

### Core Business
- **`members`** - Association members  
- **`families`** - Family groups (Spanish field names)
- **`familiars`** - Family relatives (Spanish field names)
- **`telephones`** - Phone numbers (polymorphic)

### Financial
- **`payments`** - Payment records
- **`membership_fees`** - **Annual fee definitions** ⭐ (one per year)
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

### Reset Database (Production) ⚠️

**CRITICAL: Only use if you have NO valuable data in production.**

See `POST_IMPLEMENTATION.md` for detailed instructions, or use:

```bash
# Linux/macOS
./scripts/db/reset_production_db.sh

# Windows PowerShell
.\scripts\db\reset_production_db.ps1

# SQL script (manual)
psql -U postgres -f scripts/db/reset_production_db.sql
```

### Using Helper Scripts (Development)
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
- ✅ **Annual membership fee system** (one fee per year)

## 📋 Technical Details

- **Database**: PostgreSQL with UUID extension
- **ORM Compatibility**: Schema matches GORM models exactly
- **Field Names**: Mixed English/Spanish per existing model definitions
- **Soft Deletes**: Implemented via `deleted_at` timestamp columns
- **Auto Timestamps**: `created_at`/`updated_at` with triggers
- **Indexing**: Optimized indexes for queries and foreign keys
- **Fee System**: Annual fees with UNIQUE constraint on `year`

## 💰 Membership Fee System

The system uses an **annual fee model**:

- ✅ **One fee per year** (enforced by UNIQUE constraint on `year`)
- ✅ Due date is always **December 31st** at 23:59:59
- ✅ Base fee amount configurable per year
- ✅ Extra fee for family members configurable per year
- ✅ Status tracking: `PENDING`, `PAID`, `OVERDUE`

**Important Notes:**
- Initial payments automatically create the annual fee for the current year
- The GraphQL mutation `registerFee` only accepts `year` (no `month` parameter)
- Historical migration from monthly system available as reference (`001_convert_to_annual_fees.sql.OBSOLETE`)

## 🔄 Migration History

- **v1.1 (2025-10-19)**: Annual fee system (current)
- **v1.0**: Monthly fee system (deprecated)
- Previous migration files (development artifacts) have been moved to `old_migrations/` for reference

## 🔢 Membership Numbering Convention

**IMPORTANT**: The system uses a specific convention for membership numbers:
- **Prefix 'A'**: FAMILY members (associated with a Family entity)
- **Prefix 'B'**: INDIVIDUAL members
- **Format**: `[A|B]XXXXX` (letter + at least 5 digits)

This convention is enforced at the application level. Seed data scripts create:
- Individual members: `B99001`, `B99002`, etc.
- No family members by default (to keep development simple)

**Note for Development**: When resetting data, no migration is needed. Simply:
```bash
docker-compose down -v
./start-docker.ps1  # or ./scripts/dev/fresh-database-setup.sh
```

## 🌱 Production Considerations

When transitioning to production with real users, consider:
- Incremental migrations for schema changes
- Proper backup/rollback procedures  
- Migration testing in staging environment
- Zero-downtime deployment strategies

---

> **Current Status**: ✅ **Development Phase** - Single comprehensive migration approach (v1.1 - Annual fee system)
