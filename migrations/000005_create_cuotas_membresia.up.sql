CREATE TABLE cuotas_membresia (
                                  cuota_id SERIAL PRIMARY KEY,
                                  miembro_id INTEGER NOT NULL REFERENCES miembros(miembro_id) ON DELETE RESTRICT,
                                  ano INTEGER NOT NULL,
                                  cantidad_pagada DECIMAL(10,2) NOT NULL,
                                  fecha_pago DATE NOT NULL,
                                  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
                                  updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
                                  CONSTRAINT chk_ano_valido CHECK (ano >= 2000 AND ano <= 2100),
                                  CONSTRAINT chk_cantidad_positiva CHECK (cantidad_pagada > 0),
                                  CONSTRAINT uk_cuota_miembro_ano UNIQUE (miembro_id, ano)
);

-- Trigger para actualizar updated_at
CREATE TRIGGER update_cuotas_membresia_updated_at
    BEFORE UPDATE ON cuotas_membresia
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Índices
CREATE INDEX idx_cuotas_miembro ON cuotas_membresia(miembro_id);
CREATE INDEX idx_cuotas_ano ON cuotas_membresia(ano);
CREATE INDEX idx_cuotas_fecha_pago ON cuotas_membresia(fecha_pago);