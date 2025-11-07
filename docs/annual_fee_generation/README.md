# Generación de Cuotas Anuales - Plan de Implementación

## Índice

1. [Visión General](#visión-general)
2. [Arquitectura](#arquitectura)
3. [Backend - Instrucciones Detalladas](./backend.md)
4. [Frontend - Instrucciones Detalladas](./frontend.md)
5. [Testing](./testing.md)
6. [Despliegue](./deployment.md)

---

## Visión General

### Contexto

La aplicación ASAM está casi terminada y lista para producción. La funcionalidad faltante **crítica** es la **generación masiva de cuotas anuales** para todos los socios activos.

### Requisitos Funcionales

#### RF1: Generación de Cuotas por Año
- **Actor**: Usuario administrador
- **Descripción**: El usuario debe poder generar cuotas anuales para un año específico
- **Años permitidos**: Año presente o pasado (nunca futuro)
- **Socios afectados**: Solo socios ACTIVOS en la fecha de generación
- **Comportamiento**:
  - Un socio puede tener **múltiples cuotas pendientes** de diferentes años
  - Un socio puede tener **múltiples pagos confirmados** de cuotas anuales

#### RF2: Configuración de Cuotas
- **Monto base**: Configurable por año (individual)
- **Monto familiar**: Monto base + extra familiar
- **Fecha de vencimiento**: Siempre 31 de diciembre del año
- **Idempotencia**: No se pueden generar cuotas duplicadas para el mismo año

#### RF3: Validaciones
- ✅ No generar para años futuros
- ✅ No generar cuotas duplicadas (año + socio)
- ✅ No generar para socios inactivos
- ✅ Validar que la configuración de cuota anual exista

#### RF4: Datos Históricos
- El usuario debe poder generar cuotas de años pasados
- Útil para migración desde Excel
- Ejemplo: Generar cuotas de 2020, 2021, 2022, 2023, 2024

---

## Arquitectura

### Modelo de Datos Actual

```
┌─────────────────────┐         ┌──────────────────────┐
│  MembershipFee      │         │      Payment         │
├─────────────────────┤         ├──────────────────────┤
│ ID (PK)             │◄────────│ ID (PK)              │
│ Year                │         │ MemberID (FK)        │
│ BaseFeeAmount       │         │ MembershipFeeID (FK) │
│ FamilyFeeExtra      │         │ Amount               │
│ DueDate             │         │ PaymentDate          │
│ CreatedAt           │         │ Status               │
│ UpdatedAt           │         │ PaymentMethod        │
└─────────────────────┘         │ Notes                │
                                └──────────────────────┘
                                          │
                                          │
                                          ▼
                                ┌──────────────────────┐
                                │      Member          │
                                ├──────────────────────┤
                                │ ID (PK)              │
                                │ MembershipNumber     │
                                │ MembershipType       │
                                │ State                │
                                │ Name                 │
                                │ Surnames             │
                                └──────────────────────┘
```

### Flujo de Generación de Cuotas

```
┌─────────────────┐
│   Admin User    │
└────────┬────────┘
         │
         │ 1. Selecciona año y monto
         ▼
┌─────────────────────────────────┐
│  Frontend: GenerateFeesDialog   │
│  - Validar año (≤ presente)     │
│  - Ingresar monto base          │
│  - Ingresar extra familiar      │
└────────┬────────────────────────┘
         │
         │ 2. Mutation: generateAnnualFees
         ▼
┌─────────────────────────────────────┐
│  Backend: PaymentService            │
│  - Validar año                      │
│  - Crear/verificar MembershipFee    │
│  - Obtener socios activos           │
│  - Generar payments PENDING         │
└────────┬────────────────────────────┘
         │
         │ 3. Resultado
         ▼
┌─────────────────────────────────────┐
│  Frontend: Mostrar Resultado        │
│  - N pagos generados                │
│  - Refresh lista de pagos           │
└─────────────────────────────────────┘
```

---

## Decisiones Técnicas

### 1. Estrategia de Generación

**Opción Elegida**: Generación masiva en una transacción

**Alternativas consideradas**:
- ❌ Generación bajo demanda (cuando el socio intenta pagar)
- ❌ Generación mensual automática (CRON)
- ✅ **Generación manual masiva por año**

**Rationale**:
- ✅ Control total del administrador
- ✅ Útil para migración de datos históricos
- ✅ No genera datos innecesarios anticipadamente
- ✅ Transacción atómica (todo o nada)

### 2. Idempotencia

**Comportamiento**:
- Si ya existe un pago para (MemberID + Year), **no se crea duplicado**
- La operación devuelve cuántos pagos se crearon vs cuántos ya existían

### 3. Estado de Pagos Generados

**Estado inicial**: `PENDING`

**Rationale**:
- Los pagos se crean como pendientes
- El usuario luego los confirma cuando recibe el pago real
- Permite tracking de quién debe y quién ha pagado

### 4. Cálculo de Montos

```go
// Pseudocódigo
if member.MembershipType == "familiar" {
    amount = BaseFeeAmount + FamilyFeeExtra
} else {
    amount = BaseFeeAmount
}
```

---

## Estimación

| Fase | Tiempo | Complejidad |
|------|--------|-------------|
| Backend - Servicio | 2-3 horas | Media |
| Backend - GraphQL | 1 hora | Baja |
| Frontend - UI | 2-3 horas | Media |
| Testing Manual | 1 hora | Baja |
| Testing Automatizado | 2 horas | Media |
| **TOTAL** | **8-10 horas** | **Media** |

---

## Casos de Uso

### Caso 1: Primera Generación (Año Actual)
**Escenario**: Enero 2025, primera vez que se genera cuotas

**Pasos**:
1. Admin navega a "Pagos" > "Generar Cuotas"
2. Selecciona año: 2025
3. Ingresa monto base: 40€
4. Ingresa extra familiar: 10€
5. Confirma generación
6. Sistema crea N pagos pendientes (uno por socio activo)
7. Admin ve resumen: "50 pagos generados exitosamente"

### Caso 2: Migración de Datos Históricos
**Escenario**: Necesito importar datos de 2020-2024

**Pasos**:
1. Generar cuotas de 2020 (monto: 35€, extra: 8€)
2. Generar cuotas de 2021 (monto: 35€, extra: 8€)
3. Generar cuotas de 2022 (monto: 38€, extra: 9€)
4. Generar cuotas de 2023 (monto: 40€, extra: 10€)
5. Generar cuotas de 2024 (monto: 40€, extra: 10€)
6. Cada socio activo tendrá 5 cuotas pendientes

**Luego**: Ir marcando como pagados los que sí pagaron (usando CSV import o manualmente)

### Caso 3: Regeneración (Idempotencia)
**Escenario**: Usuario ejecuta generación 2024 dos veces por error

**Resultado**:
- Primera ejecución: "50 pagos generados"
- Segunda ejecución: "0 pagos generados, 50 ya existían"
- No se crean duplicados ✅

### Caso 4: Validación - Año Futuro
**Escenario**: Usuario intenta generar cuotas de 2026

**Resultado**: ❌ Error "No se pueden generar cuotas de años futuros"

---

## Consideraciones Especiales

### Socios que se dan de Alta Durante el Año

**Ejemplo**: Juan se da de alta el 15 de Julio de 2024

**Pregunta**: ¿Debe pagar la cuota completa de 2024?

**Respuestas posibles**:
1. **Opción A**: Sí, cuota completa (más simple)
2. **Opción B**: Cuota prorrateada según meses restantes

**Decisión actual**: **Opción A** (cuota completa)
- Más simple de implementar
- Comportamiento estándar en asociaciones
- Si se requiere prorrateo, se puede ajustar manualmente el monto

### Socios que se dan de Baja Durante el Año

**Ejemplo**: María se da de baja el 20 de Marzo de 2024

**Pregunta**: ¿Debe pagar la cuota de 2024?

**Decisión**:
- Si la cuota ya está generada y pendiente → El usuario decide si la cancela
- La generación de cuotas solo afecta a socios **activos en el momento de la generación**

### Cambio de Tipo (Individual → Familiar)

**Escenario**: Pedro era individual en 2023, pero en 2024 es familiar

**Comportamiento**:
- Cuota 2023: 40€ (individual)
- Cuota 2024: 50€ (familiar)
- El monto se calcula según el tipo **en el momento de la generación**

---

## Siguientes Pasos

1. Leer [Backend - Instrucciones Detalladas](./backend.md)
2. Leer [Frontend - Instrucciones Detalladas](./frontend.md)
3. Implementar Backend
4. Implementar Frontend
5. Testing
6. Despliegue

---

**Fecha de creación**: 2025-11-07
**Versión**: 1.0
**Estado**: Listo para implementación
