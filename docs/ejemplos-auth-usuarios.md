# Ejemplos de Uso: Sistema de Autenticación y Gestión de Usuarios

Este documento proporciona ejemplos prácticos de cómo utilizar las nuevas funcionalidades de autenticación mejorada y gestión de usuarios.

## 1. Autorización por Roles

### Uso en Resolvers

```go
// En cualquier resolver que requiera permisos de administrador
func (r *Resolver) SomeAdminOnlyOperation(ctx context.Context) (*SomeType, error) {
    // Verificar permisos de administrador
    if err := middleware.MustBeAdmin(ctx); err != nil {
        return nil, err
    }
    
    // Continuar con la operación...
}

// Para operaciones que requieran solo autenticación
func (r *Resolver) SomeAuthenticatedOperation(ctx context.Context) (*SomeType, error) {
    // Verificar que el usuario esté autenticado
    user, err := middleware.GetUserFromContext(ctx)
    if err != nil {
        return nil, err
    }
    
    // Usar la información del usuario...
    log.Printf("Operation requested by user: %s", user.Username)
}
```

## 2. Rate Limiting en Login

El rate limiting se aplica automáticamente en el login. Aquí está cómo funciona:

### Comportamiento del Rate Limiter

- **Límite**: 5 intentos fallidos en 5 minutos
- **Bloqueo**: 15 minutos después de exceder el límite
- **Protección contra fuerza bruta**: 1 intento por segundo máximo

### Ejemplo de Respuesta cuando se Excede el Límite

```json
{
  "errors": [
    {
      "message": "Demasiados intentos de inicio de sesión. Cuenta bloqueada por 15m0s",
      "extensions": {
        "code": "UNAUTHORIZED"
      }
    }
  ],
  "data": null
}
```

## 3. Gestión de Usuarios

### Crear un Nuevo Usuario (Solo Admin)

```graphql
mutation CreateUser {
  createUser(input: {
    username: "nuevo_usuario"
    password: "ContraseñaSegura123!"
    role: user
  }) {
    id
    username
    role
    isActive
  }
}
```

### Listar Usuarios (Solo Admin)

```graphql
query ListUsers {
  listUsers(page: 1, pageSize: 20) {
    id
    username
    role
    isActive
    lastLogin
  }
}
```

### Obtener Usuario Actual

```graphql
query GetCurrentUser {
  getCurrentUser {
    id
    username
    role
    isActive
    lastLogin
  }
}
```

### Actualizar Usuario (Solo Admin)

```graphql
mutation UpdateUser {
  updateUser(input: {
    id: "123"
    username: "nuevo_nombre"
    role: admin
    isActive: true
  }) {
    id
    username
    role
    isActive
  }
}
```

### Cambiar Contraseña (Usuario Actual)

```graphql
mutation ChangePassword {
  changePassword(input: {
    currentPassword: "contraseñaActual"
    newPassword: "NuevaContraseñaSegura123!"
  }) {
    success
    message
    error
  }
}
```

### Resetear Contraseña de Usuario (Solo Admin)

```graphql
mutation ResetUserPassword {
  resetUserPassword(
    userId: "123"
    newPassword: "NuevaContraseñaTemporal123!"
  ) {
    success
    message
    error
  }
}
```

### Eliminar Usuario Permanentemente (Solo Admin)

**⚠️ IMPORTANTE**: Esta operación elimina el usuario de forma permanente de la base de datos. No se puede deshacer.

**Restricciones**:
- No se puede eliminar un usuario que tiene un socio (Member) asociado
- No se puede eliminar el último administrador del sistema
- Se eliminarán automáticamente todos los tokens asociados (RefreshTokens, VerificationTokens)

```graphql
mutation DeleteUser {
  deleteUser(id: "123") {
    success
    message
    error
  }
}
```

**Posibles errores**:
- `"Cannot delete user with associated member. Please remove member association first"` - El usuario tiene un socio asociado
- `"Cannot delete the last admin user"` - Intento de eliminar el último admin

## 4. Integración en Frontend

### Ejemplo con React y Apollo Client

```jsx
import { gql, useMutation, useQuery } from '@apollo/client';
import { useState } from 'react';

// Queries
const GET_CURRENT_USER = gql`
  query GetCurrentUser {
    getCurrentUser {
      id
      username
      role
      isActive
    }
  }
