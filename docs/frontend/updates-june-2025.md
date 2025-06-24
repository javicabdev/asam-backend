# 🎉 Actualizaciones del Backend - Junio 2025

## 🌐 Backend en Producción

¡El backend de ASAM ya está desplegado en producción! 🚀

**Endpoint de Producción**: `https://asam-backend-jtpswzdxuq-ew.a.run.app/graphql`

### Configuración Rápida para Frontend

```javascript
// Apollo Client
const GRAPHQL_ENDPOINT = 'https://asam-backend-jtpswzdxuq-ew.a.run.app/graphql';

const client = new ApolloClient({
  uri: GRAPHQL_ENDPOINT,
  headers: {
    authorization: localStorage.getItem('token') || '',
  },
});
```

## 🆕 Nuevas Funcionalidades

### 1. **Gestión Completa de Usuarios** (Admin)

Se ha implementado un sistema completo de gestión de usuarios que incluye:

#### Queries de Usuario
- `getCurrentUser`: Obtener el usuario actual autenticado
- `getUser(id: ID!)`: Obtener un usuario específico (Admin)
- `listUsers(page: Int, pageSize: Int)`: Listar todos los usuarios con paginación (Admin)

#### Mutations de Usuario
- `createUser(input: CreateUserInput!)`: Crear nuevo usuario (Admin)
- `updateUser(input: UpdateUserInput!)`: Actualizar usuario existente (Admin)
- `deleteUser(id: ID!)`: Eliminar usuario (Admin)
- `changePassword(input: ChangePasswordInput!)`: Cambiar contraseña propia
- `resetUserPassword(userId: ID!, newPassword: String!)`: Resetear contraseña de cualquier usuario (Admin)

### 2. **Verificación de Email**

Nuevo sistema de verificación de email que incluye:

- Campo `emailVerified: Boolean!` en el tipo User
- Campo `emailVerifiedAt: Time` para registrar cuándo se verificó
- Mutations:
  - `sendVerificationEmail`: Enviar email de verificación al usuario actual
  - `verifyEmail(token: String!)`: Verificar email con token
  - `resendVerificationEmail(email: String!)`: Reenviar email de verificación

### 3. **Recuperación de Contraseña**

Sistema completo de recuperación de contraseña:

- `requestPasswordReset(email: String!)`: Solicitar recuperación de contraseña
- `resetPasswordWithToken(token: String!, newPassword: String!)`: Resetear contraseña con token

### 4. **Mejoras en el Sistema de Autenticación**

- Los tokens JWT ahora incluyen información sobre la verificación del email
- El tipo `AuthResponse` incluye todos los campos del usuario actualizado
- Mejor manejo de errores y mensajes más descriptivos

## 📊 Tipos Actualizados

### User Type
```graphql
type User {
  id: ID!
  username: String!
  role: UserRole!
  isActive: Boolean!
  lastLogin: Time
  emailVerified: Boolean!    # NUEVO
  emailVerifiedAt: Time      # NUEVO
}
```

### Input Types Nuevos
```graphql
input CreateUserInput {
  username: String!
  password: String!
  role: UserRole!
}

input UpdateUserInput {
  id: ID!
  username: String
  password: String
  role: UserRole
  isActive: Boolean
}

input ChangePasswordInput {
  currentPassword: String!
  newPassword: String!
}
```

## 🔒 Consideraciones de Seguridad

### Roles y Permisos
- **ADMIN**: Puede gestionar todos los usuarios, resetear contraseñas, ver toda la información
- **USER**: Solo puede cambiar su propia contraseña y ver información limitada

### Validaciones de Contraseña
Las contraseñas deben cumplir con:
- Mínimo 8 caracteres
- Al menos una mayúscula
- Al menos una minúscula
- Al menos un número
- Al menos un carácter especial

### Tokens de Email
- Los tokens de verificación y recuperación expiran en 24 horas
- Solo se puede usar una vez
- Se invalidan automáticamente al crear uno nuevo

## 💡 Mejores Prácticas para el Frontend

### 1. Flujo de Login Mejorado
```javascript
// Después del login, verificar si el email está verificado
const { data } = await login({ variables: { input } });

if (!data.login.user.emailVerified) {
  // Mostrar banner o modal para verificar email
  showEmailVerificationPrompt();
}
```

