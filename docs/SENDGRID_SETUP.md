# Configuración de SendGrid para ASAM Backend

## Paso 1: Crear cuenta en SendGrid

1. Ve a https://signup.sendgrid.com/
2. Completa el formulario:
   - Email: tu email personal/trabajo
   - Password: crear una contraseña segura
3. Verifica tu email

## Paso 2: Configuración inicial en SendGrid

### 2.1 Verificar tu identidad de envío (IMPORTANTE)

SendGrid requiere que verifiques desde qué dirección enviarás emails:

1. Inicia sesión en https://app.sendgrid.com/
2. Ve a **Settings → Sender Authentication**
3. Elige una opción:

#### Opción A: Single Sender Verification (Más fácil para desarrollo)
- Click en **Get Started** bajo "Single Sender Verification"
- Click en **Create a New Sender**
- Completa el formulario:
  - **From Name**: ASAM Sistema
  - **From Email Address**: noreply@asam.org (o usa tu email real)
  - **Reply To**: javierfernandezc@gmail.com
  - **Company Address**: Tu dirección
  - **City**: Tu ciudad
  - **Country**: Tu país
- Click **Create**
- **IMPORTANTE**: Verifica el email que te llegará

#### Opción B: Domain Authentication (Para producción)
- Requiere acceso a los DNS de tu dominio
- Más profesional pero más complejo

### 2.2 Crear API Key

1. Ve a **Settings → API Keys**
2. Click en **Create API Key**
3. Configuración:
   - **API Key Name**: ASAM Backend Development
   - **API Key Permissions**: 
     - Elige **Restricted Access**
     - En **Access Details**, activa solo:
       - Mail Send → Full Access
4. Click **Create & View**
5. **COPIA LA API KEY COMPLETA** (empieza con SG.)
   - ⚠️ Solo se muestra UNA VEZ
   - Guárdala en un lugar seguro

## Paso 3: Configurar el Backend

1. Abre el archivo `.env` del backend
2. Actualiza la línea:
   ```
   SMTP_PASSWORD=REEMPLAZAR-CON-TU-API-KEY-DE-SENDGRID
   ```
   Con tu API key real:
   ```
   SMTP_PASSWORD=SG.xxxxxxxxxxxxxxxxxxxx
   ```

3. Si usaste un email diferente en "From Email", actualiza también:
   ```
   SMTP_FROM_EMAIL=tu-email-verificado@ejemplo.com
   ```

4. Reinicia el backend

## Paso 4: Probar

1. Ve a http://localhost:5173/email-verification-pending
2. Click en "Reenviar Email de Verificación"
3. Revisa:
   - Tu bandeja de entrada
   - La carpeta de SPAM
   - Los logs del backend

## Troubleshooting

### Error: "The from address does not match a verified Sender Identity"
- El email en `SMTP_FROM_EMAIL` no está verificado en SendGrid
- Ve a Settings → Sender Authentication y verifica ese email

### Error: "Invalid API Key"
- La API key es incorrecta
- Crea una nueva en Settings → API Keys

### Error: "Unauthorized"
- La API key no tiene permisos de Mail Send
- Crea una nueva API key con los permisos correctos

### Los emails llegan a SPAM
- Normal en desarrollo
- En producción, configura Domain Authentication

## Monitoreo

Puedes ver las estadísticas de emails enviados en:
- Dashboard → Statistics
- Activity → Activity Feed (para ver emails individuales)

## Límites de SendGrid Free

- 100 emails por día
- Suficiente para desarrollo y testing
