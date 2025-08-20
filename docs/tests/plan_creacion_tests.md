Voy a examinar la estructura de seeding que has recuperado para ajustar el plan de tests.## Plan de Implementación Actualizado: Suite de Tests Básica## Plan de Implementación Actualizado: Suite de Tests con Infraestructura de Seeding Existente

### Commit Propuesto: `test(setup): configurar estructura base de tests unitarios`

**Descripción**: Establecer la infraestructura básica para tests unitarios, aprovechando los generadores existentes en `test/seed`.

**Análisis de Arquitectura**: Reutilizamos la infraestructura de seeding existente, creando una capa de testing que la aprovecha sin duplicar código.

**Ficheros Afectados**:
- `test/unit/testutils/helpers.go` (nuevo)
- `test/unit/testutils/assertions.go` (nuevo)
- `test/unit/README.md` (nuevo)

**Pasos Técnicos**:
1. Crear helpers para tests unitarios que no requieren BD
2. Implementar assertions personalizadas para modelos
3. Configurar utilidades para comparación de structs
4. Documentar convenciones de testing unitario

---

### Commit Propuesto: `test(unit): añadir tests unitarios para modelo Member`

**Descripción**: Implementar tests unitarios para la entidad Member usando los generadores existentes para crear datos de prueba.

**Análisis de Arquitectura**: Tests puros de dominio, usando el `MemberGenerator` para crear instancias válidas de prueba.

**Ficheros Afectados**:
- `test/unit/domain/models/member_test.go` (nuevo)

**Pasos Técnicos**:
1. Usar `generators.NewMemberGenerator` para crear miembros de prueba
2. Test de validación de campos obligatorios
3. Test de formato de número de socio (A/B + 5 dígitos)
4. Test de validación de DNI/NIE
5. Test de transiciones de estado permitidas

---

### Commit Propuesto: `test(unit): añadir tests unitarios para modelos Family y Payment`

**Descripción**: Tests unitarios para Family y Payment, aprovechando `FamilyGenerator` y `PaymentGenerator`.

**Análisis de Arquitectura**: Continuación de tests de dominio, reutilizando generadores existentes.

**Ficheros Afectados**:
- `test/unit/domain/models/family_test.go` (nuevo)
- `test/unit/domain/models/payment_test.go` (nuevo)
- `test/unit/domain/models/cashflow_test.go` (nuevo)

**Pasos Técnicos**:
1. Test con datos generados por `FamilyGenerator`
2. Test de validaciones de Family
3. Test de Payment usando `PaymentGenerator`
4. Test de conversión Payment a CashFlow

---

### Commit Propuesto: `test(services): añadir tests para servicios con mocks`

**Descripción**: Tests unitarios de servicios usando mocks generados automáticamente.

**Análisis de Arquitectura**: Introducimos mockgen para generar mocks de las interfaces en `ports`.

**Ficheros Afectados**:
- `test/unit/domain/services/member_service_test.go` (nuevo)
- `test/unit/domain/services/payment_service_test.go` (nuevo)
- `test/mocks/generate.go` (nuevo)
- `Makefile` (modificado - añadir target para generar mocks)

**Pasos Técnicos**:
1. Configurar mockgen en tools.go
2. Crear script para generar mocks de interfaces
3. Test de MemberService con repositorio mockeado
4. Test de PaymentService con dependencias mockeadas
5. Usar datos del seeder para casos de prueba

---

### Commit Propuesto: `test(integration): configurar tests de integración con BD`

**Descripción**: Establecer infraestructura para tests de integración usando el `Seeder` existente.

**Análisis de Arquitectura**: Reutilizamos completamente el sistema de seeding para preparar datos de test.

**Ficheros Afectados**:
- `test/integration/setup_test.go` (nuevo)
- `test/integration/testutils/db.go` (nuevo)
- `docker-compose.test.yml` (verificar/actualizar)

**Pasos Técnicos**:
1. Helper para crear BD de test con contenedor
2. Integrar `seed.NewSeeder()` en el setup
3. Helper para ejecutar migraciones
4. Función para limpiar BD entre tests
5. Configurar timeout y reintentos

---

### Commit Propuesto: `test(integration): tests de repositorios con datasets`

**Descripción**: Tests de integración para repositorios usando los datasets predefinidos (minimal, full).

**Análisis de Arquitectura**: Aprovechamos los datasets existentes para probar diferentes escenarios.

**Ficheros Afectados**:
- `test/integration/repositories/member_repository_test.go` (nuevo)
- `test/integration/repositories/payment_repository_test.go` (nuevo)
- `test/integration/repositories/family_repository_test.go` (nuevo)

