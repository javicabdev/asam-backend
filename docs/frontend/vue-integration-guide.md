# Guía de Integración con Vue 3

Esta guía muestra cómo integrar el backend de ASAM con una aplicación Vue 3 usando Composition API.

## Tabla de Contenidos
1. [Configuración Inicial](#configuración-inicial)
2. [Estructura del Proyecto](#estructura-del-proyecto)
3. [Composables](#composables)
4. [Componentes](#componentes)
5. [Gestión de Estado con Pinia](#gestión-de-estado-con-pinia)
6. [Testing](#testing)

## Configuración Inicial

### 1. Instalar Dependencias

```bash
npm install @apollo/client @vue/apollo-composable graphql
npm install pinia
npm install --save-dev @graphql-codegen/cli @graphql-codegen/typescript @graphql-codegen/typescript-operations
```

### 2. Configurar Apollo Client

```javascript
// src/apollo/client.js
import { ApolloClient, InMemoryCache, createHttpLink, split } from '@apollo/client/core';
import { setContext } from '@apollo/client/link/context';
import { onError } from '@apollo/client/link/error';
import { getMainDefinition } from '@apollo/client/utilities';
import { WebSocketLink } from '@apollo/client/link/ws';
import { useAuthStore } from '@/stores/auth';

// Endpoints
const GRAPHQL_ENDPOINT = import.meta.env.PROD
  ? 'https://asam-backend-jtpswzdxuq-ew.a.run.app/graphql'
  : 'http://localhost:8080/graphql';

const WS_ENDPOINT = import.meta.env.PROD
  ? 'wss://asam-backend-jtpswzdxuq-ew.a.run.app/graphql'
  : 'ws://localhost:8080/graphql';

// HTTP Link
const httpLink = createHttpLink({
  uri: GRAPHQL_ENDPOINT,
});

// WebSocket Link (para futuras subscriptions)
const wsLink = new WebSocketLink({
  uri: WS_ENDPOINT,
  options: {
    reconnect: true,
    connectionParams: () => {
      const authStore = useAuthStore();
      return {
        authToken: authStore.accessToken,
      };
    },
  },
});

// Auth Link
const authLink = setContext((_, { headers }) => {
  const authStore = useAuthStore();
  return {
    headers: {
      ...headers,
      authorization: authStore.accessToken ? `Bearer ${authStore.accessToken}` : "",
    }
  };
});

// Error Link
const errorLink = onError(({ graphQLErrors, networkError, operation, forward }) => {
  if (graphQLErrors) {
    for (let err of graphQLErrors) {
      switch (err.extensions?.code) {
        case 'TOKEN_EXPIRED':
          // Renovar token
          const authStore = useAuthStore();
          return authStore.refreshToken().then(() => forward(operation));
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

// Split para WebSocket o HTTP
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

// Apollo Client
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
      User: {
        keyFields: ["id"]
      }
    }
  }),
  defaultOptions: {
    watchQuery: {
      fetchPolicy: 'cache-and-network',
    },
    query: {
      fetchPolicy: 'network-only',
    },
  },
});
```

### 3. Plugin de Apollo

```javascript
// src/plugins/apollo.js
import { DefaultApolloClient } from '@vue/apollo-composable';
import { apolloClient } from '@/apollo/client';

export function installApollo(app) {
  app.provide(DefaultApolloClient, apolloClient);
}
```

### 4. Main.js

```javascript
// src/main.js
import { createApp } from 'vue';
import { createPinia } from 'pinia';
import router from './router';
import { installApollo } from './plugins/apollo';
import App from './App.vue';
import './assets/main.css';

const app = createApp(App);

// Instalar Pinia
app.use(createPinia());

// Instalar Router
app.use(router);

// Instalar Apollo
installApollo(app);

app.mount('#app');
```

## Estructura del Proyecto

```
src/
├── apollo/
│   └── client.js         # Configuración de Apollo Client
├── components/
│   ├── common/           # Componentes reutilizables
│   ├── members/          # Componentes de miembros
│   ├── families/         # Componentes de familias
│   ├── payments/         # Componentes de pagos
│   ├── users/            # Componentes de usuarios (Admin)
│   └── auth/             # Componentes de autenticación
├── composables/
│   ├── useAuth.js        # Composable de autenticación
│   ├── usePagination.js  # Composable de paginación
│   ├── useDebounce.js    # Composable de debounce
│   ├── useNotification.js # Composable de notificaciones
│   └── useUser.js        # Composable de gestión de usuarios
├── graphql/
│   ├── mutations/        # Archivos .graphql con mutations
│   ├── queries/          # Archivos .graphql con queries
│   └── fragments/        # Fragmentos reutilizables
├── layouts/
│   ├── DefaultLayout.vue
│   └── AuthLayout.vue
├── pages/
│   ├── Login.vue
│   ├── Dashboard.vue
│   ├── Profile.vue
│   ├── ForgotPassword.vue
│   ├── ResetPassword.vue
│   ├── VerifyEmail.vue
│   ├── members/
│   ├── families/
│   ├── payments/
│   └── admin/
│       ├── users/
│       │   ├── index.vue
│       │   ├── [id].vue
│       │   └── create.vue
│       └── dashboard.vue
├── plugins/
│   └── apollo.js
├── router/
│   └── index.js
├── stores/
│   ├── auth.js           # Store de autenticación
│   ├── members.js        # Store de miembros
│   ├── users.js          # Store de usuarios
│   └── notifications.js  # Store de notificaciones
├── utils/
│   ├── validators.js
│   ├── formatters.js
│   └── constants.js
└── App.vue
```

## Composables

### Composable de Autenticación (Actualizado)

```javascript
// src/composables/useAuth.js
import { computed } from 'vue';
import { useRouter } from 'vue-router';
import { useMutation } from '@vue/apollo-composable';
import { useAuthStore } from '@/stores/auth';
import { 
  LOGIN_MUTATION, 
  LOGOUT_MUTATION, 
  REFRESH_TOKEN_MUTATION,
  CHANGE_PASSWORD_MUTATION,
  REQUEST_PASSWORD_RESET_MUTATION,
  SEND_VERIFICATION_EMAIL_MUTATION
} from '@/graphql/mutations/auth';

export function useAuth() {
  const router = useRouter();
  const authStore = useAuthStore();
  
  // Computed
  const user = computed(() => authStore.user);
  const isAuthenticated = computed(() => authStore.isAuthenticated);
  const isAdmin = computed(() => authStore.user?.role === 'admin');
  const isUser = computed(() => authStore.user?.role === 'user');
  const isEmailVerified = computed(() => authStore.user?.emailVerified || false);
  
  // Mutations
  const { mutate: loginMutation, loading: loginLoading } = useMutation(LOGIN_MUTATION);
  const { mutate: logoutMutation } = useMutation(LOGOUT_MUTATION);
  const { mutate: refreshMutation } = useMutation(REFRESH_TOKEN_MUTATION);
  const { mutate: changePasswordMutation, loading: changePasswordLoading } = useMutation(CHANGE_PASSWORD_MUTATION);
  const { mutate: requestResetMutation, loading: resetLoading } = useMutation(REQUEST_PASSWORD_RESET_MUTATION);
  const { mutate: sendVerificationMutation, loading: verificationLoading } = useMutation(SEND_VERIFICATION_EMAIL_MUTATION);
  
  // Methods
  const login = async (username, password) => {
    try {
      const { data } = await loginMutation({
        variables: { input: { username, password } }
      });
      
      const { user, accessToken, refreshToken, expiresAt } = data.login;
      
      // Actualizar store
      authStore.setAuth({
        user,
        accessToken,
        refreshToken,
        expiresAt
      });
      
      // Verificar si el email está verificado
      if (!user.emailVerified) {
        await router.push('/verify-email-notice');
        return { success: true, emailVerified: false };
      }
      
      // Navegar según rol
      if (user.role === 'admin') {
        await router.push('/admin/dashboard');
      } else {
        await router.push('/dashboard');
      }
      
      return { success: true, emailVerified: true };
    } catch (error) {
      console.error('Login error:', error);
      return { 
        success: false, 
        error: error.graphQLErrors?.[0]?.message || 'Error al iniciar sesión' 
      };
    }
  };
  
  const logout = async () => {
    try {
      await logoutMutation();
    } catch (error) {
      console.error('Logout error:', error);
    } finally {
      authStore.clearAuth();
      await router.push('/login');
    }
  };
  
  const refreshToken = async () => {
    const refreshToken = authStore.refreshToken;
    if (!refreshToken) {
      throw new Error('No refresh token');
    }
    
    try {
      const { data } = await refreshMutation({
        variables: { input: { refreshToken } }
      });
      
      const { accessToken, refreshToken: newRefreshToken, expiresAt } = data.refreshToken;
      
      authStore.updateTokens({
        accessToken,
        refreshToken: newRefreshToken,
        expiresAt
      });
      
      return true;
    } catch (error) {
      console.error('Refresh token error:', error);
      await logout();
      return false;
    }
  };
  
  const changePassword = async (currentPassword, newPassword) => {
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
  };
  
  const requestPasswordReset = async (email) => {
    try {
      const { data } = await requestResetMutation({
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
  };
  
  const sendVerificationEmail = async () => {
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
  };
  
  return {
    // State
    user,
    isAuthenticated,
    isAdmin,
    isUser,
    isEmailVerified,
    loginLoading,
    changePasswordLoading,
    resetLoading,
    verificationLoading,
    
    // Methods
    login,
    logout,
    refreshToken,
    changePassword,
    requestPasswordReset,
    sendVerificationEmail
  };
}
```

### Composable de Gestión de Usuarios (Admin)

```javascript
// src/composables/useUser.js
import { useMutation, useQuery } from '@vue/apollo-composable';
import {
  CREATE_USER_MUTATION,
  UPDATE_USER_MUTATION,
  DELETE_USER_MUTATION,
  RESET_USER_PASSWORD_MUTATION
} from '@/graphql/mutations/users';
import {
  GET_USER_QUERY,
  LIST_USERS_QUERY
} from '@/graphql/queries/users';

export function useUser() {
  const { mutate: createUserMutation, loading: createLoading } = useMutation(CREATE_USER_MUTATION);
  const { mutate: updateUserMutation, loading: updateLoading } = useMutation(UPDATE_USER_MUTATION);
  const { mutate: deleteUserMutation, loading: deleteLoading } = useMutation(DELETE_USER_MUTATION);
  const { mutate: resetPasswordMutation, loading: resetLoading } = useMutation(RESET_USER_PASSWORD_MUTATION);
  
  const createUser = async (input) => {
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
  };
  
  const updateUser = async (input) => {
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
  };
  
  const deleteUser = async (id) => {
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
  };
  
  const resetUserPassword = async (userId, newPassword) => {
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
  };
  
  const useUsersList = (page = 1, pageSize = 10) => {
    return useQuery(LIST_USERS_QUERY, { page, pageSize }, {
      fetchPolicy: 'network-only'
    });
  };
  
  const useUserDetail = (id) => {
    return useQuery(GET_USER_QUERY, { id }, {
      skip: !id,
      fetchPolicy: 'network-only'
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
}
```

### Composable de Paginación

```javascript
// src/composables/usePagination.js
import { ref, computed, watch } from 'vue';

export function usePagination(initialPageSize = 20) {
  const page = ref(1);
  const pageSize = ref(initialPageSize);
  
  const paginationVariables = computed(() => ({
    page: page.value,
    pageSize: pageSize.value
  }));
  
  const goToPage = (newPage) => {
    page.value = Math.max(1, newPage);
  };
  
  const nextPage = () => {
    page.value++;
  };
  
  const previousPage = () => {
    page.value = Math.max(1, page.value - 1);
  };
  
  const resetPage = () => {
    page.value = 1;
  };
  
  const changePageSize = (newPageSize) => {
    pageSize.value = newPageSize;
    page.value = 1; // Reset a primera página
  };
  
  const getPaginationInfo = (pageInfo) => {
    if (!pageInfo) {
      return {
        totalPages: 0,
        startIndex: 0,
        endIndex: 0,
        canGoNext: false,
        canGoPrevious: false
      };
    }
    
    const totalPages = Math.ceil(pageInfo.totalCount / pageSize.value);
    const startIndex = (page.value - 1) * pageSize.value + 1;
    const endIndex = Math.min(page.value * pageSize.value, pageInfo.totalCount);
    
    return {
      totalPages,
      startIndex,
      endIndex,
      canGoNext: pageInfo.hasNextPage,
      canGoPrevious: pageInfo.hasPreviousPage
    };
  };
  
  return {
    // State
    page,
    pageSize,
    paginationVariables,
    
    // Methods
    goToPage,
    nextPage,
    previousPage,
    resetPage,
    changePageSize,
    getPaginationInfo
  };
}
```

### Composable de Debounce

```javascript
// src/composables/useDebounce.js
import { ref, watch } from 'vue';

export function useDebounce(initialValue = '', delay = 300) {
  const value = ref(initialValue);
  const debouncedValue = ref(initialValue);
  let timeoutId = null;
  
  watch(value, (newValue) => {
    if (timeoutId) {
      clearTimeout(timeoutId);
    }
    
    timeoutId = setTimeout(() => {
      debouncedValue.value = newValue;
    }, delay);
  });
  
  return {
    value,
    debouncedValue
  };
}
```

### Composable de Notificaciones

```javascript
// src/composables/useNotification.js
import { useNotificationStore } from '@/stores/notifications';

export function useNotification() {
  const notificationStore = useNotificationStore();
  
  const showSuccess = (message, duration = 5000) => {
    notificationStore.add({
      type: 'success',
      message,
      duration
    });
  };
  
  const showError = (message, duration = 7000) => {
    notificationStore.add({
      type: 'error',
      message,
      duration
    });
  };
  
  const showWarning = (message, duration = 5000) => {
    notificationStore.add({
      type: 'warning',
      message,
      duration
    });
  };
  
  const showInfo = (message, duration = 5000) => {
    notificationStore.add({
      type: 'info',
      message,
      duration
    });
  };
  
  return {
    notifications: notificationStore.notifications,
    showSuccess,
    showError,
    showWarning,
    showInfo,
    remove: notificationStore.remove
  };
}
```

## Componentes

### Componente de Lista de Miembros

```vue
<!-- src/pages/members/MembersList.vue -->
<template>
  <div class="members-list">
    <div class="header">
      <h1 class="title">Gestión de Miembros</h1>
      <router-link 
        v-if="isAdmin" 
        to="/members/new" 
        class="btn-primary"
      >
        Nuevo Miembro
      </router-link>
    </div>
    
    <!-- Filtros -->
    <div class="filters">
      <div class="search-box">
        <input
          v-model="searchTerm"
          type="text"
          placeholder="Buscar por nombre, apellidos, email o DNI..."
          class="search-input"
        />
      </div>
      
      <div class="filter-group">
        <select v-model="statusFilter" class="filter-select">
          <option value="">Todos los estados</option>
          <option value="ACTIVE">Activos</option>
          <option value="INACTIVE">Inactivos</option>
        </select>
        
        <select v-model="typeFilter" class="filter-select">
          <option value="">Todos los tipos</option>
          <option value="INDIVIDUAL">Individual</option>
          <option value="FAMILY">Familiar</option>
        </select>
      </div>
      
      <button @click="exportToCSV" class="btn-secondary">
        Exportar CSV
      </button>
    </div>
    
    <!-- Tabla -->
    <DataTable
      :columns="columns"
      :data="members"
      :loading="loading"
      :sort-config="sortConfig"
      @sort="handleSort"
    >
      <template #actions="{ row }">
        <div class="actions">
          <router-link 
            :to="`/members/${row.miembro_id}`"
            class="action-link"
          >
            Ver
          </router-link>
          <router-link 
            v-if="isAdmin"
            :to="`/members/${row.miembro_id}/edit`"
            class="action-link"
          >
            Editar
          </router-link>
          <button
            v-if="isAdmin"
            @click="handleStatusChange(row)"
            class="action-button"
            :class="row.estado === 'ACTIVE' ? 'text-red-600' : 'text-green-600'"
          >
            {{ row.estado === 'ACTIVE' ? 'Desactivar' : 'Activar' }}
          </button>
        </div>
      </template>
    </DataTable>
    
    <!-- Paginación -->
    <Pagination
      v-if="pageInfo"
      :page="page"
      :page-size="pageSize"
      :total-count="pageInfo.totalCount"
      :has-next-page="pageInfo.hasNextPage"
      :has-previous-page="pageInfo.hasPreviousPage"
      @page-change="goToPage"
      @page-size-change="changePageSize"
    />
  </div>
</template>

<script setup>
import { ref, computed, watch } from 'vue';
import { useQuery, useMutation } from '@vue/apollo-composable';
import { useRouter } from 'vue-router';
import { useAuth } from '@/composables/useAuth';
import { usePagination } from '@/composables/usePagination';
import { useDebounce } from '@/composables/useDebounce';
import { useNotification } from '@/composables/useNotification';
import { LIST_MEMBERS_QUERY, CHANGE_MEMBER_STATUS_MUTATION } from '@/graphql/members';
import DataTable from '@/components/common/DataTable.vue';
import Pagination from '@/components/common/Pagination.vue';

// Composables
const router = useRouter();
const { isAdmin } = useAuth();
const { showSuccess, showError } = useNotification();
const { 
  page, 
  pageSize, 
  paginationVariables, 
  goToPage, 
  changePageSize,
  resetPage 
} = usePagination();

// State
const searchTerm = ref('');
const { debouncedValue: debouncedSearchTerm } = useDebounce(searchTerm.value, 300);
const statusFilter = ref('');
const typeFilter = ref('');
const sortConfig = ref({ field: 'NUMERO_SOCIO', direction: 'ASC' });

// Computed filter
const filter = computed(() => ({
  ...paginationVariables.value,
  search_term: debouncedSearchTerm.value,
  estado: statusFilter.value || undefined,
  tipo_membresia: typeFilter.value || undefined,
  sort: sortConfig.value
}));

// Query
const { result, loading, refetch } = useQuery(
  LIST_MEMBERS_QUERY,
  () => ({ filter: filter.value }),
  {
    fetchPolicy: 'cache-and-network',
    notifyOnNetworkStatusChange: true
  }
);

// Computed data
const members = computed(() => result.value?.listMembers?.nodes || []);
const pageInfo = computed(() => result.value?.listMembers?.pageInfo);

// Mutations
const { mutate: changeMemberStatus } = useMutation(CHANGE_MEMBER_STATUS_MUTATION);

// Watchers para resetear página
watch([statusFilter, typeFilter, debouncedSearchTerm], () => {
  resetPage();
});

// Columns configuration
const columns = [
  { 
    field: 'numero_socio', 
    label: 'Nº Socio', 
    sortable: true 
  },
  { 
    field: 'nombre', 
    label: 'Nombre', 
    sortable: true 
  },
  { 
    field: 'apellidos', 
    label: 'Apellidos', 
    sortable: true 
  },
  { 
    field: 'correo_electronico', 
    label: 'Email' 
  },
  { 
    field: 'estado', 
    label: 'Estado',
    render: (row) => ({
      text: row.estado === 'ACTIVE' ? 'Activo' : 'Inactivo',
      class: row.estado === 'ACTIVE' ? 'text-green-600' : 'text-red-600'
    })
  },
  { 
    field: 'tipo_membresia', 
    label: 'Tipo',
    render: (row) => row.tipo_membresia === 'INDIVIDUAL' ? 'Individual' : 'Familiar'
  },
  { 
    field: 'actions', 
    label: 'Acciones',
    slot: 'actions'
  }
];

// Methods
const handleSort = (field) => {
  const newDirection = 
    sortConfig.value.field === field && sortConfig.value.direction === 'ASC' 
      ? 'DESC' 
      : 'ASC';
  
  sortConfig.value = { field, direction: newDirection };
};

const handleStatusChange = async (member) => {
  const newStatus = member.estado === 'ACTIVE' ? 'INACTIVE' : 'ACTIVE';
  const action = member.estado === 'ACTIVE' ? 'desactivar' : 'activar';
  
  if (!confirm(`¿Estás seguro de ${action} a ${member.nombre} ${member.apellidos}?`)) {
    return;
  }
  
  try {
    await changeMemberStatus({
      variables: {
        id: member.miembro_id,
        status: newStatus
      }
    });
    
    showSuccess(`Miembro ${action === 'desactivar' ? 'desactivado' : 'activado'} correctamente`);
    refetch();
  } catch (error) {
    showError(`Error al ${action} el miembro`);
  }
};

const exportToCSV = async () => {
  // Implementar exportación
  try {
    // Obtener todos los datos sin paginación
    const allData = await apolloClient.query({
      query: LIST_MEMBERS_QUERY,
      variables: {
        filter: {
          ...filter.value,
          pagination: { page: 1, pageSize: 10000 }
        }
      }
    });
    
    // Convertir a CSV y descargar
    // ... implementación de CSV
    
    showSuccess('Datos exportados correctamente');
  } catch (error) {
    showError('Error al exportar los datos');
  }
};
</script>

<style scoped>
.members-list {
  @apply space-y-6;
}

.header {
  @apply flex justify-between items-center;
}

.title {
  @apply text-3xl font-bold text-gray-900;
}

.filters {
  @apply bg-white p-4 rounded-lg shadow space-y-4 md:space-y-0 md:flex md:items-center md:space-x-4;
}

.search-box {
  @apply flex-1;
}

.search-input {
  @apply w-full px-4 py-2 border border-gray-300 rounded-md focus:ring-indigo-500 focus:border-indigo-500;
}

.filter-group {
  @apply flex space-x-2;
}

.filter-select {
  @apply px-4 py-2 border border-gray-300 rounded-md focus:ring-indigo-500 focus:border-indigo-500;
}

.actions {
  @apply flex items-center space-x-2;
}

.action-link {
  @apply text-indigo-600 hover:text-indigo-900;
}

.action-button {
  @apply font-medium hover:underline;
}

.btn-primary {
  @apply inline-flex items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500;
}

.btn-secondary {
  @apply inline-flex items-center px-4 py-2 border border-gray-300 rounded-md shadow-sm text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500;
}
</style>
```

### Componente de Gestión de Usuarios (Admin)

```vue
<!-- src/pages/admin/users/index.vue -->
<template>
  <div class="users-management">
    <div class="header">
      <h1 class="title">Gestión de Usuarios</h1>
      <router-link to="/admin/users/create" class="btn-primary">
        Nuevo Usuario
      </router-link>
    </div>
    
    <!-- Lista de usuarios -->
    <div class="users-table">
      <DataTable
        :columns="columns"
        :data="users"
        :loading="loading"
        :pagination="paginationVariables"
        :page-info="pageInfo"
        @page-change="goToPage"
      >
        <template #role="{ row }">
          <span 
            class="role-badge"
            :class="row.role === 'admin' ? 'role-admin' : 'role-user'"
          >
            {{ row.role === 'admin' ? 'Administrador' : 'Usuario' }}
          </span>
        </template>
        
        <template #status="{ row }">
          <span 
            class="status-badge"
            :class="row.isActive ? 'status-active' : 'status-inactive'"
          >
            {{ row.isActive ? 'Activo' : 'Inactivo' }}
          </span>
        </template>
        
        <template #verified="{ row }">
          <span v-if="row.emailVerified" class="text-green-600">
            ✓ Verificado
          </span>
          <span v-else class="text-gray-500">
            No verificado
          </span>
        </template>
        
        <template #actions="{ row }">
          <div class="actions">
            <router-link 
              :to="`/admin/users/${row.id}`"
              class="action-link"
            >
              Ver
            </router-link>
            <button
              @click="handleResetPassword(row)"
              class="action-link"
            >
              Resetear Contraseña
            </button>
            <button
              @click="handleToggleStatus(row)"
              class="action-button"
              :class="row.isActive ? 'text-red-600' : 'text-green-600'"
            >
              {{ row.isActive ? 'Desactivar' : 'Activar' }}
            </button>
            <button
              v-if="row.id !== user?.id"
              @click="handleDelete(row)"
              class="action-button text-red-600"
            >
              Eliminar
            </button>
          </div>
        </template>
      </DataTable>
    </div>
    
    <!-- Modal de reset de contraseña -->
    <PasswordResetModal
      v-if="resetModalUser"
      :user="resetModalUser"
      @close="resetModalUser = null"
      @confirm="confirmResetPassword"
    />
  </div>
</template>

<script setup>
import { ref, computed } from 'vue';
import { useRouter } from 'vue-router';
import { useAuth } from '@/composables/useAuth';
import { useUser } from '@/composables/useUser';
import { usePagination } from '@/composables/usePagination';
import { useNotification } from '@/composables/useNotification';
import DataTable from '@/components/common/DataTable.vue';
import PasswordResetModal from '@/components/admin/PasswordResetModal.vue';

// Composables
const router = useRouter();
const { user } = useAuth();
const { showSuccess, showError } = useNotification();
const { 
  useUsersList, 
  updateUser, 
  deleteUser, 
  resetUserPassword,
  updateLoading,
  deleteLoading,
  resetLoading
} = useUser();

const { 
  page, 
  pageSize, 
  paginationVariables, 
  goToPage 
} = usePagination(10);

// Data
const resetModalUser = ref(null);

// Query
const { result, loading, refetch } = useUsersList(page.value, pageSize.value);

// Computed
const users = computed(() => result.value?.listUsers || []);
const pageInfo = computed(() => ({
  totalCount: users.value.length,
  hasNextPage: users.value.length === pageSize.value,
  hasPreviousPage: page.value > 1
}));

// Columns
const columns = [
  { field: 'username', label: 'Usuario' },
  { field: 'role', label: 'Rol', slot: 'role' },
  { field: 'status', label: 'Estado', slot: 'status' },
  { field: 'verified', label: 'Email', slot: 'verified' },
  { field: 'lastLogin', label: 'Último acceso', render: (row) => 
    row.lastLogin ? new Date(row.lastLogin).toLocaleString('es-ES') : 'Nunca'
  },
  { field: 'actions', label: 'Acciones', slot: 'actions' }
];

// Methods
const handleToggleStatus = async (targetUser) => {
  const newStatus = !targetUser.isActive;
  const action = newStatus ? 'activar' : 'desactivar';
  
  if (!confirm(`¿Estás seguro de ${action} al usuario ${targetUser.username}?`)) {
    return;
  }
  
  const result = await updateUser({
    id: targetUser.id,
    isActive: newStatus
  });
  
  if (result.success) {
    showSuccess(`Usuario ${action === 'activar' ? 'activado' : 'desactivado'} correctamente`);
    refetch();
  } else {
    showError(result.error);
  }
};

const handleResetPassword = (targetUser) => {
  resetModalUser.value = targetUser;
};

const confirmResetPassword = async (newPassword) => {
  if (!resetModalUser.value) return;
  
  const result = await resetUserPassword(resetModalUser.value.id, newPassword);
  
  if (result.success) {
    showSuccess('Contraseña reseteada correctamente');
    resetModalUser.value = null;
  } else {
    showError(result.error);
  }
};

const handleDelete = async (targetUser) => {
  if (!confirm(`¿Estás seguro de eliminar al usuario ${targetUser.username}? Esta acción no se puede deshacer.`)) {
    return;
  }
  
  const result = await deleteUser(targetUser.id);
  
  if (result.success) {
    showSuccess('Usuario eliminado correctamente');
    refetch();
  } else {
    showError(result.error);
  }
};
</script>

<style scoped>
.users-management {
  @apply space-y-6;
}

.header {
  @apply flex justify-between items-center;
}

.title {
  @apply text-3xl font-bold text-gray-900;
}

.users-table {
  @apply bg-white shadow rounded-lg overflow-hidden;
}

.role-badge {
  @apply px-2 py-1 text-xs font-medium rounded-full;
}

.role-admin {
  @apply bg-purple-100 text-purple-800;
}

.role-user {
  @apply bg-gray-100 text-gray-800;
}

.status-badge {
  @apply px-2 py-1 text-xs font-medium rounded-full;
}

.status-active {
  @apply bg-green-100 text-green-800;
}

.status-inactive {
  @apply bg-red-100 text-red-800;
}

.actions {
  @apply flex items-center space-x-2;
}

.action-link {
  @apply text-indigo-600 hover:text-indigo-900 cursor-pointer;
}

.action-button {
  @apply font-medium hover:underline cursor-pointer;
}
</style>
```

### Componente de Verificación de Email

```vue
<!-- src/components/auth/EmailVerificationNotice.vue -->
<template>
  <div v-if="!isEmailVerified" class="email-verification-notice">
    <div class="notice-icon">
      <svg class="h-5 w-5 text-yellow-400" viewBox="0 0 20 20" fill="currentColor">
        <path fill-rule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clip-rule="evenodd" />
      </svg>
    </div>
    <div class="notice-content">
      <p class="notice-text">
        Tu email aún no ha sido verificado. Por favor, revisa tu bandeja de entrada y haz click en el enlace de verificación.
      </p>
      <p class="notice-action">
        <button
          @click="handleResendEmail"
          :disabled="verificationLoading"
          class="resend-button"
        >
          {{ verificationLoading ? 'Enviando...' : 'Reenviar email de verificación' }}
        </button>
      </p>
    </div>
  </div>
</template>

<script setup>
import { useAuth } from '@/composables/useAuth';
import { useNotification } from '@/composables/useNotification';

const { isEmailVerified, sendVerificationEmail, verificationLoading } = useAuth();
const { showSuccess, showError } = useNotification();

const handleResendEmail = async () => {
  const result = await sendVerificationEmail();
  
  if (result.success) {
    showSuccess(result.message || 'Email de verificación enviado');
  } else {
    showError(result.error || 'Error al enviar el email');
  }
};
</script>

<style scoped>
.email-verification-notice {
  @apply bg-yellow-50 border-l-4 border-yellow-400 p-4 mb-6 flex;
}

.notice-icon {
  @apply flex-shrink-0;
}

.notice-content {
  @apply ml-3;
}

.notice-text {
  @apply text-sm text-yellow-700;
}

.notice-action {
  @apply mt-3 text-sm;
}

.resend-button {
  @apply font-medium text-yellow-700 underline hover:text-yellow-600 disabled:opacity-50 cursor-pointer;
}
</style>
```

## Gestión de Estado con Pinia

### Store de Autenticación (Actualizado)

```javascript
// src/stores/auth.js
import { defineStore } from 'pinia';
import { apolloClient } from '@/apollo/client';
import { GET_CURRENT_USER } from '@/graphql/queries/auth';
import { REFRESH_TOKEN_MUTATION } from '@/graphql/mutations/auth';

export const useAuthStore = defineStore('auth', {
  state: () => ({
    user: null,
    accessToken: localStorage.getItem('accessToken') || null,
    refreshToken: localStorage.getItem('refreshToken') || null,
    tokenExpiresAt: localStorage.getItem('tokenExpiresAt') || null,
    isAuthenticated: !!localStorage.getItem('accessToken')
  }),
  
  getters: {
    isAdmin: (state) => state.user?.role === 'admin',
    isUser: (state) => state.user?.role === 'user',
    isEmailVerified: (state) => state.user?.emailVerified || false,
    tokenExpired: (state) => {
      if (!state.tokenExpiresAt) return true;
      return new Date() > new Date(state.tokenExpiresAt);
    }
  },
  
  actions: {
    setAuth({ user, accessToken, refreshToken, expiresAt }) {
      this.user = user;
      this.accessToken = accessToken;
      this.refreshToken = refreshToken;
      this.tokenExpiresAt = expiresAt;
      this.isAuthenticated = true;
      
      // Guardar en localStorage
      localStorage.setItem('accessToken', accessToken);
      localStorage.setItem('refreshToken', refreshToken);
      localStorage.setItem('tokenExpiresAt', expiresAt);
    },
    
    updateTokens({ accessToken, refreshToken, expiresAt }) {
      this.accessToken = accessToken;
      this.refreshToken = refreshToken;
      this.tokenExpiresAt = expiresAt;
      
      localStorage.setItem('accessToken', accessToken);
      localStorage.setItem('refreshToken', refreshToken);
      localStorage.setItem('tokenExpiresAt', expiresAt);
    },
    
    clearAuth() {
      this.user = null;
      this.accessToken = null;
      this.refreshToken = null;
      this.tokenExpiresAt = null;
      this.isAuthenticated = false;
      
      localStorage.removeItem('accessToken');
      localStorage.removeItem('refreshToken');
      localStorage.removeItem('tokenExpiresAt');
    },
    
    async fetchCurrentUser() {
      if (!this.isAuthenticated) return;
      
      try {
        const { data } = await apolloClient.query({
          query: GET_CURRENT_USER,
          fetchPolicy: 'network-only'
        });
        
        if (data?.getCurrentUser) {
          this.user = data.getCurrentUser;
        }
      } catch (error) {
        console.error('Error fetching current user:', error);
        this.clearAuth();
      }
    },
    
    async refreshToken() {
      if (!this.refreshToken) {
        throw new Error('No refresh token available');
      }
      
      try {
        const { data } = await apolloClient.mutate({
          mutation: REFRESH_TOKEN_MUTATION,
          variables: { input: { refreshToken: this.refreshToken } }
        });
        
        const { accessToken, refreshToken, expiresAt } = data.refreshToken;
        
        this.updateTokens({ accessToken, refreshToken, expiresAt });
        
        return accessToken;
      } catch (error) {
        this.clearAuth();
        throw error;
      }
    },
    
    async checkTokenExpiry() {
      if (!this.tokenExpiresAt) return;
      
      const expiryTime = new Date(this.tokenExpiresAt).getTime();
      const currentTime = new Date().getTime();
      const timeUntilExpiry = expiryTime - currentTime;
      
      // Renovar 1 minuto antes de expirar
      if (timeUntilExpiry < 60000) {
        await this.refreshToken();
      }
    }
  }
});
```

### Store de Notificaciones

```javascript
// src/stores/notifications.js
import { defineStore } from 'pinia';

export const useNotificationStore = defineStore('notifications', {
  state: () => ({
    notifications: []
  }),
  
  actions: {
    add(notification) {
      const id = Date.now();
      const newNotification = { ...notification, id };
      
      this.notifications.push(newNotification);
      
      // Auto-eliminar después del tiempo especificado
      if (notification.duration) {
        setTimeout(() => {
          this.remove(id);
        }, notification.duration);
      }
      
      return id;
    },
    
    remove(id) {
      const index = this.notifications.findIndex(n => n.id === id);
      if (index > -1) {
        this.notifications.splice(index, 1);
      }
    },
    
    clear() {
      this.notifications = [];
    }
  }
});
```

## Testing

### Testing de Componentes

```javascript
// src/pages/members/__tests__/MembersList.spec.js
import { mount } from '@vue/test-utils';
import { createTestingPinia } from '@pinia/testing';
import { vi } from 'vitest';
import MembersList from '../MembersList.vue';
import { LIST_MEMBERS_QUERY } from '@/graphql/members';

// Mock Apollo
const mockQuery = vi.fn();
vi.mock('@vue/apollo-composable', () => ({
  useQuery: () => mockQuery(),
  useMutation: () => ({ mutate: vi.fn() })
}));

// Mock Router
const mockPush = vi.fn();
vi.mock('vue-router', () => ({
  useRouter: () => ({ push: mockPush }),
  RouterLink: { template: '<a><slot /></a>' }
}));

describe('MembersList', () => {
  let wrapper;
  
  const mockMembers = [
    {
      miembro_id: '1',
      numero_socio: '2024-001',
      nombre: 'Juan',
      apellidos: 'Pérez',
      estado: 'ACTIVE',
      tipo_membresia: 'INDIVIDUAL',
      correo_electronico: 'juan@example.com'
    },
    {
      miembro_id: '2',
      numero_socio: '2024-002',
      nombre: 'María',
      apellidos: 'García',
      estado: 'INACTIVE',
      tipo_membresia: 'FAMILY',
      correo_electronico: 'maria@example.com'
    }
  ];
  
  beforeEach(() => {
    mockQuery.mockReturnValue({
      result: {
        value: {
          listMembers: {
            nodes: mockMembers,
            pageInfo: {
              hasNextPage: false,
              hasPreviousPage: false,
              totalCount: 2
            }
          }
        }
      },
      loading: false,
      refetch: vi.fn()
    });
    
    wrapper = mount(MembersList, {
      global: {
        plugins: [
          createTestingPinia({
            initialState: {
              auth: {
                user: { role: 'admin' },
                isAuthenticated: true
              }
            }
          })
        ],
        stubs: {
          DataTable: true,
          Pagination: true
        }
      }
    });
  });
  
  afterEach(() => {
    wrapper.unmount();
  });
  
  it('renders member list', () => {
    expect(wrapper.find('.title').text()).toBe('Gestión de Miembros');
    expect(wrapper.find('.btn-primary').exists()).toBe(true);
  });
  
  it('shows new member button for admin', () => {
    const newButton = wrapper.find('.btn-primary');
    expect(newButton.text()).toBe('Nuevo Miembro');
  });
  
  it('filters members when search term changes', async () => {
    const searchInput = wrapper.find('.search-input');
    await searchInput.setValue('Juan');
    
    // Esperar debounce
    await new Promise(resolve => setTimeout(resolve, 400));
    
    expect(mockQuery).toHaveBeenCalled();
  });
  
  it('changes status filter', async () => {
    const statusSelect = wrapper.find('select').find('option[value="ACTIVE"]');
    await statusSelect.setSelected();
    
    expect(wrapper.vm.statusFilter).toBe('ACTIVE');
  });
});
```

### Testing de Composables

```javascript
// src/composables/__tests__/usePagination.spec.js
import { describe, it, expect } from 'vitest';
import { usePagination } from '../usePagination';

describe('usePagination', () => {
  it('initializes with default values', () => {
    const { page, pageSize } = usePagination();
    
    expect(page.value).toBe(1);
    expect(pageSize.value).toBe(20);
  });
  
  it('goes to specific page', () => {
    const { page, goToPage } = usePagination();
    
    goToPage(5);
    expect(page.value).toBe(5);
    
    goToPage(0);
    expect(page.value).toBe(1); // No puede ser menor que 1
  });
  
  it('navigates between pages', () => {
    const { page, nextPage, previousPage } = usePagination();
    
    nextPage();
    expect(page.value).toBe(2);
    
    nextPage();
    expect(page.value).toBe(3);
    
    previousPage();
    expect(page.value).toBe(2);
    
    previousPage();
    previousPage();
    expect(page.value).toBe(1); // No puede ser menor que 1
  });
  
  it('changes page size and resets to page 1', () => {
    const { page, pageSize, changePageSize, nextPage } = usePagination();
    
    nextPage();
    nextPage();
    expect(page.value).toBe(3);
    
    changePageSize(50);
    expect(pageSize.value).toBe(50);
    expect(page.value).toBe(1); // Se resetea a página 1
  });
  
  it('calculates pagination info correctly', () => {
    const { getPaginationInfo } = usePagination(10);
    
    const pageInfo = {
      totalCount: 95,
      hasNextPage: true,
      hasPreviousPage: false
    };
    
    const info = getPaginationInfo(pageInfo);
    
    expect(info.totalPages).toBe(10);
    expect(info.startIndex).toBe(1);
    expect(info.endIndex).toBe(10);
    expect(info.canGoNext).toBe(true);
    expect(info.canGoPrevious).toBe(false);
  });
});
```

Esta documentación proporciona una guía completa para integrar el backend de ASAM con Vue 3, incluyendo mejores prácticas, componentes reutilizables y ejemplos de testing.
