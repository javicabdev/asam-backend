--!postgresql
-- Migration Rollback: Prevent Duplicate Initial Payments
-- VERSION: 1.0
-- PURPOSE: Remove database constraints that prevent duplicate initial payments

-- =============================================================================
-- Remove unique constraints
-- =============================================================================

-- Remove constraint for members
ALTER TABLE payments
DROP CONSTRAINT IF EXISTS unique_initial_payment_per_member;

-- Remove constraint for families
ALTER TABLE payments
DROP CONSTRAINT IF EXISTS unique_initial_payment_per_family;

-- Note: This rollback does NOT restore any duplicate payments that were
-- deleted during the up migration. This is intentional to maintain data integrity.
