-- Initial schema creation based on existing Go models
-- This migration creates tables that match the GORM models exactly

-- Enable UUID extension (if needed in the future)
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create update_updated_at_column function
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create members table (from Member model)
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

-- Create families table (from Family model)
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

-- Create familiars table (from Familiar model)
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

-- Create telephones table (from Telephone model - polymorphic)
CREATE TABLE telephones (
    id SERIAL PRIMARY KEY,
    numero_telefono VARCHAR(20) NOT NULL,
    contactable_id INTEGER NOT NULL,
    contactable_type VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Create membership_fees table (from MembershipFee model)
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

-- Create payments table (from Payment model)
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

-- Create cash_flows table (from CashFlow model)
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

-- Create users table (from User model)
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(255) NOT NULL,
    password VARCHAR(255) NOT NULL,
    role VARCHAR(20) NOT NULL DEFAULT 'user',
    last_login TIMESTAMP WITH TIME ZONE,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    CONSTRAINT uni_users_username UNIQUE (username)
);

-- Create refresh_tokens table for JWT authentication
CREATE TABLE refresh_tokens (
    id SERIAL PRIMARY KEY,
    uuid VARCHAR(255) NOT NULL,
    user_id INTEGER NOT NULL,
    expires_at BIGINT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    device_name VARCHAR(255),
    ip_address VARCHAR(45),
    user_agent TEXT,
    last_used_at TIMESTAMP WITH TIME ZONE,
    CONSTRAINT uni_refresh_tokens_uuid UNIQUE (uuid)
);

-- Add foreign key constraints
ALTER TABLE families 
    ADD CONSTRAINT families_miembro_origen_id_fkey 
    FOREIGN KEY (miembro_origen_id) REFERENCES members(id);

ALTER TABLE familiars 
    ADD CONSTRAINT familiars_familia_id_fkey 
    FOREIGN KEY (familia_id) REFERENCES families(id);

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

ALTER TABLE cash_flows 
    ADD CONSTRAINT cash_flows_member_id_fkey 
    FOREIGN KEY (member_id) REFERENCES members(id) ON DELETE SET NULL;

ALTER TABLE cash_flows 
    ADD CONSTRAINT cash_flows_family_id_fkey 
    FOREIGN KEY (family_id) REFERENCES families(id) ON DELETE SET NULL;

ALTER TABLE cash_flows 
    ADD CONSTRAINT cash_flows_payment_id_fkey 
    FOREIGN KEY (payment_id) REFERENCES payments(id) ON DELETE SET NULL;

ALTER TABLE refresh_tokens 
    ADD CONSTRAINT refresh_tokens_user_id_fkey 
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

-- Create indexes for better performance
CREATE INDEX idx_members_membership_number ON members(membership_number);
CREATE INDEX idx_members_state ON members(state);
CREATE INDEX idx_members_identity_card ON members(identity_card);

CREATE INDEX idx_families_numero_socio ON families(numero_socio);
CREATE INDEX idx_families_miembro_origen_id ON families(miembro_origen_id);
CREATE INDEX idx_families_deleted_at ON families(deleted_at);

CREATE INDEX idx_familiars_familia_id ON familiars(familia_id);
CREATE INDEX idx_familiars_deleted_at ON familiars(deleted_at);

CREATE INDEX idx_telephones_contactable ON telephones(contactable_id, contactable_type);
CREATE INDEX idx_telephones_deleted_at ON telephones(deleted_at);

CREATE INDEX idx_payments_member_id ON payments(member_id);
CREATE INDEX idx_payments_family_id ON payments(family_id);
CREATE INDEX idx_payments_status ON payments(status);
CREATE INDEX idx_payments_payment_date ON payments(payment_date);
CREATE INDEX idx_payments_deleted_at ON payments(deleted_at);

CREATE INDEX idx_cash_flows_member_id ON cash_flows(member_id);
CREATE INDEX idx_cash_flows_family_id ON cash_flows(family_id);
CREATE INDEX idx_cash_flows_payment_id ON cash_flows(payment_id);
CREATE INDEX idx_cash_flows_operation_type ON cash_flows(operation_type);
CREATE INDEX idx_cash_flows_date ON cash_flows(date);
CREATE INDEX idx_cash_flows_deleted_at ON cash_flows(deleted_at);

CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_deleted_at ON users(deleted_at);

CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX idx_refresh_tokens_expires_at ON refresh_tokens(expires_at);
CREATE INDEX idx_refresh_tokens_uuid ON refresh_tokens(uuid);
CREATE INDEX idx_refresh_tokens_last_used_at ON refresh_tokens(last_used_at);

-- Create triggers for updated_at
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

-- Add comments
COMMENT ON SCHEMA public IS 'ASAM (Asociación de Senegaleses y Amigos de Montmeló) database schema';
COMMENT ON TABLE members IS 'Individual members of the association';
COMMENT ON TABLE families IS 'Family groups with mixed Spanish field names per model';
COMMENT ON TABLE familiars IS 'Family relatives with Spanish field names per model';
COMMENT ON TABLE telephones IS 'Polymorphic phone numbers with Spanish field name';
