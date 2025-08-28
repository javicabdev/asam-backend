--!postgresql
-- ASAM Backend - Complete Schema Rollback
-- Drops all tables, indexes, triggers, and functions created by 000001_complete_schema.up.sql

-- =============================================================================
-- DROP TABLES (in correct order to respect foreign key constraints)
-- =============================================================================

-- Drop tables with foreign keys first
DROP TABLE IF EXISTS verification_tokens CASCADE;
DROP TABLE IF EXISTS refresh_tokens CASCADE;
DROP TABLE IF EXISTS cash_flows CASCADE;
DROP TABLE IF EXISTS payments CASCADE;
DROP TABLE IF EXISTS membership_fees CASCADE;
DROP TABLE IF EXISTS telephones CASCADE;
DROP TABLE IF EXISTS familiars CASCADE;
DROP TABLE IF EXISTS families CASCADE;
DROP TABLE IF EXISTS users CASCADE;
DROP TABLE IF EXISTS members CASCADE;

-- =============================================================================
-- DROP FUNCTIONS
-- =============================================================================

DROP FUNCTION IF EXISTS update_updated_at_column() CASCADE;

-- =============================================================================
-- DROP EXTENSIONS (optional - leave uuid-ossp as it might be used elsewhere)
-- =============================================================================

-- Uncomment if you want to remove the extension completely
-- DROP EXTENSION IF EXISTS "uuid-ossp";

-- Schema cleanup complete
