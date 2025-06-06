# Guía de Manejo de Errores para Frontend

Esta guía proporciona estrategias y ejemplos detallados para manejar correctamente los errores del backend de ASAM en aplicaciones frontend.

## Tabla de Contenidos
1. [Tipos de Errores](#tipos-de-errores)
2. [Estructura de Errores](#estructura-de-errores)
3. [Estrategias de Manejo](#estrategias-de-manejo)
4. [Implementación por Framework](#implementación-por-framework)
5. [Mensajes de Usuario](#mensajes-de-usuario)
6. [Recuperación de Errores](#recuperación-de-errores)
7. [Logging y Monitoreo](#logging-y-monitoreo)

## Tipos de Errores

### 1. Errores de Red
Ocurren cuando no se puede establecer conexión con el servidor.

```javascript
// Detectar errores de red
if (error.networkError) {
  if (error.networkError.statusCode === 0) {
    // Sin conexión a internet
    return "No hay conexión a internet. Por favor, verifica tu conexión.";
  } else if (error.networkError.statusCode >= 500) {
    // Error del servidor
    return "Error del servidor. Por favor, intenta más tarde.";
  }
}
```

### 2. Errores de GraphQL
Errores que vienen del servidor con información estructurada.

```javascript
// Estructura típica de error GraphQL
{
  "errors": [
    {
      "message": "Validation failed",
      "path": ["createMember"],
      "extensions": {
        "code": "VALIDATION_ERROR",
        "details": {
          "nombre": "El nombre es requerido",
          "correo_electronico": "Formato de email inválido"
        }
      }
    }
  ]
}
```

### 3. Errores de Autenticación
Relacionados con tokens expirados o permisos insuficientes.

```javascript
// Códigos de error de autenticación
const AUTH_ERROR_CODES = {
  UNAUTHORIZED: 'UNAUTHORIZED',
  FORBIDDEN: 'FORBIDDEN',
  TOKEN_EXPIRED: 'TOKEN_EXPIRED',
  INVALID_TOKEN: 'INVALID_TOKEN',
  INVALID_CREDENTIALS: 'INVALID_CREDENTIALS'
};
```

### 4. Errores de Validación
Datos de entrada que no cumplen los requisitos.

```javascript
// Códigos de error de validación
const VALIDATION_ERROR_CODES = {
  VALIDATION_ERROR: 'VALIDATION_ERROR',
  REQUIRED_FIELD: 'REQUIRED_FIELD',
  INVALID_FORMAT: 'INVALID_FORMAT',
  DUPLICATE_ENTRY: 'DUPLICATE_ENTRY',
  CONSTRAINT_VIOLATION: 'CONSTRAINT_VIOLATION'
};
```

### 5. Errores de Negocio
Relacionados con reglas de negocio específicas.

```javascript
// Códigos de error de negocio
const BUSINESS_ERROR_CODES = {
  MEMBER_NOT_FOUND: 'MEMBER_NOT_FOUND',
  INSUFFICIENT_FUNDS: 'INSUFFICIENT_FUNDS',
  PAYMENT_ALREADY_PROCESSED: 'PAYMENT_ALREADY_PROCESSED',
  INVALID_STATUS_TRANSITION: 'INVALID_STATUS_TRANSITION',
  QUOTA_EXCEEDED: 'QUOTA_EXCEEDED'
};
```

## Estructura de Errores

### Error Handler Genérico

```javascript
// utils/errorHandler.js
export class ErrorHandler {
  constructor() {
    this.errorMessages = {
      // Errores de red
      NETWORK_ERROR: 'Error de conexión. Por favor, verifica tu conexión a internet.',
      SERVER_ERROR: 'Error del servidor. Por favor, intenta más tarde.',
      
      // Errores de autenticación
      UNAUTHORIZED: 'No tienes autorización para realizar esta acción.',
      FORBIDDEN: 'No tienes permisos suficientes.',
      TOKEN_EXPIRED: 'Tu sesión ha expirado. Por favor, inicia sesión nuevamente.',
      INVALID_CREDENTIALS: 'Usuario o contraseña incorrectos.',
      
      // Errores de validación
      VALIDATION_ERROR: 'Por favor, revisa los datos ingresados.',
      REQUIRED_FIELD: 'Este campo es requerido.',
      INVALID_FORMAT: 'Formato inválido.',
      DUPLICATE_ENTRY: 'Este registro ya existe.',
      
      // Errores de negocio
      MEMBER_NOT_FOUND: 'Miembro no encontrado.',
      INSUFFICIENT_FUNDS: 'Fondos insuficientes para realizar la operación.',
      PAYMENT_ALREADY_PROCESSED: 'Este pago ya ha sido procesado.',
      INVALID_STATUS_TRANSITION: 'No se puede cambiar a este estado.',
      QUOTA_EXCEEDED: 'Se ha excedido el límite permitido.',
      
      // Errores de límites
      RATE_LIMIT_EXCEEDED: 'Demasiadas solicitudes. Por favor, espera un momento.',
      
      // Error por defecto
      UNKNOWN_ERROR: 'Ha ocurrido un error inesperado.'
    };
  }
  
  /**
   * Procesa un error y devuelve información estructurada
   */
  handle(error) {
    // Error de red
    if (error.networkError) {
      return this.handleNetworkError(error.networkError);
    }
    
    // Errores GraphQL
    if (error.graphQLErrors && error.graphQLErrors.length > 0) {
      return this.handleGraphQLErrors(error.graphQLErrors);
    }
    
    // Error genérico
    return {
      type: 'error',
      code: 'UNKNOWN_ERROR',
      message: this.errorMessages.UNKNOWN_ERROR,
      details: {}
    };
  }
  
  /**
   * Maneja errores de red
   */
  handleNetworkError(networkError) {
    if (networkError.statusCode === 0) {
      return {
        type: 'network',
        code: 'NETWORK_ERROR',
        message: this.errorMessages.NETWORK_ERROR,
        details: {}
      };
    }
    
    if (networkError.statusCode >= 500) {
      return {
        type: 'server',
        code: 'SERVER_ERROR',
        message: this.errorMessages.SERVER_ERROR,
        details: { statusCode: networkError.statusCode }
      };
    }
    
    return {
      type: 'network',
      code: 'NETWORK_ERROR',
      message: networkError.message || this.errorMessages.NETWORK_ERROR,
      details: { statusCode: networkError.statusCode }
    };
  }
  
  /**
   * Maneja errores GraphQL
   */
  handleGraphQLErrors(graphQLErrors) {
    const firstError = graphQLErrors[0];
    const code = firstError.extensions?.code || 'UNKNOWN_ERROR';
    
    // Errores de validación con detalles
    if (code === 'VALIDATION_ERROR' && firstError.extensions?.details) {
      return {
        type: 'validation',
        code,
        message: this.errorMessages[code] || firstError.message,
        details: firstError.extensions.details
      };
    }
    
    // Otros errores
    return {
      type: 'business',
      code,
      message: this.errorMessages[code] || firstError.message,
      details: firstError.extensions || {}
    };
  }
  
  /**
   * Obtiene un mensaje de error user-friendly
   */
  getMessage(code, defaultMessage = null) {
    return this.errorMessages[code] || defaultMessage || this.errorMessages.UNKNOWN_ERROR;
  }
  
  /**
   * Formatea errores de validación para formularios
   */
  formatValidationErrors(details) {
    const formatted = {};
    for (const [field, message] of Object.entries(details)) {
      formatted[field] = Array.isArray(message) ? message[0] : message;
    }
    return formatted;
  }
}

export const errorHandler = new ErrorHandler();
```

## Estrategias de Manejo

### 1. Interceptor Global

```javascript
// apollo/errorLink.js
import { onError } from '@apollo/client/link/error';
import { errorHandler } from '@/utils/errorHandler';
import { notificationStore } from '@/stores/notifications';
import { authStore } from '@/stores/auth';

export const errorLink = onError(({ graphQLErrors, networkError, operation, forward }) => {
  const error = errorHandler.handle({ graphQLErrors, networkError });
  
  // Log para desarrollo
  if (process.env.NODE_ENV === 'development') {
    console.error('Apollo Error:', error);
  }
  
  // Manejo específico por tipo de error
  switch (error.code) {
    case 'TOKEN_EXPIRED':
      // Intentar renovar token
      return authStore.refreshToken()
        .then(() => forward(operation))
        .catch(() => {
          authStore.logout();
          window.location.href = '/login';
        });
    
    case 'UNAUTHORIZED':
    case 'INVALID_TOKEN':
      // Logout inmediato
      authStore.logout();
      window.location.href = '/login';
      break;
    
    case 'RATE_LIMIT_EXCEEDED':
      // Mostrar notificación con tiempo de espera
      const retryAfter = error.details.retryAfter || 60;
      notificationStore.error(
        `Por favor espera ${retryAfter} segundos antes de intentar nuevamente.`,
        10000
      );
      break;
    
    case 'NETWORK_ERROR':
      // Mostrar notificación de error de red
      notificationStore.error(error.message, 7000);
      break;
    
    default:
      // Para otros errores, dejar que se propaguen
      break;
  }
});
```

### 2. Manejo en Componentes

```javascript
// composables/useErrorHandler.js
import { ref } from 'vue';
import { errorHandler } from '@/utils/errorHandler';
import { useNotification } from './useNotification';

export function useErrorHandler() {
  const errors = ref({});
  const generalError = ref('');
  const { showError, showWarning } = useNotification();
  
  /**
   * Maneja un error y actualiza el estado
   */
  const handleError = (error, options = {}) => {
    const { 
      showNotification = true, 
      notificationDuration = 7000,
      fieldMapping = {} 
    } = options;
    
    const handled = errorHandler.handle(error);
    
    // Limpiar errores previos
    errors.value = {};
    generalError.value = '';
    
    switch (handled.type) {
      case 'validation':
        // Mapear errores a campos del formulario
        errors.value = mapFieldErrors(handled.details, fieldMapping);
        if (showNotification) {
          showWarning(handled.message, notificationDuration);
        }
        break;
      
      case 'network':
      case 'server':
        generalError.value = handled.message;
        if (showNotification) {
          showError(handled.message, notificationDuration);
        }
        break;
      
      default:
        generalError.value = handled.message;
        if (showNotification) {
          showError(handled.message, notificationDuration);
        }
    }
    
    return handled;
  };
  
  /**
   * Mapea errores del servidor a campos del formulario
   */
  const mapFieldErrors = (serverErrors, mapping) => {
    const mapped = {};
    
    for (const [serverField, message] of Object.entries(serverErrors)) {
      const clientField = mapping[serverField] || serverField;
      mapped[clientField] = message;
    }
    
    return mapped;
  };
  
  /**
   * Limpia todos los errores
   */
  const clearErrors = () => {
    errors.value = {};
    generalError.value = '';
  };
  
  /**
   * Limpia error de un campo específico
   */
  const clearFieldError = (field) => {
    delete errors.value[field];
  };
  
  return {
    errors,
    generalError,
    handleError,
    clearErrors,
    clearFieldError
  };
}
```

## Implementación por Framework

### React - Hook de Manejo de Errores

```javascript
// hooks/useApiCall.js
import { useState, useCallback } from 'react';
import { errorHandler } from '../utils/errorHandler';
import { useNotification } from './useNotification';

export function useApiCall() {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const { showError, showSuccess } = useNotification();
  
  const execute = useCallback(async (
    apiFunction,
    {
      onSuccess,
      onError,
      successMessage,
      errorMessage,
      showNotifications = true
    } = {}
  ) => {
    setLoading(true);
    setError(null);
    
    try {
      const result = await apiFunction();
      
      if (showNotifications && successMessage) {
        showSuccess(successMessage);
      }
      
      if (onSuccess) {
        await onSuccess(result);
      }
      
      return { success: true, data: result };
    } catch (err) {
      const handled = errorHandler.handle(err);
      setError(handled);
      
      if (showNotifications) {
        const message = errorMessage || handled.message;
        showError(message);
      }
      
      if (onError) {
        await onError(handled);
      }
      
      return { success: false, error: handled };
    } finally {
      setLoading(false);
    }
  }, [showError, showSuccess]);
  
  return {
    loading,
    error,
    execute
  };
}

// Uso en componente
function CreateMemberForm() {
  const [createMember] = useMutation(CREATE_MEMBER_MUTATION);
  const { execute, loading, error } = useApiCall();
  
  const handleSubmit = async (formData) => {
    await execute(
      () => createMember({ variables: { input: formData } }),
      {
        successMessage: 'Miembro creado exitosamente',
        onSuccess: (result) => {
          navigate(`/members/${result.data.createMember.miembro_id}`);
        },
        onError: (error) => {
          if (error.type === 'validation') {
            // Mostrar errores en campos
            setFieldErrors(error.details);
          }
        }
      }
    );
  };
  
  return (
    <form onSubmit={handleSubmit}>
      {error?.type === 'validation' && (
        <ValidationErrors errors={error.details} />
      )}
      {/* Campos del formulario */}
    </form>
  );
}
```

### Vue - Composable de API

```vue
<script setup>
import { ref } from 'vue';
import { useMutation } from '@vue/apollo-composable';
import { useApiCall } from '@/composables/useApiCall';
import { CREATE_MEMBER_MUTATION } from '@/graphql/members';

const { execute, loading, errors, generalError } = useApiCall();

const form = ref({
  nombre: '',
  apellidos: '',
  // ... otros campos
});

const { mutate: createMember } = useMutation(CREATE_MEMBER_MUTATION);

const handleSubmit = async () => {
  const result = await execute(
    () => createMember({ variables: { input: form.value } }),
    {
      successMessage: 'Miembro creado exitosamente',
      fieldMapping: {
        'member_name': 'nombre',  // Mapear errores del servidor
        'member_email': 'correo_electronico'
      }
    }
  );
  
  if (result.success) {
    router.push('/members');
  }
};
</script>

<template>
  <form @submit.prevent="handleSubmit">
    <div v-if="generalError" class="error-message">
      {{ generalError }}
    </div>
    
    <FormField
      v-model="form.nombre"
      label="Nombre"
      :error="errors.nombre"
    />
    
    <FormField
      v-model="form.apellidos"
      label="Apellidos"
      :error="errors.apellidos"
    />
    
    <button :disabled="loading">
      {{ loading ? 'Guardando...' : 'Guardar' }}
    </button>
  </form>
</template>
```

## Mensajes de Usuario

### Diccionario de Mensajes

```javascript
// utils/userMessages.js
export const userMessages = {
  // Mensajes de éxito
  success: {
    memberCreated: 'Miembro creado exitosamente',
    memberUpdated: 'Miembro actualizado correctamente',
    memberDeleted: 'Miembro eliminado correctamente',
    paymentRegistered: 'Pago registrado exitosamente',
    loginSuccess: 'Bienvenido de nuevo',
    logoutSuccess: 'Sesión cerrada correctamente',
    dataExported: 'Datos exportados correctamente'
  },
  
  // Mensajes de error por contexto
  errors: {
    // Login
    login: {
      INVALID_CREDENTIALS: 'Email o contraseña incorrectos',
      ACCOUNT_LOCKED: 'Tu cuenta ha sido bloqueada. Contacta al administrador.',
      ACCOUNT_INACTIVE: 'Tu cuenta está inactiva.'
    },
    
    // Miembros
    members: {
      MEMBER_NOT_FOUND: 'No se encontró el miembro solicitado',
      DUPLICATE_MEMBER_NUMBER: 'Ya existe un miembro con ese número de socio',
      CANNOT_DELETE_ACTIVE_MEMBER: 'No se puede eliminar un miembro activo',
      INVALID_DNI_FORMAT: 'El formato del DNI/NIE no es válido'
    },
    
    // Pagos
    payments: {
      INSUFFICIENT_FUNDS: 'El balance es insuficiente para realizar esta operación',
      PAYMENT_ALREADY_PROCESSED: 'Este pago ya ha sido procesado anteriormente',
      INVALID_AMOUNT: 'El monto ingresado no es válido',
      PAYMENT_NOT_FOUND: 'No se encontró el pago solicitado'
    },
    
    // Validación
    validation: {
      required: 'Este campo es obligatorio',
      email: 'Ingresa un email válido',
      minLength: 'Mínimo {min} caracteres',
      maxLength: 'Máximo {max} caracteres',
      pattern: 'Formato inválido',
      date: 'Fecha inválida',
      futureDate: 'La fecha no puede ser futura',
      pastDate: 'La fecha debe ser futura'
    }
  },
  
  // Mensajes de confirmación
  confirmations: {
    deleteMember: '¿Estás seguro de eliminar a {name}? Esta acción no se puede deshacer.',
    cancelPayment: '¿Estás seguro de cancelar este pago?',
    logout: '¿Estás seguro de cerrar sesión?'
  },
  
  // Mensajes informativos
  info: {
    noResults: 'No se encontraron resultados',
    loading: 'Cargando...',
    saving: 'Guardando...',
    processing: 'Procesando...',
    sessionExpiringSoon: 'Tu sesión expirará en {minutes} minutos',
    offlineMode: 'Estás trabajando sin conexión'
  }
};

// Helper para obtener mensajes con interpolación
export function getMessage(path, params = {}) {
  const keys = path.split('.');
  let message = userMessages;
  
  for (const key of keys) {
    message = message[key];
    if (!message) return path; // Retornar path si no encuentra el mensaje
  }
  
  // Interpolar parámetros
  if (typeof message === 'string' && params) {
    return message.replace(/{(\w+)}/g, (match, key) => params[key] || match);
  }
  
  return message;
}
```

## Recuperación de Errores

### Estrategias de Retry

```javascript
// utils/retryStrategy.js
export class RetryStrategy {
  constructor(options = {}) {
    this.maxRetries = options.maxRetries || 3;
    this.initialDelay = options.initialDelay || 1000;
    this.maxDelay = options.maxDelay || 30000;
    this.backoffMultiplier = options.backoffMultiplier || 2;
    this.retryCondition = options.retryCondition || this.defaultRetryCondition;
  }
  
  /**
   * Condición por defecto para reintentar
   */
  defaultRetryCondition(error) {
    // Reintentar en errores de red y errores 5xx
    if (error.networkError) {
      const status = error.networkError.statusCode;
      return status === 0 || (status >= 500 && status < 600);
    }
    
    // Reintentar en algunos errores específicos
    const retryableCodes = ['RATE_LIMIT_EXCEEDED', 'TIMEOUT', 'SERVICE_UNAVAILABLE'];
    if (error.graphQLErrors?.length > 0) {
      const code = error.graphQLErrors[0].extensions?.code;
      return retryableCodes.includes(code);
    }
    
    return false;
  }
  
  /**
   * Ejecuta una función con reintentos
   */
  async execute(fn, context = {}) {
    let lastError;
    
    for (let attempt = 0; attempt <= this.maxRetries; attempt++) {
      try {
        return await fn();
      } catch (error) {
        lastError = error;
        
        // Verificar si se debe reintentar
        if (attempt < this.maxRetries && this.retryCondition(error)) {
          const delay = this.calculateDelay(attempt, error);
          
          // Callback opcional antes de reintentar
          if (context.onRetry) {
            context.onRetry({ attempt, delay, error });
          }
          
          await this.sleep(delay);
        } else {
          // No reintentar, lanzar error
          throw error;
        }
      }
    }
    
    throw lastError;
  }
  
  /**
   * Calcula el delay para el siguiente reintento
   */
  calculateDelay(attempt, error) {
    // Si el error incluye retry-after, usarlo
    if (error.graphQLErrors?.[0]?.extensions?.retryAfter) {
      return error.graphQLErrors[0].extensions.retryAfter * 1000;
    }
    
    // Exponential backoff
    const delay = Math.min(
      this.initialDelay * Math.pow(this.backoffMultiplier, attempt),
      this.maxDelay
    );
    
    // Añadir jitter para evitar thundering herd
    const jitter = delay * 0.2 * Math.random();
    
    return Math.round(delay + jitter);
  }
  
  sleep(ms) {
    return new Promise(resolve => setTimeout(resolve, ms));
  }
}

// Hook para usar retry
export function useRetry() {
  const retry = new RetryStrategy();
  const { showWarning } = useNotification();
  
  const executeWithRetry = async (fn, options = {}) => {
    return retry.execute(fn, {
      onRetry: ({ attempt, delay }) => {
        if (options.showRetryNotification) {
          showWarning(
            `Reintentando... (intento ${attempt + 1}/${retry.maxRetries})`,
            delay
          );
        }
      }
    });
  };
  
  return { executeWithRetry };
}
```

### Componente de Boundary Error

```javascript
// components/ErrorBoundary.jsx
import React from 'react';
import { ErrorBoundary as ReactErrorBoundary } from 'react-error-boundary';

function ErrorFallback({ error, resetErrorBoundary }) {
  return (
    <div className="error-boundary">
      <div className="error-content">
        <h2>Oops! Algo salió mal</h2>
        <p>Ha ocurrido un error inesperado.</p>
        
        {process.env.NODE_ENV === 'development' && (
          <details className="error-details">
            <summary>Detalles del error</summary>
            <pre>{error.message}</pre>
            <pre>{error.stack}</pre>
          </details>
        )}
        
        <div className="error-actions">
          <button onClick={resetErrorBoundary} className="btn-primary">
            Intentar de nuevo
          </button>
          <button onClick={() => window.location.href = '/'} className="btn-secondary">
            Ir al inicio
          </button>
        </div>
      </div>
    </div>
  );
}

export function AppErrorBoundary({ children }) {
  return (
    <ReactErrorBoundary
      FallbackComponent={ErrorFallback}
      onError={(error, errorInfo) => {
        // Log to error reporting service
        console.error('Error caught by boundary:', error, errorInfo);
        
        // Enviar a servicio de monitoreo
        if (window.Sentry) {
          window.Sentry.captureException(error, {
            contexts: { react: errorInfo }
          });
        }
      }}
      onReset={() => {
        // Limpiar estado si es necesario
        window.location.reload();
      }}
    >
      {children}
    </ReactErrorBoundary>
  );
}
```

## Logging y Monitoreo

### Servicio de Logging

```javascript
// services/logger.js
class Logger {
  constructor() {
    this.isDevelopment = process.env.NODE_ENV === 'development';
    this.queue = [];
    this.batchSize = 10;
    this.flushInterval = 5000;
    
    // Iniciar flush automático
    if (!this.isDevelopment) {
      setInterval(() => this.flush(), this.flushInterval);
    }
  }
  
  /**
   * Log de error
   */
  error(message, error, context = {}) {
    const errorLog = {
      level: 'error',
      message,
      timestamp: new Date().toISOString(),
      error: this.serializeError(error),
      context,
      userAgent: navigator.userAgent,
      url: window.location.href
    };
    
    if (this.isDevelopment) {
      console.error(message, error, context);
    } else {
      this.queue.push(errorLog);
      
      // Flush inmediato para errores críticos
      if (error.severity === 'critical') {
        this.flush();
      }
    }
  }
  
  /**
   * Log de advertencia
   */
  warn(message, context = {}) {
    const warnLog = {
      level: 'warn',
      message,
      timestamp: new Date().toISOString(),
      context,
      url: window.location.href
    };
    
    if (this.isDevelopment) {
      console.warn(message, context);
    } else {
      this.queue.push(warnLog);
    }
  }
  
  /**
   * Log informativo
   */
  info(message, context = {}) {
    if (this.isDevelopment) {
      console.info(message, context);
    }
  }
  
  /**
   * Serializa un error para logging
   */
  serializeError(error) {
    if (!error) return null;
    
    const serialized = {
      message: error.message,
      stack: error.stack,
      name: error.name
    };
    
    // GraphQL errors
    if (error.graphQLErrors) {
      serialized.graphQLErrors = error.graphQLErrors.map(e => ({
        message: e.message,
        path: e.path,
        extensions: e.extensions
      }));
    }
    
    // Network errors
    if (error.networkError) {
      serialized.networkError = {
        message: error.networkError.message,
        statusCode: error.networkError.statusCode
      };
    }
    
    return serialized;
  }
  
  /**
   * Envía logs al servidor
   */
  async flush() {
    if (this.queue.length === 0) return;
    
    const logs = [...this.queue];
    this.queue = [];
    
    try {
      await fetch('/api/logs', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${getAccessToken()}`
        },
        body: JSON.stringify({ logs })
      });
    } catch (error) {
      // Si falla, volver a añadir a la cola
      this.queue.unshift(...logs);
    }
  }
}

export const logger = new Logger();
```

### Integración con Sentry

```javascript
// services/monitoring.js
import * as Sentry from '@sentry/react';
import { BrowserTracing } from '@sentry/tracing';

export function initMonitoring() {
  if (process.env.NODE_ENV === 'production') {
    Sentry.init({
      dsn: process.env.REACT_APP_SENTRY_DSN,
      environment: process.env.NODE_ENV,
      integrations: [
        new BrowserTracing(),
      ],
      tracesSampleRate: 0.1,
      
      beforeSend(event, hint) {
        // Filtrar errores que no queremos enviar
        if (event.exception) {
          const error = hint.originalException;
          
          // No enviar errores de red esperados
          if (error?.networkError?.statusCode === 0) {
            return null;
          }
          
          // No enviar errores de validación
          if (error?.graphQLErrors?.[0]?.extensions?.code === 'VALIDATION_ERROR') {
            return null;
          }
        }
        
        return event;
      }
    });
  }
}

// Hook para tracking de errores
export function useErrorTracking() {
  const trackError = (error, context = {}) => {
    // Log local
    logger.error('Application error', error, context);
    
    // Enviar a Sentry
    if (window.Sentry) {
      Sentry.captureException(error, {
        extra: context
      });
    }
  };
  
  return { trackError };
}
```

## Ejemplos de Uso Completo

### Formulario con Manejo Completo de Errores

```jsx
// components/MemberFormWithErrorHandling.jsx
import React, { useState } from 'react';
import { useMutation } from '@apollo/client';
import { useApiCall } from '../hooks/useApiCall';
import { useRetry } from '../hooks/useRetry';
import { useErrorTracking } from '../services/monitoring';
import { CREATE_MEMBER_MUTATION } from '../graphql/mutations';
import { userMessages } from '../utils/userMessages';

function MemberFormWithErrorHandling() {
  const [formData, setFormData] = useState({
    nombre: '',
    apellidos: '',
    correo_electronico: ''
  });
  
  const [createMember] = useMutation(CREATE_MEMBER_MUTATION);
  const { execute, loading, errors, generalError } = useApiCall();
  const { executeWithRetry } = useRetry();
  const { trackError } = useErrorTracking();
  
  const handleSubmit = async (e) => {
    e.preventDefault();
    
    const result = await execute(
      () => executeWithRetry(
        () => createMember({ variables: { input: formData } }),
        { showRetryNotification: true }
      ),
      {
        successMessage: userMessages.success.memberCreated,
        onError: (error) => {
          // Tracking adicional para errores específicos
          if (error.code === 'DUPLICATE_ENTRY') {
            trackError(new Error('Duplicate member attempt'), {
              numero_socio: formData.numero_socio
            });
          }
        }
      }
    );
    
    if (result.success) {
      // Navegar o limpiar formulario
      navigate('/members');
    }
  };
  
  // Limpiar error de campo cuando el usuario empieza a escribir
  const handleFieldChange = (field, value) => {
    setFormData(prev => ({ ...prev, [field]: value }));
    
    if (errors[field]) {
      clearFieldError(field);
    }
  };
  
  return (
    <form onSubmit={handleSubmit} className="member-form">
      {/* Error general */}
      {generalError && (
        <Alert type="error" onClose={() => clearErrors()}>
          {generalError}
        </Alert>
      )}
      
      {/* Campos del formulario */}
      <FormField
        label="Nombre"
        value={formData.nombre}
        onChange={(value) => handleFieldChange('nombre', value)}
        error={errors.nombre}
        required
      />
      
      <FormField
        label="Apellidos"
        value={formData.apellidos}
        onChange={(value) => handleFieldChange('apellidos', value)}
        error={errors.apellidos}
        required
      />
      
      <FormField
        label="Email"
        type="email"
        value={formData.correo_electronico}
        onChange={(value) => handleFieldChange('correo_electronico', value)}
        error={errors.correo_electronico}
      />
      
      <div className="form-actions">
        <Button
          type="submit"
          loading={loading}
          disabled={loading}
        >
          Crear Miembro
        </Button>
      </div>
    </form>
  );
}
```

Esta guía proporciona un sistema completo y robusto para manejar errores en aplicaciones frontend que consumen el backend de ASAM, asegurando una experiencia de usuario consistente y profesional.
