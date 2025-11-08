# Feature: Creación Automática de Usuario al Crear Socio

**Estado:** 📋 Pendiente  
**Prioridad:** Media  
**Grado de Dificultad:** MEDIO (6/10)  
**Estimación Total:** 26-35 horas (~1.5-2 semanas)

---

## 📋 Descripción

Implementar funcionalidad para crear automáticamente un usuario tipo `user` asociado a un socio nuevo o existente, con envío de correo electrónico de bienvenida con credenciales temporales.

### Casos de Uso

1. **Durante la creación de un socio nuevo:**
   - El admin marca un checkbox "Crear cuenta de usuario"
   - Al guardar el socio, automáticamente se crea su usuario asociado
   - Se envía email con credenciales al socio

2. **Para socios existentes sin usuario:**
   - En la tabla de socios, aparece un botón "Crear usuario" para socios sin cuenta
   - Al hacer clic, se crea el usuario y se envían las credenciales
   - Solo visible para administradores

### Flujo de Trabajo

```
┌─────────────────────────────────────────────────────────────┐
│ Admin crea socio nuevo                                       │
│ └─ Marca checkbox "Crear cuenta de usuario"                 │
└─────────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│ Backend                                                      │
│ 1. Crea el socio en la BD                                   │
│ 2. Genera username (basado en email)                        │
│ 3. Genera contraseña temporal segura (12 chars)             │
│ 4. Crea usuario en BD (role=user, linked al socio)          │
│ 5. Marca requirePasswordChange=true                         │
│ 6. Envía email de bienvenida con credenciales               │
└─────────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│ Frontend muestra diálogo con:                               │
│ ✓ Usuario creado exitosamente                              │
│ ✓ Username: johndoe1234                                     │
│ ✓ Contraseña temporal: Abc123!@#xyz                         │
│ ✓ Email enviado ✅                                          │
│ └─ Botón para copiar contraseña                             │
└─────────────────────────────────────────────────────────────┘
```

---

## 🎯 Objetivos

- ✅ Simplificar la incorporación de nuevos socios al sistema
- ✅ Automatizar la creación de cuentas de usuario
- ✅ Garantizar envío de credenciales de forma segura
- ✅ Mantener buena experiencia de usuario tanto para admin como para socio
- ✅ Aplicar mejores prácticas de seguridad

---

## 🔧 BACKEND - Tareas Detalladas

### Backend Tarea 1: Crear Mutación `createUserForMember`
**Dificultad:** Media (5/10) | **Estimación:** 3-4 horas

**Objetivo:** Crear una mutación GraphQL específica para crear usuarios asociados a socios.

#### Archivos a modificar
- `/internal/adapters/gql/schema/schema.graphql`
- `/internal/adapters/gql/resolvers/user_resolver.go`
- `/internal/domain/services/user_service.go`

#### Actualización del Schema GraphQL

```graphql
# Agregar al schema.graphql

input CreateUserForMemberInput {
    """
    ID del socio al que se asociará el usuario
    """
    memberId: ID!
    
    """
    Si se debe enviar email de bienvenida con credenciales.
    Default: true
    """
    sendWelcomeEmail: Boolean! = true
}

type CreateUserForMemberResponse {
    """
    Usuario creado
    """
    user: User!
    
    """
    Contraseña temporal generada (solo visible en esta respuesta)
    """
    temporaryPassword: String!
    
    """
    Indica si el email fue enviado exitosamente
    """
    emailSent: Boolean!
    
    """
    Mensaje adicional (ej: "Email enviado" o "Error al enviar email")
    """
    message: String
}

type Mutation {
    # ... mutaciones existentes ...
    
    """
    Crea un usuario tipo 'user' asociado a un socio.
    Genera username y contraseña temporal automáticamente.
    Envía email de bienvenida si sendWelcomeEmail=true.
    Solo disponible para administradores.
    
    Validaciones:
    - El socio debe existir
    - El socio no debe tener ya un usuario asociado
    - El socio debe tener un email válido
    """
    createUserForMember(input: CreateUserForMemberInput!): CreateUserForMemberResponse!
}
```

#### Implementación del Resolver

```go
// internal/adapters/gql/resolvers/user_resolver.go

func (r *mutationResolver) CreateUserForMember(
    ctx context.Context,
    input model.CreateUserForMemberInput,
) (*model.CreateUserForMemberResponse, error) {
    // 1. Validar permisos (solo admin puede crear usuarios)
    currentUser := middleware.GetUserFromContext(ctx)
    if currentUser == nil || currentUser.Role != domain.RoleAdmin {
        return nil, fmt.Errorf("unauthorized: only admins can create users")
    }

    // 2. Validar que el socio existe
    member, err := r.MemberService.GetByID(ctx, input.MemberId)
    if err != nil {
        return nil, fmt.Errorf("member not found: %w", err)
    }

    // 3. Validar que el socio no tiene usuario ya asociado
    existingUser, err := r.UserService.GetUserByMemberID(ctx, input.MemberId)
    if err == nil && existingUser != nil {
        return nil, fmt.Errorf("member already has an associated user")
    }

    // 4. Validar que el socio tiene email
    if member.Email == "" {
        return nil, fmt.Errorf("member must have an email address")
    }

    // 5. Crear usuario usando el servicio
    user, tempPassword, err := r.UserService.CreateUserForMember(
        ctx,
        input.MemberId,
        input.SendWelcomeEmail,
    )
    if err != nil {
        return nil, fmt.Errorf("failed to create user: %w", err)
    }

    // 6. Enviar email si se solicitó
    emailSent := false
    message := "User created successfully"
    
    if input.SendWelcomeEmail {
        memberName := fmt.Sprintf("%s %s", member.FirstName, member.LastName)
        err = r.EmailService.SendUserWelcomeEmail(
            ctx,
            user,
            memberName,
            tempPassword,
        )
        if err != nil {
            // Log el error pero no fallar la operación
            log.Printf("Failed to send welcome email: %v", err)
            message = "User created but email could not be sent"
        } else {
            emailSent = true
            message = "User created and welcome email sent"
        }
    }

    // 7. Retornar respuesta
    return &model.CreateUserForMemberResponse{
        User:              convertUserToModel(user),
        TemporaryPassword: tempPassword,
        EmailSent:         emailSent,
        Message:           &message,
    }, nil
}
```

#### Implementación del Servicio

