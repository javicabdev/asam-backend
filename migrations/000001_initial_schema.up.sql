--!postgresql
-- ASAM Backend - Initial Database Schema (Consolidated)
-- Single migration containing the complete final schema
-- Includes all features: members, families, users, authentication, email verification, payments, cash flows
-- VERSION: 2.0 - Consolidated from 9 previous migrations
-- IDEMPOTENT: Safe to run multiple times

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto"; -- For password hashing

-- Create update_updated_at_column function for automatic timestamp updates
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- =============================================================================
-- CORE BUSINESS TABLES
-- =============================================================================

-- Members table - Individual members of the association
CREATE TABLE IF NOT EXISTS members (
    id SERIAL PRIMARY KEY,
    membership_number VARCHAR(255) UNIQUE NOT NULL,
    membership_type VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    surnames VARCHAR(255) NOT NULL,
    address VARCHAR(255) NOT NULL,
    postcode VARCHAR(255) NOT NULL,
    city VARCHAR(255) NOT NULL,
    province VARCHAR(255) NOT NULL DEFAULT 'Barcelona',
    country VARCHAR(255) NOT NULL DEFAULT 'España',
    state VARCHAR(255) NOT NULL,
    registration_date DATE NOT NULL,
    leaving_date DATE,
    birth_date DATE,
    identity_card VARCHAR(255),
    email VARCHAR(255),
    profession VARCHAR(255),
    nationality VARCHAR(255) DEFAULT 'Senegal',
    remarks TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Families table - Family groups (Spanish field names per existing model)
CREATE TABLE IF NOT EXISTS families (
    id SERIAL PRIMARY KEY,
    numero_socio VARCHAR(255) UNIQUE NOT NULL,
    miembro_origen_id INTEGER,
    esposo_nombre VARCHAR(100),
    esposo_apellidos VARCHAR(100),
    esposa_nombre VARCHAR(100),
    esposa_apellidos VARCHAR(100),
    esposo_fecha_nacimiento DATE,
    esposo_documento_identidad VARCHAR(20),
    esposo_correo_electronico VARCHAR(100),
    esposa_fecha_nacimiento DATE,
    esposa_documento_identidad VARCHAR(20),
    esposa_correo_electronico VARCHAR(100),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Familiars table - Family relatives (Spanish field names per existing model)
CREATE TABLE IF NOT EXISTS familiars (
    id SERIAL PRIMARY KEY,
    familia_id INTEGER NOT NULL,
    nombre VARCHAR(100) NOT NULL,
    apellidos VARCHAR(100) NOT NULL,
    fecha_nacimiento DATE,
    dni_nie VARCHAR(20),
    correo_electronico VARCHAR(100),
    parentesco VARCHAR(50) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Telephones table - Polymorphic phone numbers (Spanish field name per model)
CREATE TABLE IF NOT EXISTS telephones (
    id SERIAL PRIMARY KEY,
    numero_telefono VARCHAR(20) NOT NULL,
    contactable_id INTEGER NOT NULL,
    contactable_type VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- =============================================================================
-- FINANCIAL TABLES
-- =============================================================================

-- Membership fees table - Annual fee definitions
-- IMPORTANT: This is an ANNUAL system. One fee per year, due date is always December 31st
-- CHANGES from v1: Removed 'status' and 'payment_id' fields (migration 006 & 007)
CREATE TABLE IF NOT EXISTS membership_fees (
    id SERIAL PRIMARY KEY,
    year INTEGER NOT NULL UNIQUE,  -- UNIQUE constraint ensures one fee per year
    base_fee_amount DECIMAL(10,2) NOT NULL,
    family_fee_extra DECIMAL(10,2) NOT NULL DEFAULT 0,
    due_date DATE NOT NULL,  -- Always December 31st of the year
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Payments table - Payment records
-- CHANGES from v1:
-- - member_id is NOT NULL (migration 008: removed family_id, now all payments must have member)
-- - payment_date is nullable (migration 005: pending payments don't have a date yet)
-- - membership_fee_id tracks which annual fee this payment is for
-- - Unique constraint: one initial payment per member (migration 004)
CREATE TABLE IF NOT EXISTS payments (
    id SERIAL PRIMARY KEY,
    member_id INTEGER NOT NULL,  -- Required: for families, use family.miembro_origen_id
    amount DECIMAL(10,2) NOT NULL,
    payment_date DATE,  -- NULL for pending payments
    status VARCHAR(255) NOT NULL,
    payment_method VARCHAR(255) NOT NULL,
    notes TEXT,
    membership_fee_id INTEGER,  -- Links to annual fee definition
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Cash flows table - Financial movements tracking
-- CHANGES from v1:
-- - Removed family_id (migration 008)
-- - Added unique constraint on payment_id (migration 009)
-- - member_id can be NULL for non-payment transactions
CREATE TABLE IF NOT EXISTS cash_flows (
    id SERIAL PRIMARY KEY,
    member_id INTEGER,  -- Can be NULL for non-payment transactions
    payment_id INTEGER,  -- Must be unique when not NULL
    operation_type VARCHAR(20) NOT NULL,
    amount DECIMAL(10,2) NOT NULL,
    date TIMESTAMP NOT NULL,
    detail VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- =============================================================================
-- AUTHENTICATION & USER MANAGEMENT
-- =============================================================================

-- Users table - System users with email authentication
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(255) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    role VARCHAR(20) NOT NULL DEFAULT 'user',
    member_id INTEGER NULL,
    last_login TIMESTAMP WITH TIME ZONE,
    is_active BOOLEAN NOT NULL DEFAULT true,
    email_verified BOOLEAN NOT NULL DEFAULT false,
    email_verified_at TIMESTAMP WITH TIME ZONE NULL,
    email_verification_sent_at TIMESTAMP WITH TIME ZONE NULL,
    password_reset_sent_at TIMESTAMP WITH TIME ZONE NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Refresh tokens table - JWT refresh token management
CREATE TABLE IF NOT EXISTS refresh_tokens (
    id SERIAL PRIMARY KEY,
    uuid VARCHAR(255) NOT NULL UNIQUE,
    user_id INTEGER NOT NULL,
    expires_at BIGINT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    device_name VARCHAR(255),
    ip_address VARCHAR(45),
    user_agent TEXT,
    last_used_at TIMESTAMP WITH TIME ZONE
);

-- Verification tokens table - Email verification and password reset
CREATE TABLE IF NOT EXISTS verification_tokens (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    token VARCHAR(64) NOT NULL UNIQUE,
    type VARCHAR(20) NOT NULL DEFAULT 'email_verification',
    email VARCHAR(100) NOT NULL,
    used BOOLEAN NOT NULL DEFAULT false,
    used_at TIMESTAMP WITH TIME ZONE NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- =============================================================================
-- FOREIGN KEY CONSTRAINTS (Idempotent)
-- =============================================================================

-- Family relationships
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'families_miembro_origen_id_fkey'
    ) THEN
        ALTER TABLE families
            ADD CONSTRAINT families_miembro_origen_id_fkey
            FOREIGN KEY (miembro_origen_id) REFERENCES members(id) ON DELETE SET NULL;
    END IF;
END $$;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'familiars_familia_id_fkey'
    ) THEN
        ALTER TABLE familiars
            ADD CONSTRAINT familiars_familia_id_fkey
            FOREIGN KEY (familia_id) REFERENCES families(id) ON DELETE CASCADE;
    END IF;
END $$;

-- Payment relationships
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'payments_member_id_fkey'
    ) THEN
        ALTER TABLE payments
            ADD CONSTRAINT payments_member_id_fkey
            FOREIGN KEY (member_id) REFERENCES members(id) ON DELETE RESTRICT;
    END IF;
END $$;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'payments_membership_fee_id_fkey'
    ) THEN
        ALTER TABLE payments
            ADD CONSTRAINT payments_membership_fee_id_fkey
            FOREIGN KEY (membership_fee_id) REFERENCES membership_fees(id) ON DELETE SET NULL;
    END IF;
END $$;

-- Cash flow relationships
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'cash_flows_member_id_fkey'
    ) THEN
        ALTER TABLE cash_flows
            ADD CONSTRAINT cash_flows_member_id_fkey
            FOREIGN KEY (member_id) REFERENCES members(id) ON DELETE SET NULL;
    END IF;
END $$;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'cash_flows_payment_id_fkey'
    ) THEN
        ALTER TABLE cash_flows
            ADD CONSTRAINT cash_flows_payment_id_fkey
            FOREIGN KEY (payment_id) REFERENCES payments(id) ON DELETE SET NULL;
    END IF;
END $$;

-- Authentication relationships
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'users_member_id_fkey'
    ) THEN
        ALTER TABLE users
            ADD CONSTRAINT users_member_id_fkey
            FOREIGN KEY (member_id) REFERENCES members(id) ON DELETE SET NULL;
    END IF;
END $$;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'refresh_tokens_user_id_fkey'
    ) THEN
        ALTER TABLE refresh_tokens
            ADD CONSTRAINT refresh_tokens_user_id_fkey
            FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
    END IF;
END $$;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'verification_tokens_user_id_fkey'
    ) THEN
        ALTER TABLE verification_tokens
            ADD CONSTRAINT verification_tokens_user_id_fkey
            FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
    END IF;
