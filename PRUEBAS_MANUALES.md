# 🧪 Guía de Pruebas Manuales - Generación de Cuotas Anuales

## ✅ Implementación Completada

Se ha implementado la funcionalidad completa de generación de cuotas anuales ASAM con los siguientes componentes:

### Backend
- ✅ Repositorio: `GetAllActive()` en MemberRepository
- ✅ DTOs: `GenerateAnnualFeesRequest/Response` en `internal/ports/input/payment_service.go`
- ✅ Servicio: `GenerateAnnualFees()` en PaymentService con validaciones e idempotencia
- ✅ GraphQL Schema: Tipos, Inputs y Mutation `generateAnnualFees`
- ✅ Resolver: Implementado con validación de permisos ADMIN

## 📋 Opciones para Probar

### Opción 1: GraphQL Playground (Más Fácil)

1. **Abrir GraphQL Playground**
   ```
   http://localhost:8080/graphql
   ```

2. **Hacer Login y obtener token**
   ```graphql
   mutation Login {
     login(input: {
       username: "admin@ejemplo.com"
       password: "tu_contraseña"
     }) {
       accessToken
       user {
         email
         role
       }
     }
   }
   ```

3. **Configurar Headers** (en la pestaña HTTP HEADERS del Playground):
   ```json
   {
     "Authorization": "Bearer TU_TOKEN_AQUI"
   }
   ```

4. **Ejecutar Mutation de Generación**
   ```graphql
   mutation GenerateAnnualFees2025 {
     generateAnnualFees(input: {
       year: 2025
       base_fee_amount: 100.00
       family_fee_extra: 50.00
     }) {
       year
       membership_fee_id
       payments_generated
       payments_existing
       total_members
       total_expected_amount
       details {
         member_number
         member_name
         amount
         was_created
         error
       }
     }
   }
   ```

5. **Verificar resultados esperados**:
   - `year`: 2025
   - `payments_generated`: 2 (hay 2 miembros activos)
   - `total_members`: 2
   - `total_expected_amount`: 200.00 (2 individuales × 100.00)
   - `details`: Array con 2 elementos, ambos con `was_created: true`

6. **Probar Idempotencia** (ejecutar la mutation de nuevo):
   - `payments_generated`: 0
   - `payments_existing`: 2
   - `details`: Array con 2 elementos, ambos con `was_created: false`

### Opción 2: Verificación Directa en Base de Datos

```bash
# 1. Ver miembros activos
docker-compose exec -T postgres psql -U postgres -d asam_db -c \
  "SELECT id, membership_number, name, surnames, membership_type, state
   FROM members WHERE state = 'active';"

# 2. Verificar cuotas de membresía
docker-compose exec -T postgres psql -U postgres -d asam_db -c \
  "SELECT id, year, base_fee_amount, family_fee_extra, due_date
   FROM membership_fees
   ORDER BY year DESC
   LIMIT 5;"

# 3. Ver pagos generados para 2025
docker-compose exec -T postgres psql -U postgres -d asam_db -c \
  "SELECT
     p.id,
     m.membership_number,
     m.name,
     m.surnames,
     p.amount,
     p.status,
     p.notes,
     mf.year
   FROM payments p
   JOIN members m ON p.member_id = m.id
   JOIN membership_fees mf ON p.membership_fee_id = mf.id
   WHERE mf.year = 2025
   ORDER BY m.membership_number;"

# 4. Contar pagos por estado
docker-compose exec -T postgres psql -U postgres -d asam_db -c \
  "SELECT status, COUNT(*) as total
   FROM payments
   WHERE membership_fee_id = (SELECT id FROM membership_fees WHERE year = 2025)
   GROUP BY status;"
```

### Opción 3: Usando curl (Línea de Comandos)

```bash
# 1. Login y obtener token
TOKEN=$(curl -s -X POST http://localhost:8080/graphql \
  -H "Content-Type: application/json" \
  -d '{"operationName":"Login","query":"mutation Login{login(input:{username:\"TU_EMAIL\",password:\"TU_PASSWORD\"}){accessToken}}"}' \
  | grep -o '"accessToken":"[^"]*' | sed 's/"accessToken":"//')

echo "Token: $TOKEN"

# 2. Generar cuotas
curl -X POST http://localhost:8080/graphql \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"operationName":"GenerateFees","query":"mutation GenerateFees{generateAnnualFees(input:{year:2025,base_fee_amount:100.00,family_fee_extra:50.00}){year membership_fee_id payments_generated payments_existing total_members total_expected_amount}}"}' \
  | jq '.'
```

