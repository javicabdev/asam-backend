--!postgresql
-- ASAM Backend - Rollback Initial Schema
-- Drops all tables and objects created in the initial migration
-- WARNING: This will delete ALL data in the database

-- Drop triggers first
DROP TRIGGER IF EXISTS update_verification_tokens_updated_at ON verification_tokens;
DROP TRIGGER IF EXISTS update_users_updated_at ON users;
DROP TRIGGER IF EXISTS update_cash_flows_updated_at ON cash_flows;
DROP TRIGGER IF EXISTS update_payments_updated_at ON payments;
DROP TRIGGER IF EXISTS update_membership_fees_updated_at ON membership_fees;
DROP TRIGGER IF EXISTS update_telephones_updated_at ON telephones;
DROP TRIGGER IF EXISTS update_familiars_updated_at ON familiars;
DROP TRIGGER IF EXISTS update_families_updated_at ON families;
DROP TRIGGER IF EXISTS update_members_updated_at ON members;

-- Drop tables in reverse dependency order
DROP TABLE IF EXISTS verification_tokens CASCADE;
DROP TABLE IF EXISTS refresh_tokens CASCADE;
DROP TABLE IF EXISTS users CASCADE;
DROP TABLE IF EXISTS cash_flows CASCADE;
DROP TABLE IF EXISTS payments CASCADE;
DROP TABLE IF EXISTS membership_fees CASCADE;
DROP TABLE IF EXISTS telephones CASCADE;
DROP TABLE IF EXISTS familiars CASCADE;
DROP TABLE IF EXISTS families CASCADE;
DROP TABLE IF EXISTS members CASCADE;

-- Drop functions
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop extensions (optional - may be used by other databases)
-- DROP EXTENSION IF EXISTS "pgcrypto";
-- DROP EXTENSION IF EXISTS "uuid-ossp";