`;

const LIST_USERS = gql`
  query ListUsers($page: Int, $pageSize: Int) {
    listUsers(page: $page, pageSize: $pageSize) {
      id
      username
      role
      isActive
      lastLogin
    }
  }
`;

// Mutations
const CREATE_USER = gql`
  mutation CreateUser($input: CreateUserInput!) {
    createUser(input: $input) {
      id
      username
      role
    }
  }
`;

const CHANGE_PASSWORD = gql`
  mutation ChangePassword($input: ChangePasswordInput!) {
    changePassword(input: $input) {
      success
      message
      error
    }
  }
`;

// Component Examples
function UserProfile() {
  const { data, loading, error } = useQuery(GET_CURRENT_USER);
  
  if (loading) return <p>Cargando...</p>;
  if (error) return <p>Error: {error.message}</p>;
  
  return (
    <div>
      <h2>Mi Perfil</h2>
      <p>Usuario: {data.getCurrentUser.username}</p>
      <p>Rol: {data.getCurrentUser.role}</p>
      <ChangePasswordForm />
    </div>
  );
}

function ChangePasswordForm() {
  const [currentPassword, setCurrentPassword] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [changePassword, { loading, error }] = useMutation(CHANGE_PASSWORD);
  
  const handleSubmit = async (e) => {
    e.preventDefault();
    
    try {
      const { data } = await changePassword({
        variables: {
          input: {
            currentPassword,
            newPassword
          }
        }
      });
      
      if (data.changePassword.success) {
        alert('Contraseña cambiada exitosamente');
        setCurrentPassword('');
        setNewPassword('');
      } else {
        alert(`Error: ${data.changePassword.error}`);
      }
    } catch (err) {
      console.error('Error changing password:', err);
    }
  };
  
  return (
    <form onSubmit={handleSubmit}>
      <h3>Cambiar Contraseña</h3>
      {error && <p className="error">{error.message}</p>}
      
      <input
        type="password"
        placeholder="Contraseña actual"
        value={currentPassword}
        onChange={(e) => setCurrentPassword(e.target.value)}
        required
      />
      
      <input
        type="password"
        placeholder="Nueva contraseña"
        value={newPassword}
        onChange={(e) => setNewPassword(e.target.value)}
        required
        pattern="^(?=.*[a-z])(?=.*[A-Z])(?=.*\d).{8,}$"
        title="Mínimo 8 caracteres, una mayúscula, una minúscula y un número"
      />
      
      <button type="submit" disabled={loading}>
        {loading ? 'Cambiando...' : 'Cambiar Contraseña'}
      </button>
    </form>
  );
}

function AdminUserList() {
  const { data, loading, error } = useQuery(LIST_USERS, {
    variables: { page: 1, pageSize: 20 }
  });
  
  const [createUser] = useMutation(CREATE_USER);
  
  if (loading) return <p>Cargando usuarios...</p>;
  if (error) {
    // Si el error es de permisos, mostrar mensaje apropiado
    if (error.message.includes('Insufficient permissions')) {
      return <p>No tienes permisos para ver esta página.</p>;
    }
    return <p>Error: {error.message}</p>;
  }
  
  return (
    <div>
      <h2>Gestión de Usuarios</h2>
      <CreateUserForm onUserCreated={() => refetch()} />
      
      <table>
        <thead>
          <tr>
            <th>ID</th>
            <th>Usuario</th>
            <th>Rol</th>
            <th>Estado</th>
            <th>Último Login</th>
            <th>Acciones</th>
          </tr>
        </thead>
        <tbody>
          {data.listUsers.map(user => (
            <tr key={user.id}>
              <td>{user.id}</td>
              <td>{user.username}</td>
              <td>{user.role}</td>
              <td>{user.isActive ? 'Activo' : 'Inactivo'}</td>
              <td>{user.lastLogin ? new Date(user.lastLogin).toLocaleString() : 'Nunca'}</td>
              <td>
                <button onClick={() => handleEdit(user)}>Editar</button>
                <button onClick={() => handleDelete(user.id)}>Eliminar</button>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
```

### Manejo de Errores de Rate Limiting

