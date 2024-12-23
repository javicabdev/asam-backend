CREATE TABLE telefonos (
                           telefono_id SERIAL PRIMARY KEY,
                           miembro_id INTEGER REFERENCES miembros(miembro_id) ON DELETE CASCADE,
                           familia_id INTEGER REFERENCES familias(familia_id) ON DELETE CASCADE,
                           numero_telefono VARCHAR(20) NOT NULL,
                           created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
                           updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    -- Asegurar que el teléfono está asociado a un miembro O a una familia, pero no a ambos
                           CONSTRAINT chk_telefono_asociacion CHECK (
                               (miembro_id IS NOT NULL AND familia_id IS NULL) OR
                               (miembro_id IS NULL AND familia_id IS NOT NULL)
                               ),
                           CONSTRAINT uk_telefono_numero UNIQUE (numero_telefono)
);

-- Trigger para actualizar updated_at
CREATE TRIGGER update_telefonos_updated_at
    BEFORE UPDATE ON telefonos
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Índices
CREATE INDEX idx_telefonos_miembro ON telefonos(miembro_id);
CREATE INDEX idx_telefonos_familia ON telefonos(familia_id);
CREATE INDEX idx_telefonos_numero ON telefonos(numero_telefono);