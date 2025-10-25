-- Rollback: Restore NOT NULL constraint on member_id
-- WARNING: This migration will fail if there are payments with NULL member_id

-- First, remove the check constraint
ALTER TABLE payments DROP CONSTRAINT IF EXISTS payments_member_or_family_check;

-- Delete any payments that have NULL member_id (family-only payments)
-- This is necessary to restore the NOT NULL constraint
DELETE FROM payments WHERE member_id IS NULL;

-- Restore the NOT NULL constraint
ALTER TABLE payments ALTER COLUMN member_id SET NOT NULL;

-- Restore original comments
COMMENT ON COLUMN payments.member_id IS 'Member associated with the payment';
COMMENT ON COLUMN payments.family_id IS 'Optional family associated with the payment';
