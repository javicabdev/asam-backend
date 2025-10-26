--!postgresql
-- Migration: Prevent Duplicate Initial Payments
-- VERSION: 1.1
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
-- STEP 2: Add unique constraints for initial payments - IDEMPOTENT
-- =============================================================================

-- Constraint: Only one initial payment per member
-- Check if constraint already exists before creating
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint 
        WHERE conname = 'unique_initial_payment_per_member'
    ) THEN
        ALTER TABLE payments
        ADD CONSTRAINT unique_initial_payment_per_member
        UNIQUE (member_id)
        WHERE membership_fee_id IS NOT NULL AND member_id IS NOT NULL;
        
        RAISE NOTICE 'Created constraint: unique_initial_payment_per_member';
    ELSE
        RAISE NOTICE 'Constraint already exists: unique_initial_payment_per_member';
    END IF;
END $$;

-- Constraint: Only one initial payment per family
-- Check if constraint already exists before creating
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint 
        WHERE conname = 'unique_initial_payment_per_family'
    ) THEN
        ALTER TABLE payments
        ADD CONSTRAINT unique_initial_payment_per_family
        UNIQUE (family_id)
        WHERE membership_fee_id IS NOT NULL AND family_id IS NOT NULL;
        
        RAISE NOTICE 'Created constraint: unique_initial_payment_per_family';
    ELSE
        RAISE NOTICE 'Constraint already exists: unique_initial_payment_per_family';
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

-- Verify constraints exist:
-- SELECT conname, contype, pg_get_constraintdef(oid) 
-- FROM pg_constraint 
-- WHERE conname IN ('unique_initial_payment_per_member', 'unique_initial_payment_per_family');
