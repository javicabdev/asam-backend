-- Drop all tables and related objects in reverse dependency order

-- Drop triggers first
DROP TRIGGER IF EXISTS update_users_updated_at ON users;
DROP TRIGGER IF EXISTS update_cash_flows_updated_at ON cash_flows;
DROP TRIGGER IF EXISTS update_payments_updated_at ON payments;
DROP TRIGGER IF EXISTS update_membership_fees_updated_at ON membership_fees;
DROP TRIGGER IF EXISTS update_telephones_updated_at ON telephones;
DROP TRIGGER IF EXISTS update_familiars_updated_at ON familiars;
DROP TRIGGER IF EXISTS update_families_updated_at ON families;
DROP TRIGGER IF EXISTS update_members_updated_at ON members;

-- Drop tables in reverse order of dependencies
DROP TABLE IF EXISTS refresh_tokens CASCADE;
DROP TABLE IF EXISTS cash_flows CASCADE;
DROP TABLE IF EXISTS payments CASCADE;
DROP TABLE IF EXISTS membership_fees CASCADE;
DROP TABLE IF EXISTS telephones CASCADE;
DROP TABLE IF EXISTS familiars CASCADE;
DROP TABLE IF EXISTS families CASCADE;
DROP TABLE IF EXISTS users CASCADE;
DROP TABLE IF EXISTS members CASCADE;

-- Drop the update function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop extensions
DROP EXTENSION IF EXISTS "uuid-ossp";
