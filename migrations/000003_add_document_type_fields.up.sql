-- Add document_type fields to support multiple document types (DNI/NIE, Senegal Passport, Other)
-- This migration adds columns to track the type of identity document for members, families, and familiars
-- IDEMPOTENT: Safe to run multiple times

DO $$
BEGIN
    -- Add document_type to members table
    IF NOT EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_name = 'members'
          AND column_name = 'document_type'
    ) THEN
        ALTER TABLE members
        ADD COLUMN document_type VARCHAR(20);

        RAISE NOTICE 'Added document_type column to members table';
    ELSE
        RAISE NOTICE 'document_type column already exists in members table, skipping';
    END IF;

    -- Add esposo_document_type to families table
    IF NOT EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_name = 'families'
          AND column_name = 'esposo_document_type'
    ) THEN
        ALTER TABLE families
        ADD COLUMN esposo_document_type VARCHAR(20);

        RAISE NOTICE 'Added esposo_document_type column to families table';
    ELSE
        RAISE NOTICE 'esposo_document_type column already exists in families table, skipping';
    END IF;

    -- Add esposa_document_type to families table
    IF NOT EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_name = 'families'
          AND column_name = 'esposa_document_type'
    ) THEN
        ALTER TABLE families
        ADD COLUMN esposa_document_type VARCHAR(20);

        RAISE NOTICE 'Added esposa_document_type column to families table';
    ELSE
        RAISE NOTICE 'esposa_document_type column already exists in families table, skipping';
    END IF;

    -- Add document_type to familiars table
    IF NOT EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_name = 'familiars'
          AND column_name = 'document_type'
    ) THEN
        ALTER TABLE familiars
        ADD COLUMN document_type VARCHAR(20);

        RAISE NOTICE 'Added document_type column to familiars table';
    ELSE
        RAISE NOTICE 'document_type column already exists in familiars table, skipping';
    END IF;
END $$;

-- Add comments explaining the new columns
COMMENT ON COLUMN members.document_type IS 'Type of identity document: DNI_NIE, SENEGAL_PASSPORT, or OTHER';
COMMENT ON COLUMN families.esposo_document_type IS 'Type of identity document for husband: DNI_NIE, SENEGAL_PASSPORT, or OTHER';
COMMENT ON COLUMN families.esposa_document_type IS 'Type of identity document for wife: DNI_NIE, SENEGAL_PASSPORT, or OTHER';
COMMENT ON COLUMN familiars.document_type IS 'Type of identity document: DNI_NIE, SENEGAL_PASSPORT, or OTHER';