END $$;

-- =============================================================================
-- UNIQUE CONSTRAINTS (From migration 002 & 004)
-- =============================================================================

-- Unique constraint on identity_card (migration 002)
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'members_identity_card_unique'
    ) THEN
        ALTER TABLE members
        ADD CONSTRAINT members_identity_card_unique
        UNIQUE (identity_card);
    END IF;
END $$;

-- Partial unique indexes for initial payments (migration 004)
-- Only one initial payment per member
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_class c
        JOIN pg_namespace n ON n.oid = c.relnamespace
        WHERE c.relname = 'unique_initial_payment_per_member'
          AND n.nspname = 'public'
    ) THEN
        CREATE UNIQUE INDEX unique_initial_payment_per_member
        ON payments (member_id)
        WHERE membership_fee_id IS NOT NULL AND deleted_at IS NULL;
    END IF;
END $$;

-- Unique constraint on payment_id in cash_flows (migration 009)
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'unique_payment_id_not_null'
          AND conrelid = 'cash_flows'::regclass
    ) THEN
        CREATE UNIQUE INDEX unique_payment_id_not_null
        ON cash_flows(payment_id)
        WHERE payment_id IS NOT NULL AND deleted_at IS NULL;
    END IF;
END $$;

-- =============================================================================
-- INDEXES FOR PERFORMANCE (Idempotent)
-- =============================================================================

