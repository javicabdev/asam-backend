-- Rollback: Restore NOT NULL constraint on member_id
-- WARNING: This migration will fail if there are payments with NULL member_id
-- IDEMPOTENT: Safe to run multiple times

-- First, remove the check constraint (idempotent)
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM pg_constraint 
        WHERE conname = 'payments_member_or_family_check'
    ) THEN
        ALTER TABLE payments DROP CONSTRAINT payments_member_or_family_check;
    END IF;
END $$;

-- Delete any payments that have NULL member_id (family-only payments)
-- This is necessary to restore the NOT NULL constraint
-- This operation is idempotent: only deletes if records exist
DELETE FROM payments WHERE member_id IS NULL;

-- Restore the NOT NULL constraint (idempotent)
DO $$
BEGIN
    -- Check if the column already has NOT NULL constraint
    IF EXISTS (
        SELECT 1 
        FROM information_schema.columns 
        WHERE table_name = 'payments' 
        AND column_name = 'member_id' 
        AND is_nullable = 'YES'
    ) THEN
        ALTER TABLE payments ALTER COLUMN member_id SET NOT NULL;
    END IF;
END $$;

-- Restore original comments
COMMENT ON COLUMN payments.member_id IS 'Member associated with the payment';
COMMENT ON COLUMN payments.family_id IS 'Optional family associated with the payment';