```go
// internal/domain/services/user_service.go

// CreateUserForMember crea un usuario asociado a un socio con contraseña temporal
func (s *UserService) CreateUserForMember(
    ctx context.Context,
    memberID string,
    sendEmail bool,
) (*domain.User, string, error) {
    // 1. Obtener información del socio
    member, err := s.memberRepo.GetByID(ctx, memberID)
    if err != nil {
        return nil, "", fmt.Errorf("failed to get member: %w", err)
    }

    // 2. Validar que el email es válido
    if member.Email == "" {
        return nil, "", fmt.Errorf("member must have an email address")
    }

    // 3. Validar que no existe usuario con ese email
    existingUser, _ := s.userRepo.GetByEmail(ctx, member.Email)
    if existingUser != nil {
        return nil, "", fmt.Errorf("email already in use by another user")
    }

    // 4. Generar username basado en email
    username := generateUsernameFromEmail(member.Email)
    
    // 5. Asegurar que el username es único
    username, err = s.ensureUniqueUsername(ctx, username)
    if err != nil {
        return nil, "", fmt.Errorf("failed to generate unique username: %w", err)
    }

    // 6. Generar contraseña temporal
    tempPassword := utils.GenerateTemporaryPassword()

    // 7. Hash de la contraseña
    hashedPassword, err := bcrypt.GenerateFromPassword(
        []byte(tempPassword),
        bcrypt.DefaultCost,
    )
    if err != nil {
        return nil, "", fmt.Errorf("failed to hash password: %w", err)
    }

    // 8. Crear usuario
    user := &domain.User{
        Username:              username,
        Email:                 member.Email,
        PasswordHash:          string(hashedPassword),
        Role:                  domain.RoleUser,
        IsActive:              true,
        RequirePasswordChange: true,  // Forzar cambio en primer login
        EmailVerified:         false,
        MemberID:              &memberID,
    }

    // 9. Guardar en BD
    createdUser, err := s.userRepo.Create(ctx, user)
    if err != nil {
        return nil, "", fmt.Errorf("failed to create user: %w", err)
    }

    return createdUser, tempPassword, nil
}

// generateUsernameFromEmail genera username basado en email
// Ejemplo: john.doe@example.com -> johndoe
func generateUsernameFromEmail(email string) string {
    parts := strings.Split(email, "@")
    if len(parts) == 0 {
        return "user"
    }
    
    // Tomar parte antes del @
    username := parts[0]
    
    // Remover caracteres especiales
    username = strings.ReplaceAll(username, ".", "")
    username = strings.ReplaceAll(username, "_", "")
    username = strings.ReplaceAll(username, "-", "")
    
    // Convertir a minúsculas
    username = strings.ToLower(username)
    
    // Limitar a 8 caracteres
    if len(username) > 8 {
        username = username[:8]
    }
    
    return username
}

// ensureUniqueUsername añade sufijo numérico si es necesario
func (s *UserService) ensureUniqueUsername(ctx context.Context, baseUsername string) (string, error) {
    username := baseUsername
    
    // Intentar hasta 1000 veces
    for i := 0; i < 1000; i++ {
        exists, err := s.userRepo.UsernameExists(ctx, username)
        if err != nil {
            return "", err
        }
        
        if !exists {
            return username, nil
        }
        
        // Añadir sufijo numérico de 4 dígitos
        suffix := fmt.Sprintf("%04d", rand.Intn(10000))
        username = baseUsername + suffix
    }
    
    return "", fmt.Errorf("could not generate unique username after 1000 attempts")
}
```

---

### Backend Tarea 2: Implementar Generación de Contraseña Temporal Segura
**Dificultad:** Baja (3/10) | **Estimación:** 1 hora

**Objetivo:** Crear utilidad para generar contraseñas temporales seguras.

#### Archivos a crear
- `/internal/domain/utils/password.go`

#### Implementación

```go
// internal/domain/utils/password.go

package utils

import (
    "crypto/rand"
    "math/big"
)

const (
    uppercaseLetters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
    lowercaseLetters = "abcdefghijklmnopqrstuvwxyz"
    digits          = "0123456789"
    symbols         = "!@#$%&*"
)

// GenerateTemporaryPassword genera una contraseña temporal segura de 12 caracteres:
// - 3 mayúsculas
// - 3 minúsculas
// - 3 números
// - 3 símbolos
func GenerateTemporaryPassword() string {
    password := make([]byte, 0, 12)
    
    // Añadir 3 mayúsculas
    password = append(password, randomChars(uppercaseLetters, 3)...)
    
    // Añadir 3 minúsculas
    password = append(password, randomChars(lowercaseLetters, 3)...)
    
    // Añadir 3 números
    password = append(password, randomChars(digits, 3)...)
    
    // Añadir 3 símbolos
    password = append(password, randomChars(symbols, 3)...)
    
    // Mezclar aleatoriamente
    return shuffle(string(password))
}

// randomChars selecciona n caracteres aleatorios de la cadena dada
func randomChars(charset string, n int) []byte {
    result := make([]byte, n)
    for i := 0; i < n; i++ {
        idx, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
        result[i] = charset[idx.Int64()]
    }
    return result
}

// shuffle mezcla aleatoriamente los caracteres de una cadena
func shuffle(s string) string {
    runes := []rune(s)
    n := len(runes)
    
    for i := n - 1; i > 0; i-- {
        j, _ := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
        runes[i], runes[j.Int64()] = runes[j.Int64()], runes[i]
    }
    
    return string(runes)
}
```

#### Tests Unitarios

```go
// internal/domain/utils/password_test.go

package utils

import (
    "strings"
    "testing"
    "unicode"
)

func TestGenerateTemporaryPassword(t *testing.T) {
    password := GenerateTemporaryPassword()
    
    // Verificar longitud
    if len(password) != 12 {
        t.Errorf("Expected password length 12, got %d", len(password))
    }
    
    // Contar tipos de caracteres
    var uppercase, lowercase, digits, symbols int
    
    for _, char := range password {
        switch {
        case unicode.IsUpper(char):
            uppercase++
        case unicode.IsLower(char):
            lowercase++
        case unicode.IsDigit(char):
            digits++
        case strings.ContainsRune("!@#$%&*", char):
            symbols++
        }
    }
    
    // Verificar que tiene al menos de cada tipo
    if uppercase < 3 {
        t.Errorf("Expected at least 3 uppercase, got %d", uppercase)
    }
    if lowercase < 3 {
        t.Errorf("Expected at least 3 lowercase, got %d", lowercase)
    }
    if digits < 3 {
        t.Errorf("Expected at least 3 digits, got %d", digits)
    }
    if symbols < 3 {
        t.Errorf("Expected at least 3 symbols, got %d", symbols)
    }
}

func TestGenerateTemporaryPassword_Uniqueness(t *testing.T) {
    // Generar 100 contraseñas y verificar que son únicas
    passwords := make(map[string]bool)
    
    for i := 0; i < 100; i++ {
        password := GenerateTemporaryPassword()
        if passwords[password] {
            t.Errorf("Generated duplicate password: %s", password)
        }
        passwords[password] = true
    }
}
```

