-- Migration Rollback: Restore NOT NULL constraint on payment_date
-- WARNING: This will fail if there are any existing records with NULL payment_date
-- You must set payment_date for all NULL records before running this rollback

-- First, update any NULL payment_date values to a default value
-- (using the created_at date as a fallback)
UPDATE payments
SET payment_date = created_at::DATE
WHERE payment_date IS NULL;

-- Restore NOT NULL constraint on payment_date column
ALTER TABLE payments ALTER COLUMN payment_date SET NOT NULL;

-- Restore original comment
COMMENT ON COLUMN payments.payment_date IS 'Date when the payment was made.';
