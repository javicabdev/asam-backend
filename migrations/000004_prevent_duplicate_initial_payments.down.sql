--!postgresql
-- Migration Rollback: Prevent Duplicate Initial Payments
-- VERSION: 1.1
-- PURPOSE: Remove partial unique indexes that prevent duplicate initial payments
-- IDEMPOTENT: Safe to run multiple times without errors

-- =============================================================================
-- Remove partial unique indexes - IDEMPOTENT
-- =============================================================================

-- Remove index for members
DROP INDEX IF EXISTS unique_initial_payment_per_member;

-- Remove index for families
DROP INDEX IF EXISTS unique_initial_payment_per_family;

-- Note: This rollback does NOT restore any duplicate payments that were
-- deleted during the up migration. This is intentional to maintain data integrity.