---

### Backend Tarea 3: Implementar Plantilla de Email de Bienvenida
**Dificultad:** Media (4/10) | **Estimación:** 2-3 horas

**Objetivo:** Crear plantilla HTML de correo con credenciales y envío del mismo.

#### Archivos a crear/modificar
- `/internal/adapters/email/templates/user_welcome.html`
- `/internal/adapters/email/email_service.go`

#### Plantilla HTML

```html
<!-- internal/adapters/email/templates/user_welcome.html -->

<!DOCTYPE html>
<html lang="es">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Bienvenido a Mutua ASAM</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            line-height: 1.6;
            color: #333;
            max-width: 600px;
            margin: 0 auto;
            padding: 20px;
        }
        .header {
            background-color: #2c3e50;
            color: white;
            padding: 20px;
            text-align: center;
            border-radius: 8px 8px 0 0;
        }
        .content {
            background-color: #f8f9fa;
            padding: 30px;
            border-radius: 0 0 8px 8px;
        }
        .credentials-box {
            background: white;
            padding: 20px;
            border-radius: 8px;
            margin: 20px 0;
            border-left: 4px solid #3498db;
        }
        .credentials-box p {
            margin: 10px 0;
        }
        .password {
            font-family: 'Courier New', monospace;
            font-size: 16px;
            font-weight: bold;
            color: #2c3e50;
            background: #ecf0f1;
            padding: 8px 12px;
            border-radius: 4px;
            display: inline-block;
        }
        .warning-box {
            background: #fff3cd;
            padding: 15px;
            border-radius: 8px;
            border-left: 4px solid #ffc107;
            margin: 20px 0;
        }
        .warning-box strong {
            color: #856404;
        }
        .warning-box ul {
            margin: 10px 0;
            padding-left: 20px;
        }
        .button {
            display: inline-block;
            background-color: #3498db;
            color: white;
            padding: 12px 30px;
            text-decoration: none;
            border-radius: 4px;
            margin: 20px 0;
        }
        .footer {
            text-align: center;
            margin-top: 30px;
            padding-top: 20px;
            border-top: 1px solid #ddd;
            color: #777;
            font-size: 14px;
        }
    </style>
</head>
<body>
    <div class="header">
        <h1>¡Bienvenido a Mutua ASAM!</h1>
    </div>
    
    <div class="content">
        <p>Estimado/a <strong>{{.MemberName}}</strong>,</p>
        
        <p>Nos complace informarte que se ha creado tu cuenta de usuario para acceder al portal de socios de Mutua ASAM.</p>
        
        <div class="credentials-box">
            <h2 style="margin-top: 0; color: #2c3e50;">Tus credenciales de acceso:</h2>
            
            <p><strong>Usuario:</strong> <span class="password">{{.Username}}</span></p>
            <p><strong>Contraseña temporal:</strong> <span class="password">{{.TemporaryPassword}}</span></p>
        </div>
        
        <div class="warning-box">
            <p><strong>⚠️ IMPORTANTE - Por tu seguridad:</strong></p>
            <ul>
                <li>Debes <strong>cambiar tu contraseña</strong> en cuanto inicies sesión por primera vez.</li>
                <li>No compartas tus credenciales con nadie.</li>
                <li>Esta contraseña es temporal y única para ti.</li>
                <li>Si no solicitaste esta cuenta, contacta con nosotros inmediatamente.</li>
            </ul>
        </div>
        
        <p style="text-align: center;">
            <a href="{{.PortalURL}}" class="button">Acceder al Portal</a>
        </p>
        
        <p>Una vez dentro del portal podrás:</p>
        <ul>
            <li>Consultar tu historial de pagos</li>
            <li>Actualizar tus datos personales</li>
            <li>Realizar pagos de cuotas</li>
            <li>Acceder a información importante</li>
        </ul>
        
        <p>Si tienes alguna pregunta o necesitas ayuda, no dudes en contactarnos.</p>
        
        <p>Saludos cordiales,<br>
        <strong>Equipo de Mutua ASAM</strong></p>
    </div>
    
    <div class="footer">
        <p>Mutua ASAM - Asociación Senegalesa de Ayuda Mutua</p>
        <p>www.mutuaasam.org</p>
    </div>
</body>
</html>
```

#### Implementación del Servicio de Email

```go
// internal/adapters/email/email_service.go

// SendUserWelcomeEmail envía email de bienvenida con credenciales
func (s *EmailService) SendUserWelcomeEmail(
    ctx context.Context,
    user *domain.User,
    memberName string,
    temporaryPassword string,
) error {
    // 1. Preparar datos para la plantilla
    data := map[string]interface{}{
        "MemberName":        memberName,
        "Username":          user.Username,
        "TemporaryPassword": temporaryPassword,
        "PortalURL":         s.config.PortalURL, // ej: https://www.mutuaasam.org
    }

    // 2. Renderizar plantilla
    htmlBody, err := s.renderTemplate("user_welcome.html", data)
    if err != nil {
        return fmt.Errorf("failed to render email template: %w", err)
    }

    // 3. Crear versión de texto plano
    textBody := fmt.Sprintf(`
¡Bienvenido a Mutua ASAM!

Estimado/a %s,

Se ha creado tu cuenta de usuario para acceder al portal de socios.

Tus credenciales de acceso:
- Usuario: %s
- Contraseña temporal: %s

⚠️ IMPORTANTE:
Por tu seguridad, debes cambiar tu contraseña en cuanto inicies sesión por primera vez.

Puedes acceder al portal en: %s

Saludos cordiales,
Equipo de Mutua ASAM
    `, memberName, user.Username, temporaryPassword, s.config.PortalURL)

    // 4. Preparar mensaje
    msg := &EmailMessage{
        To:          []string{user.Email},
        Subject:     "Bienvenido a Mutua ASAM - Tus credenciales de acceso",
        TextBody:    textBody,
        HTMLBody:    htmlBody,
        From:        s.config.FromEmail,
        FromName:    "Mutua ASAM",
        ReplyTo:     s.config.SupportEmail,
    }

    // 5. Enviar
    if err := s.Send(ctx, msg); err != nil {
        return fmt.Errorf("failed to send welcome email: %w", err)
    }

    log.Printf("Welcome email sent to %s (username: %s)", user.Email, user.Username)
    return nil
}

// renderTemplate renderiza una plantilla HTML con los datos proporcionados
func (s *EmailService) renderTemplate(templateName string, data map[string]interface{}) (string, error) {
    templatePath := filepath.Join(s.config.TemplatesPath, templateName)
    
    tmpl, err := template.ParseFiles(templatePath)
    if err != nil {
        return "", fmt.Errorf("failed to parse template: %w", err)
    }

    var buf bytes.Buffer
    if err := tmpl.Execute(&buf, data); err != nil {
        return "", fmt.Errorf("failed to execute template: %w", err)
    }

    return buf.String(), nil
}
```

