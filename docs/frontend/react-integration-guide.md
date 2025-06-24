# Guía de Integración con React

Esta guía muestra cómo integrar el backend de ASAM con una aplicación React moderna.

## Tabla de Contenidos
1. [Configuración Inicial](#configuración-inicial)
2. [Estructura del Proyecto](#estructura-del-proyecto)
3. [Hooks Personalizados](#hooks-personalizados)
4. [Componentes Reutilizables](#componentes-reutilizables)
5. [Manejo de Estado Global](#manejo-de-estado-global)
6. [Testing](#testing)

## Configuración Inicial

### 1. Instalar Dependencias

```bash
npm install @apollo/client graphql
npm install --save-dev @graphql-codegen/cli @graphql-codegen/typescript @graphql-codegen/typescript-operations @graphql-codegen/typescript-react-apollo
```

### 2. Configurar Apollo Client

```javascript
// src/apollo/client.js
import { ApolloClient, InMemoryCache, createHttpLink, split } from '@apollo/client';
import { setContext } from '@apollo/client/link/context';
import { onError } from '@apollo/client/link/error';
import { getMainDefinition } from '@apollo/client/utilities';
import { WebSocketLink } from '@apollo/client/link/ws';

// Endpoints
const GRAPHQL_ENDPOINT = process.env.NODE_ENV === 'production'
  ? 'https://asam-backend-jtpswzdxuq-ew.a.run.app/graphql'
  : 'http://localhost:8080/graphql';

const WS_ENDPOINT = process.env.NODE_ENV === 'production'
  ? 'wss://asam-backend-jtpswzdxuq-ew.a.run.app/graphql'
  : 'ws://localhost:8080/graphql';

// HTTP Link
const httpLink = createHttpLink({
  uri: GRAPHQL_ENDPOINT,
});

// WebSocket Link para subscriptions (si las implementas en el futuro)
const wsLink = new WebSocketLink({
  uri: WS_ENDPOINT,
  options: {
    reconnect: true,
    connectionParams: () => ({
      authToken: localStorage.getItem('accessToken'),
    }),
  },
});

// Auth Link
const authLink = setContext((_, { headers }) => {
  const token = localStorage.getItem('accessToken');
  return {
    headers: {
      ...headers,
      authorization: token ? `Bearer ${token}` : "",
    }
  };
});

// Error Link
const errorLink = onError(({ graphQLErrors, networkError, operation, forward }) => {
  if (graphQLErrors) {
    for (let err of graphQLErrors) {
      switch (err.extensions.code) {
        case 'TOKEN_EXPIRED':
          // Lógica para renovar token
          return refreshToken().then(() => forward(operation));
        case 'UNAUTHORIZED':
          // Redireccionar a login
          window.location.href = '/login';
          break;
        default:
          console.error(`[GraphQL error]: Message: ${err.message}`);
      }
    }
  }
  
  if (networkError) {
    console.error(`[Network error]: ${networkError}`);
  }
});

// Split para usar WebSocket o HTTP según el tipo de operación
const splitLink = split(
  ({ query }) => {
    const definition = getMainDefinition(query);
    return (
      definition.kind === 'OperationDefinition' &&
      definition.operation === 'subscription'
    );
  },
  wsLink,
  authLink.concat(httpLink),
);

// Cliente Apollo
export const apolloClient = new ApolloClient({
  link: errorLink.concat(splitLink),
  cache: new InMemoryCache({
    typePolicies: {
      Query: {
        fields: {
          listMembers: {
            keyArgs: ["filter"],
            merge(existing = { nodes: [] }, incoming) {
              return incoming;
            }
          },
          listFamilies: {
            keyArgs: ["filter"],
            merge(existing = { nodes: [] }, incoming) {
              return incoming;
            }
          },
          getTransactions: {
            keyArgs: ["filter"],
            merge(existing = { nodes: [] }, incoming) {
              return incoming;
            }
          },
          listUsers: {
            keyArgs: ["page", "pageSize"],
            merge(existing = [], incoming) {
              return incoming;
            }
          }
        }
      },
      Member: {
        keyFields: ["miembro_id"]
      },
      Family: {
        keyFields: ["id"]
      },
      Payment: {
        keyFields: ["id"]
      },
      User: {
        keyFields: ["id"]
      }
    }
  }),
  defaultOptions: {
    watchQuery: {
      fetchPolicy: 'cache-and-network',
      errorPolicy: 'all',
    },
    query: {
      fetchPolicy: 'network-only',
      errorPolicy: 'all',
    },
  },
});

// Función para renovar token
async function refreshToken() {
  const refreshToken = localStorage.getItem('refreshToken');
  if (!refreshToken) {
    throw new Error('No refresh token available');
  }
  
  try {
    const { data } = await apolloClient.mutate({
      mutation: REFRESH_TOKEN_MUTATION,
      variables: { input: { refreshToken } }
    });
    
    const { accessToken, refreshToken: newRefreshToken, expiresAt } = data.refreshToken;
    
    localStorage.setItem('accessToken', accessToken);
    localStorage.setItem('refreshToken', newRefreshToken);
    localStorage.setItem('tokenExpiresAt', expiresAt);
    
    return accessToken;
  } catch (error) {
    // Si falla la renovación, limpiar tokens y redireccionar
    localStorage.clear();
    window.location.href = '/login';
    throw error;
  }
}
```

### 3. Provider Principal

```javascript
// src/App.js
import React from 'react';
import { ApolloProvider } from '@apollo/client';
import { BrowserRouter } from 'react-router-dom';
import { apolloClient } from './apollo/client';
import { AuthProvider } from './contexts/AuthContext';
import { NotificationProvider } from './contexts/NotificationContext';
import Routes from './routes';

function App() {
  return (
    <ApolloProvider client={apolloClient}>
      <BrowserRouter>
        <AuthProvider>
          <NotificationProvider>
            <Routes />
          </NotificationProvider>
        </AuthProvider>
      </BrowserRouter>
    </ApolloProvider>
  );
}

export default App;
```

## Estructura del Proyecto

```
src/
├── apollo/
│   ├── client.js         # Configuración de Apollo Client
│   └── cache.js          # Configuración del cache
├── components/
│   ├── common/           # Componentes reutilizables
│   ├── members/          # Componentes de miembros
│   ├── families/         # Componentes de familias
│   ├── payments/         # Componentes de pagos
│   ├── users/            # Componentes de usuarios (Admin)
│   └── layout/           # Componentes de layout
├── contexts/
│   ├── AuthContext.js    # Context de autenticación
│   └── NotificationContext.js
├── graphql/
│   ├── mutations/        # Archivos .graphql con mutations
│   ├── queries/          # Archivos .graphql con queries
│   └── fragments/        # Fragmentos reutilizables
├── hooks/
│   ├── useAuth.js        # Hook de autenticación
│   ├── usePagination.js  # Hook de paginación
│   ├── useDebounce.js    # Hook de debounce
│   └── useUser.js        # Hook de gestión de usuarios
├── pages/
│   ├── Login.js
│   ├── Dashboard.js
│   ├── Profile.js
│   ├── ForgotPassword.js
│   ├── ResetPassword.js
│   ├── VerifyEmail.js
│   ├── Members/
│   ├── Families/
│   ├── Payments/
│   └── Admin/
│       ├── Users.js
│       └── UserDetail.js
├── services/
│   ├── auth.service.js
│   ├── members.service.js
│   ├── payments.service.js
│   └── users.service.js
├── utils/
│   ├── validators.js
│   ├── formatters.js
│   └── constants.js
└── routes/
    └── index.js
```

## Hooks Personalizados

### Hook de Autenticación (Actualizado)

```javascript
// src/hooks/useAuth.js
import { useContext, useCallback, useEffect } from 'react';
import { useMutation } from '@apollo/client';
import { useNavigate } from 'react-router-dom';
import { AuthContext } from '../contexts/AuthContext';
import { 
  LOGIN_MUTATION, 
  LOGOUT_MUTATION, 
  REFRESH_TOKEN_MUTATION,
  CHANGE_PASSWORD_MUTATION,
  REQUEST_PASSWORD_RESET_MUTATION,
  SEND_VERIFICATION_EMAIL_MUTATION
} from '../graphql/mutations/auth';

export const useAuth = () => {
  const context = useContext(AuthContext);
  const navigate = useNavigate();
  
  if (!context) {
    throw new Error('useAuth must be used within AuthProvider');
  }
  
  const { user, setUser, isAuthenticated, setIsAuthenticated } = context;
  
  const [loginMutation, { loading: loginLoading }] = useMutation(LOGIN_MUTATION);
  const [logoutMutation] = useMutation(LOGOUT_MUTATION);
  const [refreshMutation] = useMutation(REFRESH_TOKEN_MUTATION);
  const [changePasswordMutation, { loading: changePasswordLoading }] = useMutation(CHANGE_PASSWORD_MUTATION);
  const [requestPasswordResetMutation, { loading: resetLoading }] = useMutation(REQUEST_PASSWORD_RESET_MUTATION);
  const [sendVerificationMutation, { loading: verificationLoading }] = useMutation(SEND_VERIFICATION_EMAIL_MUTATION);
  
  const login = useCallback(async (username, password) => {
    try {
      const { data } = await loginMutation({
        variables: { input: { username, password } }
      });
      
      const { user, accessToken, refreshToken, expiresAt } = data.login;
      
      // Guardar tokens
      localStorage.setItem('accessToken', accessToken);
      localStorage.setItem('refreshToken', refreshToken);
      localStorage.setItem('tokenExpiresAt', expiresAt);
      
      // Actualizar estado
      setUser(user);
      setIsAuthenticated(true);
      
      // Verificar si el email está verificado
      if (!user.emailVerified) {
        navigate('/verify-email-notice');
        return { success: true, emailVerified: false };
      }
      
      // Navegar según rol
      if (user.role === 'admin') {
        navigate('/admin/dashboard');
      } else {
        navigate('/dashboard');
      }
      
      return { success: true, emailVerified: true };
    } catch (error) {
      console.error('Login error:', error);
      return { 
        success: false, 
        error: error.graphQLErrors?.[0]?.message || 'Error al iniciar sesión' 
      };
    }
  }, [loginMutation, setUser, setIsAuthenticated, navigate]);
  
  const logout = useCallback(async () => {
    try {
      await logoutMutation();
    } catch (error) {
      console.error('Logout error:', error);
    } finally {
      // Limpiar estado local siempre
      localStorage.clear();
      setUser(null);
      setIsAuthenticated(false);
      navigate('/login');
    }
  }, [logoutMutation, setUser, setIsAuthenticated, navigate]);
  
  const refreshToken = useCallback(async () => {
    const refreshToken = localStorage.getItem('refreshToken');
    if (!refreshToken) {
      throw new Error('No refresh token');
    }
    
    try {
      const { data } = await refreshMutation({
        variables: { input: { refreshToken } }
      });
      
      const { accessToken, refreshToken: newRefreshToken, expiresAt } = data.refreshToken;
      
      localStorage.setItem('accessToken', accessToken);
      localStorage.setItem('refreshToken', newRefreshToken);
      localStorage.setItem('tokenExpiresAt', expiresAt);
      
      return true;
    } catch (error) {
      console.error('Refresh token error:', error);
      await logout();
      return false;
    }
  }, [refreshMutation, logout]);
  
  const changePassword = useCallback(async (currentPassword, newPassword) => {
    try {
      const { data } = await changePasswordMutation({
        variables: { input: { currentPassword, newPassword } }
      });
      
      return {
        success: data.changePassword.success,
        message: data.changePassword.message,
        error: data.changePassword.error
      };
    } catch (error) {
      console.error('Change password error:', error);
      return {
        success: false,
        error: error.graphQLErrors?.[0]?.message || 'Error al cambiar la contraseña'
      };
    }
  }, [changePasswordMutation]);
  
  const requestPasswordReset = useCallback(async (email) => {
    try {
      const { data } = await requestPasswordResetMutation({
        variables: { email }
      });
      
      return {
        success: data.requestPasswordReset.success,
        message: data.requestPasswordReset.message
      };
    } catch (error) {
      console.error('Request password reset error:', error);
      return {
        success: false,
        error: error.graphQLErrors?.[0]?.message || 'Error al solicitar reseteo'
      };
    }
  }, [requestPasswordResetMutation]);
  
  const sendVerificationEmail = useCallback(async () => {
    try {
      const { data } = await sendVerificationMutation();
      
      return {
        success: data.sendVerificationEmail.success,
        message: data.sendVerificationEmail.message
      };
    } catch (error) {
      console.error('Send verification email error:', error);
      return {
        success: false,
        error: error.graphQLErrors?.[0]?.message || 'Error al enviar email'
      };
    }
  }, [sendVerificationMutation]);
  
  // Auto-refresh token
  useEffect(() => {
    if (!isAuthenticated) return;
    
    const checkTokenExpiry = () => {
      const expiresAt = localStorage.getItem('tokenExpiresAt');
      if (!expiresAt) return;
      
      const expiryTime = new Date(expiresAt).getTime();
      const currentTime = new Date().getTime();
      const timeUntilExpiry = expiryTime - currentTime;
      
      // Renovar 1 minuto antes de expirar
      if (timeUntilExpiry < 60000) {
        refreshToken();
      }
    };
    
    // Verificar cada 30 segundos
    const interval = setInterval(checkTokenExpiry, 30000);
    
    // Verificar inmediatamente
    checkTokenExpiry();
    
    return () => clearInterval(interval);
  }, [isAuthenticated, refreshToken]);
  
  return {
    user,
    isAuthenticated,
    login,
    logout,
    refreshToken,
    changePassword,
    requestPasswordReset,
    sendVerificationEmail,
    loginLoading,
    changePasswordLoading,
    resetLoading,
    verificationLoading,
    isAdmin: user?.role === 'admin',
    isUser: user?.role === 'user',
    isEmailVerified: user?.emailVerified || false,
  };
};
```

### Hook de Gestión de Usuarios (Admin)

```javascript
// src/hooks/useUser.js
import { useMutation, useQuery } from '@apollo/client';
import { useCallback } from 'react';
import {
  CREATE_USER_MUTATION,
  UPDATE_USER_MUTATION,
  DELETE_USER_MUTATION,
  RESET_USER_PASSWORD_MUTATION
} from '../graphql/mutations/users';
import {
  GET_USER_QUERY,
  LIST_USERS_QUERY
} from '../graphql/queries/users';

export const useUser = () => {
  const [createUserMutation, { loading: createLoading }] = useMutation(CREATE_USER_MUTATION);
  const [updateUserMutation, { loading: updateLoading }] = useMutation(UPDATE_USER_MUTATION);
  const [deleteUserMutation, { loading: deleteLoading }] = useMutation(DELETE_USER_MUTATION);
  const [resetPasswordMutation, { loading: resetLoading }] = useMutation(RESET_USER_PASSWORD_MUTATION);
  
  const createUser = useCallback(async (input) => {
    try {
      const { data } = await createUserMutation({
        variables: { input },
        refetchQueries: [{ query: LIST_USERS_QUERY }]
      });
      
      return {
        success: true,
        user: data.createUser
      };
    } catch (error) {
      console.error('Create user error:', error);
      return {
        success: false,
        error: error.graphQLErrors?.[0]?.message || 'Error al crear usuario'
      };
    }
  }, [createUserMutation]);
  
  const updateUser = useCallback(async (input) => {
    try {
      const { data } = await updateUserMutation({
        variables: { input },
        refetchQueries: [{ query: LIST_USERS_QUERY }]
      });
      
      return {
        success: true,
        user: data.updateUser
      };
    } catch (error) {
      console.error('Update user error:', error);
      return {
        success: false,
        error: error.graphQLErrors?.[0]?.message || 'Error al actualizar usuario'
      };
    }
  }, [updateUserMutation]);
  
  const deleteUser = useCallback(async (id) => {
    try {
      const { data } = await deleteUserMutation({
        variables: { id },
        refetchQueries: [{ query: LIST_USERS_QUERY }]
      });
      
      return {
        success: data.deleteUser.success,
        message: data.deleteUser.message
      };
    } catch (error) {
      console.error('Delete user error:', error);
      return {
        success: false,
        error: error.graphQLErrors?.[0]?.message || 'Error al eliminar usuario'
      };
    }
  }, [deleteUserMutation]);
  
  const resetUserPassword = useCallback(async (userId, newPassword) => {
    try {
      const { data } = await resetPasswordMutation({
        variables: { userId, newPassword }
      });
      
      return {
        success: data.resetUserPassword.success,
        message: data.resetUserPassword.message
      };
    } catch (error) {
      console.error('Reset password error:', error);
      return {
        success: false,
        error: error.graphQLErrors?.[0]?.message || 'Error al resetear contraseña'
      };
    }
  }, [resetPasswordMutation]);
  
  const useUsersList = (page = 1, pageSize = 10) => {
    return useQuery(LIST_USERS_QUERY, {
      variables: { page, pageSize },
      fetchPolicy: 'network-only'
    });
  };
  
  const useUserDetail = (id) => {
    return useQuery(GET_USER_QUERY, {
      variables: { id },
      skip: !id
    });
  };
  
  return {
    createUser,
    updateUser,
    deleteUser,
    resetUserPassword,
    useUsersList,
    useUserDetail,
    createLoading,
    updateLoading,
    deleteLoading,
    resetLoading
  };
};
```

### Hook de Paginación

```javascript
// src/hooks/usePagination.js
import { useState, useCallback, useMemo } from 'react';

export const usePagination = (initialPageSize = 20) => {
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(initialPageSize);
  
  const paginationVariables = useMemo(() => ({
    page,
    pageSize
  }), [page, pageSize]);
  
  const goToPage = useCallback((newPage) => {
    setPage(Math.max(1, newPage));
  }, []);
  
  const nextPage = useCallback(() => {
    setPage(prev => prev + 1);
  }, []);
  
  const previousPage = useCallback(() => {
    setPage(prev => Math.max(1, prev - 1));
  }, []);
  
  const resetPage = useCallback(() => {
    setPage(1);
  }, []);
  
  const changePageSize = useCallback((newPageSize) => {
    setPageSize(newPageSize);
    setPage(1); // Reset a primera página
  }, []);
  
  const getPaginationInfo = useCallback((pageInfo) => {
    if (!pageInfo) {
      return {
        totalPages: 0,
        startIndex: 0,
        endIndex: 0,
        canGoNext: false,
        canGoPrevious: false
      };
    }
    
    const totalPages = Math.ceil(pageInfo.totalCount / pageSize);
    const startIndex = (page - 1) * pageSize + 1;
    const endIndex = Math.min(page * pageSize, pageInfo.totalCount);
    
    return {
      totalPages,
      startIndex,
      endIndex,
      canGoNext: pageInfo.hasNextPage,
      canGoPrevious: pageInfo.hasPreviousPage
    };
  }, [page, pageSize]);
  
  return {
    // Estado
    page,
    pageSize,
    paginationVariables,
    
    // Acciones
    goToPage,
    nextPage,
    previousPage,
    resetPage,
    changePageSize,
    
    // Helpers
    getPaginationInfo
  };
};
```

### Hook de Debounce

```javascript
// src/hooks/useDebounce.js
import { useState, useEffect } from 'react';

export const useDebounce = (value, delay = 300) => {
  const [debouncedValue, setDebouncedValue] = useState(value);
  
  useEffect(() => {
    const handler = setTimeout(() => {
      setDebouncedValue(value);
    }, delay);
    
    return () => {
      clearTimeout(handler);
    };
  }, [value, delay]);
  
  return debouncedValue;
};
```

### Hook de Notificaciones

```javascript
// src/hooks/useNotification.js
import { useContext, useCallback } from 'react';
import { NotificationContext } from '../contexts/NotificationContext';

export const useNotification = () => {
  const context = useContext(NotificationContext);
  
  if (!context) {
    throw new Error('useNotification must be used within NotificationProvider');
  }
  
  const { notifications, addNotification, removeNotification } = context;
  
  const showSuccess = useCallback((message, duration = 5000) => {
    addNotification({
      type: 'success',
      message,
      duration
    });
  }, [addNotification]);
  
  const showError = useCallback((message, duration = 7000) => {
    addNotification({
      type: 'error',
      message,
      duration
    });
  }, [addNotification]);
  
  const showWarning = useCallback((message, duration = 5000) => {
    addNotification({
      type: 'warning',
      message,
      duration
    });
  }, [addNotification]);
  
  const showInfo = useCallback((message, duration = 5000) => {
    addNotification({
      type: 'info',
      message,
      duration
    });
  }, [addNotification]);
  
  return {
    notifications,
    showSuccess,
    showError,
    showWarning,
    showInfo,
    removeNotification
  };
};
```

## Componentes Reutilizables

### Componente de Tabla con Paginación

```javascript
// src/components/common/DataTable.js
import React from 'react';
import PropTypes from 'prop-types';
import { ChevronUpIcon, ChevronDownIcon } from '@heroicons/react/24/outline';
import Pagination from './Pagination';
import LoadingSpinner from './LoadingSpinner';

const DataTable = ({
  columns,
  data,
  loading,
  sortConfig,
  onSort,
  pagination,
  onPageChange,
  pageInfo,
  emptyMessage = 'No hay datos para mostrar',
  className = ''
}) => {
  const getSortIcon = (column) => {
    if (!sortConfig || sortConfig.field !== column.field) {
      return null;
    }
    
    return sortConfig.direction === 'ASC' 
      ? <ChevronUpIcon className="w-4 h-4" />
      : <ChevronDownIcon className="w-4 h-4" />;
  };
  
  const handleSort = (column) => {
    if (!column.sortable || !onSort) return;
    onSort(column.field);
  };
  
  if (loading && (!data || data.length === 0)) {
    return (
      <div className="flex justify-center items-center h-64">
        <LoadingSpinner />
      </div>
    );
  }
  
  return (
    <div className={`bg-white shadow rounded-lg overflow-hidden ${className}`}>
      <div className="overflow-x-auto">
        <table className="min-w-full divide-y divide-gray-200">
          <thead className="bg-gray-50">
            <tr>
              {columns.map((column) => (
                <th
                  key={column.field}
                  onClick={() => handleSort(column)}
                  className={`
                    px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider
                    ${column.sortable ? 'cursor-pointer hover:bg-gray-100' : ''}
                  `}
                >
                  <div className="flex items-center space-x-1">
                    <span>{column.label}</span>
                    {column.sortable && getSortIcon(column)}
                  </div>
                </th>
              ))}
            </tr>
          </thead>
          
          <tbody className="bg-white divide-y divide-gray-200">
            {data && data.length > 0 ? (
              data.map((row, index) => (
                <tr key={row.id || index} className="hover:bg-gray-50">
                  {columns.map((column) => (
                    <td key={column.field} className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                      {column.render ? column.render(row) : row[column.field]}
                    </td>
                  ))}
                </tr>
              ))
            ) : (
              <tr>
                <td colSpan={columns.length} className="px-6 py-12 text-center text-gray-500">
                  {emptyMessage}
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
      
      {pagination && pageInfo && (
        <div className="px-6 py-4 border-t border-gray-200">
          <Pagination
            page={pagination.page}
            pageSize={pagination.pageSize}
            totalCount={pageInfo.totalCount}
            hasNextPage={pageInfo.hasNextPage}
            hasPreviousPage={pageInfo.hasPreviousPage}
            onPageChange={onPageChange}
          />
        </div>
      )}
      
      {loading && (
        <div className="absolute inset-0 bg-white bg-opacity-50 flex items-center justify-center">
          <LoadingSpinner />
        </div>
      )}
    </div>
  );
};

DataTable.propTypes = {
  columns: PropTypes.arrayOf(PropTypes.shape({
    field: PropTypes.string.isRequired,
    label: PropTypes.string.isRequired,
    sortable: PropTypes.bool,
    render: PropTypes.func
  })).isRequired,
  data: PropTypes.array,
  loading: PropTypes.bool,
  sortConfig: PropTypes.shape({
    field: PropTypes.string,
    direction: PropTypes.oneOf(['ASC', 'DESC'])
  }),
  onSort: PropTypes.func,
  pagination: PropTypes.shape({
    page: PropTypes.number,
    pageSize: PropTypes.number
  }),
  onPageChange: PropTypes.func,
  pageInfo: PropTypes.shape({
    totalCount: PropTypes.number,
    hasNextPage: PropTypes.bool,
    hasPreviousPage: PropTypes.bool
  }),
  emptyMessage: PropTypes.string,
  className: PropTypes.string
};

export default DataTable;
```

### Componente de Formulario Inteligente

```javascript
// src/components/common/SmartForm.js
import React, { useState, useCallback } from 'react';
import PropTypes from 'prop-types';
import * as Yup from 'yup';

const SmartForm = ({
  fields,
  initialValues = {},
  validationSchema,
  onSubmit,
  submitText = 'Enviar',
  cancelText = 'Cancelar',
  onCancel,
  loading = false
}) => {
  const [values, setValues] = useState(() => {
    const defaultValues = {};
    fields.forEach(field => {
      defaultValues[field.name] = initialValues[field.name] || field.defaultValue || '';
    });
    return defaultValues;
  });
  
  const [errors, setErrors] = useState({});
  const [touched, setTouched] = useState({});
  
  const handleChange = useCallback((fieldName, value) => {
    setValues(prev => ({ ...prev, [fieldName]: value }));
    
    // Limpiar error si existe
    if (errors[fieldName] && touched[fieldName]) {
      validateField(fieldName, value);
    }
  }, [errors, touched]);
  
  const handleBlur = useCallback((fieldName) => {
    setTouched(prev => ({ ...prev, [fieldName]: true }));
    validateField(fieldName, values[fieldName]);
  }, [values]);
  
  const validateField = async (fieldName, value) => {
    if (!validationSchema) return;
    
    try {
      await validationSchema.validateAt(fieldName, { [fieldName]: value });
      setErrors(prev => {
        const newErrors = { ...prev };
        delete newErrors[fieldName];
        return newErrors;
      });
    } catch (error) {
      setErrors(prev => ({ ...prev, [fieldName]: error.message }));
    }
  };
  
  const validateAll = async () => {
    if (!validationSchema) return true;
    
    try {
      await validationSchema.validate(values, { abortEarly: false });
      setErrors({});
      return true;
    } catch (error) {
      const validationErrors = {};
      error.inner.forEach(err => {
        validationErrors[err.path] = err.message;
      });
      setErrors(validationErrors);
      
      // Marcar todos los campos como touched
      const allTouched = {};
      fields.forEach(field => {
        allTouched[field.name] = true;
      });
      setTouched(allTouched);
      
      return false;
    }
  };
  
  const handleSubmit = async (e) => {
    e.preventDefault();
    
    const isValid = await validateAll();
    if (!isValid) return;
    
    await onSubmit(values);
  };
  
  const renderField = (field) => {
    const error = touched[field.name] && errors[field.name];
    const commonProps = {
      id: field.name,
      name: field.name,
      value: values[field.name],
      onChange: (e) => handleChange(field.name, e.target.value),
      onBlur: () => handleBlur(field.name),
      disabled: loading || field.disabled,
      className: `
        mt-1 block w-full rounded-md shadow-sm
        ${error 
          ? 'border-red-300 focus:border-red-500 focus:ring-red-500' 
          : 'border-gray-300 focus:border-indigo-500 focus:ring-indigo-500'
        }
        sm:text-sm
      `
    };
    
    switch (field.type) {
      case 'select':
        return (
          <select {...commonProps}>
            <option value="">Seleccionar...</option>
            {field.options.map(option => (
              <option key={option.value} value={option.value}>
                {option.label}
              </option>
            ))}
          </select>
        );
        
      case 'textarea':
        return (
          <textarea 
            {...commonProps} 
            rows={field.rows || 3}
          />
        );
        
      case 'date':
        return (
          <input 
            {...commonProps} 
            type="date"
            max={field.max}
            min={field.min}
          />
        );
        
      default:
        return (
          <input 
            {...commonProps} 
            type={field.type || 'text'}
            placeholder={field.placeholder}
          />
        );
    }
  };
  
  return (
    <form onSubmit={handleSubmit} className="space-y-6">
      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        {fields.map(field => (
          <div key={field.name} className={field.fullWidth ? 'md:col-span-2' : ''}>
            <label htmlFor={field.name} className="block text-sm font-medium text-gray-700">
              {field.label}
              {field.required && <span className="text-red-500 ml-1">*</span>}
            </label>
            
            {renderField(field)}
            
            {touched[field.name] && errors[field.name] && (
              <p className="mt-1 text-sm text-red-600">{errors[field.name]}</p>
            )}
            
            {field.helpText && (
              <p className="mt-1 text-sm text-gray-500">{field.helpText}</p>
            )}
          </div>
        ))}
      </div>
      
      <div className="flex justify-end space-x-4 pt-6 border-t">
        {onCancel && (
          <button
            type="button"
            onClick={onCancel}
            disabled={loading}
            className="px-4 py-2 border border-gray-300 rounded-md shadow-sm text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 disabled:opacity-50"
          >
            {cancelText}
          </button>
        )}
        
        <button
          type="submit"
          disabled={loading}
          className="px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 disabled:opacity-50"
        >
          {loading ? 'Procesando...' : submitText}
        </button>
      </div>
    </form>
  );
};

SmartForm.propTypes = {
  fields: PropTypes.arrayOf(PropTypes.shape({
    name: PropTypes.string.isRequired,
    label: PropTypes.string.isRequired,
    type: PropTypes.string,
    required: PropTypes.bool,
    disabled: PropTypes.bool,
    fullWidth: PropTypes.bool,
    placeholder: PropTypes.string,
    helpText: PropTypes.string,
    defaultValue: PropTypes.any,
    options: PropTypes.arrayOf(PropTypes.shape({
      value: PropTypes.string,
      label: PropTypes.string
    })),
    rows: PropTypes.number,
    min: PropTypes.string,
    max: PropTypes.string
  })).isRequired,
  initialValues: PropTypes.object,
  validationSchema: PropTypes.object,
  onSubmit: PropTypes.func.isRequired,
  onCancel: PropTypes.func,
  submitText: PropTypes.string,
  cancelText: PropTypes.string,
  loading: PropTypes.bool
};

export default SmartForm;
```

### Componente de Verificación de Email

```javascript
// src/components/auth/EmailVerificationNotice.js
import React from 'react';
import { useAuth } from '../../hooks/useAuth';
import { useNotification } from '../../hooks/useNotification';

const EmailVerificationNotice = () => {
  const { user, sendVerificationEmail, verificationLoading } = useAuth();
  const { showSuccess, showError } = useNotification();
  
  const handleResendEmail = async () => {
    const result = await sendVerificationEmail();
    
    if (result.success) {
      showSuccess(result.message || 'Email de verificación enviado');
    } else {
      showError(result.error || 'Error al enviar el email');
    }
  };
  
  if (!user || user.emailVerified) {
    return null;
  }
  
  return (
    <div className="bg-yellow-50 border-l-4 border-yellow-400 p-4 mb-6">
      <div className="flex">
        <div className="flex-shrink-0">
          <svg className="h-5 w-5 text-yellow-400" viewBox="0 0 20 20" fill="currentColor">
            <path fillRule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clipRule="evenodd" />
          </svg>
        </div>
        <div className="ml-3">
          <p className="text-sm text-yellow-700">
            Tu email aún no ha sido verificado. Por favor, revisa tu bandeja de entrada y haz click en el enlace de verificación.
          </p>
          <p className="mt-3 text-sm">
            <button
              onClick={handleResendEmail}
              disabled={verificationLoading}
              className="font-medium text-yellow-700 underline hover:text-yellow-600 disabled:opacity-50"
            >
              {verificationLoading ? 'Enviando...' : 'Reenviar email de verificación'}
            </button>
          </p>
        </div>
      </div>
    </div>
  );
};

export default EmailVerificationNotice;
```

## Manejo de Estado Global

### Context de Autenticación

```javascript
// src/contexts/AuthContext.js
import React, { createContext, useState, useEffect } from 'react';
import { useQuery } from '@apollo/client';
import { GET_CURRENT_USER } from '../graphql/queries/auth';

export const AuthContext = createContext(null);

export const AuthProvider = ({ children }) => {
  const [user, setUser] = useState(null);
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [isLoading, setIsLoading] = useState(true);
  
  // Verificar si hay token al cargar
  useEffect(() => {
    const token = localStorage.getItem('accessToken');
    if (token) {
      setIsAuthenticated(true);
    }
    setIsLoading(false);
  }, []);
  
  // Obtener datos del usuario si está autenticado
  const { data, loading } = useQuery(GET_CURRENT_USER, {
    skip: !isAuthenticated,
    onCompleted: (data) => {
      if (data?.getCurrentUser) {
        setUser(data.getCurrentUser);
      }
    },
    onError: () => {
      // Si falla, limpiar autenticación
      localStorage.clear();
      setIsAuthenticated(false);
      setUser(null);
    }
  });
  
  const value = {
    user,
    setUser,
    isAuthenticated,
    setIsAuthenticated,
    isLoading: isLoading || loading
  };
  
  return (
    <AuthContext.Provider value={value}>
      {children}
    </AuthContext.Provider>
  );
};
```

### Context de Notificaciones

```javascript
// src/contexts/NotificationContext.js
import React, { createContext, useState, useCallback } from 'react';

export const NotificationContext = createContext(null);

export const NotificationProvider = ({ children }) => {
  const [notifications, setNotifications] = useState([]);
  
  const addNotification = useCallback((notification) => {
    const id = Date.now();
    const newNotification = { ...notification, id };
    
    setNotifications(prev => [...prev, newNotification]);
    
    // Auto-eliminar después del tiempo especificado
    if (notification.duration) {
      setTimeout(() => {
        removeNotification(id);
      }, notification.duration);
    }
  }, []);
  
  const removeNotification = useCallback((id) => {
    setNotifications(prev => prev.filter(n => n.id !== id));
  }, []);
  
  const value = {
    notifications,
    addNotification,
    removeNotification
  };
  
  return (
    <NotificationContext.Provider value={value}>
      {children}
      <NotificationContainer 
        notifications={notifications} 
        onRemove={removeNotification} 
      />
    </NotificationContext.Provider>
  );
};

// Componente para mostrar notificaciones
const NotificationContainer = ({ notifications, onRemove }) => {
  return (
    <div className="fixed top-4 right-4 z-50 space-y-2">
      {notifications.map(notification => (
        <Notification
          key={notification.id}
          notification={notification}
          onRemove={() => onRemove(notification.id)}
        />
      ))}
    </div>
  );
};

const Notification = ({ notification, onRemove }) => {
  const bgColor = {
    success: 'bg-green-500',
    error: 'bg-red-500',
    warning: 'bg-yellow-500',
    info: 'bg-blue-500'
  }[notification.type] || 'bg-gray-500';
  
  return (
    <div className={`${bgColor} text-white px-6 py-4 rounded-lg shadow-lg flex items-center justify-between min-w-[300px]`}>
      <span>{notification.message}</span>
      <button
        onClick={onRemove}
        className="ml-4 text-white hover:text-gray-200"
      >
        ×
      </button>
    </div>
  );
};
```

## Testing

### Testing de Componentes

```javascript
// src/components/members/__tests__/MembersList.test.js
import React from 'react';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import { MockedProvider } from '@apollo/client/testing';
import { MemoryRouter } from 'react-router-dom';
import MembersList from '../MembersList';
import { LIST_MEMBERS_QUERY } from '../../../graphql/queries/members';

const mocks = [
  {
    request: {
      query: LIST_MEMBERS_QUERY,
      variables: {
        filter: {
          pagination: { page: 1, pageSize: 20 }
        }
      }
    },
    result: {
      data: {
        listMembers: {
          nodes: [
            {
              miembro_id: '1',
              numero_socio: '2023-001',
              nombre: 'Juan',
              apellidos: 'Pérez',
              estado: 'ACTIVE',
              tipo_membresia: 'INDIVIDUAL',
              fecha_alta: '2023-01-15T00:00:00Z',
              correo_electronico: 'juan@ejemplo.com'
            },
            {
              miembro_id: '2',
              numero_socio: '2023-002',
              nombre: 'María',
              apellidos: 'García',
              estado: 'ACTIVE',
              tipo_membresia: 'FAMILY',
              fecha_alta: '2023-02-01T00:00:00Z',
              correo_electronico: 'maria@ejemplo.com'
            }
          ],
          pageInfo: {
            hasNextPage: false,
            hasPreviousPage: false,
            totalCount: 2
          }
        }
      }
    }
  }
];

const renderWithProviders = (component) => {
  return render(
    <MockedProvider mocks={mocks} addTypename={false}>
      <MemoryRouter>
        {component}
      </MemoryRouter>
    </MockedProvider>
  );
};

describe('MembersList', () => {
  it('renders loading state initially', () => {
    renderWithProviders(<MembersList />);
    expect(screen.getByTestId('loading-spinner')).toBeInTheDocument();
  });
  
  it('renders members list after loading', async () => {
    renderWithProviders(<MembersList />);
    
    await waitFor(() => {
      expect(screen.getByText('Juan Pérez')).toBeInTheDocument();
      expect(screen.getByText('María García')).toBeInTheDocument();
    });
  });
  
  it('shows correct member count', async () => {
    renderWithProviders(<MembersList />);
    
    await waitFor(() => {
      expect(screen.getByText('Total: 2 miembros')).toBeInTheDocument();
    });
  });
  
  it('filters members by status', async () => {
    renderWithProviders(<MembersList />);
    
    await waitFor(() => {
      expect(screen.getByText('Juan Pérez')).toBeInTheDocument();
    });
    
    // Cambiar filtro
    const statusSelect = screen.getByLabelText('Estado');
    fireEvent.change(statusSelect, { target: { value: 'INACTIVE' } });
    
    // Verificar que se hace nueva query
    await waitFor(() => {
      expect(screen.getByTestId('loading-spinner')).toBeInTheDocument();
    });
  });
});
```

### Testing de Hooks

```javascript
// src/hooks/__tests__/useAuth.test.js
import { renderHook, act } from '@testing-library/react';
import { MockedProvider } from '@apollo/client/testing';
import { useAuth } from '../useAuth';
import { AuthProvider } from '../../contexts/AuthContext';
import { LOGIN_MUTATION } from '../../graphql/mutations/auth';

const mockNavigate = jest.fn();
jest.mock('react-router-dom', () => ({
  ...jest.requireActual('react-router-dom'),
  useNavigate: () => mockNavigate
}));

const wrapper = ({ children }) => (
  <MockedProvider mocks={[]}>
    <AuthProvider>
      {children}
    </AuthProvider>
  </MockedProvider>
);

describe('useAuth', () => {
  beforeEach(() => {
    localStorage.clear();
    mockNavigate.mockClear();
  });
  
  it('initial state is not authenticated', () => {
    const { result } = renderHook(() => useAuth(), { wrapper });
    
    expect(result.current.isAuthenticated).toBe(false);
    expect(result.current.user).toBe(null);
  });
  
  it('login successful', async () => {
    const mocks = [
      {
        request: {
          query: LOGIN_MUTATION,
          variables: {
            input: { username: 'test@ejemplo.com', password: 'password123' }
          }
        },
        result: {
          data: {
            login: {
              user: {
                id: '1',
                username: 'test@ejemplo.com',
                role: 'user',
                isActive: true,
                lastLogin: null,
                emailVerified: true,
                emailVerifiedAt: new Date().toISOString()
              },
              accessToken: 'fake-access-token',
              refreshToken: 'fake-refresh-token',
              expiresAt: new Date(Date.now() + 900000).toISOString()
            }
          }
        }
      }
    ];
    
    const { result } = renderHook(() => useAuth(), { 
      wrapper: ({ children }) => (
        <MockedProvider mocks={mocks}>
          <AuthProvider>
            {children}
          </AuthProvider>
        </MockedProvider>
      )
    });
    
    await act(async () => {
      const response = await result.current.login('test@ejemplo.com', 'password123');
      expect(response.success).toBe(true);
      expect(response.emailVerified).toBe(true);
    });
    
    expect(localStorage.getItem('accessToken')).toBe('fake-access-token');
    expect(mockNavigate).toHaveBeenCalledWith('/dashboard');
  });
  
  it('login with unverified email', async () => {
    const mocks = [
      {
        request: {
          query: LOGIN_MUTATION,
          variables: {
            input: { username: 'test@ejemplo.com', password: 'password123' }
          }
        },
        result: {
          data: {
            login: {
              user: {
                id: '1',
                username: 'test@ejemplo.com',
                role: 'user',
                isActive: true,
                lastLogin: null,
                emailVerified: false,
                emailVerifiedAt: null
              },
              accessToken: 'fake-access-token',
              refreshToken: 'fake-refresh-token',
              expiresAt: new Date(Date.now() + 900000).toISOString()
            }
          }
        }
      }
    ];
    
    const { result } = renderHook(() => useAuth(), { 
      wrapper: ({ children }) => (
        <MockedProvider mocks={mocks}>
          <AuthProvider>
            {children}
          </AuthProvider>
        </MockedProvider>
      )
    });
    
    await act(async () => {
      const response = await result.current.login('test@ejemplo.com', 'password123');
      expect(response.success).toBe(true);
      expect(response.emailVerified).toBe(false);
    });
    
    expect(mockNavigate).toHaveBeenCalledWith('/verify-email-notice');
  });
});
```

Estos archivos proporcionan una base sólida para que los desarrolladores frontend puedan integrar el backend de ASAM con React. La estructura es modular, reutilizable y sigue las mejores prácticas de desarrollo.
