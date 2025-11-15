-- Rollback: Remove document_type fields from members, families, and familiares tables

DO $$
BEGIN
    -- Remove document_type from members table
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_name = 'members'
          AND column_name = 'document_type'
    ) THEN
        ALTER TABLE members
        DROP COLUMN document_type;

        RAISE NOTICE 'Removed document_type column from members table';
    ELSE
        RAISE NOTICE 'document_type column does not exist in members table, skipping';
    END IF;

    -- Remove esposo_document_type from families table
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_name = 'families'
          AND column_name = 'esposo_document_type'
    ) THEN
        ALTER TABLE families
        DROP COLUMN esposo_document_type;

        RAISE NOTICE 'Removed esposo_document_type column from families table';
    ELSE
        RAISE NOTICE 'esposo_document_type column does not exist in families table, skipping';
    END IF;

    -- Remove esposa_document_type from families table
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_name = 'families'
          AND column_name = 'esposa_document_type'
    ) THEN
        ALTER TABLE families
        DROP COLUMN esposa_document_type;

        RAISE NOTICE 'Removed esposa_document_type column from families table';
    ELSE
        RAISE NOTICE 'esposa_document_type column does not exist in families table, skipping';
    END IF;

    -- Remove document_type from familiares table
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_name = 'familiares'
          AND column_name = 'document_type'
    ) THEN
        ALTER TABLE familiares
        DROP COLUMN document_type;

        RAISE NOTICE 'Removed document_type column from familiares table';
    ELSE
        RAISE NOTICE 'document_type column does not exist in familiares table, skipping';
    END IF;
END $$;
