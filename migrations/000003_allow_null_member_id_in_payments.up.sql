-- Migration: Allow NULL member_id in payments table
-- Purpose: Enable family-only payments without requiring a member_id
-- This allows payments to be associated with either a member OR a family
-- IDEMPOTENT: Safe to run multiple times

-- Remove NOT NULL constraint from member_id column (idempotent)
DO $$
BEGIN
    -- Check if the column already allows NULL
    IF EXISTS (
        SELECT 1 
        FROM information_schema.columns 
        WHERE table_name = 'payments' 
        AND column_name = 'member_id' 
        AND is_nullable = 'NO'
    ) THEN
        ALTER TABLE payments ALTER COLUMN member_id DROP NOT NULL;
    END IF;
END $$;

-- Add a check constraint to ensure at least one of member_id or family_id is present (idempotent)
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint 
        WHERE conname = 'payments_member_or_family_check'
    ) THEN
        ALTER TABLE payments 
            ADD CONSTRAINT payments_member_or_family_check 
            CHECK (member_id IS NOT NULL OR family_id IS NOT NULL);
    END IF;
END $$;

-- Update comment to reflect the change
COMMENT ON COLUMN payments.member_id IS 'Member associated with the payment. NULL if payment is family-only. At least one of member_id or family_id must be present.';
COMMENT ON COLUMN payments.family_id IS 'Family associated with the payment. NULL if payment is member-only. At least one of member_id or family_id must be present.';