---

### Backend Tarea 4: Añadir Campo `requirePasswordChange` al Modelo User
**Dificultad:** Baja (2/10) | **Estimación:** 1 hora

**Objetivo:** Añadir flag para forzar cambio de contraseña en primer login.

#### Migración de Base de Datos

```sql
-- migrations/YYYYMMDDHHMMSS_add_require_password_change.sql

-- Add require_password_change column to users table
ALTER TABLE users ADD COLUMN require_password_change BOOLEAN DEFAULT FALSE;

-- Set existing users to not require password change
UPDATE users SET require_password_change = FALSE WHERE require_password_change IS NULL;

-- Make it NOT NULL
ALTER TABLE users ALTER COLUMN require_password_change SET NOT NULL;

-- Add comment
COMMENT ON COLUMN users.require_password_change IS 'Flag indicating if user must change password on next login';
```

#### Actualización del Schema GraphQL

```graphql
type User {
    id: ID!
    username: String!
    email: String!
    role: UserRole!
    member: Member
    isActive: Boolean!
    lastLogin: Time
    emailVerified: Boolean!
    emailVerifiedAt: Time
    requirePasswordChange: Boolean!  # NUEVO CAMPO
}

type AuthResponse {
    user: User!
    accessToken: JWT!
    refreshToken: JWT!
    expiresAt: Time!
    requirePasswordChange: Boolean!  # NUEVO CAMPO para facilitar frontend
}
```

#### Actualización del Modelo de Dominio

```go
// internal/domain/user.go

type User struct {
    ID                    string
    Username              string
    Email                 string
    PasswordHash          string
    Role                  UserRole
    IsActive              bool
    LastLogin             *time.Time
    EmailVerified         bool
    EmailVerifiedAt       *time.Time
    MemberID              *string
    RequirePasswordChange bool      // NUEVO CAMPO
    CreatedAt             time.Time
    UpdatedAt             time.Time
}
```

#### Actualización del Resolver de Login

```go
// internal/adapters/gql/resolvers/auth_resolver.go

func (r *mutationResolver) Login(ctx context.Context, input model.LoginInput) (*model.AuthResponse, error) {
    // ... lógica existente de login ...

    // Incluir requirePasswordChange en la respuesta
    return &model.AuthResponse{
        User:                  convertUserToModel(user),
        AccessToken:           accessToken,
        RefreshToken:          refreshToken,
        ExpiresAt:             expiresAt,
        RequirePasswordChange: user.RequirePasswordChange,  // NUEVO
    }, nil
}
```

---

### Backend Tarea 5: Añadir Query `getMemberUserStatus`
**Dificultad:** Baja (3/10) | **Estimación:** 1 hora

**Objetivo:** Permitir al frontend verificar si un socio tiene usuario asociado.

#### Actualización del Schema

```graphql
type MemberUserStatus {
    memberId: ID!
    hasUser: Boolean!
    userId: ID
    username: String
}

type Query {
    # ... queries existentes ...
    
    """
    Obtiene el estado de usuario asociado a un socio.
    Útil para determinar si mostrar botón "Crear usuario".
    """
    getMemberUserStatus(memberId: ID!): MemberUserStatus!
}
```

#### Implementación del Resolver

```go
// internal/adapters/gql/resolvers/user_resolver.go

func (r *queryResolver) GetMemberUserStatus(
    ctx context.Context,
    memberID string,
) (*model.MemberUserStatus, error) {
    // Obtener usuario asociado al socio (si existe)
    user, err := r.UserService.GetUserByMemberID(ctx, memberID)
    
    status := &model.MemberUserStatus{
        MemberID: memberID,
        HasUser:  user != nil,
    }
    
    if user != nil {
        status.UserID = &user.ID
        status.Username = &user.Username
    }
    
    return status, nil
}
```

#### Implementación en el Servicio

```go
// internal/domain/services/user_service.go

// GetUserByMemberID obtiene el usuario asociado a un socio
func (s *UserService) GetUserByMemberID(ctx context.Context, memberID string) (*domain.User, error) {
    return s.userRepo.GetByMemberID(ctx, memberID)
}
```

#### Actualización del Repositorio

```go
// internal/domain/repositories/user_repository.go

type UserRepository interface {
    // ... métodos existentes ...
    
    // GetByMemberID obtiene usuario por ID de socio
    GetByMemberID(ctx context.Context, memberID string) (*User, error)
}
```

```go
// internal/adapters/db/user_repository.go

func (r *userRepository) GetByMemberID(ctx context.Context, memberID string) (*domain.User, error) {
    query := `
        SELECT id, username, email, password_hash, role, is_active,
               last_login, email_verified, email_verified_at, member_id,
               require_password_change, created_at, updated_at
        FROM users
        WHERE member_id = $1 AND is_active = true
    `
    
    var user domain.User
    err := r.db.QueryRowContext(ctx, query, memberID).Scan(
        &user.ID,
        &user.Username,
        &user.Email,
        &user.PasswordHash,
        &user.Role,
        &user.IsActive,
        &user.LastLogin,
        &user.EmailVerified,
        &user.EmailVerifiedAt,
        &user.MemberID,
        &user.RequirePasswordChange,
        &user.CreatedAt,
        &user.UpdatedAt,
    )
    
    if err == sql.ErrNoRows {
        return nil, nil  // No error, simplemente no hay usuario
    }
    
    if err != nil {
        return nil, fmt.Errorf("failed to get user by member ID: %w", err)
    }
    
    return &user, nil
}
```

---

## 🎨 FRONTEND - Tareas Detalladas

### Frontend Tarea 1: Añadir Checkbox en Formulario de Creación de Socio
**Dificultad:** Baja (3/10) | **Estimación:** 1-2 horas

**Objetivo:** Permitir crear usuario al crear socio.

#### Archivos a modificar
- `/src/features/members/components/MemberForm.tsx`

#### Cambios en el Componente

