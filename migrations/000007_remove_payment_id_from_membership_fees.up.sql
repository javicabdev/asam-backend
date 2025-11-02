-- Migration: Remove payment_id column from membership_fees table
-- Reason: The payment_id field creates an incorrect one-to-one relationship
-- from MembershipFee to Payment. This is backwards - a single membership fee
-- (e.g., 2024 annual fee) should have MANY payments (one per member/family),
-- not point to a single payment.
--
-- Correct relationship already exists:
-- - payments.membership_fee_id → membership_fees.id (many-to-one)
--
-- Problems with payment_id:
-- 1. Creates confusing bidirectional relationship
-- 2. Suggests one fee = one payment (incorrect)
-- 3. Not used in production code (only in one test builder)
-- 4. Could lead to bugs similar to the removed 'status' field

-- Drop the foreign key constraint first (idempotent)
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'membership_fees_payment_id_fkey'
    ) THEN
        ALTER TABLE membership_fees
        DROP CONSTRAINT membership_fees_payment_id_fkey;
        RAISE NOTICE 'Dropped foreign key constraint membership_fees_payment_id_fkey';
    ELSE
        RAISE NOTICE 'Foreign key constraint membership_fees_payment_id_fkey does not exist';
    END IF;
END $$;

-- Remove the payment_id column (idempotent)
ALTER TABLE membership_fees DROP COLUMN IF EXISTS payment_id;

-- Note: No data migration needed because:
-- 1. This field was never properly used in production
-- 2. The correct relationship (payments.membership_fee_id) contains the actual data
