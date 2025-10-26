--!postgresql
-- Migration: Prevent Duplicate Initial Payments
-- VERSION: 1.2
-- PURPOSE: Add database constraints to prevent duplicate initial payments for members and families
-- IDEMPOTENT: Safe to run multiple times without errors

-- =============================================================================
-- STEP 1: Clean up existing duplicates (if any) - IDEMPOTENT
-- =============================================================================

-- For members: Keep only the oldest payment (by created_at) for each member with initial payment
-- Only delete if duplicates actually exist
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM payments
        WHERE member_id IS NOT NULL 
          AND membership_fee_id IS NOT NULL
        GROUP BY member_id
        HAVING COUNT(*) > 1
    ) THEN
        WITH duplicates AS (
            SELECT 
                id,
                ROW_NUMBER() OVER (
                    PARTITION BY member_id 
                    ORDER BY created_at ASC
                ) as rn
            FROM payments
            WHERE member_id IS NOT NULL 
              AND membership_fee_id IS NOT NULL
        )
        DELETE FROM payments
        WHERE id IN (
            SELECT id FROM duplicates WHERE rn > 1
        );
        
        RAISE NOTICE 'Cleaned up duplicate initial payments for members';
    ELSE
        RAISE NOTICE 'No duplicate initial payments found for members';
    END IF;
END $$;

-- For families: Keep only the oldest payment (by created_at) for each family with initial payment
-- Only delete if duplicates actually exist
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM payments
        WHERE family_id IS NOT NULL 
          AND membership_fee_id IS NOT NULL
        GROUP BY family_id
        HAVING COUNT(*) > 1
    ) THEN
        WITH duplicates AS (
            SELECT 
                id,
                ROW_NUMBER() OVER (
                    PARTITION BY family_id 
                    ORDER BY created_at ASC
                ) as rn
            FROM payments
            WHERE family_id IS NOT NULL 
              AND membership_fee_id IS NOT NULL
        )
        DELETE FROM payments
        WHERE id IN (
            SELECT id FROM duplicates WHERE rn > 1
        );
        
        RAISE NOTICE 'Cleaned up duplicate initial payments for families';
    ELSE
        RAISE NOTICE 'No duplicate initial payments found for families';
    END IF;
END $$;

-- =============================================================================
-- STEP 2: Add partial unique indexes for initial payments - IDEMPOTENT
-- =============================================================================

-- Index: Only one initial payment per member
-- Check if index already exists before creating
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
        WHERE membership_fee_id IS NOT NULL AND member_id IS NOT NULL;
        
        RAISE NOTICE 'Created index: unique_initial_payment_per_member';
    ELSE
        RAISE NOTICE 'Index already exists: unique_initial_payment_per_member';
    END IF;
END $$;

-- Index: Only one initial payment per family
-- Check if index already exists before creating
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 
        FROM pg_class c
        JOIN pg_namespace n ON n.oid = c.relnamespace
        WHERE c.relname = 'unique_initial_payment_per_family'
          AND n.nspname = 'public'
    ) THEN
        CREATE UNIQUE INDEX unique_initial_payment_per_family 
        ON payments (family_id)
        WHERE membership_fee_id IS NOT NULL AND family_id IS NOT NULL;
        
        RAISE NOTICE 'Created index: unique_initial_payment_per_family';
    ELSE
        RAISE NOTICE 'Index already exists: unique_initial_payment_per_family';
    END IF;
END $$;

-- =============================================================================
-- VERIFICATION QUERIES (for manual testing after migration)
-- =============================================================================

-- Check for remaining duplicates (should return 0 rows):
-- SELECT member_id, COUNT(*) 
-- FROM payments 
-- WHERE membership_fee_id IS NOT NULL AND member_id IS NOT NULL
-- GROUP BY member_id 
-- HAVING COUNT(*) > 1;

-- SELECT family_id, COUNT(*) 
-- FROM payments 
-- WHERE membership_fee_id IS NOT NULL AND family_id IS NOT NULL
-- GROUP BY family_id 
-- HAVING COUNT(*) > 1;

-- Verify indexes exist:
-- SELECT schemaname, tablename, indexname, indexdef
-- FROM pg_indexes
-- WHERE indexname IN ('unique_initial_payment_per_member', 'unique_initial_payment_per_family');