-- Member indexes
CREATE INDEX IF NOT EXISTS idx_members_membership_number ON members(membership_number);
CREATE INDEX IF NOT EXISTS idx_members_state ON members(state);
CREATE INDEX IF NOT EXISTS idx_members_identity_card ON members(identity_card) WHERE identity_card IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_members_email ON members(email);

-- Family indexes
CREATE INDEX IF NOT EXISTS idx_families_numero_socio ON families(numero_socio);
CREATE INDEX IF NOT EXISTS idx_families_miembro_origen_id ON families(miembro_origen_id);
CREATE INDEX IF NOT EXISTS idx_families_deleted_at ON families(deleted_at);

-- Familiar indexes
CREATE INDEX IF NOT EXISTS idx_familiars_familia_id ON familiars(familia_id);
CREATE INDEX IF NOT EXISTS idx_familiars_deleted_at ON familiars(deleted_at);

-- Telephone indexes
CREATE INDEX IF NOT EXISTS idx_telephones_contactable ON telephones(contactable_id, contactable_type);
CREATE INDEX IF NOT EXISTS idx_telephones_deleted_at ON telephones(deleted_at);

-- Payment indexes
CREATE INDEX IF NOT EXISTS idx_payments_member_id ON payments(member_id);
CREATE INDEX IF NOT EXISTS idx_payments_status ON payments(status);
CREATE INDEX IF NOT EXISTS idx_payments_payment_date ON payments(payment_date);
CREATE INDEX IF NOT EXISTS idx_payments_membership_fee_id ON payments(membership_fee_id);
CREATE INDEX IF NOT EXISTS idx_payments_deleted_at ON payments(deleted_at);

