# Checklist de Limpieza de Base de Datos

**Fecha programada:** Domingo ___/___/2025
**Hora programada:** _____:_____
**Ejecutado por:** _____________________

---

## ANTES DEL DOMINGO (Durante la semana)

### Preparación

- [ ] Verificar que usuarios Babacar y javi existen en producción
  ```bash
  # Ver usuarios actuales
  SELECT username, email, role FROM users WHERE username IN ('Babacar', 'javi');
  ```

- [ ] **CRÍTICO - Plan Gratuito:** Crear backup manual con pg_dump
  ```bash
  pg_dump -h pg-asam-asam-backend-db.l.aivencloud.com \
    -p 14276 -U avnadmin -d defaultdb \
    --no-owner --no-acl -F c \
    -f backup_before_cleanup_$(date +%Y%m%d_%H%M%S).dump
  ```
  - [ ] Backup creado exitosamente
  - [ ] Archivo guardado en: _____________________
  - [ ] Verificar integridad: `pg_restore -l backup*.dump | head -20`
  - [ ] Copiar a Google Drive / lugar seguro: _______________

- [ ] Verificar backup automático en Aiven Console
  - [ ] Entrar a https://console.aiven.io
  - [ ] Verificar backup existe (últimas 24h)
  - [ ] Anotar fecha del último backup: _______________
  - [ ] **NOTA**: Solo puedes hacer Restore (sobrescribe), NO Fork

- [ ] Ejecutar dry-run del script de limpieza
  ```bash
  go run cmd/cleanup-db/main.go -env production -keep-users "Babacar,javi" -dry-run
  ```
  - [ ] Revisar números mostrados
  - [ ] Anotar cantidad de registros a eliminar:
    - Members: _______
    - Payments: _______
    - Families: _______
    - Users (a eliminar): _______

- [ ] **OPCIONAL:** Crear backup manual
  ```bash
  pg_dump ... -f backup_before_cleanup_$(date +%Y%m%d).dump
  ```
  - [ ] Backup guardado en: _____________________

---

## EL DOMINGO (Día de ejecución)

### Hora de inicio: _____:_____

### Pre-verificación

- [ ] Backup reciente confirmado (últimas 24h)
- [ ] Nadie más usando el sistema
- [ ] Ambiente: PRODUCCIÓN ⚠️

### Ejecución

- [ ] **PASO 1:** Ejecutar script de limpieza
  ```bash
  go run cmd/cleanup-db/main.go -env production -keep-users "Babacar,javi"
  ```

- [ ] Escribir `DELETE ALL` cuando se solicite
- [ ] Escribir `YES` cuando se solicite confirmación de backup

- [ ] Script completado sin errores
  - Hora de finalización: _____:_____

### Verificación Post-Limpieza

- [ ] Verificar solo 2 usuarios en BD
  ```sql
  SELECT COUNT(*) FROM users;  -- Debe ser 2
  SELECT username FROM users;  -- Babacar, javi
  ```
  Cantidad de usuarios: _______

- [ ] Verificar tablas vacías
  ```sql
  SELECT COUNT(*) FROM members;          -- Debe ser 0
  SELECT COUNT(*) FROM payments;         -- Debe ser 0
  SELECT COUNT(*) FROM families;         -- Debe ser 0
  SELECT COUNT(*) FROM cash_flows;       -- Debe ser 0
  SELECT COUNT(*) FROM membership_fees;  -- Debe ser 0
  ```
  - Members: _______
  - Payments: _______
  - Families: _______
  - Cash flows: _______
  - Membership fees: _______

### Pruebas de funcionalidad

- [ ] Probar login de Babacar
  - [ ] Login exitoso ✓
  - [ ] Token recibido ✓

- [ ] Probar login de javi
  - [ ] Login exitoso ✓
  - [ ] Token recibido ✓

- [ ] Probar endpoint GraphQL
  - [ ] `/graphql` responde ✓

- [ ] Verificar aplicación en Cloud Run
  - [ ] URL: https://asam-backend-jtpswzdxuq-ew.a.run.app
  - [ ] Estado: [ ] Running [ ] Error

---

## POST-LIMPIEZA

- [ ] Todo funcionó correctamente
- [ ] Documentar resultado:
  ```
  Resultado: [ ] ÉXITO [ ] FALLO

  Notas:
  ___________________________________________________
  ___________________________________________________
  ___________________________________________________
  ```

- [ ] Actualizar historial en DATABASE_CLEANUP_PROCEDURE.md

- [ ] Archivar backup en lugar seguro

---

## EN CASO DE ERROR

Si algo falla, **DETENTE** y:

1. [ ] Revisar mensaje de error
2. [ ] Verificar estado de la BD
3. [ ] Si es necesario, restaurar desde backup:
   ```bash
   # Desde Aiven Console > Backups > Restore
   # O desde backup manual con pg_restore
   ```

---

## Números de Referencia Esperados (aprox.)

Después de las pruebas beta, se espera eliminar aproximadamente:
- Members: ~10-50 (datos de prueba)
- Families: ~5-20 (datos de prueba)
- Payments: ~20-100 (datos de prueba)
- Users: ~5-15 (solo de prueba, manteniendo Babacar y javi)

**Si los números del dry-run son MUY diferentes, revisar antes de continuar.**

---

## Contactos de Emergencia

- Aiven Support: https://console.aiven.io
- Documentación: /docs/DATABASE_CLEANUP_PROCEDURE.md
- Script: /cmd/cleanup-db/main.go

---

**Firma de ejecución:**

Ejecutado por: _____________________
Fecha: ___/___/2025
Hora: _____:_____
Resultado: [ ] ÉXITO [ ] FALLO