```tsx
// Añadir estado para el checkbox
const [createUser, setCreateUser] = React.useState(false)

// Añadir nueva sección antes de los botones de acción
<Grid item xs={12}>
  <Divider sx={{ my: 3 }} />
  <Typography variant="h6" sx={{ mb: 2 }}>
    {t('memberForm.sections.userAccount')}
  </Typography>
  
  <FormControlLabel
    control={
      <Checkbox
        checked={createUser}
        onChange={(e) => setCreateUser(e.target.checked)}
      />
    }
    label={
      <Box>
        <Typography variant="body1">
          {t('memberForm.fields.createUserAccount')}
        </Typography>
        <Typography variant="caption" color="text.secondary">
          {t('memberForm.helpers.createUserAccountHelp')}
        </Typography>
      </Box>
    }
  />
  
  {createUser && (
    <Alert severity="info" sx={{ mt: 2 }}>
      {t('memberForm.helpers.createUserAccountInfo')}
    </Alert>
  )}
</Grid>

// Modificar el callback de submit para incluir el flag
const handleFormSubmit = React.useCallback(
  async (data: MemberFormData) => {
    // ... validaciones existentes ...
    
    const formattedData = {
      ...data,
      // ... resto del formateo ...
      createUser,  // AÑADIR ESTE FLAG
    }
    
    if (onSubmit) {
      await onSubmit(formattedData)
    }
  },
  [/* dependencias */, createUser]  // Añadir createUser a las dependencias
)
```

---

### Frontend Tarea 2: Implementar Mutación GraphQL `createUserForMember`
**Dificultad:** Baja (2/10) | **Estimación:** 30 min

#### Archivos a crear
- `/src/features/users/api/mutations.ts`

```typescript
import { gql } from '@apollo/client'

export const CREATE_USER_FOR_MEMBER = gql`
  mutation CreateUserForMember($input: CreateUserForMemberInput!) {
    createUserForMember(input: $input) {
      user {
        id
        username
        email
        role
        member {
          miembro_id
          nombre
          apellidos
        }
      }
      temporaryPassword
      emailSent
      message
    }
  }
`

export const GET_MEMBER_USER_STATUS = gql`
  query GetMemberUserStatus($memberId: ID!) {
    getMemberUserStatus(memberId: $memberId) {
      memberId
      hasUser
      userId
      username
    }
  }
`
```

Luego ejecutar codegen:
```bash
npm run codegen
```

---

### Frontend Tarea 3: Crear Hook `useCreateUserForMember`
**Dificultad:** Media (4/10) | **Estimación:** 1-2 horas

#### Archivos a crear
- `/src/features/users/hooks/useCreateUserForMember.ts`

```typescript
import { useMutation } from '@apollo/client'
import { useTranslation } from 'react-i18next'
import { CREATE_USER_FOR_MEMBER } from '../api/mutations'
import type {
  CreateUserForMemberMutation,
  CreateUserForMemberMutationVariables,
} from '@/graphql/generated/operations'

interface CreateUserResult {
  user: {
    id: string
    username: string
    email: string
  }
  temporaryPassword: string
  emailSent: boolean
  message?: string
}

interface UseCreateUserForMemberResult {
  createUserForMember: (memberId: string) => Promise<CreateUserResult>
  loading: boolean
  error: Error | undefined
}

export const useCreateUserForMember = (): UseCreateUserForMemberResult => {
  const { t } = useTranslation('users')
  
  const [createUserMutation, { loading, error }] = useMutation<
    CreateUserForMemberMutation,
    CreateUserForMemberMutationVariables
  >(CREATE_USER_FOR_MEMBER, {
    onError: (error) => {
      console.error('Error creating user for member:', error)
    },
  })

  const createUserForMember = async (memberId: string): Promise<CreateUserResult> => {
    try {
      const result = await createUserMutation({
        variables: {
          input: {
            memberId,
            sendWelcomeEmail: true,
          },
        },
      })

      if (!result.data?.createUserForMember) {
        throw new Error(t('errors.userCreationFailed'))
      }

      return {
        user: {
          id: result.data.createUserForMember.user.id,
          username: result.data.createUserForMember.user.username,
          email: result.data.createUserForMember.user.email,
        },
        temporaryPassword: result.data.createUserForMember.temporaryPassword,
        emailSent: result.data.createUserForMember.emailSent,
        message: result.data.createUserForMember.message || undefined,
      }
    } catch (err) {
      console.error('Failed to create user for member:', err)
      throw err
    }
  }

  return {
    createUserForMember,
    loading,
    error: error as Error | undefined,
  }
}
```

---

### Frontend Tarea 4: Crear Diálogo de Confirmación de Credenciales
**Dificultad:** Media (5/10) | **Estimación:** 2-3 horas

#### Archivos a crear
- `/src/features/users/components/UserCredentialsDialog.tsx`

