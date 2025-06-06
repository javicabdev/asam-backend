-- Script para ejecutar todas las migraciones manualmente
-- Copiar el contenido de 000001_initial_schema.up.sql aquí

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

-- Create families table (from Family model)
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

-- Create familiars table (from Familiar model)
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

-- Create telephones table (from Telephone model - polymorphic)
CREATE TABLE IF NOT EXISTS telephones (
    id SERIAL PRIMARY KEY,
    numero_telefono VARCHAR(20) NOT NULL,
    contactable_id INTEGER NOT NULL,
    contactable_type VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Create membership_fees table (from MembershipFee model)
CREATE TABLE IF NOT EXISTS membership_fees (
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
CREATE TABLE IF NOT EXISTS payments (
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
CREATE TABLE IF NOT EXISTS cash_flows (
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
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    role VARCHAR(20) NOT NULL DEFAULT 'user',
    last_login TIMESTAMP WITH TIME ZONE,
    is_active BOOLEAN NOT NULL DEFAULT true,
    refresh_token VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Add foreign key constraints (only if they don't exist)
DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'families_miembro_origen_id_fkey') THEN
        ALTER TABLE families ADD CONSTRAINT families_miembro_origen_id_fkey 
        FOREIGN KEY (miembro_origen_id) REFERENCES members(id);
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'familiars_familia_id_fkey') THEN
        ALTER TABLE familiars ADD CONSTRAINT familiars_familia_id_fkey 
        FOREIGN KEY (familia_id) REFERENCES families(id);
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'payments_member_id_fkey') THEN
        ALTER TABLE payments ADD CONSTRAINT payments_member_id_fkey 
        FOREIGN KEY (member_id) REFERENCES members(id) ON DELETE RESTRICT;
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'payments_family_id_fkey') THEN
        ALTER TABLE payments ADD CONSTRAINT payments_family_id_fkey 
        FOREIGN KEY (family_id) REFERENCES families(id) ON DELETE RESTRICT;
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'payments_membership_fee_id_fkey') THEN
        ALTER TABLE payments ADD CONSTRAINT payments_membership_fee_id_fkey 
        FOREIGN KEY (membership_fee_id) REFERENCES membership_fees(id) ON DELETE SET NULL;
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'membership_fees_payment_id_fkey') THEN
        ALTER TABLE membership_fees ADD CONSTRAINT membership_fees_payment_id_fkey 
        FOREIGN KEY (payment_id) REFERENCES payments(id) ON DELETE SET NULL;
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'cash_flows_member_id_fkey') THEN
        ALTER TABLE cash_flows ADD CONSTRAINT cash_flows_member_id_fkey 
        FOREIGN KEY (member_id) REFERENCES members(id) ON DELETE SET NULL;
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'cash_flows_family_id_fkey') THEN
        ALTER TABLE cash_flows ADD CONSTRAINT cash_flows_family_id_fkey 
        FOREIGN KEY (family_id) REFERENCES families(id) ON DELETE SET NULL;
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'cash_flows_payment_id_fkey') THEN
        ALTER TABLE cash_flows ADD CONSTRAINT cash_flows_payment_id_fkey 
        FOREIGN KEY (payment_id) REFERENCES payments(id) ON DELETE SET NULL;
    END IF;
END $$;

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_members_membership_number ON members(membership_number);
CREATE INDEX IF NOT EXISTS idx_members_state ON members(state);
CREATE INDEX IF NOT EXISTS idx_members_identity_card ON members(identity_card);

CREATE INDEX IF NOT EXISTS idx_families_numero_socio ON families(numero_socio);
CREATE INDEX IF NOT EXISTS idx_families_miembro_origen_id ON families(miembro_origen_id);
CREATE INDEX IF NOT EXISTS idx_families_deleted_at ON families(deleted_at);

CREATE INDEX IF NOT EXISTS idx_familiars_familia_id ON familiars(familia_id);
CREATE INDEX IF NOT EXISTS idx_familiars_deleted_at ON familiars(deleted_at);

CREATE INDEX IF NOT EXISTS idx_telephones_contactable ON telephones(contactable_id, contactable_type);
CREATE INDEX IF NOT EXISTS idx_telephones_deleted_at ON telephones(deleted_at);

CREATE INDEX IF NOT EXISTS idx_payments_member_id ON payments(member_id);
CREATE INDEX IF NOT EXISTS idx_payments_family_id ON payments(family_id);
CREATE INDEX IF NOT EXISTS idx_payments_status ON payments(status);
CREATE INDEX IF NOT EXISTS idx_payments_payment_date ON payments(payment_date);
CREATE INDEX IF NOT EXISTS idx_payments_deleted_at ON payments(deleted_at);

CREATE INDEX IF NOT EXISTS idx_cash_flows_member_id ON cash_flows(member_id);
CREATE INDEX IF NOT EXISTS idx_cash_flows_family_id ON cash_flows(family_id);
CREATE INDEX IF NOT EXISTS idx_cash_flows_payment_id ON cash_flows(payment_id);
CREATE INDEX IF NOT EXISTS idx_cash_flows_operation_type ON cash_flows(operation_type);
CREATE INDEX IF NOT EXISTS idx_cash_flows_date ON cash_flows(date);
CREATE INDEX IF NOT EXISTS idx_cash_flows_deleted_at ON cash_flows(deleted_at);

CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users(deleted_at);

-- Create triggers for updated_at
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

-- Add comments
COMMENT ON SCHEMA public IS 'ASAM (Asociación de Senegaleses y Amigos de Montmeló) database schema';
COMMENT ON TABLE members IS 'Individual members of the association';
COMMENT ON TABLE families IS 'Family groups with mixed Spanish field names per model';
COMMENT ON TABLE familiars IS 'Family relatives with Spanish field names per model';
COMMENT ON TABLE telephones IS 'Polymorphic phone numbers with Spanish field name';

-- Show created tables
SELECT 'Tables created successfully!' as message;
SELECT table_name FROM information_schema.tables WHERE table_schema = 'public' ORDER BY table_name;