-- Membership fee indexes
CREATE INDEX IF NOT EXISTS idx_membership_fees_year ON membership_fees(year);
CREATE INDEX IF NOT EXISTS idx_membership_fees_due_date ON membership_fees(due_date);

-- Cash flow indexes (migration 009: optimized indexes)
CREATE INDEX IF NOT EXISTS idx_cashflows_member ON cash_flows(member_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_cashflows_date ON cash_flows(date) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_cashflows_operation_type ON cash_flows(operation_type) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_cashflows_payment ON cash_flows(payment_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_cashflows_member_date ON cash_flows(member_id, date) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_cashflows_type_date ON cash_flows(operation_type, date) WHERE deleted_at IS NULL;

-- User and authentication indexes
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_member_id ON users(member_id);
CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users(deleted_at);

-- Refresh token indexes (including cleanup optimization)
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_expires_at ON refresh_tokens(expires_at);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_uuid ON refresh_tokens(uuid);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_last_used_at ON refresh_tokens(last_used_at);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_created_at ON refresh_tokens(created_at);

-- Verification token indexes
CREATE INDEX IF NOT EXISTS idx_verification_tokens_user_id ON verification_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_verification_tokens_token ON verification_tokens(token);
CREATE INDEX IF NOT EXISTS idx_verification_tokens_expires_at ON verification_tokens(expires_at);
CREATE INDEX IF NOT EXISTS idx_verification_tokens_type ON verification_tokens(type);
CREATE INDEX IF NOT EXISTS idx_verification_tokens_used ON verification_tokens(used);
CREATE INDEX IF NOT EXISTS idx_verification_tokens_deleted_at ON verification_tokens(deleted_at);

-- =============================================================================
-- TRIGGERS FOR AUTOMATIC TIMESTAMP UPDATES (Idempotent)
-- =============================================================================

DROP TRIGGER IF EXISTS update_members_updated_at ON members;
CREATE TRIGGER update_members_updated_at
    BEFORE UPDATE ON members
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_families_updated_at ON families;
CREATE TRIGGER update_families_updated_at
    BEFORE UPDATE ON families
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_familiars_updated_at ON familiars;
CREATE TRIGGER update_familiars_updated_at
    BEFORE UPDATE ON familiars
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_telephones_updated_at ON telephones;
CREATE TRIGGER update_telephones_updated_at
    BEFORE UPDATE ON telephones
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_membership_fees_updated_at ON membership_fees;
CREATE TRIGGER update_membership_fees_updated_at
    BEFORE UPDATE ON membership_fees
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_payments_updated_at ON payments;
CREATE TRIGGER update_payments_updated_at
    BEFORE UPDATE ON payments
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_cash_flows_updated_at ON cash_flows;
CREATE TRIGGER update_cash_flows_updated_at
    BEFORE UPDATE ON cash_flows
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_users_updated_at ON users;
CREATE TRIGGER update_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_verification_tokens_updated_at ON verification_tokens;
CREATE TRIGGER update_verification_tokens_updated_at
    BEFORE UPDATE ON verification_tokens
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- =============================================================================
-- COMMENTS AND DOCUMENTATION
-- =============================================================================

COMMENT ON SCHEMA public IS 'ASAM (Asociación de Senegaleses y Amigos de Montmeló) - Complete database schema v2.0';

-- Table comments
COMMENT ON TABLE members IS 'Individual members of the association';
COMMENT ON TABLE families IS 'Family groups with mixed Spanish field names per existing model';
COMMENT ON TABLE familiars IS 'Family relatives with Spanish field names per existing model';
COMMENT ON TABLE telephones IS 'Polymorphic phone numbers with Spanish field name per existing model';
COMMENT ON TABLE membership_fees IS 'Annual membership fee definitions. One fee per year, due date is always December 31st';
COMMENT ON TABLE payments IS 'Payment records from members. For families, use family.miembro_origen_id as member_id';
COMMENT ON TABLE cash_flows IS 'Financial movements and transaction tracking';
COMMENT ON TABLE users IS 'System users with authentication and email verification';
COMMENT ON TABLE refresh_tokens IS 'JWT refresh token management for secure authentication';
COMMENT ON TABLE verification_tokens IS 'Email verification and password reset tokens';

-- Key column comments
COMMENT ON COLUMN members.identity_card IS 'Identity card number (DNI/NIE). Must be unique when present';
COMMENT ON COLUMN membership_fees.year IS 'Year of the membership fee. UNIQUE constraint ensures only one fee per year';
COMMENT ON COLUMN membership_fees.due_date IS 'Due date for the fee payment. Always set to December 31st of the year';
COMMENT ON COLUMN payments.member_id IS 'Member associated with the payment. For family payments, use family.miembro_origen_id';
COMMENT ON COLUMN payments.payment_date IS 'Date when the payment was made. NULL for pending payments, set when payment status changes to paid';
COMMENT ON COLUMN payments.membership_fee_id IS 'Links to the annual fee this payment is for. Initial payments will have this set';
COMMENT ON COLUMN cash_flows.payment_id IS 'Links to payment record. Must be unique when not NULL (one cash_flow per payment)';
COMMENT ON COLUMN users.email IS 'User email address, required for notifications and authentication';
COMMENT ON COLUMN users.email_verified_at IS 'Timestamp when email was verified, NULL if not verified';
COMMENT ON COLUMN users.email_verification_sent_at IS 'Last time verification email was sent';
COMMENT ON COLUMN users.password_reset_sent_at IS 'Last time password reset email was sent';
COMMENT ON COLUMN users.member_id IS 'Optional link to member record for members who have user accounts';
COMMENT ON COLUMN verification_tokens.type IS 'Type of token: email_verification, password_reset, etc.';
COMMENT ON COLUMN verification_tokens.email IS 'Email address associated with the token';
COMMENT ON COLUMN verification_tokens.used IS 'Whether the token has been used';
COMMENT ON COLUMN verification_tokens.used_at IS 'Timestamp when the token was used';
COMMENT ON COLUMN refresh_tokens.uuid IS 'Unique identifier for the refresh token';
COMMENT ON COLUMN cash_flows.operation_type IS 'Type of operation: income, expense, transfer, etc.';

-- =============================================================================
-- MIGRATION HISTORY
-- =============================================================================
-- This consolidated migration includes changes from:
-- - 000001: Complete initial schema
-- - 000002: Add unique constraint on identity_card
-- - 000003: Allow NULL member_id (later reverted in 008)
-- - 000004: Prevent duplicate initial payments
-- - 000005: Make payment_date nullable
-- - 000006: Remove status from membership_fees
-- - 000007: Remove payment_id from membership_fees
-- - 000008: Remove family_id from payments and cash_flows
-- - 000009: Add cash_flow constraints and optimized indexes

-- =============================================================================
-- NOTA SOBRE USUARIOS INICIALES
-- =============================================================================
-- Los usuarios administradores iniciales deben crearse usando el comando seed:
-- ADMIN_EMAIL=admin@example.com ADMIN_PASSWORD=SecurePass123! go run cmd/seed/main.go
--
-- NUNCA incluyas contraseñas en los archivos de migración, incluso si están hasheadas.
-- Los usuarios creados con el comando seed tendrán email_verified=false y deberán
-- verificar su email en el primer acceso.

-- Schema version for reference
COMMENT ON EXTENSION "uuid-ossp" IS 'ASAM Schema v2.0 - Consolidated single migration - IDEMPOTENT';
