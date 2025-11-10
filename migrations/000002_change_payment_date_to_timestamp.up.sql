-- Change payment_date from DATE to TIMESTAMP WITH TIME ZONE
-- This allows storing the complete date and time (including hours, minutes, and seconds)
-- instead of just the date
-- IDEMPOTENT: Safe to run multiple times

DO $$
BEGIN
    -- Check if the column is not already TIMESTAMP WITH TIME ZONE
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_name = 'payments'
          AND column_name = 'payment_date'
          AND data_type != 'timestamp with time zone'
    ) THEN
        -- Change the column type
        ALTER TABLE payments
        ALTER COLUMN payment_date TYPE TIMESTAMP WITH TIME ZONE
        USING payment_date::TIMESTAMP WITH TIME ZONE;

        RAISE NOTICE 'Changed payment_date column type to TIMESTAMP WITH TIME ZONE';
    ELSE
        RAISE NOTICE 'payment_date column is already TIMESTAMP WITH TIME ZONE, skipping';
    END IF;
END $$;

COMMENT ON COLUMN payments.payment_date IS 'Date and time when the payment was made. NULL for pending payments, set when payment status changes to paid. Stores complete timestamp including hours, minutes, and seconds';
