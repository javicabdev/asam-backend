# Database Migrations

This directory contains database migrations for the ASAM backend application.

## Current Schema

The database schema is defined in the following migrations:
- `000001_initial_schema.up.sql` - Creates all tables matching the GORM models (includes refresh_tokens table)
- `000001_initial_schema.down.sql` - Drops all tables
- `000004_add_email_verification.up.sql` - Adds email verification fields and verification_tokens table
- `000004_add_email_verification.down.sql` - Removes email verification features
- `000006_add_token_cleanup_indexes.up.sql` - Adds indexes for token cleanup optimization
- `000006_add_token_cleanup_indexes.down.sql` - Removes token cleanup indexes

**Note:** Migration numbers 000002, 000003, and 000005 were not used.

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
9. **refresh_tokens** - JWT refresh tokens (separate table approach)
10. **verification_tokens** - Email verification and password reset tokens

## Important Notes

- The migrations respect the exact field names defined in the Go models
- Some models use Spanish field names (families, familiars, telephones) while others use English
- All timestamps use `TIMESTAMP WITH TIME ZONE`
- Soft deletes are implemented with `deleted_at` columns  
- The `updated_at` column is automatically updated via triggers
- All foreign key relationships match the GORM model definitions
- The project uses PostgreSQL database

## Running Migrations

To run migrations:

```bash
# Apply all migrations
go run cmd/migrate/main.go -cmd up

# Rollback all migrations  
go run cmd/migrate/main.go -cmd down

# Check migration status
go run cmd/migrate/main.go -cmd status
```

For development with Docker:

```bash
# Apply migrations
make db-migrate

# Reset database (rollback and re-apply)
make db-reset
```

## Model-Database Mapping

The migrations are designed to match the existing Go models exactly:

- `Member` model → `members` table (English fields)
- `Family` model → `families` table (Spanish fields like `numero_socio`, `miembro_origen_id`)
- `Familiar` model → `familiars` table (Spanish fields like `familia_id`, `nombre`, `apellidos`)
- `Telephone` model → `telephones` table (Spanish field `numero_telefono`)
- `User` model → `users` table
- `RefreshToken` model → `refresh_tokens` table
- `VerificationToken` model → `verification_tokens` table
- Other models use English field names

This ensures compatibility with the existing codebase without requiring model changes.
