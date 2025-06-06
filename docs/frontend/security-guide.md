# Guía de Seguridad para Frontend

Esta guía proporciona las mejores prácticas de seguridad para aplicaciones frontend que consumen el backend de ASAM.

## Tabla de Contenidos
1. [Autenticación y Autorización](#autenticación-y-autorización)
2. [Almacenamiento Seguro](#almacenamiento-seguro)
3. [Protección contra XSS](#protección-contra-xss)
4. [Protección contra CSRF](#protección-contra-csrf)
5. [Comunicación Segura](#comunicación-segura)
6. [Validación de Datos](#validación-de-datos)
7. [Manejo de Errores Seguro](#manejo-de-errores-seguro)
8. [Auditoría y Logging](#auditoría-y-logging)

## Autenticación y Autorización

### 1. Gestión Segura de Tokens

```javascript
// services/tokenService.js
class TokenService {
  constructor() {
    this.ACCESS_TOKEN_KEY = 'asam_access_token';
    this.REFRESH_TOKEN_KEY = 'asam_refresh_token';
    this.TOKEN_EXPIRY_KEY = 'asam_token_expiry';
  }

  // Guardar tokens de forma segura
  setTokens(accessToken, refreshToken, expiresAt) {
    // Para mayor seguridad, considera usar sessionStorage
    // o mejor aún, mantenerlos solo en memoria
    sessionStorage.setItem(this.ACCESS_TOKEN_KEY, accessToken);
    
    // Refresh token en httpOnly cookie sería más seguro
    // pero si debe estar en el cliente:
    sessionStorage.setItem(this.REFRESH_TOKEN_KEY, refreshToken);
    sessionStorage.setItem(this.TOKEN_EXPIRY_KEY, expiresAt);
  }

  getAccessToken() {
    const token = sessionStorage.getItem(this.ACCESS_TOKEN_KEY);
    const expiry = sessionStorage.getItem(this.TOKEN_EXPIRY_KEY);
    
    // Verificar expiración
    if (token && expiry && new Date(expiry) > new Date()) {
      return token;
    }
    
    this.clearTokens();
    return null;
  }

  clearTokens() {
    sessionStorage.removeItem(this.ACCESS_TOKEN_KEY);
    sessionStorage.removeItem(this.REFRESH_TOKEN_KEY);
    sessionStorage.removeItem(this.TOKEN_EXPIRY_KEY);
  }

  // Verificar si el token está próximo a expirar
  isTokenExpiringSoon() {
    const expiry = sessionStorage.getItem(this.TOKEN_EXPIRY_KEY);
    if (!expiry) return true;
    
    const expiryDate = new Date(expiry);
    const now = new Date();
    const timeUntilExpiry = expiryDate - now;
    
    // Renovar si quedan menos de 5 minutos
    return timeUntilExpiry < 5 * 60 * 1000;
  }
}

export const tokenService = new TokenService();
```

### 2. Implementación de Auto-Refresh

```javascript
// services/authService.js
import { tokenService } from './tokenService';

class AuthService {
  constructor() {
    this.refreshTimer = null;
  }

  async login(username, password) {
    try {
      const response = await apolloClient.mutate({
        mutation: LOGIN_MUTATION,
        variables: { input: { username, password } }
      });

      const { accessToken, refreshToken, expiresAt, user } = response.data.login;
      
      // Guardar tokens
      tokenService.setTokens(accessToken, refreshToken, expiresAt);
      
      // Guardar info del usuario (sin datos sensibles)
      this.setCurrentUser(user);
      
      // Iniciar auto-refresh
      this.scheduleTokenRefresh();
      
      return { success: true, user };
    } catch (error) {
      return { success: false, error: error.message };
    }
  }

  scheduleTokenRefresh() {
    // Cancelar timer anterior
    if (this.refreshTimer) {
      clearTimeout(this.refreshTimer);
    }

    // Verificar cada minuto
    this.refreshTimer = setInterval(async () => {
      if (tokenService.isTokenExpiringSoon()) {
        await this.refreshToken();
      }
    }, 60000); // 1 minuto
  }

  async refreshToken() {
    const refreshToken = tokenService.getRefreshToken();
    if (!refreshToken) {
      this.logout();
      return;
    }

    try {
      const response = await apolloClient.mutate({
        mutation: REFRESH_TOKEN_MUTATION,
        variables: { input: { refreshToken } }
      });

      const { accessToken, refreshToken: newRefreshToken, expiresAt } = 
        response.data.refreshToken;
      
      tokenService.setTokens(accessToken, newRefreshToken, expiresAt);
    } catch (error) {
      // Si falla el refresh, hacer logout
      this.logout();
    }
  }

  logout() {
    // Limpiar timer
    if (this.refreshTimer) {
      clearInterval(this.refreshTimer);
      this.refreshTimer = null;
    }

    // Limpiar tokens
    tokenService.clearTokens();
    
    // Limpiar usuario
    this.clearCurrentUser();
    
    // Limpiar cache de Apollo
    apolloClient.clearStore();
    
    // Redirigir a login
    window.location.href = '/login';
  }

  setCurrentUser(user) {
    // No guardar información sensible
    const safeUser = {
      id: user.id,
      username: user.username,
      role: user.role
    };
    
    localStorage.setItem('current_user', JSON.stringify(safeUser));
  }

  getCurrentUser() {
    const userStr = localStorage.getItem('current_user');
    return userStr ? JSON.parse(userStr) : null;
  }

  clearCurrentUser() {
    localStorage.removeItem('current_user');
  }

  hasPermission(requiredRole) {
    const user = this.getCurrentUser();
    if (!user) return false;
    
    // Jerarquía de roles
    const roleHierarchy = {
      'ADMIN': 2,
      'USER': 1
    };
    
    return roleHierarchy[user.role] >= roleHierarchy[requiredRole];
  }
}

export const authService = new AuthService();
```

### 3. Protección de Rutas

```javascript
// components/ProtectedRoute.jsx
import { Navigate } from 'react-router-dom';
import { authService } from '../services/authService';

export function ProtectedRoute({ children, requiredRole = 'USER' }) {
  const user = authService.getCurrentUser();
  
  if (!user) {
    // No autenticado
    return <Navigate to="/login" replace />;
  }
  
  if (!authService.hasPermission(requiredRole)) {
    // Sin permisos suficientes
    return <Navigate to="/unauthorized" replace />;
  }
  
  return children;
}

// Uso
<Route 
  path="/admin" 
  element={
    <ProtectedRoute requiredRole="ADMIN">
      <AdminPanel />
    </ProtectedRoute>
  } 
/>
```

## Almacenamiento Seguro

### 1. Estrategias de Almacenamiento

```javascript
// services/secureStorage.js
class SecureStorage {
  constructor() {
    this.encryptionKey = this.generateKey();
  }

  generateKey() {
    // En producción, usar una clave derivada del servidor
    return 'temp-key-' + Date.now();
  }

  // Para datos sensibles temporales
  setSecureSession(key, value) {
    const encrypted = this.encrypt(JSON.stringify(value));
    sessionStorage.setItem(key, encrypted);
  }

  getSecureSession(key) {
    const encrypted = sessionStorage.getItem(key);
    if (!encrypted) return null;
    
    try {
      const decrypted = this.decrypt(encrypted);
      return JSON.parse(decrypted);
    } catch {
      return null;
    }
  }

  // Encriptación simple (usar una librería robusta en producción)
  encrypt(text) {
    // En producción usar CryptoJS o similar
    return btoa(text);
  }

  decrypt(encrypted) {
    return atob(encrypted);
  }

  // Limpiar datos sensibles
  clearSensitiveData() {
    // Limpiar todos los datos de sesión
    sessionStorage.clear();
    
    // Limpiar datos específicos de localStorage
    const sensitiveKeys = ['current_user', 'temp_data'];
    sensitiveKeys.forEach(key => localStorage.removeItem(key));
  }

  // Auto-limpieza por inactividad
  setupAutoCleanup(timeoutMinutes = 30) {
    let timer;
    
    const resetTimer = () => {
      clearTimeout(timer);
      timer = setTimeout(() => {
        this.clearSensitiveData();
        authService.logout();
      }, timeoutMinutes * 60 * 1000);
    };

    // Eventos de actividad
    ['mousedown', 'keypress', 'scroll', 'touchstart'].forEach(event => {
      document.addEventListener(event, resetTimer, true);
    });

    resetTimer();
  }
}

export const secureStorage = new SecureStorage();
```

### 2. Sanitización de Datos

```javascript
// utils/sanitizer.js
import DOMPurify from 'dompurify';

export const sanitizer = {
  // Sanitizar HTML
  html(dirty) {
    return DOMPurify.sanitize(dirty, {
      ALLOWED_TAGS: ['b', 'i', 'em', 'strong', 'p', 'br'],
      ALLOWED_ATTR: []
    });
  },

  // Sanitizar para atributos
  attribute(value) {
    return value.replace(/['"<>]/g, '');
  },

  // Sanitizar URLs
  url(url) {
    try {
      const parsed = new URL(url);
      // Solo permitir http/https
      if (!['http:', 'https:'].includes(parsed.protocol)) {
        return '';
      }
      return parsed.href;
    } catch {
      return '';
    }
  },

  // Sanitizar entrada de usuario
  userInput(input) {
    return input
      .trim()
      .replace(/[<>]/g, '') // Remover tags
      .slice(0, 1000); // Limitar longitud
  }
};
```

## Protección contra XSS

### 1. Renderizado Seguro

```javascript
// components/SafeRender.jsx
import DOMPurify from 'dompurify';

export function SafeHTML({ html, className }) {
  const sanitized = DOMPurify.sanitize(html);
  
  return (
    <div 
      className={className}
      dangerouslySetInnerHTML={{ __html: sanitized }}
    />
  );
}

// Componente para texto de usuario
export function UserContent({ content }) {
  // Escapar caracteres especiales
  const escaped = content
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#x27;');
  
  return <div className="user-content">{escaped}</div>;
}
```

### 2. Validación de Props

```javascript
// components/MemberCard.jsx
import PropTypes from 'prop-types';
import { sanitizer } from '../utils/sanitizer';

export function MemberCard({ member }) {
  // Sanitizar datos antes de renderizar
  const safeMember = {
    nombre: sanitizer.userInput(member.nombre || ''),
    apellidos: sanitizer.userInput(member.apellidos || ''),
    email: sanitizer.attribute(member.correo_electronico || ''),
    observaciones: sanitizer.html(member.observaciones || '')
  };

  return (
    <div className="member-card">
      <h3>{safeMember.nombre} {safeMember.apellidos}</h3>
      <p>Email: {safeMember.email}</p>
      <SafeHTML html={safeMember.observaciones} />
    </div>
  );
}

MemberCard.propTypes = {
  member: PropTypes.shape({
    nombre: PropTypes.string.isRequired,
    apellidos: PropTypes.string.isRequired,
    correo_electronico: PropTypes.string,
    observaciones: PropTypes.string
  }).isRequired
};
```

## Protección contra CSRF

### 1. Implementación de CSRF Token

```javascript
// services/csrfService.js
class CSRFService {
  constructor() {
    this.tokenKey = 'csrf_token';
  }

  async fetchToken() {
    try {
      const response = await fetch('/api/csrf-token', {
        credentials: 'include'
      });
      const { token } = await response.json();
      this.setToken(token);
      return token;
    } catch (error) {
      console.error('Failed to fetch CSRF token:', error);
      return null;
    }
  }

  setToken(token) {
    sessionStorage.setItem(this.tokenKey, token);
  }

  getToken() {
    return sessionStorage.getItem(this.tokenKey);
  }

  // Añadir token a headers
  getHeaders() {
    const token = this.getToken();
    return token ? { 'X-CSRF-Token': token } : {};
  }
}

export const csrfService = new CSRFService();

// Integración con Apollo Client
const csrfLink = new ApolloLink((operation, forward) => {
  operation.setContext({
    headers: {
      ...csrfService.getHeaders()
    }
  });
  
  return forward(operation);
});
```

## Comunicación Segura

### 1. Configuración HTTPS

```javascript
// utils/secureRedirect.js
export function enforceHTTPS() {
  if (process.env.NODE_ENV === 'production' && 
      window.location.protocol !== 'https:') {
    window.location.href = 
      'https:' + window.location.href.substring(window.location.protocol.length);
  }
}

// En index.js
import { enforceHTTPS } from './utils/secureRedirect';
enforceHTTPS();
```

### 2. Headers de Seguridad

```javascript
// server/securityHeaders.js (para Express)
export const securityHeaders = {
  'Content-Security-Policy': [
    "default-src 'self'",
    "script-src 'self' 'unsafe-inline' https://trusted-cdn.com",
    "style-src 'self' 'unsafe-inline'",
    "img-src 'self' data: https:",
    "font-src 'self'",
    "connect-src 'self' https://api.asam.org",
    "frame-ancestors 'none'",
    "base-uri 'self'",
    "form-action 'self'"
  ].join('; '),
  'X-Frame-Options': 'DENY',
  'X-Content-Type-Options': 'nosniff',
  'X-XSS-Protection': '1; mode=block',
  'Referrer-Policy': 'strict-origin-when-cross-origin',
  'Permissions-Policy': 'geolocation=(), microphone=(), camera=()'
};
```

## Validación de Datos

### 1. Esquemas de Validación

```javascript
// validation/memberSchemas.js
import * as yup from 'yup';

export const memberSchema = yup.object({
  numero_socio: yup
    .string()
    .required('Número de socio es requerido')
    .matches(/^\d{4}-\d{3}$/, 'Formato inválido (YYYY-NNN)'),
  
  nombre: yup
    .string()
    .required('Nombre es requerido')
    .min(2, 'Mínimo 2 caracteres')
    .max(50, 'Máximo 50 caracteres')
    .matches(/^[a-zA-ZáéíóúÁÉÍÓÚñÑ\s]+$/, 'Solo letras permitidas'),
  
  correo_electronico: yup
    .string()
    .email('Email inválido')
    .max(100, 'Máximo 100 caracteres'),
  
  documento_identidad: yup
    .string()
    .matches(/^[0-9]{8}[A-Z]$/, 'Formato DNI inválido'),
  
  codigo_postal: yup
    .string()
    .matches(/^\d{5}$/, 'Código postal debe tener 5 dígitos')
});

// Hook de validación
export function useValidation(schema) {
  const [errors, setErrors] = useState({});
  
  const validate = async (data) => {
    try {
      await schema.validate(data, { abortEarly: false });
      setErrors({});
      return true;
    } catch (err) {
      const validationErrors = {};
      err.inner.forEach(error => {
        validationErrors[error.path] = error.message;
      });
      setErrors(validationErrors);
      return false;
    }
  };
  
  return { errors, validate };
}
```

### 2. Sanitización de Entrada

```javascript
// utils/inputSanitizer.js
export const inputSanitizer = {
  // Limpiar espacios y caracteres invisibles
  text(value) {
    return value
      .trim()
      .replace(/\s+/g, ' ') // Múltiples espacios a uno
      .replace(/[\u0000-\u001F\u007F-\u009F]/g, ''); // Caracteres de control
  },

  // Sanitizar números
  number(value) {
    const cleaned = value.replace(/[^\d.-]/g, '');
    const parsed = parseFloat(cleaned);
    return isNaN(parsed) ? 0 : parsed;
  },

  // Sanitizar teléfonos
  phone(value) {
    return value.replace(/[^\d+\s()-]/g, '');
  },

  // Sanitizar para SQL (aunque uses prepared statements)
  sql(value) {
    return value
      .replace(/['";\\]/g, '') // Remover caracteres peligrosos
      .slice(0, 255); // Limitar longitud
  }
};
```

## Manejo de Errores Seguro

### 1. Filtrado de Información Sensible

```javascript
// utils/errorHandler.js
export class SecureErrorHandler {
  handle(error) {
    // No exponer stack traces en producción
    if (process.env.NODE_ENV === 'production') {
      // Mapear errores técnicos a mensajes genéricos
      const errorMap = {
        'NetworkError': 'Error de conexión. Por favor, intenta nuevamente.',
        'GraphQLError': 'Error al procesar la solicitud.',
        'TypeError': 'Ha ocurrido un error inesperado.',
        'ReferenceError': 'Ha ocurrido un error inesperado.'
      };
      
      const genericMessage = errorMap[error.constructor.name] || 
                           'Ha ocurrido un error inesperado.';
      
      // Log completo para debugging (no visible al usuario)
      this.logError(error);
      
      return {
        message: genericMessage,
        code: 'GENERIC_ERROR'
      };
    }
    
    // En desarrollo, mostrar más detalles
    return {
      message: error.message,
      code: error.code || 'UNKNOWN_ERROR',
      details: error.stack
    };
  }

  logError(error) {
    // Enviar a servicio de logging
    if (window.Sentry) {
      window.Sentry.captureException(error);
    }
    
    // Log local (sin exponer a consola en producción)
    if (process.env.NODE_ENV !== 'production') {
      console.error('Error:', error);
    }
  }
}
```

## Auditoría y Logging

### 1. Logging de Actividades Sensibles

```javascript
// services/auditService.js
class AuditService {
  constructor() {
    this.queue = [];
    this.flushInterval = 5000; // 5 segundos
    this.startFlushing();
  }

  log(action, details = {}) {
    const entry = {
      action,
      timestamp: new Date().toISOString(),
      user: authService.getCurrentUser()?.id,
      details: this.sanitizeDetails(details),
      sessionId: this.getSessionId(),
      ip: 'client-side' // El servidor debe añadir la IP real
    };
    
    this.queue.push(entry);
    
    // Flush inmediato para acciones críticas
    if (this.isCriticalAction(action)) {
      this.flush();
    }
  }

  sanitizeDetails(details) {
    // Remover información sensible
    const sanitized = { ...details };
    delete sanitized.password;
    delete sanitized.token;
    delete sanitized.refreshToken;
    return sanitized;
  }

  isCriticalAction(action) {
    const criticalActions = [
      'LOGIN',
      'LOGOUT',
      'DELETE_MEMBER',
      'CHANGE_PERMISSIONS',
      'EXPORT_DATA'
    ];
    return criticalActions.includes(action);
  }

  async flush() {
    if (this.queue.length === 0) return;
    
    const entries = [...this.queue];
    this.queue = [];
    
    try {
      await fetch('/api/audit', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          ...authService.getAuthHeaders()
        },
        body: JSON.stringify({ entries })
      });
    } catch (error) {
      // Re-añadir a la cola si falla
      this.queue.unshift(...entries);
    }
  }

  startFlushing() {
    setInterval(() => this.flush(), this.flushInterval);
  }

  getSessionId() {
    let sessionId = sessionStorage.getItem('session_id');
    if (!sessionId) {
      sessionId = this.generateSessionId();
      sessionStorage.setItem('session_id', sessionId);
    }
    return sessionId;
  }

  generateSessionId() {
    return 'session_' + Date.now() + '_' + Math.random().toString(36);
  }
}

export const auditService = new AuditService();

// Uso
auditService.log('VIEW_MEMBER', { memberId: '123' });
auditService.log('UPDATE_MEMBER', { memberId: '123', fields: ['email'] });
auditService.log('DELETE_MEMBER', { memberId: '123' });
```

## Checklist de Seguridad

### Autenticación
- [ ] Tokens almacenados de forma segura (sessionStorage o memoria)
- [ ] Auto-refresh de tokens implementado
- [ ] Logout limpia todos los datos sensibles
- [ ] Protección de rutas según roles
- [ ] Timeout por inactividad

### Datos y Comunicación
- [ ] HTTPS forzado en producción
- [ ] Headers de seguridad configurados
- [ ] CSRF protection implementado
- [ ] Sanitización de entrada de usuario
- [ ] Validación en cliente y servidor

### XSS y Renderizado
- [ ] DOMPurify para contenido HTML
- [ ] Props validation en componentes
- [ ] No usar dangerouslySetInnerHTML sin sanitizar
- [ ] Escapar caracteres especiales en texto

### Errores y Logging
- [ ] No exponer stack traces en producción
- [ ] Mensajes de error genéricos para usuarios
- [ ] Logging de actividades sensibles
- [ ] No loggear información sensible

### Almacenamiento
- [ ] No guardar información sensible en localStorage
- [ ] Limpiar datos al cerrar sesión
- [ ] Encriptar datos sensibles temporales
- [ ] Auto-limpieza por inactividad

Esta guía proporciona una base sólida para implementar seguridad en aplicaciones frontend con el backend de ASAM.