```tsx
import React, { useState } from 'react'
import {
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Button,
  Typography,
  Box,
  Alert,
  IconButton,
  Tooltip,
} from '@mui/material'
import {
  ContentCopy as ContentCopyIcon,
  CheckCircle as CheckCircleIcon,
  Close as CloseIcon,
} from '@mui/icons-material'
import { useTranslation } from 'react-i18next'

interface UserCredentialsDialogProps {
  open: boolean
  username: string
  temporaryPassword: string
  memberName: string
  emailSent: boolean
  message?: string
  onClose: () => void
}

export const UserCredentialsDialog: React.FC<UserCredentialsDialogProps> = ({
  open,
  username,
  temporaryPassword,
  memberName,
  emailSent,
  message,
  onClose,
}) => {
  const { t } = useTranslation('users')
  const [copiedPassword, setCopiedPassword] = useState(false)
  const [copiedUsername, setCopiedUsername] = useState(false)

  const handleCopyPassword = async () => {
    try {
      await navigator.clipboard.writeText(temporaryPassword)
      setCopiedPassword(true)
      setTimeout(() => setCopiedPassword(false), 2000)
    } catch (err) {
      console.error('Failed to copy password:', err)
    }
  }

  const handleCopyUsername = async () => {
    try {
      await navigator.clipboard.writeText(username)
      setCopiedUsername(true)
      setTimeout(() => setCopiedUsername(false), 2000)
    } catch (err) {
      console.error('Failed to copy username:', err)
    }
  }

  return (
    <Dialog open={open} onClose={onClose} maxWidth="sm" fullWidth>
      <DialogTitle>
        <Box display="flex" alignItems="center" justifyContent="space-between">
          <Box display="flex" alignItems="center" gap={1}>
            <CheckCircleIcon color="success" />
            <Typography variant="h6">
              {t('credentials.title')}
            </Typography>
          </Box>
          <IconButton onClick={onClose} size="small">
            <CloseIcon />
          </IconButton>
        </Box>
      </DialogTitle>

      <DialogContent>
        {/* Mensaje de éxito */}
        <Alert severity="success" sx={{ mb: 2 }}>
          {t('credentials.description', { memberName })}
        </Alert>

        {/* Estado del email */}
        {emailSent ? (
          <Alert severity="info" sx={{ mb: 2 }}>
            ✅ {t('credentials.emailSent')}
          </Alert>
        ) : (
          <Alert severity="warning" sx={{ mb: 2 }}>
            ⚠️ {message || t('credentials.emailNotSent')}
          </Alert>
        )}

        {/* Credenciales */}
        <Box
          sx={{
            bgcolor: 'grey.100',
            p: 2,
            borderRadius: 1,
            mb: 2,
          }}
        >
          {/* Username */}
          <Box mb={2}>
            <Typography variant="body2" color="text.secondary" gutterBottom>
              <strong>{t('credentials.username')}</strong>
            </Typography>
            <Box display="flex" alignItems="center" gap={1}>
              <Typography
                variant="body1"
                sx={{
                  fontFamily: 'monospace',
                  bgcolor: 'white',
                  p: 1,
                  borderRadius: 0.5,
                  flex: 1,
                }}
              >
                {username}
              </Typography>
              <Tooltip title={copiedUsername ? t('credentials.copied') : t('credentials.copyButton')}>
                <IconButton
                  size="small"
                  onClick={handleCopyUsername}
                  color={copiedUsername ? 'success' : 'default'}
                >
                  <ContentCopyIcon />
                </IconButton>
              </Tooltip>
            </Box>
          </Box>

          {/* Password */}
          <Box>
            <Typography variant="body2" color="text.secondary" gutterBottom>
              <strong>{t('credentials.temporaryPassword')}</strong>
            </Typography>
            <Box display="flex" alignItems="center" gap={1}>
              <Typography
                variant="body1"
                sx={{
                  fontFamily: 'monospace',
                  bgcolor: 'white',
                  p: 1,
                  borderRadius: 0.5,
                  flex: 1,
                  fontSize: '1.1rem',
                  fontWeight: 'bold',
                  letterSpacing: '0.05em',
                }}
              >
                {temporaryPassword}
              </Typography>
              <Tooltip title={copiedPassword ? t('credentials.copied') : t('credentials.copyButton')}>
                <IconButton
                  size="small"
                  onClick={handleCopyPassword}
                  color={copiedPassword ? 'success' : 'default'}
                >
                  <ContentCopyIcon />
                </IconButton>
              </Tooltip>
            </Box>
          </Box>
        </Box>

        {/* Advertencia de seguridad */}
        <Alert severity="warning">
          <Typography variant="body2">
            <strong>{t('credentials.important')}</strong>{' '}
            {t('credentials.changePasswordWarning')}
          </Typography>
        </Alert>
      </DialogContent>

      <DialogActions>
        <Button onClick={onClose} variant="contained" fullWidth>
          {t('credentials.close')}
        </Button>
      </DialogActions>
    </Dialog>
  )
}
```

---

### Frontend Tarea 5: Integrar Creación de Usuario en Flujo de Creación de Socio
**Dificultad:** Media (5/10) | **Estimación:** 2-3 horas

#### Archivos a modificar
- `/src/pages/members/NewMemberPage.tsx` (o donde se maneje la creación)

```tsx
import { useState } from 'react'
import { useCreateUserForMember } from '@/features/users/hooks/useCreateUserForMember'
import { UserCredentialsDialog } from '@/features/users/components/UserCredentialsDialog'

// Estado para el diálogo de credenciales
const [credentialsDialog, setCredentialsDialog] = useState<{
  open: boolean
  username: string
  temporaryPassword: string
  memberName: string
  emailSent: boolean
  message?: string
} | null>(null)

const { createUserForMember, loading: creatingUser } = useCreateUserForMember()

const handleSubmit = async (data: MemberFormData & { createUser?: boolean }) => {
  setIsSubmitting(true)
  try {
    // 1. Crear el socio
    const result = await createMember({
      variables: {
        input: {
          // ... datos del socio
        },
      },
    })

    const newMember = result.data?.createMember

    if (!newMember) {
      throw new Error('Failed to create member')
    }

    // 2. Si se marcó el checkbox, crear usuario
    if (data.createUser) {
      try {
        const userResult = await createUserForMember(newMember.miembro_id)

        // 3. Mostrar diálogo con credenciales
        setCredentialsDialog({
          open: true,
          username: userResult.user.username,
          temporaryPassword: userResult.temporaryPassword,
          memberName: `${newMember.nombre} ${newMember.apellidos}`,
          emailSent: userResult.emailSent,
          message: userResult.message,
        })

        // No navegar automáticamente, esperar a que el admin cierre el diálogo
      } catch (userError) {
        // Si falla la creación del usuario, mostrar error pero
        // el socio ya fue creado exitosamente
        console.error('Failed to create user:', userError)
        enqueueSnackbar(
          t('errors.userCreationFailedButMemberCreated'),
          { variant: 'warning' }
        )
        navigate('/members')
      }
    } else {
      // Si no se crea usuario, navegar directamente
      enqueueSnackbar(t('success.memberCreated'), { variant: 'success' })
      navigate('/members')
    }
  } catch (error) {
    console.error('Failed to create member:', error)
    enqueueSnackbar(t('errors.memberCreationFailed'), { variant: 'error' })
  } finally {
    setIsSubmitting(false)
  }
}

const handleCloseCredentialsDialog = () => {
  setCredentialsDialog(null)
  // Navegar a la lista de socios después de cerrar el diálogo
  navigate('/members')
}

// En el return del componente
return (
  <>
    <MemberForm
      onSubmit={handleSubmit}
      // ... otras props
    />

    {credentialsDialog && (
      <UserCredentialsDialog
        open={credentialsDialog.open}
        username={credentialsDialog.username}
        temporaryPassword={credentialsDialog.temporaryPassword}
        memberName={credentialsDialog.memberName}
        emailSent={credentialsDialog.emailSent}
        message={credentialsDialog.message}
        onClose={handleCloseCredentialsDialog}
      />
    )}
  </>
)
```

---

### Frontend Tarea 6: Añadir Acción "Crear Usuario" en Tabla de Miembros
**Dificultad:** Media (6/10) | **Estimación:** 3-4 horas

#### Archivos a modificar
- `/src/features/members/components/MembersTable.tsx`

#### Archivos a crear
- `/src/features/members/components/CreateUserAction.tsx`

**Componente CreateUserAction:**

