# Documentación Frontend - ASAM Backend

Bienvenido a la documentación completa del backend de ASAM para desarrolladores frontend. Esta documentación está organizada para proporcionar toda la información necesaria para integrar exitosamente tu aplicación frontend con nuestro backend GraphQL.

## 🌐 Endpoints de la API

- **Producción**: `https://asam-backend-jtpswzdxuq-ew.a.run.app/graphql`
- **Desarrollo**: `http://localhost:8080/graphql`
- **GraphQL Playground**: `http://localhost:8080/playground` (solo desarrollo)

## 🎉 Novedades

### [Actualizaciones de Junio 2025](./updates-june-2025.md) 🆕
Resumen de todas las nuevas funcionalidades añadidas al backend:
- Sistema completo de gestión de usuarios
- Verificación de email
- Recuperación de contraseña
- Backend desplegado en producción
- Y mucho más...

## 📚 Documentación Disponible

### 1. [Referencia Rápida](./quick-reference.md)
Una guía concisa con la información esencial para comenzar rápidamente:
- Endpoints principales
- Autenticación básica
- Queries y mutations más comunes
- Códigos de error
- Configuración inicial

### 2. [Documentación Completa](../guia-frontend.md)
La documentación exhaustiva del sistema que incluye:
- Arquitectura del sistema
- Configuración detallada del entorno
- Schema GraphQL completo
- Sistema de autenticación y seguridad
- Rate limiting y límites
- Ejemplos de implementación
- Mejores prácticas

### 3. [Colección de Queries y Mutations](./graphql-queries.md)
Todas las queries y mutations del sistema listas para usar:
- Queries y mutations organizadas por dominio
- Gestión de usuarios y autenticación
- Verificación de email y recuperación de contraseña
- Ejemplos de variables
- Respuestas esperadas
- Casos de uso comunes

### 4. [Guía de Integración con React](./react-integration-guide.md)
Guía completa para integrar el backend con aplicaciones React:
- Configuración de Apollo Client
- Hooks personalizados
- Componentes reutilizables
- Gestión de estado
- Testing

### 5. [Guía de Integración con Vue 3](./vue-integration-guide.md)
Guía completa para integrar el backend con aplicaciones Vue 3:
- Configuración de Apollo Client para Vue
- Composables
- Componentes con Composition API
- Gestión de estado con Pinia
- Testing con Vitest

### 6. [Guía de Manejo de Errores](./error-handling-guide.md)
Estrategias completas para manejar errores:
- Tipos de errores del sistema
- Estructura de errores
- Estrategias de manejo
- Mensajes user-friendly
- Recuperación y retry
- Logging y monitoreo

## 🚀 Inicio Rápido

### 1. Configuración Básica

```javascript
// Instalar dependencias
npm install @apollo/client graphql

// Configurar Apollo Client
import { ApolloClient, InMemoryCache, createHttpLink } from '@apollo/client';
import { setContext } from '@apollo/client/link/context';

const httpLink = createHttpLink({
  uri: process.env.NODE_ENV === 'production' 
    ? 'https://asam-backend-jtpswzdxuq-ew.a.run.app/graphql'
    : 'http://localhost:8080/graphql'
});

const authLink = setContext((_, { headers }) => ({
  headers: {
    ...headers,
    authorization: localStorage.getItem('accessToken') 
      ? `Bearer ${localStorage.getItem('accessToken')}` 
      : "",
  }
}));

const client = new ApolloClient({
  link: authLink.concat(httpLink),
  cache: new InMemoryCache()
});
```

### 2. Primera Query

```javascript
import { gql, useQuery } from '@apollo/client';

const LIST_MEMBERS = gql`
  query ListMembers {
    listMembers {
      nodes {
        miembro_id
        numero_socio
        nombre
        apellidos
        estado
      }
      pageInfo {
        totalCount
      }
    }
  }
`;

function MembersList() {
  const { data, loading, error } = useQuery(LIST_MEMBERS);
  
  if (loading) return <p>Cargando...</p>;
  if (error) return <p>Error: {error.message}</p>;
  
  return (
    <ul>
      {data.listMembers.nodes.map(member => (
        <li key={member.miembro_id}>
          {member.nombre} {member.apellidos}
        </li>
      ))}
    </ul>
  );
}
```