```jsx
function LoginForm() {
  const [login] = useMutation(LOGIN_MUTATION);
  const [error, setError] = useState('');
  const [isBlocked, setIsBlocked] = useState(false);
  const [blockTimeRemaining, setBlockTimeRemaining] = useState(0);
  
  const handleLogin = async (username, password) => {
    try {
      const { data } = await login({
        variables: { input: { username, password } }
      });
      
      // Login exitoso
      localStorage.setItem('accessToken', data.login.accessToken);
      localStorage.setItem('refreshToken', data.login.refreshToken);
      
    } catch (err) {
      // Verificar si es un error de rate limiting
      if (err.message.includes('Demasiados intentos')) {
        setIsBlocked(true);
        
        // Extraer tiempo de bloqueo del mensaje
        const match = err.message.match(/bloqueada por (\d+)m/);
        if (match) {
          const minutes = parseInt(match[1]);
          setBlockTimeRemaining(minutes * 60); // convertir a segundos
          
          // Iniciar countdown
          const interval = setInterval(() => {
            setBlockTimeRemaining(prev => {
              if (prev <= 1) {
                clearInterval(interval);
                setIsBlocked(false);
                return 0;
              }
              return prev - 1;
            });
          }, 1000);
        }
      } else {
        setError(err.message);
      }
    }
  };
  
  if (isBlocked) {
    const minutes = Math.floor(blockTimeRemaining / 60);
    const seconds = blockTimeRemaining % 60;
    
    return (
      <div className="error-message">
        <p>Demasiados intentos de inicio de sesión.</p>
        <p>Por favor, espera {minutes}:{seconds.toString().padStart(2, '0')} antes de intentar nuevamente.</p>
      </div>
    );
  }
  
  // ... resto del formulario de login
}
```

## 5. Validación de Contraseñas

Las contraseñas deben cumplir con los siguientes requisitos:

- Mínimo 8 caracteres
- Máximo 100 caracteres
- Al menos una letra mayúscula
- Al menos una letra minúscula
- Al menos un número
- Opcionalmente caracteres especiales (!@#$%^&*()_+-=[]{}|;:,.<>?)

### Ejemplo de Validación en Frontend

```javascript
function validatePassword(password) {
  const errors = [];
  
  if (password.length < 8) {
    errors.push('La contraseña debe tener al menos 8 caracteres');
  }
  
  if (password.length > 100) {
    errors.push('La contraseña no puede exceder 100 caracteres');
  }
  
  if (!/[A-Z]/.test(password)) {
    errors.push('La contraseña debe contener al menos una letra mayúscula');
  }
  
  if (!/[a-z]/.test(password)) {
    errors.push('La contraseña debe contener al menos una letra minúscula');
  }
  
  if (!/[0-9]/.test(password)) {
    errors.push('La contraseña debe contener al menos un número');
  }
  
  return errors;
}
```

## 6. Validación de Nombres de Usuario

Los nombres de usuario deben cumplir con:

- Mínimo 3 caracteres
- Máximo 50 caracteres
- Solo letras, números, guión bajo (_), guión (-) y punto (.)

### Ejemplo de Validación

```javascript
function validateUsername(username) {
  const errors = [];
  
  if (username.length < 3) {
    errors.push('El nombre de usuario debe tener al menos 3 caracteres');
  }
  
  if (username.length > 50) {
    errors.push('El nombre de usuario no puede exceder 50 caracteres');
  }
  
  if (!/^[a-zA-Z0-9._-]+$/.test(username)) {
    errors.push('El nombre de usuario solo puede contener letras, números, punto, guión y guión bajo');
  }
  
  return errors;
}
```

## Notas de Seguridad

1. **Rate Limiting**: El sistema bloquea automáticamente las cuentas después de 5 intentos fallidos en 5 minutos.

2. **Autorización**: Todas las operaciones de gestión de usuarios requieren rol de administrador, excepto cambiar la propia contraseña.

3. **Eliminación de Usuarios**: La eliminación es "soft delete" - el usuario se marca como inactivo pero no se elimina de la base de datos.

4. **Auto-eliminación**: Los usuarios no pueden eliminar su propia cuenta para evitar que un administrador se quede sin acceso al sistema.

5. **Sesiones**: El sistema mantiene un registro de tokens de refresco activos y puede limitar el número de sesiones concurrentes por usuario.
