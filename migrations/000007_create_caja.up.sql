CREATE TABLE caja (
                      caja_id SERIAL PRIMARY KEY,
                      miembro_id INTEGER REFERENCES miembros(miembro_id) ON DELETE RESTRICT,
                      familia_id INTEGER REFERENCES familias(familia_id) ON DELETE RESTRICT,
                      tipo_operacion VARCHAR(50) CHECK (
                          tipo_operacion IN (
                                             'ingreso_cuota',
                                             'gasto_corriente',
                                             'entrega_fondo',
                                             'otros_ingresos'
                              )
                          ) NOT NULL,
                      monto DECIMAL(10,2) NOT NULL,
                      fecha DATE NOT NULL,
                      detalle TEXT,
                      created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
                      updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
                      CONSTRAINT chk_origen_movimiento CHECK (
                          (miembro_id IS NOT NULL AND familia_id IS NULL) OR
                          (miembro_id IS NULL AND familia_id IS NOT NULL) OR
                          (miembro_id IS NULL AND familia_id IS NULL)
                          )
);

-- Trigger para actualizar updated_at
CREATE TRIGGER update_caja_updated_at
    BEFORE UPDATE ON caja
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Índices
CREATE INDEX idx_caja_miembro ON caja(miembro_id);
CREATE INDEX idx_caja_familia ON caja(familia_id);
CREATE INDEX idx_caja_tipo ON caja(tipo_operacion);
CREATE INDEX idx_caja_fecha ON caja(fecha);