**Pasos Técnicos**:
1. Test con `seeder.SeedMinimalDataset()` para casos básicos
2. Test con `seeder.SeedFullDataset()` para paginación
3. Usar escenarios específicos para edge cases
4. Verificar integridad referencial
5. Test de queries complejas con JOINs

---

### Commit Propuesto: `test(integration): tests de servicios con BD real`

**Descripción**: Tests de integración de servicios usando BD real y datos del seeder.

**Análisis de Arquitectura**: Verificación de la capa de servicios con todas sus dependencias reales.

**Ficheros Afectados**:
- `test/integration/services/member_service_integration_test.go` (nuevo)
- `test/integration/services/payment_service_integration_test.go` (nuevo)

**Pasos Técnicos**:
1. Setup con `data.NewMinimalDataset()`
2. Test de flujos completos de negocio
3. Verificar transacciones y rollbacks
4. Test de concurrencia con el seeder
5. Usar escenarios predefinidos para casos específicos

---

### Commit Propuesto: `test(graphql): tests de resolvers GraphQL`

**Descripción**: Tests de la API GraphQL usando el cliente generado por gqlgen.

**Análisis de Arquitectura**: Verificación end-to-end de la API GraphQL.

**Ficheros Afectados**:
- `test/integration/graphql/member_resolver_test.go` (nuevo)
- `test/integration/graphql/auth_test.go` (nuevo)
- `test/integration/graphql/client_test.go` (nuevo)

**Pasos Técnicos**:
1. Cliente GraphQL de test con autenticación
2. Seed con dataset minimal para cada test
3. Test de queries con diferentes permisos
4. Test de mutations con validaciones
5. Test de paginación y filtros

---

### Commit Propuesto: `test(scenarios): tests basados en escenarios de negocio`

**Descripción**: Tests que verifican los escenarios de negocio predefinidos en el seeder.

**Análisis de Arquitectura**: Aprovechamos los escenarios existentes (payment_overdue, membership_expired, etc.).

**Ficheros Afectados**:
- `test/integration/scenarios/payment_overdue_test.go` (nuevo)
- `test/integration/scenarios/membership_lifecycle_test.go` (nuevo)
- `test/integration/scenarios/family_management_test.go` (nuevo)

**Pasos Técnicos**:
1. Cargar escenario específico con el seeder
2. Ejecutar operaciones sobre los datos del escenario
3. Verificar comportamiento esperado
4. Test de notificaciones para morosos
5. Test de cálculo de cuotas familiares

---

### Commit Propuesto: `test(benchmark): añadir benchmarks de rendimiento`

**Descripción**: Benchmarks para operaciones críticas usando el dataset completo.

**Análisis de Arquitectura**: Medición de rendimiento con volúmenes realistas de datos.

**Ficheros Afectados**:
- `test/benchmark/member_operations_bench_test.go` (nuevo)
- `test/benchmark/payment_processing_bench_test.go` (nuevo)
- `test/benchmark/report_generation_bench_test.go` (nuevo)

**Pasos Técnicos**:
1. Usar `seeder.SeedFullDataset()` para volumen real
2. Benchmark de búsquedas con paginación
3. Benchmark de cálculo de morosos
4. Benchmark de generación de reportes
5. Comparar con métricas objetivo

---

### Commit Propuesto: `test(ci): integrar suite de tests en CI/CD`

**Descripción**: Configurar ejecución automática aprovechando los comandos del seeder.

**Análisis de Arquitectura**: Automatización completa del testing.

**Ficheros Afectados**:
- `Makefile` (modificado)
- `.github/workflows/test.yml` (modificado)
- `scripts/run_tests.sh` (nuevo)

**Pasos Técnicos**:
1. Target `make test-unit` (sin BD)
2. Target `make test-integration` (con seed minimal)
3. Target `make test-scenarios` (con escenarios)
4. Target `make test-coverage` con reporte
5. GitHub Actions con matriz de tests paralelos

---

**Ventajas de este enfoque actualizado**:
- Reutilización total de la infraestructura de seeding existente
- No duplicamos generadores de datos
- Aprovechamos datasets y escenarios ya definidos
- Tests más realistas con datos consistentes
- Fácil mantenimiento al centralizar la generación de datos

**Comandos de ejemplo**:
```bash
# Tests unitarios (sin BD)
make test-unit

# Tests de integración con dataset minimal
make test-integration

# Tests de escenarios específicos
make test-scenarios SCENARIO=payment_overdue

# Suite completa con cobertura
make test-all
```