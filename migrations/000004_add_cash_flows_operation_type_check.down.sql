-- Rollback: drop the operation_type CHECK constraint and restore the previous column comment

ALTER TABLE cash_flows DROP CONSTRAINT IF EXISTS cash_flows_operation_type_check;

COMMENT ON COLUMN cash_flows.operation_type IS 'Type of operation: income, expense, transfer, etc.';