## 🔑 Conceptos Clave

### Autenticación
- Sistema de doble token (access + refresh)
- Access token: 15 minutos
- Refresh token: 7 días
- Renovación automática
- Verificación de email
- Recuperación de contraseña

### Permisos
- **ADMIN**: Acceso completo (gestión de usuarios, miembros, pagos)
- **USER**: Solo lectura y cambio de su propia contraseña

### Rate Limiting
- 10 requests/segundo por IP
- Burst máximo: 20 requests

### Paginación
- Basada en cursor
- Tamaño de página configurable
- Información de página completa

## 🛠️ Herramientas Recomendadas

### Desarrollo
- [GraphQL Playground](http://localhost:8080/playground) - Para explorar la API
- [Apollo DevTools](https://www.apollographql.com/docs/react/development-testing/developer-tools/) - Para debugging
- [GraphQL Code Generator](https://graphql-code-generator.com/) - Para generar tipos TypeScript

### Testing
- [Mock Service Worker](https://mswjs.io/) - Para mockear requests
- [Apollo MockedProvider](https://www.apollographql.com/docs/react/development-testing/testing/) - Para testing de componentes

## 📊 Flujos de Trabajo Comunes

### Login Flow
1. Enviar credenciales → `login` mutation
2. Recibir tokens y datos de usuario
3. Guardar tokens de forma segura
4. Configurar Apollo Client con token
5. Verificar si el email está verificado
6. Redirigir según rol

### Gestión de Usuarios (Admin)
1. **Listar usuarios**: `listUsers` query
2. **Crear usuario**: `createUser` mutation
3. **Actualizar usuario**: `updateUser` mutation
4. **Resetear contraseña**: `resetUserPassword` mutation
5. **Eliminar usuario**: `deleteUser` mutation

### CRUD de Miembros
1. **Listar**: `listMembers` query con filtros
2. **Ver detalle**: `getMember` query
3. **Crear**: `createMember` mutation
4. **Actualizar**: `updateMember` mutation
5. **Cambiar estado**: `changeMemberStatus` mutation

### Gestión de Pagos
1. **Registrar pago**: `registerPayment` mutation
2. **Ver pagos**: `getMemberPayments` query
3. **Ver balance**: `getBalance` query
4. **Registrar cuotas masivas**: `registerFee` mutation

### Flujo de Verificación de Email
1. Usuario se registra o hace login
2. Verificar campo `emailVerified`
3. Si no está verificado, mostrar aviso
4. Enviar email → `sendVerificationEmail` mutation
5. Usuario hace click en link con token
6. Verificar → `verifyEmail` mutation

### Flujo de Recuperación de Contraseña
1. Usuario solicita recuperación → `requestPasswordReset` mutation
2. Usuario recibe email con token
3. Usuario ingresa nueva contraseña → `resetPasswordWithToken` mutation

## 🔍 Depuración

### Logs en Desarrollo
```javascript
// Habilitar logs de Apollo
const client = new ApolloClient({
  // ... configuración
  connectToDevTools: true,
  defaultOptions: {
    watchQuery: {
      errorPolicy: 'all',
    }
  }
});
```

### Errores Comunes
1. **Token expirado**: Implementar renovación automática
2. **CORS**: Verificar configuración del servidor
3. **Rate limit**: Implementar debounce y cache
4. **Permisos**: Verificar rol del usuario
5. **Email no verificado**: Mostrar botón para reenviar verificación

## 📞 Soporte

- **Issues**: [GitHub](https://github.com/javicabdev/asam-backend/issues)
- **Email**: soporte@asam.org
- **Documentación API**: [GraphQL Schema](./graphql-queries.md)

## 🎯 Próximos Pasos

1. Lee la [Referencia Rápida](./quick-reference.md) para comenzar
2. Revisa los ejemplos en tu framework ([React](./react-integration-guide.md) o [Vue](./vue-integration-guide.md))
3. Implementa el [manejo de errores](./error-handling-guide.md)
4. Explora el [schema completo](./graphql-queries.md)

---

*Esta documentación se mantiene actualizada con cada release del backend. Última actualización: Junio 2025*