## 🔍 Casos de Prueba

### Caso 1: Primera Generación (Happy Path)
- **Input**: año=2025, base=100, extra=50
- **Esperado**:
  - Se crea MembershipFee para 2025
  - Se generan 2 pagos PENDING
  - payments_generated = 2
  - payments_existing = 0

### Caso 2: Idempotencia
- **Input**: Mismos datos del Caso 1
- **Esperado**:
  - MembershipFee se actualiza (mismos valores)
  - NO se crean nuevos pagos
  - payments_generated = 0
  - payments_existing = 2

### Caso 3: Año Futuro (Validación)
- **Input**: año=2026
- **Esperado**: Error "No se pueden generar cuotas para años futuros"

### Caso 4: Monto Negativo (Validación)
- **Input**: base_fee_amount = -100
- **Esperado**: Error "El monto base debe ser positivo"

### Caso 5: Sin Permisos (Seguridad)
- **Sin token ADMIN**
- **Esperado**: Error 401 Unauthorized

## 📊 Estructura de Datos Generada

### Tabla `membership_fees`
```sql
id | year | base_fee_amount | family_fee_extra | due_date
---+------+-----------------+------------------+-----------
1  | 2025 | 100.00          | 50.00            | 2025-12-31
```

### Tabla `payments`
```sql
id | member_id | amount  | status  | membership_fee_id | notes
---+-----------+---------+---------+-------------------+-------
1  | 1         | 100.00  | pending | 1                 | Cuota anual...
2  | 2         | 100.00  | pending | 1                 | Cuota anual...
```

## 🐛 Troubleshooting

### Problema: "Could not determine GraphQL operation"
- **Solución**: Añadir `operationName` en la request
- **Ejemplo**: `{"operationName":"GenerateFees","query":"mutation GenerateFees{...}"}`

### Problema: "credenciales inválidas"
- **Solución**: Verificar credenciales de admin en BD
- **Comando**:
  ```bash
  docker-compose exec -T postgres psql -U postgres -d asam_db -c \
    "SELECT email, role FROM users WHERE role = 'admin';"
  ```

### Problema: El servidor no responde
- **Solución**: Reconstruir y reiniciar contenedor
  ```bash
  docker-compose build api
  docker-compose up -d api
  docker-compose logs api --tail=20
  ```

### Problema: "membership_fee_id column does not exist"
- **Solución**: Verificar que las migraciones se ejecutaron
  ```bash
  docker-compose logs api | grep migration
  ```

## ✅ Checklist de Verificación

- [ ] El servidor está corriendo (http://localhost:8080/health)
- [ ] GraphQL Playground accesible (http://localhost:8080/graphql)
- [ ] Hay al menos 1 usuario ADMIN en la BD
- [ ] Hay miembros activos en la BD
- [ ] Puedes hacer login y obtener un token
- [ ] La mutation `generateAnnualFees` aparece en el schema
- [ ] La mutation ejecuta exitosamente
- [ ] Se crean registros en `membership_fees`
- [ ] Se crean registros en `payments` con status PENDING
- [ ] La segunda ejecución no crea duplicados

## 📝 Próximos Pasos

Una vez verificado que funciona:

1. **Tests Unitarios**: Escribir tests para el servicio
2. **Tests de Integración**: Tests end-to-end con BD de prueba
3. **Frontend**: Implementar UI siguiendo `docs/annual_fee_generation/frontend.md`
4. **Documentación**: Actualizar README con la nueva funcionalidad
5. **Deploy**: Preparar para producción siguiendo `docs/annual_fee_generation/deployment.md`

## 📞 Ayuda

Si encuentras problemas, verifica:
1. Logs del servidor: `docker-compose logs api --tail=50`
2. Estado de contenedores: `docker-compose ps`
3. Conexión a BD: `docker-compose exec postgres psql -U postgres -d asam_db -c '\dt'`
