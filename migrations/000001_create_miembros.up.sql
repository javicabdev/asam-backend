CREATE TABLE miembros (
                          miembro_id SERIAL PRIMARY KEY,
                          numero_socio VARCHAR(50) UNIQUE NOT NULL,
                          tipo_membresia VARCHAR(20) CHECK (tipo_membresia IN ('individual', 'familiar')) NOT NULL,
                          nombre VARCHAR(100) NOT NULL,
                          apellidos VARCHAR(100) NOT NULL,
                          calle_numero_piso VARCHAR(200) NOT NULL,
                          codigo_postal VARCHAR(10) NOT NULL,
                          poblacion VARCHAR(100) NOT NULL,
                          provincia VARCHAR(100) DEFAULT 'Barcelona' NOT NULL,
                          pais VARCHAR(100) DEFAULT 'España' NOT NULL,
                          estado VARCHAR(20) CHECK (estado IN ('activo', 'inactivo')) NOT NULL,
                          fecha_alta DATE NOT NULL,
                          fecha_baja DATE,
                          fecha_nacimiento DATE,
                          documento_identidad VARCHAR(50),
                          correo_electronico VARCHAR(100),
                          profesion VARCHAR(100),
                          nacionalidad VARCHAR(100) DEFAULT 'Senegal',
                          observaciones TEXT,
                          created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
                          updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Trigger para actualizar updated_at automáticamente
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_miembros_updated_at
    BEFORE UPDATE ON miembros
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Índices
CREATE INDEX idx_miembros_numero_socio ON miembros(numero_socio);
CREATE INDEX idx_miembros_estado ON miembros(estado);
CREATE INDEX idx_miembros_documento_identidad ON miembros(documento_identidad);