```tsx
// src/features/members/components/CreateUserAction.tsx

import React, { useState } from 'react'
import { IconButton, Tooltip, CircularProgress } from '@mui/material'
import { PersonAdd as PersonAddIcon } from '@mui/icons-material'
import { useTranslation } from 'react-i18next'
import { useSnackbar } from 'notistack'
import { useCreateUserForMember } from '@/features/users/hooks/useCreateUserForMember'
import { UserCredentialsDialog } from '@/features/users/components/UserCredentialsDialog'
import type { Member } from '../types'

interface CreateUserActionProps {
  member: Member
  onSuccess?: () => void
}

export const CreateUserAction: React.FC<CreateUserActionProps> = ({
  member,
  onSuccess,
}) => {
  const { t } = useTranslation(['members', 'users'])
  const { enqueueSnackbar } = useSnackbar()
  const { createUserForMember, loading } = useCreateUserForMember()

  const [dialogOpen, setDialogOpen] = useState(false)
  const [credentials, setCredentials] = useState<{
    username: string
    temporaryPassword: string
    emailSent: boolean
    message?: string
  } | null>(null)

  const handleClick = async () => {
    try {
      const result = await createUserForMember(member.miembro_id)

      setCredentials({
        username: result.user.username,
        temporaryPassword: result.temporaryPassword,
        emailSent: result.emailSent,
        message: result.message,
      })
      setDialogOpen(true)

      enqueueSnackbar(
        t('users:success.userCreated'),
        { variant: 'success' }
      )

      if (onSuccess) {
        onSuccess()
      }
    } catch (error) {
      console.error('Error creating user:', error)
      enqueueSnackbar(
        t('users:errors.userCreationFailed'),
        { variant: 'error' }
      )
    }
  }

  const handleCloseDialog = () => {
    setDialogOpen(false)
  }

  return (
    <>
      <Tooltip title={t('members:list.actions.createUser')}>
        <span>
          <IconButton
            size="small"
            onClick={handleClick}
            disabled={loading}
          >
            {loading ? (
              <CircularProgress size={20} />
            ) : (
              <PersonAddIcon fontSize="small" />
            )}
          </IconButton>
        </span>
      </Tooltip>

      {credentials && (
        <UserCredentialsDialog
          open={dialogOpen}
          username={credentials.username}
          temporaryPassword={credentials.temporaryPassword}
          memberName={`${member.nombre} ${member.apellidos}`}
          emailSent={credentials.emailSent}
          message={credentials.message}
          onClose={handleCloseDialog}
        />
      )}
    </>
  )
}
```

**Modificación de MembersTable:**

```tsx
// src/features/members/components/MembersTable.tsx

import { useQuery } from '@apollo/client'
import { GET_MEMBER_USER_STATUS } from '@/features/users/api/mutations'
import { CreateUserAction } from './CreateUserAction'

// En la definición de columnas, modificar la columna de acciones:
{
  field: 'actions',
  headerName: t('list.columns.actions'),
  width: isAdmin ? 180 : 80,  // Aumentar ancho para nueva acción
  sortable: false,
  filterable: false,
  disableColumnMenu: true,
  renderCell: (params) => {
    // Hook para verificar si el socio tiene usuario
    const { data: userStatusData } = useQuery(GET_MEMBER_USER_STATUS, {
      variables: { memberId: params.row?.miembro_id },
      skip: !params.row || !isAdmin,
    })

    const hasUser = userStatusData?.getMemberUserStatus?.hasUser

    return (
      <Box sx={{ display: 'flex', gap: 0.5 }}>
        {/* Acción: Ver */}
        <Tooltip title={t('list.actions.view')}>
          <IconButton
            size="small"
            onClick={(e) => {
              e.stopPropagation()
              if (params.row) {
                onRowClick(params.row)
              }
            }}
          >
            <VisibilityIcon fontSize="small" />
          </IconButton>
        </Tooltip>

        {isAdmin && (
          <>
            {/* Acción: Editar */}
            <Tooltip title={t('list.actions.edit')}>
              <IconButton
                size="small"
                onClick={(e) => {
                  e.stopPropagation()
                  if (params.row) {
                    onEditClick(params.row)
                  }
                }}
              >
                <EditIcon fontSize="small" />
              </IconButton>
            </Tooltip>

            {/* NUEVA ACCIÓN: Crear usuario (solo si no tiene) */}
            {!hasUser && params.row && (
              <CreateUserAction
                member={params.row}
                onSuccess={() => {
                  // Opcional: refrescar la lista
                  // refetch()
                }}
              />
            )}

            {/* Acción: Desactivar */}
            <Tooltip title={t('table.deactivate')}>
              <span>
                <IconButton
                  size="small"
                  disabled={params.row?.estado === MemberStatus.INACTIVE}
                  onClick={(e) => {
                    e.stopPropagation()
                    if (params.row) {
                      onDeactivateClick(params.row)
                    }
                  }}
                >
                  <PersonRemoveIcon fontSize="small" />
                </IconButton>
              </span>
            </Tooltip>
          </>
        )}
      </Box>
    )
  },
}
```

---

### Frontend Tarea 7: Añadir Traducciones
**Dificultad:** Baja (2/10) | **Estimación:** 30 min

#### Archivos a modificar
- `/src/locales/es/members.json`
- `/src/locales/es/users.json`

**members.json:**

```json
{
  "memberForm": {
    "sections": {
      "userAccount": "Cuenta de Usuario"
    },
    "fields": {
      "createUserAccount": "Crear cuenta de usuario para este socio"
    },
    "helpers": {
      "createUserAccountHelp": "Se enviará un correo electrónico con las credenciales de acceso",
      "createUserAccountInfo": "Se generará automáticamente un nombre de usuario basado en el email del socio y una contraseña temporal segura."
    }
  },
  "list": {
    "actions": {
      "createUser": "Crear usuario"
    }
  }
}
```

**users.json:**

```json
{
  "credentials": {
    "title": "Usuario creado exitosamente",
    "description": "Se ha creado un usuario para {memberName}",
    "emailSent": "Se ha enviado un correo electrónico con las credenciales",
    "emailNotSent": "No se pudo enviar el correo electrónico. Por favor, comparte las credenciales manualmente.",
    "username": "Usuario",
    "temporaryPassword": "Contraseña temporal",
    "copyButton": "Copiar",
    "copied": "¡Copiado!",
    "important": "⚠️ Importante:",
    "changePasswordWarning": "El usuario deberá cambiar esta contraseña en su primer inicio de sesión por motivos de seguridad.",
    "close": "Cerrar"
  },
  "success": {
    "userCreated": "Usuario creado exitosamente"
  },
  "errors": {
    "userCreationFailed": "Error al crear el usuario",
    "userCreationFailedButMemberCreated": "El socio se creó correctamente, pero hubo un error al crear su usuario. Por favor, créalo manualmente desde la tabla de socios."
  }
}
```

---

## 📝 Orden de Implementación Sugerido

