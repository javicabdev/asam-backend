# Despliegue - Generación de Cuotas Anuales

## Índice

1. [Pre-Despliegue](#pre-despliegue)
2. [Despliegue Backend](#despliegue-backend)
3. [Despliegue Frontend](#despliegue-frontend)
4. [Post-Despliegue](#post-despliegue)
5. [Rollback](#rollback)

---

## Pre-Despliegue

### Checklist de Preparación

#### Backend
- [ ] Todos los tests unitarios pasan: `go test ./...`
- [ ] Compilación exitosa: `go build ./...`
- [ ] Código GraphQL regenerado: `go run github.com/99designs/gqlgen generate`
- [ ] Sin warnings de lint: `make lint`
- [ ] Migración de BD preparada (si aplica)
- [ ] Variables de entorno documentadas

#### Frontend
- [ ] Todos los tests pasan: `npm test`
- [ ] Build exitoso: `npm run build`
- [ ] Sin errores de TypeScript: `npm run type-check`
- [ ] Sin warnings de lint: `npm run lint`
- [ ] Bundle size verificado: `npm run analyze`

#### Documentación
- [ ] README actualizado
- [ ] CHANGELOG actualizado con nuevas funcionalidades
- [ ] Documentación de API actualizada
- [ ] Manual de usuario actualizado (si aplica)

---

## Despliegue Backend

### Paso 1: Preparar Rama de Release

```bash
cd /Users/javierfernandezcabanas/repos/asam-backend

# Asegurar que estamos en main y actualizado
git checkout main
git pull origin main

# Crear rama de release (opcional, depende de tu workflow)
git checkout -b release/annual-fee-generation

# Verificar cambios
git log --oneline -10
```

### Paso 2: Ejecutar Tests Finales

```bash
# Tests unitarios
go test ./internal/domain/services/... -v

# Tests de integración (si existen)
go test ./test/integration/... -v

# Build
go build -o bin/asam-backend ./cmd/api
```

### Paso 3: Commit y Push

```bash
# Si hay cambios pendientes
git add .
git commit -m "feat: add annual fee generation functionality

- Añadir GenerateAnnualFees en PaymentService
- Añadir mutation generateAnnualFees en GraphQL
- Implementar idempotencia para evitar duplicados
- Validar año no futuro y montos positivos
- Solo generar para socios activos
- Tests unitarios y de integración

Closes #XXX"

git push origin release/annual-fee-generation

# Crear Pull Request en GitHub/GitLab
# Esperar aprobación y merge a main
```

### Paso 4: Desplegar a Staging

```bash
# Método depende de tu infraestructura

# Opción A: Docker Compose
cd /Users/javierfernandezcabanas/repos/asam-backend
docker-compose build api
docker-compose up -d api

# Opción B: Manual
ssh staging-server
cd /app/asam-backend
git pull origin main
go build -o bin/asam-backend ./cmd/api
systemctl restart asam-backend

# Opción C: CI/CD automático
# Push a branch 'staging' y esperar deployment automático
git push origin main:staging
```

### Paso 5: Verificar en Staging

```bash
# Healthcheck
curl https://staging.asam.com/health

# GraphQL introspection
curl https://staging.asam.com/graphql \
  -H "Content-Type: application/json" \
  -d '{"query": "{ __schema { mutationType { fields { name } } } }"}'

# Verificar que existe generateAnnualFees
```

### Paso 6: Probar en Staging

**Test Manual**:
1. Login como admin en staging
2. Ir a Pagos > Generar Cuotas
3. Generar cuotas de un año de prueba
4. Verificar que se crean correctamente
5. Verificar idempotencia (ejecutar dos veces)

**Test con Script**:

```bash
# Crear script de test
cat > test_staging.sh <<'EOF'
#!/bin/bash

STAGING_URL="https://staging.asam.com/graphql"
ADMIN_TOKEN="<obtener_token_admin>"

# Test 1: Generar cuotas
curl -X POST $STAGING_URL \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "query": "mutation { generateAnnualFees(input: { year: 2024, baseFeeAmount: 40, familyFeeExtra: 10 }) { year paymentsGenerated paymentsExisting totalMembers } }"
  }'

echo "\n\nTest 1 completado"

# Test 2: Idempotencia (ejecutar dos veces)
curl -X POST $STAGING_URL \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "query": "mutation { generateAnnualFees(input: { year: 2024, baseFeeAmount: 40, familyFeeExtra: 10 }) { year paymentsGenerated paymentsExisting } }"
  }'

echo "\n\nTest 2 completado (debe mostrar 0 generados, N existentes)"
EOF

chmod +x test_staging.sh
./test_staging.sh
```

### Paso 7: Desplegar a Producción

**Solo si staging está OK**:

```bash
# Opción A: Tag y deploy automático
git tag v1.5.0
git push origin v1.5.0

# Opción B: Deploy manual
ssh production-server
cd /app/asam-backend
git pull origin main
go build -o bin/asam-backend ./cmd/api

# Backup del binario anterior
cp bin/asam-backend bin/asam-backend.backup

# Restart con nuevo binario
systemctl restart asam-backend

# Verificar logs
journalctl -u asam-backend -f
```

### Paso 8: Verificar en Producción

```bash
# Healthcheck
curl https://asam.com/health

# Verificar logs (sin errores)
tail -f /var/log/asam-backend/app.log
```

---

## Despliegue Frontend

### Paso 1: Preparar Build

```bash
cd /Users/javierfernandezcabanas/repos/asam-frontend

# Pull latest
git checkout main
git pull origin main

# Install dependencies (si hay cambios)
npm install

# Build
npm run build

# Verificar que el build es correcto
ls -lh dist/
```

### Paso 2: Tests Pre-Deploy

```bash
# Tests
npm test

# Lint
npm run lint

# Type check
npm run type-check

# Build size analysis (opcional)
npm run analyze
```

### Paso 3: Deploy a Staging

```bash
# Método depende de tu hosting

# Opción A: Vercel/Netlify
vercel --prod

# Opción B: S3 + CloudFront
aws s3 sync dist/ s3://staging-asam-frontend/
aws cloudfront create-invalidation --distribution-id XXXXX --paths "/*"

# Opción C: Servidor tradicional
scp -r dist/* user@staging-server:/var/www/asam-frontend/
```

### Paso 4: Verificar en Staging

1. Abrir https://staging.asam.com
2. Login como admin
3. Navegar a Pagos
4. **Verificar que aparece botón "Generar Cuotas Anuales"**
5. Click y verificar que se abre el diálogo
6. Realizar generación de prueba
7. Verificar resultado

### Paso 5: Deploy a Producción

**Solo si staging está OK**:

```bash
# Crear tag
git tag frontend-v1.5.0
git push origin frontend-v1.5.0

# Deploy
# (Mismo proceso que staging pero a URLs de producción)

# Opción A: Automatic
git push origin main  # Si tienes CD configurado

# Opción B: Manual
npm run build
aws s3 sync dist/ s3://prod-asam-frontend/
aws cloudfront create-invalidation --distribution-id YYYYY --paths "/*"
```

### Paso 6: Smoke Test en Producción

1. Abrir https://asam.com
2. Login como admin
3. Ir a Pagos
4. Verificar que botón existe
5. **NO generar cuotas aún** (solo verificar que la UI funciona)

---

## Post-Despliegue

### Monitoreo Inicial

**Primeras 2 horas después del deploy**:

```bash
# Monitorear logs del backend
tail -f /var/log/asam-backend/app.log | grep -i "error\|panic\|fatal"

# Monitorear métricas
# - Latencia de requests
# - Tasa de error
# - CPU/Memory del servidor

# Verificar que no hay errores 500 en logs de acceso
tail -f /var/log/nginx/access.log | grep " 500 "
```

### Primera Ejecución en Producción

**Ejecutar con cautela la primera vez**:

1. **Backup de la BD antes de generar**:
```bash
pg_dump asam_db > backup_before_annual_fees_$(date +%Y%m%d).sql
```

2. **Generar cuotas en producción**:
   - Login como admin
   - Ir a Pagos > Generar Cuotas
   - Generar cuotas del **año actual** primero
   - Verificar resultado
   - Verificar en BD que los datos son correctos

3. **Verificar integridad**:
```sql
-- Verificar cuotas generadas
SELECT
    mf.year,
    COUNT(p.id) as total_payments,
    SUM(CASE WHEN p.status = 'pending' THEN 1 ELSE 0 END) as pending,
    SUM(CASE WHEN p.status = 'paid' THEN 1 ELSE 0 END) as paid
FROM membership_fees mf
LEFT JOIN payments p ON p.membership_fee_id = mf.id
GROUP BY mf.year
ORDER BY mf.year DESC;

-- Verificar que no hay duplicados
SELECT
    member_id,
    membership_fee_id,
    COUNT(*) as count
FROM payments
GROUP BY member_id, membership_fee_id
HAVING COUNT(*) > 1;
```

### Migración de Datos Históricos

**Si necesitas importar años pasados**:

```bash
# Script para generar múltiples años
cat > generate_historical_fees.sh <<'EOF'
#!/bin/bash

GRAPHQL_URL="https://asam.com/graphql"
ADMIN_TOKEN="<obtener_token>"

# Años y montos históricos
declare -A years=(
    [2020]="35:8"    # base:extra
    [2021]="35:8"
    [2022]="38:9"
    [2023]="40:10"
    [2024]="40:10"
)

for year in "${!years[@]}"; do
    IFS=':' read -r base extra <<< "${years[$year]}"

    echo "Generando cuotas de $year (base: $base, extra: $extra)..."

    curl -X POST $GRAPHQL_URL \
      -H "Authorization: Bearer $ADMIN_TOKEN" \
      -H "Content-Type: application/json" \
      -d "{
        \"query\": \"mutation { generateAnnualFees(input: { year: $year, baseFeeAmount: $base, familyFeeExtra: $extra }) { year paymentsGenerated paymentsExisting } }\"
      }"

    echo "\n"
    sleep 2  # Pequeña pausa entre requests
done

echo "Migración completada"
EOF

chmod +x generate_historical_fees.sh
./generate_historical_fees.sh
```

### Documentación de Producción

**Actualizar documentación interna**:

1. **Wiki/Confluence**: Añadir sección de "Generación de Cuotas Anuales"
2. **Runbook**: Documentar procedimiento operativo
3. **FAQ**: Respuestas a preguntas comunes

**Ejemplo de Runbook**:

```markdown
# Generación de Cuotas Anuales

## Cuándo Ejecutar
- Principios de enero de cada año para el año actual
- Bajo demanda para años pasados (migración)

## Quién Puede Ejecutar
- Solo usuarios con rol ADMIN

## Pasos
1. Login a https://asam.com con cuenta admin
2. Navegar a Pagos > Generar Cuotas Anuales
3. Ingresar:
   - Año: [año actual o pasado]
   - Monto base: [consultar tarifario vigente]
   - Extra familiar: [consultar tarifario vigente]
4. Click "Generar"
5. Verificar resultado

## Verificaciones
- Número de pagos generados = número de socios activos
- Montos correctos (individual vs familiar)
- No hay errores en logs

## Troubleshooting
- Si falla: Verificar logs del backend
- Si hay duplicados: Investigar idempotencia
- Si montos incorrectos: Verificar input de formulario
```

---

## Rollback

### Cuándo Hacer Rollback

- ❌ Error crítico en producción que afecta usuarios
- ❌ Pagos duplicados detectados
- ❌ Cálculos de montos incorrectos
- ❌ Performance inaceptable (> 30s para generar)

### Procedimiento de Rollback

#### Backend Rollback

**Opción A: Revertir a versión anterior**

```bash
ssh production-server

# Stop servicio
systemctl stop asam-backend

# Restaurar binario anterior
cd /app/asam-backend
cp bin/asam-backend.backup bin/asam-backend

# Start servicio
systemctl start asam-backend

# Verificar
systemctl status asam-backend
curl https://asam.com/health
```

**Opción B: Git revert**

```bash
# Identificar commit a revertir
git log --oneline

# Revert
git revert <commit-hash>
git push origin main

# Redeploy
# (seguir proceso de deploy normal)
```

#### Frontend Rollback

```bash
# Opción A: Deploy versión anterior
git checkout frontend-v1.4.0
npm run build
# Deploy a producción

# Opción B: Rollback en Vercel/Netlify
vercel rollback <deployment-url>
```

#### Base de Datos Rollback

**Solo si se generaron datos incorrectos**:

```sql
-- Identificar cuotas problemáticas
SELECT * FROM membership_fees WHERE year = 2024;

-- Eliminar pagos generados incorrectamente
DELETE FROM payments
WHERE membership_fee_id = <id_cuota_problematica>
AND status = 'pending'
AND created_at > '2024-11-07 12:00:00';

-- Eliminar cuota si es necesario
DELETE FROM membership_fees WHERE id = <id_cuota_problematica>;
```

**Restaurar desde backup**:

```bash
# Si es necesario restaurar completamente
pg_restore -d asam_db backup_before_annual_fees_20241107.sql
```

---

## Comunicación

### Anuncio a Usuarios

**Antes del Deploy (Email/Notificación)**:

```
Asunto: Nueva Funcionalidad: Generación de Cuotas Anuales

Estimados administradores,

El [FECHA] implementaremos una nueva funcionalidad que permitirá generar
las cuotas anuales de todos los socios de forma automatizada.

Esta funcionalidad facilitará:
- Generación masiva de cuotas por año
- Importación de datos históricos
- Control de pagos pendientes

La aplicación estará disponible durante el deployment, sin tiempo de caída.

Cualquier duda, contactar a soporte@asam.com
```

**Después del Deploy**:

```
Asunto: Nueva Funcionalidad Disponible: Generación de Cuotas Anuales

La funcionalidad de generación de cuotas anuales ya está disponible.

Para usarla:
1. Ir a Pagos > Generar Cuotas Anuales
2. Seleccionar año y montos
3. Confirmar generación

Documentación completa: [LINK]

Manual de usuario: [LINK]
```

---

## Métricas de Éxito

### KPIs Post-Deploy

**Primera semana**:
- [ ] 0 errores críticos reportados
- [ ] Tiempo de generación < 10 segundos para 100 socios
- [ ] 100% de idempotencia (0 duplicados)
- [ ] 0 rollbacks necesarios

**Primer mes**:
- [ ] Cuotas de todos los años históricos migradas
- [ ] Al menos 1 ciclo completo de generación anual ejecutado
- [ ] Feedback positivo de usuarios admin
- [ ] Reducción del 90% en tiempo de gestión de cuotas

---

## Checklist Final de Deploy

### Pre-Deploy
- [ ] Tests pasan en backend
- [ ] Tests pasan en frontend
- [ ] Code review aprobado
- [ ] Documentación actualizada
- [ ] Backup de BD realizado
- [ ] Stakeholders notificados

### Deploy
- [ ] Backend deployed a staging
- [ ] Frontend deployed a staging
- [ ] Tests manuales en staging OK
- [ ] Backend deployed a producción
- [ ] Frontend deployed a producción
- [ ] Smoke tests en producción OK

### Post-Deploy
- [ ] Monitoreo activo primeras 2 horas
- [ ] Logs verificados (sin errores)
- [ ] Primera generación en producción exitosa
- [ ] Métricas de performance OK
- [ ] Usuarios notificados
- [ ] Documentación publicada

---

**¡Deployment completado exitosamente!** 🚀
