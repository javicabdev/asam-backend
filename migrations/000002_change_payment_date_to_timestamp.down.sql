-- Revert payment_date from TIMESTAMP WITH TIME ZONE back to DATE
-- This will lose time information (hours, minutes, seconds)

ALTER TABLE payments
ALTER COLUMN payment_date TYPE DATE
USING payment_date::DATE;

COMMENT ON COLUMN payments.payment_date IS 'Date when the payment was made. NULL for pending payments, set when payment status changes to paid';
