-- Migration rollback: 000009_add_cashflow_constraints_and_indexes
-- Description: Elimina constraint único e índices de cash_flows

-- Eliminar índices compuestos
DROP INDEX IF EXISTS idx_cashflows_type_date;
DROP INDEX IF EXISTS idx_cashflows_member_date;

-- Eliminar índices simples
DROP INDEX IF EXISTS idx_cashflows_payment;
DROP INDEX IF EXISTS idx_cashflows_operation_type;
DROP INDEX IF EXISTS idx_cashflows_date;
DROP INDEX IF EXISTS idx_cashflows_member;

-- Eliminar constraint único de payment_id
DROP INDEX IF EXISTS unique_payment_id_not_null;
