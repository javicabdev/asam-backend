CREATE TABLE historial_membresia (
                                     historial_id SERIAL PRIMARY KEY,
                                     miembro_id INTEGER NOT NULL REFERENCES miembros(miembro_id) ON DELETE CASCADE,
                                     tipo_membresia VARCHAR(20) CHECK (tipo_membresia IN ('individual', 'familiar')) NOT NULL,
                                     fecha_inicio DATE NOT NULL,
                                     fecha_fin DATE,
                                     created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
                                     updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
                                     CONSTRAINT chk_fechas_validas CHECK (
                                         fecha_fin IS NULL OR fecha_fin > fecha_inicio
                                         )
);

-- Trigger para actualizar updated_at
CREATE TRIGGER update_historial_membresia_updated_at
    BEFORE UPDATE ON historial_membresia
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Índices
CREATE INDEX idx_historial_miembro ON historial_membresia(miembro_id);
CREATE INDEX idx_historial_tipo ON historial_membresia(tipo_membresia);
CREATE INDEX idx_historial_fechas ON historial_membresia(fecha_inicio, fecha_fin);