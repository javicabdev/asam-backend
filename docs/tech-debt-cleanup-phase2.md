# Limpieza de Deuda Técnica - Fase 2

## Resumen de Cambios

Esta es la segunda fase de limpieza de deuda técnica, enfocada en elementos de prioridad media.

### Cambios a Realizar

#### 1. Mover mocks a la carpeta de tests

**Motivo**: Los mocks no deben estar en el código de producción
- `internal/adapters/email/mock_notification_adapter.go` está en el código de producción
- Los mocks deben estar en la carpeta de tests para evitar que se incluyan en el binario final

**Acción**: 
- Crear estructura `test/mocks/email/`
- Mover el mock ajustando el package a `mocks`
- Actualizar todos los imports en los tests

#### 2. Reorganizar scripts de prueba

**Motivo**: Scripts de prueba mal ubicados
- `scripts/test-mock-email.go` es un script de prueba manual

**Acción**: 
- Crear estructura `test/manual/`
- Mover el script de prueba

#### 3. Consolidar archivos .env

**Motivo**: Exceso de archivos de configuración causa confusión
- 17 archivos .env* diferentes
- Muchos son redundantes o ejemplos innecesarios

**Acción**: Consolidar a un conjunto mínimo:
- `.env.example` - Plantilla con todas las variables
- `.env.development` - Configuración de desarrollo
- `.env.production` - Configuración de producción
- `.env.test` - Configuración para tests
- `.env.aiven` - Configuración específica de Aiven (temporal)

**Archivos a eliminar**:
- `.env.production.free`
- `.env.production.test`
- `.env.docker.example`
- `.env.email.example`
- `.env.complete.example`
- `.env.local`
- `.env` (si está vacío o es redundante)

### Impacto

- **Archivos movidos**: 2
- **Archivos eliminados**: ~7 archivos .env redundantes
- **Mejora en organización**: Separación clara entre código de producción y tests
- **Reducción de confusión**: Un conjunto claro de archivos de configuración

### Prerequisitos

**IMPORTANTE**: Antes de ejecutar este script, asegúrate de:
1. Haber ejecutado la Fase 1 de limpieza
2. Hacer commit de todos los cambios pendientes
3. Estar en una rama limpia

### Cómo Ejecutar

Para aplicar estos cambios, ejecuta uno de los siguientes scripts desde la raíz del proyecto:

**En Windows (PowerShell):**
```powershell
.\scripts\cleanup-tech-debt-phase2.ps1
```

**En Linux/Mac:**
```bash
chmod +x scripts/cleanup-tech-debt-phase2.sh
./scripts/cleanup-tech-debt-phase2.sh
```

### Validación Post-Ejecución

Después de ejecutar el script:
1. Verifica que los tests pasen: `go test ./...`
2. Revisa que no haya imports rotos
3. Confirma que los archivos .env esenciales sigan presentes

### Próximos Pasos

**Fase 3 (Prioridad Baja):**
- Organizar scripts de utilidades en carpetas apropiadas
- Limpiar scripts obsoletos
- Mejorar documentación de scripts
