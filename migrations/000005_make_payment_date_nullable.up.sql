-- Migration: Make payment_date nullable in payments table
-- Purpose: Allow pending payments to have NULL payment_date
-- This makes the model more semantically correct: pending payments don't have a payment date yet
-- IDEMPOTENT: Safe to run multiple times

-- Remove NOT NULL constraint from payment_date column (idempotent)
DO $$
BEGIN
    -- Check if the column has NOT NULL constraint
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_name = 'payments'
        AND column_name = 'payment_date'
        AND is_nullable = 'NO'
    ) THEN
        ALTER TABLE payments ALTER COLUMN payment_date DROP NOT NULL;
    END IF;
END $$;

-- Update comment to reflect the change
COMMENT ON COLUMN payments.payment_date IS 'Date when the payment was made. NULL for pending payments, set when payment status changes to paid.';
