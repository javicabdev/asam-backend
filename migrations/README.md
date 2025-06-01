# Database Migrations

This directory contains database migrations for the ASAM backend application.

## Current Schema

The database schema is defined in a single initial migration that matches the existing Go models:
- `000001_initial_schema.up.sql` - Creates all tables matching the GORM models
- `000001_initial_schema.down.sql` - Drops all tables

## Tables

The schema includes the following tables (names match GORM conventions):

1. **members** - Association members (model: Member)
2. **families** - Family groups (model: Family) - uses Spanish field names
3. **familiars** - Family relatives (model: Familiar) - uses Spanish field names  
4. **telephones** - Phone numbers (model: Telephone) - polymorphic, uses Spanish field name
5. **membership_fees** - Monthly membership fee definitions (model: MembershipFee)
6. **payments** - Payment records (model: Payment)
7. **cash_flows** - Financial movements (model: CashFlow)
8. **users** - System users for authentication (model: User)

## Important Notes

- The migrations respect the exact field names defined in the Go models
- Some models use Spanish field names (families, familiars, telephones) while others use English
- All timestamps use `TIMESTAMP WITH TIME ZONE`
- Soft deletes are implemented with `deleted_at` columns
- The `updated_at` column is automatically updated via triggers
- All foreign key relationships match the GORM model definitions

## Running Migrations

To recreate the database from scratch:

```powershell
# Recreate the database (WARNING: This will delete all data!)
.\scripts\recreate_database.ps1

# Or manually:
.\migrate.ps1 local down  # Roll back all migrations
.\migrate.ps1 local up     # Apply all migrations
```

## Model-Database Mapping

The migrations are designed to match the existing Go models exactly:

- `Member` model → `members` table (English fields)
- `Family` model → `families` table (Spanish fields like `numero_socio`, `miembro_origen_id`)
- `Familiar` model → `familiars` table (Spanish fields like `familia_id`, `nombre`, `apellidos`)
- `Telephone` model → `telephones` table (Spanish field `numero_telefono`)
- Other models use English field names

This ensures compatibility with the existing codebase without requiring model changes.
