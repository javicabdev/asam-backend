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

// HTTP Link
const httpLink = createHttpLink({
  uri: import.meta.env.VITE_GRAPHQL_URL || 'http://localhost:8080/graphql',
});

// WebSocket Link (para futuras subscriptions)
const wsLink = new WebSocketLink({
  uri: import.meta.env.VITE_WS_URL || 'ws://localhost:8080/graphql',
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
          }
        }
      },
      Member: {
        keyFields: ["miembro_id"]
      },
      Family: {
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
│   └── payments/         # Componentes de pagos
├── composables/
│   ├── useAuth.js        # Composable de autenticación
│   ├── usePagination.js  # Composable de paginación
│   ├── useDebounce.js    # Composable de debounce
│   └── useNotification.js # Composable de notificaciones
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
│   ├── members/
│   ├── families/
│   └── payments/
├── plugins/
│   └── apollo.js
├── router/
│   └── index.js
├── stores/
│   ├── auth.js           # Store de autenticación
│   ├── members.js        # Store de miembros
│   └── notifications.js  # Store de notificaciones
├── utils/
│   ├── validators.js
│   ├── formatters.js
│   └── constants.js
└── App.vue
```

## Composables

### Composable de Autenticación

```javascript
// src/composables/useAuth.js
import { computed } from 'vue';
import { useRouter } from 'vue-router';
import { useMutation } from '@vue/apollo-composable';
import { useAuthStore } from '@/stores/auth';
import { LOGIN_MUTATION, LOGOUT_MUTATION, REFRESH_TOKEN_MUTATION } from '@/graphql/mutations/auth';

export function useAuth() {
  const router = useRouter();
  const authStore = useAuthStore();
  
  // Computed
  const user = computed(() => authStore.user);
  const isAuthenticated = computed(() => authStore.isAuthenticated);
  const isAdmin = computed(() => authStore.user?.role === 'ADMIN');
  const isUser = computed(() => authStore.user?.role === 'USER');
  
  // Mutations
  const { mutate: loginMutation, loading: loginLoading } = useMutation(LOGIN_MUTATION);
  const { mutate: logoutMutation } = useMutation(LOGOUT_MUTATION);
  const { mutate: refreshMutation } = useMutation(REFRESH_TOKEN_MUTATION);
  
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
      
      // Navegar según rol
      if (user.role === 'ADMIN') {
        await router.push('/admin/dashboard');
      } else {
        await router.push('/dashboard');
      }
      
      return { success: true };
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
  
  return {
    // State
    user,
    isAuthenticated,
    isAdmin,
    isUser,
    loginLoading,
    
    // Methods
    login,
    logout,
    refreshToken
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

### Componente de Formulario de Miembro

```vue
<!-- src/pages/members/MemberForm.vue -->
<template>
  <div class="member-form">
    <div class="header">
      <h1 class="title">
        {{ isEdit ? 'Editar Miembro' : 'Nuevo Miembro' }}
      </h1>
    </div>
    
    <form @submit.prevent="handleSubmit" class="form-container">
      <!-- Información Básica -->
      <fieldset class="fieldset">
        <legend class="legend">Información Básica</legend>
        
        <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
          <FormField
            v-model="form.numero_socio"
            label="Número de Socio"
            name="numero_socio"
            :error="errors.numero_socio"
            required
            placeholder="2024-001"
          />
          
          <FormField
            v-model="form.tipo_membresia"
            label="Tipo de Membresía"
            name="tipo_membresia"
            type="select"
            :options="membershipTypes"
            :error="errors.tipo_membresia"
            required
          />
          
          <FormField
            v-model="form.nombre"
            label="Nombre"
            name="nombre"
            :error="errors.nombre"
            required
          />
          
          <FormField
            v-model="form.apellidos"
            label="Apellidos"
            name="apellidos"
            :error="errors.apellidos"
            required
          />
          
          <FormField
            v-model="form.documento_identidad"
            label="DNI/NIE"
            name="documento_identidad"
            :error="errors.documento_identidad"
            placeholder="12345678X"
            @input="form.documento_identidad = $event.toUpperCase()"
          />
          
          <FormField
            v-model="form.fecha_nacimiento"
            label="Fecha de Nacimiento"
            name="fecha_nacimiento"
            type="date"
            :error="errors.fecha_nacimiento"
            :max="maxDate"
          />
          
          <FormField
            v-model="form.correo_electronico"
            label="Email"
            name="correo_electronico"
            type="email"
            :error="errors.correo_electronico"
            placeholder="ejemplo@email.com"
          />
          
          <FormField
            v-model="form.profesion"
            label="Profesión"
            name="profesion"
            :error="errors.profesion"
          />
        </div>
      </fieldset>
      
      <!-- Dirección -->
      <fieldset class="fieldset">
        <legend class="legend">Dirección</legend>
        
        <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
          <div class="md:col-span-2">
            <FormField
              v-model="form.calle_numero_piso"
              label="Calle, Número, Piso"
              name="calle_numero_piso"
              :error="errors.calle_numero_piso"
              required
              placeholder="Calle Principal 123, 2º A"
            />
          </div>
          
          <FormField
            v-model="form.codigo_postal"
            label="Código Postal"
            name="codigo_postal"
            :error="errors.codigo_postal"
            required
            placeholder="07001"
            maxlength="5"
          />
          
          <FormField
            v-model="form.poblacion"
            label="Población"
            name="poblacion"
            :error="errors.poblacion"
            required
          />
          
          <FormField
            v-model="form.provincia"
            label="Provincia"
            name="provincia"
            :error="errors.provincia"
          />
          
          <FormField
            v-model="form.pais"
            label="País"
            name="pais"
            :error="errors.pais"
          />
        </div>
      </fieldset>
      
      <!-- Información Adicional -->
      <fieldset class="fieldset">
        <legend class="legend">Información Adicional</legend>
        
        <div class="space-y-6">
          <FormField
            v-model="form.nacionalidad"
            label="Nacionalidad"
            name="nacionalidad"
            :error="errors.nacionalidad"
          />
          
          <FormField
            v-model="form.observaciones"
            label="Observaciones"
            name="observaciones"
            type="textarea"
            :error="errors.observaciones"
            rows="4"
            placeholder="Notas adicionales sobre el miembro..."
          />
        </div>
      </fieldset>
      
      <!-- Botones -->
      <div class="form-actions">
        <button
          type="button"
          @click="handleCancel"
          class="btn-secondary"
          :disabled="loading"
        >
          Cancelar
        </button>
        
        <button
          type="submit"
          class="btn-primary"
          :disabled="loading || !isValid"
        >
          {{ loading ? 'Guardando...' : (isEdit ? 'Actualizar' : 'Crear') }}
        </button>
      </div>
    </form>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useQuery, useMutation } from '@vue/apollo-composable';
import { useNotification } from '@/composables/useNotification';
import { 
  GET_MEMBER_QUERY, 
  CREATE_MEMBER_MUTATION, 
  UPDATE_MEMBER_MUTATION 
} from '@/graphql/members';
import { memberValidationSchema } from '@/utils/validators';
import FormField from '@/components/common/FormField.vue';

// Props & Setup
const route = useRoute();
const router = useRouter();
const { showSuccess, showError } = useNotification();

// State
const isEdit = computed(() => !!route.params.id);
const memberId = computed(() => route.params.id);
const loading = ref(false);
const errors = reactive({});

// Form data
const form = reactive({
  numero_socio: '',
  tipo_membresia: 'INDIVIDUAL',
  nombre: '',
  apellidos: '',
  calle_numero_piso: '',
  codigo_postal: '',
  poblacion: '',
  provincia: '',
  pais: 'España',
  fecha_nacimiento: null,
  documento_identidad: '',
  correo_electronico: '',
  profesion: '',
  nacionalidad: 'Española',
  observaciones: ''
});

// Constants
const membershipTypes = [
  { value: 'INDIVIDUAL', label: 'Individual' },
  { value: 'FAMILY', label: 'Familiar' }
];

const maxDate = new Date().toISOString().split('T')[0];

// Load member data if editing
if (isEdit.value) {
  const { onResult, onError } = useQuery(
    GET_MEMBER_QUERY,
    { id: memberId.value },
    { fetchPolicy: 'network-only' }
  );
  
  onResult(({ data }) => {
    if (data?.getMember) {
      Object.assign(form, {
        ...data.getMember,
        fecha_nacimiento: data.getMember.fecha_nacimiento?.split('T')[0] || null
      });
    }
  });
  
  onError(() => {
    showError('Error al cargar los datos del miembro');
    router.push('/members');
  });
}

// Mutations
const { mutate: createMember } = useMutation(CREATE_MEMBER_MUTATION);
const { mutate: updateMember } = useMutation(UPDATE_MEMBER_MUTATION);

// Validation
const isValid = computed(() => {
  // Simple validation check
  return form.numero_socio && 
         form.nombre && 
         form.apellidos && 
         form.calle_numero_piso && 
         form.codigo_postal && 
         form.poblacion;
});

const validate = async () => {
  try {
    await memberValidationSchema.validate(form, { abortEarly: false });
    Object.keys(errors).forEach(key => delete errors[key]);
    return true;
  } catch (error) {
    error.inner.forEach(err => {
      errors[err.path] = err.message;
    });
    return false;
  }
};

// Methods
const handleSubmit = async () => {
  if (!await validate()) return;
  
  loading.value = true;
  
  try {
    const input = {
      ...form,
      fecha_nacimiento: form.fecha_nacimiento 
        ? new Date(form.fecha_nacimiento).toISOString() 
        : null
    };
    
    if (isEdit.value) {
      await updateMember({
        variables: {
          input: {
            miembro_id: memberId.value,
            ...input
          }
        }
      });
      showSuccess('Miembro actualizado correctamente');
    } else {
      await createMember({
        variables: { input }
      });
      showSuccess('Miembro creado correctamente');
    }
    
    router.push('/members');
  } catch (error) {
    const message = error.graphQLErrors?.[0]?.message || 'Error al guardar';
    showError(message);
    
    // Manejar errores de validación del servidor
    if (error.graphQLErrors?.[0]?.extensions?.code === 'VALIDATION_ERROR') {
      const serverErrors = error.graphQLErrors[0].extensions.details;
      Object.assign(errors, serverErrors);
    }
  } finally {
    loading.value = false;
  }
};

const handleCancel = () => {
  router.push('/members');
};
</script>

<style scoped>
.member-form {
  @apply max-w-4xl mx-auto space-y-6;
}

.header {
  @apply mb-8;
}

.title {
  @apply text-3xl font-bold text-gray-900;
}

.form-container {
  @apply bg-white shadow rounded-lg p-6 space-y-8;
}

.fieldset {
  @apply space-y-6;
}

.legend {
  @apply text-lg font-medium text-gray-900 mb-4;
}

.form-actions {
  @apply flex justify-end space-x-4 pt-6 border-t;
}
</style>
```

## Gestión de Estado con Pinia

### Store de Autenticación

```javascript
// src/stores/auth.js
import { defineStore } from 'pinia';
import { apolloClient } from '@/apollo/client';
import { GET_CURRENT_USER } from '@/graphql/queries/auth';

export const useAuthStore = defineStore('auth', {
  state: () => ({
    user: null,
    accessToken: localStorage.getItem('accessToken') || null,
    refreshToken: localStorage.getItem('refreshToken') || null,
    tokenExpiresAt: localStorage.getItem('tokenExpiresAt') || null,
    isAuthenticated: !!localStorage.getItem('accessToken')
  }),
  
  getters: {
    isAdmin: (state) => state.user?.role === 'ADMIN',
    isUser: (state) => state.user?.role === 'USER',
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
                user: { role: 'ADMIN' },
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
