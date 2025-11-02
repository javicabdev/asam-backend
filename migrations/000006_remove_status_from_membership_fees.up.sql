-- Migration: Remove status column from membership_fees table
-- Reason: The status field was incorrectly used. Individual payment status
-- should be tracked in the payments table, not in the membership_fee configuration table.
--
-- Background:
-- - membership_fees defines annual fee configuration (amounts, year, due date)
-- - payments tracks individual member/family payment status
-- - Having status in membership_fees caused bugs where marking one payment as "paid"
--   would mark the entire annual fee as paid for all members

-- Drop the index first
DROP INDEX IF EXISTS idx_membership_fees_status;

-- Remove the status column
ALTER TABLE membership_fees DROP COLUMN IF EXISTS status;

-- Note: No data migration needed because:
-- 1. Payment status is already correctly stored in payments.status
-- 2. The membership_fees.status field was not being used correctly
