--!postgresql
-- ASAM Backend - Complete Database Schema
-- Single migration containing the complete final schema for development phase
-- Includes all features: members, families, users, authentication, email verification, payments, etc.

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
CREATE TABLE members (
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
CREATE TABLE families (
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
CREATE TABLE familiars (
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
CREATE TABLE telephones (
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

-- Membership fees table - Monthly fee definitions
CREATE TABLE membership_fees (
    id SERIAL PRIMARY KEY,
    year INTEGER NOT NULL,
    month INTEGER NOT NULL,
    base_fee_amount DECIMAL(10,2) NOT NULL,
    family_fee_extra DECIMAL(10,2) NOT NULL DEFAULT 0,
    status VARCHAR(255) NOT NULL DEFAULT 'pending',
    due_date DATE NOT NULL,
    payment_id INTEGER,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Payments table - Payment records
CREATE TABLE payments (
    id SERIAL PRIMARY KEY,
    member_id INTEGER NOT NULL,
    family_id INTEGER,
    amount DECIMAL(10,2) NOT NULL,
    payment_date DATE NOT NULL,
    status VARCHAR(255) NOT NULL,
    payment_method VARCHAR(255) NOT NULL,
    notes TEXT,
    membership_fee_id INTEGER,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Cash flows table - Financial movements tracking
CREATE TABLE cash_flows (
    id SERIAL PRIMARY KEY,
    member_id INTEGER,
    family_id INTEGER,
    payment_id INTEGER,
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
CREATE TABLE users (
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
CREATE TABLE refresh_tokens (
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
CREATE TABLE verification_tokens (
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
-- FOREIGN KEY CONSTRAINTS
-- =============================================================================

-- Family relationships
ALTER TABLE families 
    ADD CONSTRAINT families_miembro_origen_id_fkey 
    FOREIGN KEY (miembro_origen_id) REFERENCES members(id) ON DELETE SET NULL;

ALTER TABLE familiars 
    ADD CONSTRAINT familiars_familia_id_fkey 
    FOREIGN KEY (familia_id) REFERENCES families(id) ON DELETE CASCADE;

-- Payment relationships
ALTER TABLE payments 
    ADD CONSTRAINT payments_member_id_fkey 
    FOREIGN KEY (member_id) REFERENCES members(id) ON DELETE RESTRICT;

ALTER TABLE payments 
    ADD CONSTRAINT payments_family_id_fkey 
    FOREIGN KEY (family_id) REFERENCES families(id) ON DELETE RESTRICT;

ALTER TABLE payments 
    ADD CONSTRAINT payments_membership_fee_id_fkey 
    FOREIGN KEY (membership_fee_id) REFERENCES membership_fees(id) ON DELETE SET NULL;

ALTER TABLE membership_fees 
    ADD CONSTRAINT membership_fees_payment_id_fkey 
    FOREIGN KEY (payment_id) REFERENCES payments(id) ON DELETE SET NULL;

-- Cash flow relationships
ALTER TABLE cash_flows 
    ADD CONSTRAINT cash_flows_member_id_fkey 
    FOREIGN KEY (member_id) REFERENCES members(id) ON DELETE SET NULL;

ALTER TABLE cash_flows 
    ADD CONSTRAINT cash_flows_family_id_fkey 
    FOREIGN KEY (family_id) REFERENCES families(id) ON DELETE SET NULL;

ALTER TABLE cash_flows 
    ADD CONSTRAINT cash_flows_payment_id_fkey 
    FOREIGN KEY (payment_id) REFERENCES payments(id) ON DELETE SET NULL;

-- Authentication relationships
ALTER TABLE users 
    ADD CONSTRAINT users_member_id_fkey 
    FOREIGN KEY (member_id) REFERENCES members(id) ON DELETE SET NULL;

ALTER TABLE refresh_tokens 
    ADD CONSTRAINT refresh_tokens_user_id_fkey 
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE verification_tokens 
    ADD CONSTRAINT verification_tokens_user_id_fkey 
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

-- =============================================================================
-- INDEXES FOR PERFORMANCE
-- =============================================================================

-- Member indexes
CREATE INDEX idx_members_membership_number ON members(membership_number);
CREATE INDEX idx_members_state ON members(state);
CREATE INDEX idx_members_identity_card ON members(identity_card);
CREATE INDEX idx_members_email ON members(email);

-- Family indexes  
CREATE INDEX idx_families_numero_socio ON families(numero_socio);
CREATE INDEX idx_families_miembro_origen_id ON families(miembro_origen_id);
CREATE INDEX idx_families_deleted_at ON families(deleted_at);

-- Familiar indexes
CREATE INDEX idx_familiars_familia_id ON familiars(familia_id);
CREATE INDEX idx_familiars_deleted_at ON familiars(deleted_at);

-- Telephone indexes
CREATE INDEX idx_telephones_contactable ON telephones(contactable_id, contactable_type);
CREATE INDEX idx_telephones_deleted_at ON telephones(deleted_at);

-- Payment indexes
CREATE INDEX idx_payments_member_id ON payments(member_id);
CREATE INDEX idx_payments_family_id ON payments(family_id);
CREATE INDEX idx_payments_status ON payments(status);
CREATE INDEX idx_payments_payment_date ON payments(payment_date);
CREATE INDEX idx_payments_deleted_at ON payments(deleted_at);

-- Cash flow indexes
CREATE INDEX idx_cash_flows_member_id ON cash_flows(member_id);
CREATE INDEX idx_cash_flows_family_id ON cash_flows(family_id);
CREATE INDEX idx_cash_flows_payment_id ON cash_flows(payment_id);
CREATE INDEX idx_cash_flows_operation_type ON cash_flows(operation_type);
CREATE INDEX idx_cash_flows_date ON cash_flows(date);
CREATE INDEX idx_cash_flows_deleted_at ON cash_flows(deleted_at);

-- User and authentication indexes
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_member_id ON users(member_id);
CREATE INDEX idx_users_deleted_at ON users(deleted_at);

-- Refresh token indexes (including cleanup optimization)
CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX idx_refresh_tokens_expires_at ON refresh_tokens(expires_at);
CREATE INDEX idx_refresh_tokens_uuid ON refresh_tokens(uuid);
CREATE INDEX idx_refresh_tokens_last_used_at ON refresh_tokens(last_used_at);
CREATE INDEX idx_refresh_tokens_created_at ON refresh_tokens(created_at);

-- Verification token indexes
CREATE INDEX idx_verification_tokens_user_id ON verification_tokens(user_id);
CREATE INDEX idx_verification_tokens_token ON verification_tokens(token);
CREATE INDEX idx_verification_tokens_expires_at ON verification_tokens(expires_at);
CREATE INDEX idx_verification_tokens_type ON verification_tokens(type);
CREATE INDEX idx_verification_tokens_used ON verification_tokens(used);
CREATE INDEX idx_verification_tokens_deleted_at ON verification_tokens(deleted_at);

-- =============================================================================
-- TRIGGERS FOR AUTOMATIC TIMESTAMP UPDATES
-- =============================================================================

CREATE TRIGGER update_members_updated_at
    BEFORE UPDATE ON members
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_families_updated_at
    BEFORE UPDATE ON families
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_familiars_updated_at
    BEFORE UPDATE ON familiars
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_telephones_updated_at
    BEFORE UPDATE ON telephones
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_membership_fees_updated_at
    BEFORE UPDATE ON membership_fees
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_payments_updated_at
    BEFORE UPDATE ON payments
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_cash_flows_updated_at
    BEFORE UPDATE ON cash_flows
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_verification_tokens_updated_at
    BEFORE UPDATE ON verification_tokens
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- =============================================================================
-- COMMENTS AND DOCUMENTATION
-- =============================================================================

COMMENT ON SCHEMA public IS 'ASAM (Asociación de Senegaleses y Amigos de Montmeló) - Complete database schema';

-- Table comments
COMMENT ON TABLE members IS 'Individual members of the association';
COMMENT ON TABLE families IS 'Family groups with mixed Spanish field names per existing model';
COMMENT ON TABLE familiars IS 'Family relatives with Spanish field names per existing model';
COMMENT ON TABLE telephones IS 'Polymorphic phone numbers with Spanish field name per existing model';
COMMENT ON TABLE membership_fees IS 'Monthly membership fee definitions';
COMMENT ON TABLE payments IS 'Payment records from members and families';
COMMENT ON TABLE cash_flows IS 'Financial movements and transaction tracking';
COMMENT ON TABLE users IS 'System users with authentication and email verification';
COMMENT ON TABLE refresh_tokens IS 'JWT refresh token management for secure authentication';
COMMENT ON TABLE verification_tokens IS 'Email verification and password reset tokens';

-- Key column comments
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
-- NOTA SOBRE USUARIOS INICIALES
-- =============================================================================
-- Los usuarios administradores iniciales deben crearse usando el comando seed:
-- ADMIN_EMAIL=admin@example.com ADMIN_PASSWORD=SecurePass123! go run cmd/seed/main.go
-- 
-- NUNCA incluyas contraseñas en los archivos de migración, incluso si están hasheadas.
-- Los usuarios creados con el comando seed tendrán email_verified=false y deberán
-- verificar su email en el primer acceso.

-- =============================================================================
-- SCHEMA CREATION COMPLETE
-- =============================================================================

-- Schema version for reference
COMMENT ON EXTENSION "uuid-ossp" IS 'ASAM Schema v1.0 - Complete consolidated schema for development phase';
