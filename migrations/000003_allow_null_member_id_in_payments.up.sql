-- Migration: Allow NULL member_id in payments table
-- Purpose: Enable family-only payments without requiring a member_id
-- This allows payments to be associated with either a member OR a family

-- Remove NOT NULL constraint from member_id column
ALTER TABLE payments ALTER COLUMN member_id DROP NOT NULL;

-- Add a check constraint to ensure at least one of member_id or family_id is present
ALTER TABLE payments 
    ADD CONSTRAINT payments_member_or_family_check 
    CHECK (member_id IS NOT NULL OR family_id IS NOT NULL);

-- Update comment to reflect the change
COMMENT ON COLUMN payments.member_id IS 'Member associated with the payment. NULL if payment is family-only. At least one of member_id or family_id must be present.';
COMMENT ON COLUMN payments.family_id IS 'Family associated with the payment. NULL if payment is member-only. At least one of member_id or family_id must be present.';