### **Fase 1: Backend Core** (Estimación: 5-7 horas)
1. ✅ Backend Tarea 2: Generación de contraseñas (1h)
2. ✅ Backend Tarea 4: Campo `requirePasswordChange` (1h)
3. ✅ Backend Tarea 1: Mutación `createUserForMember` (3-4h)

### **Fase 2: Backend Email** (Estimación: 2-3 horas)
4. ✅ Backend Tarea 3: Plantilla y envío de email (2-3h)

### **Fase 3: Backend Query** (Estimación: 1 hora)
5. ✅ Backend Tarea 5: Query `getMemberUserStatus` (1h)

### **Fase 4: Frontend Core** (Estimación: 3-4 horas)
6. ✅ Frontend Tarea 2: Mutación GraphQL (30min)
7. ✅ Frontend Tarea 3: Hook `useCreateUserForMember` (1-2h)
8. ✅ Frontend Tarea 4: Diálogo de credenciales (2-3h)

### **Fase 5: Frontend Integración** (Estimación: 6-8 horas)
9. ✅ Frontend Tarea 1: Checkbox en formulario (1-2h)
10. ✅ Frontend Tarea 5: Integración en flujo de creación (2-3h)
11. ✅ Frontend Tarea 6: Acción en tabla (3-4h)
12. ✅ Frontend Tarea 7: Traducciones (30min)

---

## 🧪 Testing

### Backend
- [ ] Test unitario: generación de contraseñas seguras
- [ ] Test unitario: generación de username único
- [ ] Test de integración: creación de usuario para socio
- [ ] Test de integración: envío de email (con mock)
- [ ] Test: validación de socio sin usuario previo
- [ ] Test: validación de email único
- [ ] Test: error cuando socio ya tiene usuario

### Frontend
- [ ] Test: renderizado del checkbox en formulario
- [ ] Test: integración en flujo de creación
- [ ] Test: diálogo de credenciales
- [ ] Test: botón copy to clipboard
- [ ] Test: acción en tabla solo visible cuando no hay usuario
- [ ] Test: loading states
- [ ] Test: error handling

---

## ⚠️ Consideraciones Importantes

### 1. **Seguridad**
- ✅ Contraseñas temporales deben ser únicas, aleatorias y seguras (12+ caracteres)
- ✅ Usuario debe cambiar contraseña en primer login (`requirePasswordChange=true`)
- ✅ Solo administradores pueden crear usuarios
- ✅ Validar que email del socio es único en tabla users
- ✅ Username generado debe ser único

### 2. **Email**
- ✅ Verificar que el socio tenga email válido antes de crear usuario
- ✅ Manejar fallos en envío de email (crear usuario pero avisar que email falló)
- ✅ Incluir todas las instrucciones necesarias en el email
- ✅ Email debe ser responsive y verse bien en dispositivos móviles

### 3. **UX (Experiencia de Usuario)**
- ✅ Mostrar claramente las credenciales al admin con posibilidad de copiar
- ✅ Informar si el email fue enviado exitosamente o si hubo error
- ✅ No navegar automáticamente después de crear usuario, esperar a que admin cierre diálogo
- ✅ Si falla creación de usuario, el socio debe quedar creado (no hacer rollback)
- ✅ Loading states claros durante las operaciones asíncronas

### 4. **Validaciones**
- ✅ Un socio no puede tener más de un usuario
- ✅ El email del socio debe ser único en la tabla users
- ✅ Username generado debe ser único (usar sufijos si es necesario)
- ✅ Contraseña temporal debe cumplir requisitos de complejidad
- ✅ Solo socios activos pueden tener usuarios asociados

### 5. **Manejo de Errores**
- ✅ Si falla creación de usuario después de crear socio, NO hacer rollback del socio
- ✅ Mostrar mensajes de error claros y accionables
- ✅ Logging apropiado en backend para debugging
- ✅ Capturar todos los errores posibles (BD, email, validaciones)

---

## 📊 Desglose de Estimación

| Fase | Backend | Frontend | Testing | Total |
|------|---------|----------|---------|-------|
| Fase 1: Backend Core | 5-7h | - | 2h | 7-9h |
| Fase 2: Backend Email | 2-3h | - | 1h | 3-4h |
| Fase 3: Backend Query | 1h | - | 0.5h | 1.5h |
| Fase 4: Frontend Core | - | 3-4h | 1h | 4-5h |
| Fase 5: Frontend Integración | - | 6-8h | 2h | 8-10h |
| **TOTAL** | **8-11h** | **9-12h** | **6.5h** | **23.5-29.5h** |

**Estimación conservadora: 26-35 horas (~1.5-2 semanas con dedicación parcial)**

---

## 🎯 Criterios de Aceptación

### Backend
- [ ] Mutación `createUserForMember` implementada y funcional
- [ ] Generación de contraseñas temporales seguras (12 chars, 4 tipos de caracteres)
- [ ] Generación de username único basado en email
- [ ] Campo `requirePasswordChange` añadido a modelo User y BD
- [ ] Email de bienvenida se envía con plantilla HTML responsive
- [ ] Query `getMemberUserStatus` retorna estado correcto
- [ ] Todas las validaciones de negocio implementadas
- [ ] Tests unitarios y de integración pasando

### Frontend
- [ ] Checkbox "Crear usuario" visible en formulario de nuevo socio
- [ ] Diálogo de credenciales muestra username y contraseña con opción de copiar
- [ ] Botón "Crear usuario" visible en tabla solo para socios sin usuario
- [ ] Loading states apropiados durante operaciones asíncronas
- [ ] Mensajes de error claros cuando falla alguna operación
- [ ] Traducciones completas en español
- [ ] Tests de componentes pasando

### General
- [ ] Documentación actualizada
- [ ] Sin errores de TypeScript
- [ ] Sin errores de linting
- [ ] Sin errores de compilación
- [ ] Flujo end-to-end probado manualmente

---

## 📚 Referencias

- [Go crypto/rand](https://pkg.go.dev/crypto/rand)
- [GraphQL Best Practices](https://graphql.org/learn/best-practices/)
- [React Hook Form](https://react-hook-form.com/)
- [Material-UI Dialogs](https://mui.com/material-ui/react-dialog/)
- [Email Template Best Practices](https://www.campaignmonitor.com/blog/email-marketing/best-practices-for-email-design/)

---

## 🚀 Próximos Pasos

1. Revisar y aprobar este plan
2. Crear rama de feature: `feature/auto-create-user-for-member`
3. Implementar tareas en orden sugerido
4. Realizar code review
5. Testing manual exhaustivo
6. Merge a develop
7. Deploy a staging para pruebas finales
8. Deploy a producción

---

**Última actualización:** 2025-01-19  
**Estado:** 📋 Pendiente de aprobación e implementación
