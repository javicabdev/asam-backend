-- Migration: 000009_add_cashflow_constraints_and_indexes
-- Description: Agrega constraint único en payment_id y verifica índices para optimización de queries en cash_flows

-- Crear constraint único para payment_id (idempotencia)
-- Permite que múltiples registros tengan payment_id NULL, pero payment_id no-NULL debe ser único
-- Esto previene crear múltiples cash_flows para el mismo pago
DO $$
BEGIN
    -- Solo crear el constraint si no existe
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'unique_payment_id_not_null'
          AND conrelid = 'cash_flows'::regclass
    ) THEN
        -- Crear índice único parcial (solo para payment_id NOT NULL)
        CREATE UNIQUE INDEX unique_payment_id_not_null
        ON cash_flows(payment_id)
        WHERE payment_id IS NOT NULL AND deleted_at IS NULL;
    END IF;
END $$;

-- Verificar y crear índices necesarios para optimización

-- Índice en member_id (usado en filtros por miembro)
CREATE INDEX IF NOT EXISTS idx_cashflows_member
ON cash_flows(member_id)
WHERE deleted_at IS NULL;

-- Índice en date (usado en queries de balance y estadísticas por periodo)
CREATE INDEX IF NOT EXISTS idx_cashflows_date
ON cash_flows(date)
WHERE deleted_at IS NULL;

-- Índice en operation_type (usado en filtros por categoría)
CREATE INDEX IF NOT EXISTS idx_cashflows_operation_type
ON cash_flows(operation_type)
WHERE deleted_at IS NULL;

-- Índice en payment_id (usado para buscar cash_flow por pago)
CREATE INDEX IF NOT EXISTS idx_cashflows_payment
ON cash_flows(payment_id)
WHERE deleted_at IS NULL;

-- Índice compuesto para queries de balance por miembro y fecha
-- Optimiza queries como: WHERE member_id = X AND date BETWEEN Y AND Z
CREATE INDEX IF NOT EXISTS idx_cashflows_member_date
ON cash_flows(member_id, date)
WHERE deleted_at IS NULL;

-- Índice compuesto para queries de estadísticas por tipo y fecha
-- Optimiza queries como: WHERE operation_type = X AND date BETWEEN Y AND Z
CREATE INDEX IF NOT EXISTS idx_cashflows_type_date
ON cash_flows(operation_type, date)
WHERE deleted_at IS NULL;