### 2. Interceptor para Renovación Automática de Token
```javascript
// Configurar Apollo Client con auto-refresh
import { TokenRefreshLink } from 'apollo-link-token-refresh';

const refreshLink = new TokenRefreshLink({
  isTokenValidOrUndefined: () => {
    const expiresAt = localStorage.getItem('expiresAt');
    if (!expiresAt) return true;
    
    return Date.now() < new Date(expiresAt).getTime();
  },
  fetchAccessToken: async () => {
    const refreshToken = localStorage.getItem('refreshToken');
    const { data } = await client.mutate({
      mutation: REFRESH_TOKEN_MUTATION,
      variables: { input: { refreshToken } }
    });
    
    return data.refreshToken;
  },
  handleFetch: (accessToken) => {
    localStorage.setItem('accessToken', accessToken);
  },
  handleError: (err) => {
    console.error('Token refresh error:', err);
    // Redirigir a login
    window.location.href = '/login';
  }
});
```

### 3. Gestión de Estados de Usuario
```javascript
// Hook para gestión de estado del usuario
export function useUserStatus() {
  const { data } = useQuery(GET_CURRENT_USER);
  
  const userStatus = {
    isLoggedIn: !!data?.getCurrentUser,
    isAdmin: data?.getCurrentUser?.role === 'admin',
    isEmailVerified: data?.getCurrentUser?.emailVerified,
    needsPasswordChange: checkPasswordAge(data?.getCurrentUser?.lastLogin),
    canManageUsers: data?.getCurrentUser?.role === 'admin',
    canManagePayments: data?.getCurrentUser?.role === 'admin'
  };
  
  return userStatus;
}
```

## 🚨 Breaking Changes

### 1. Campos Obligatorios en User
- `emailVerified` es ahora obligatorio (Boolean!)
- Los usuarios creados antes de esta actualización tienen `emailVerified: false` por defecto

### 2. Respuesta de Login
La respuesta de login ahora incluye más información del usuario:
```javascript
// Antes
{ id, username, role }

// Ahora
{ id, username, role, isActive, lastLogin, emailVerified, emailVerifiedAt }
```

## 📝 Ejemplos de Implementación

### Componente de Gestión de Usuarios (Admin)
```javascript
function UserManagement() {
  const { data, loading } = useQuery(LIST_USERS, {
    variables: { page: 1, pageSize: 20 }
  });
  
  const [createUser] = useMutation(CREATE_USER);
  const [updateUser] = useMutation(UPDATE_USER);
  const [deleteUser] = useMutation(DELETE_USER);
  const [resetPassword] = useMutation(RESET_USER_PASSWORD);
  
  const handleCreateUser = async (input) => {
    try {
      await createUser({ 
        variables: { input },
        refetchQueries: [{ query: LIST_USERS }]
      });
      showSuccess('Usuario creado exitosamente');
    } catch (error) {
      showError(`Error al crear usuario: ${error.message}`);
    }
  };
  
  const handleResetPassword = async (userId) => {
    const newPassword = generateSecurePassword();
    try {
      await resetPassword({ 
        variables: { userId, newPassword }
      });
      showSuccess(`Nueva contraseña: ${newPassword}`);
    } catch (error) {
      showError(`Error al resetear contraseña: ${error.message}`);
    }
  };
  
  // ... resto del componente
}
```

### Flujo de Verificación de Email
```javascript
function EmailVerification() {
  const [sendVerification] = useMutation(SEND_VERIFICATION_EMAIL);
  const [verifyEmail] = useMutation(VERIFY_EMAIL);
  const { token } = useParams(); // Si viene de un link
  
  useEffect(() => {
    if (token) {
      handleVerification(token);
    }
  }, [token]);
  
  const handleSendVerification = async () => {
    try {
      await sendVerification();
      showInfo('Email de verificación enviado');
    } catch (error) {
      showError('Error al enviar email');
    }
  };
  
  const handleVerification = async (token) => {
    try {
      await verifyEmail({ variables: { token } });
      showSuccess('Email verificado exitosamente');
      navigate('/dashboard');
    } catch (error) {
      showError('Token inválido o expirado');
    }
  };
  
  // ... resto del componente
}
```

## 🎯 Próximos Pasos Recomendados

1. **Actualizar los tipos de TypeScript** si estás usando TypeScript
2. **Implementar el flujo de verificación de email** en el proceso de registro/login
3. **Añadir la gestión de usuarios** al panel de administración
4. **Actualizar los interceptores** para manejar los nuevos campos
5. **Implementar notificaciones** para estados del usuario (email no verificado, etc.)

## 📞 Soporte

Si tienes alguna pregunta sobre estas actualizaciones:
- **Email**: soporte@asam.org
- **Documentación completa**: [Ver toda la documentación](./README.md)
- **Colección de Queries**: [GraphQL Queries](./graphql-queries.md)

---

*Última actualización: Junio 2025*
*Backend versión: 1.0.0*
*Desplegado en: Google Cloud Run*
