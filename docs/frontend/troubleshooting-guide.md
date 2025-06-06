# Guía de Troubleshooting y Problemas Comunes

Esta guía ayuda a resolver los problemas más frecuentes al desarrollar aplicaciones frontend que consumen el backend de ASAM.

## Tabla de Contenidos
1. [Problemas de Conexión](#problemas-de-conexión)
2. [Errores de Autenticación](#errores-de-autenticación)
3. [Problemas con GraphQL](#problemas-con-graphql)
4. [Errores de CORS](#errores-de-cors)
5. [Problemas de Performance](#problemas-de-performance)
6. [Errores de Build](#errores-de-build)
7. [Problemas de Estado](#problemas-de-estado)
8. [Debugging Avanzado](#debugging-avanzado)

## Problemas de Conexión

### 1. Error: "Failed to fetch" o "Network Error"

**Síntomas:**
- Las peticiones fallan inmediatamente
- Console muestra errores de red
- No se puede conectar al backend

**Soluciones:**

```javascript
// 1. Verificar que el backend esté corriendo
// En terminal:
curl http://localhost:8080/health

// 2. Verificar la URL en las variables de entorno
console.log('GraphQL URL:', process.env.REACT_APP_GRAPHQL_URL);

// 3. Verificar configuración de Apollo Client
const httpLink = createHttpLink({
  uri: process.env.REACT_APP_GRAPHQL_URL || 'http://localhost:8080/graphql',
  // Añadir timeout
  fetchOptions: {
    timeout: 10000 // 10 segundos
  }
});

// 4. Implementar retry logic
import { RetryLink } from '@apollo/client/link/retry';

const retryLink = new RetryLink({
  delay: {
    initial: 300,
    max: Infinity,
    jitter: true
  },
  attempts: {
    max: 5,
    retryIf: (error, _operation) => {
      // Reintentar en errores de red
      return !!error && error.networkError?.statusCode === 500;
    }
  }
});
```

### 2. WebSocket Connection Failed

**Síntomas:**
- Subscriptions no funcionan
- Console muestra "WebSocket connection failed"

**Soluciones:**

```javascript
// 1. Verificar URL de WebSocket
const wsUrl = process.env.REACT_APP_WS_URL || 'ws://localhost:8080/graphql';

// 2. Configurar WebSocket con reconexión
import { GraphQLWsLink } from '@apollo/client/link/subscriptions';
import { createClient } from 'graphql-ws';

const wsLink = new GraphQLWsLink(
  createClient({
    url: wsUrl,
    connectionParams: () => ({
      authToken: localStorage.getItem('accessToken'),
    }),
    // Reconexión automática
    shouldRetry: () => true,
    retryAttempts: 5,
    retryWait: async (retryCount) => {
      // Exponential backoff
      await new Promise(resolve => 
        setTimeout(resolve, Math.min(1000 * 2 ** retryCount, 30000))
      );
    },
    on: {
      connected: () => console.log('WebSocket connected'),
      error: (error) => console.error('WebSocket error:', error),
      closed: () => console.log('WebSocket closed')
    }
  })
);
```

## Errores de Autenticación

### 1. Token Expired

**Síntomas:**
- Error 401 Unauthorized
- Usuario deslogueado inesperadamente

**Soluciones:**

```javascript
// 1. Implementar refresh token automático
import { setContext } from '@apollo/client/link/context';
import { onError } from '@apollo/client/link/error';
import { Observable } from '@apollo/client';

let isRefreshing = false;
let pendingRequests = [];

const resolvePendingRequests = () => {
  pendingRequests.map(callback => callback());
  pendingRequests = [];
};

const errorLink = onError(({ graphQLErrors, operation, forward }) => {
  if (graphQLErrors) {
    for (let err of graphQLErrors) {
      if (err.extensions.code === 'UNAUTHENTICATED' || 
          err.extensions.code === 'TOKEN_EXPIRED') {
        
        // Si ya estamos refrescando, encolar la petición
        if (isRefreshing) {
          return new Observable(observer => {
            pendingRequests.push(() => {
              forward(operation).subscribe(observer);
            });
          });
        }
        
        isRefreshing = true;
        
        return new Observable(observer => {
          refreshToken()
            .then(newToken => {
              // Actualizar token en el storage
              localStorage.setItem('accessToken', newToken);
              
              // Actualizar header de la operación
              operation.setContext({
                headers: {
                  ...operation.getContext().headers,
                  authorization: `Bearer ${newToken}`
                }
              });
              
              // Resolver peticiones pendientes
              resolvePendingRequests();
              
              // Reintentar la operación original
              forward(operation).subscribe(observer);
            })
            .catch(error => {
              // Refresh falló, hacer logout
              pendingRequests = [];
              authService.logout();
              observer.error(error);
            })
            .finally(() => {
              isRefreshing = false;
            });
        });
      }
    }
  }
});

// 2. Función de refresh token robusta
async function refreshToken() {
  const refreshToken = localStorage.getItem('refreshToken');
  
  if (!refreshToken) {
    throw new Error('No refresh token available');
  }
  
  try {
    const response = await fetch(`${API_URL}/graphql`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        query: `
          mutation RefreshToken($input: RefreshTokenInput!) {
            refreshToken(input: $input) {
              accessToken
              refreshToken
              expiresAt
            }
          }
        `,
        variables: {
          input: { refreshToken }
        }
      })
    });
    
    const data = await response.json();
    
    if (data.errors) {
      throw new Error('Refresh token failed');
    }
    
    const { accessToken, refreshToken: newRefreshToken } = data.data.refreshToken;
    
    localStorage.setItem('accessToken', accessToken);
    localStorage.setItem('refreshToken', newRefreshToken);
    
    return accessToken;
  } catch (error) {
    console.error('Refresh token error:', error);
    throw error;
  }
}
```

### 2. Invalid Credentials

**Síntomas:**
- Login falla con "Invalid credentials"
- Usuario no puede acceder

**Soluciones:**

```javascript
// 1. Verificar formato de credenciales
const validateCredentials = (username, password) => {
  const errors = {};
  
  // Validar email
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
  if (!emailRegex.test(username)) {
    errors.username = 'Email inválido';
  }
  
  // Validar contraseña
  if (password.length < 8) {
    errors.password = 'La contraseña debe tener al menos 8 caracteres';
  }
  
  return errors;
};

// 2. Manejar errores específicos
const handleLoginError = (error) => {
  if (error.graphQLErrors?.length > 0) {
    const code = error.graphQLErrors[0].extensions?.code;
    
    switch (code) {
      case 'INVALID_CREDENTIALS':
        return 'Email o contraseña incorrectos';
      case 'ACCOUNT_LOCKED':
        return 'Tu cuenta ha sido bloqueada. Contacta al administrador.';
      case 'ACCOUNT_INACTIVE':
        return 'Tu cuenta está inactiva.';
      default:
        return 'Error al iniciar sesión. Por favor, intenta nuevamente.';
    }
  }
  
  return 'Error de conexión. Verifica tu internet.';
};
```

## Problemas con GraphQL

### 1. Query Variables Undefined

**Síntomas:**
- Error "Variable $id of required type ID! was not provided"
- Queries fallan con variables undefined

**Soluciones:**

```javascript
// 1. Validar variables antes de ejecutar query
const MEMBER_QUERY = gql`
  query GetMember($id: ID!) {
    getMember(id: $id) {
      miembro_id
      nombre
    }
  }
`;

function useMember(id) {
  const { data, loading, error } = useQuery(MEMBER_QUERY, {
    variables: { id },
    // Skip si no hay ID
    skip: !id,
    // O manejar con onError
    onError: (error) => {
      console.error('Query error:', error);
    }
  });
  
  // Validación adicional
  useEffect(() => {
    if (!id) {
      console.warn('Member ID is required');
    }
  }, [id]);
  
  return { data, loading, error };
}

// 2. Usar default values
function useMemberList(filters = {}) {
  const defaultFilters = {
    pagination: { page: 1, pageSize: 20 },
    sort: { field: 'NOMBRE', direction: 'ASC' },
    ...filters
  };
  
  return useQuery(LIST_MEMBERS_QUERY, {
    variables: { filter: defaultFilters }
  });
}
```

### 2. Cache Issues

**Síntomas:**
- Datos no se actualizan después de mutation
- Queries devuelven datos obsoletos

**Soluciones:**

```javascript
// 1. Configurar cache correctamente
const cache = new InMemoryCache({
  typePolicies: {
    Query: {
      fields: {
        listMembers: {
          // Merge para paginación
          keyArgs: ['filter', ['estado', 'tipo_membresia']],
          merge(existing = { nodes: [] }, incoming) {
            return {
              ...incoming,
              nodes: [...existing.nodes, ...incoming.nodes]
            };
          }
        }
      }
    },
    Member: {
      keyFields: ['miembro_id'],
      fields: {
        // Campo computado
        fullName: {
          read(_, { readField }) {
            const nombre = readField('nombre');
            const apellidos = readField('apellidos');
            return `${nombre} ${apellidos}`;
          }
        }
      }
    }
  }
});

// 2. Actualizar cache después de mutation
const [createMember] = useMutation(CREATE_MEMBER_MUTATION, {
  update(cache, { data: { createMember } }) {
    // Leer query existente
    const existing = cache.readQuery({
      query: LIST_MEMBERS_QUERY,
      variables: { filter: {} }
    });
    
    if (existing) {
      // Escribir nueva data
      cache.writeQuery({
        query: LIST_MEMBERS_QUERY,
        variables: { filter: {} },
        data: {
          listMembers: {
            ...existing.listMembers,
            nodes: [createMember, ...existing.listMembers.nodes]
          }
        }
      });
    }
  },
  // O usar refetchQueries
  refetchQueries: [
    { query: LIST_MEMBERS_QUERY, variables: { filter: {} } }
  ]
});

// 3. Limpiar cache cuando sea necesario
const clearMemberCache = () => {
  client.cache.evict({ 
    id: 'ROOT_QUERY',
    fieldName: 'listMembers'
  });
  client.cache.gc();
};
```

## Errores de CORS

### 1. CORS Policy Block

**Síntomas:**
- "Access to fetch at 'http://localhost:8080' from origin 'http://localhost:3000' has been blocked by CORS policy"

**Soluciones:**

```javascript
// 1. Configurar proxy en desarrollo (React)
// package.json
{
  "proxy": "http://localhost:8080"
}

// 2. O usar setupProxy.js
// src/setupProxy.js
const { createProxyMiddleware } = require('http-proxy-middleware');

module.exports = function(app) {
  app.use(
    '/graphql',
    createProxyMiddleware({
      target: 'http://localhost:8080',
      changeOrigin: true,
      ws: true, // Para WebSocket
      logLevel: 'debug'
    })
  );
};

// 3. Para Vite
// vite.config.js
export default {
  server: {
    proxy: {
      '/graphql': {
        target: 'http://localhost:8080',
        changeOrigin: true,
        secure: false,
        ws: true
      }
    }
  }
};

// 4. Headers en Apollo Client
const httpLink = createHttpLink({
  uri: '/graphql', // Usar ruta relativa con proxy
  credentials: 'include', // Para cookies
  headers: {
    'Content-Type': 'application/json',
  }
});
```

## Problemas de Performance

### 1. Renders Excesivos

**Síntomas:**
- Componentes se renderizan múltiples veces
- UI lenta o con lag

**Soluciones:**

```javascript
// 1. Usar React.memo
const MemberCard = React.memo(({ member }) => {
  console.log('MemberCard render:', member.miembro_id);
  
  return (
    <div className="member-card">
      <h3>{member.nombre} {member.apellidos}</h3>
    </div>
  );
}, (prevProps, nextProps) => {
  // Comparación personalizada
  return prevProps.member.miembro_id === nextProps.member.miembro_id &&
         prevProps.member.nombre === nextProps.member.nombre;
});

// 2. Optimizar queries
const { data } = useQuery(MEMBER_QUERY, {
  // Fetch policy para evitar re-fetches innecesarios
  fetchPolicy: 'cache-first',
  // Next fetch policy después del primer fetch
  nextFetchPolicy: 'cache-only',
  // Evitar polling innecesario
  pollInterval: 0,
  // No refetch on window focus
  refetchOnWindowFocus: false
});

// 3. Usar useMemo y useCallback
function MemberList({ members }) {
  // Memoizar cálculos costosos
  const sortedMembers = useMemo(() => {
    return [...members].sort((a, b) => 
      a.nombre.localeCompare(b.nombre)
    );
  }, [members]);
  
  // Memoizar callbacks
  const handleClick = useCallback((id) => {
    console.log('Clicked:', id);
  }, []);
  
  return sortedMembers.map(member => (
    <MemberCard 
      key={member.miembro_id}
      member={member}
      onClick={handleClick}
    />
  ));
}
```

### 2. Bundle Size Grande

**Síntomas:**
- Carga inicial lenta
- Bundle size > 500KB

**Soluciones:**

```javascript
// 1. Lazy loading de rutas
const MemberManagement = lazy(() => 
  import(/* webpackChunkName: "members" */ './pages/MemberManagement')
);

const PaymentManagement = lazy(() => 
  import(/* webpackChunkName: "payments" */ './pages/PaymentManagement')
);

// 2. Tree shaking de imports
// ❌ Evitar
import _ from 'lodash';
const result = _.debounce(fn, 300);

// ✅ Mejor
import debounce from 'lodash/debounce';
const result = debounce(fn, 300);

// 3. Analizar bundle
// package.json
{
  "scripts": {
    "analyze": "source-map-explorer 'build/static/js/*.js'",
    "bundle-report": "webpack-bundle-analyzer build/stats.json"
  }
}
```

## Errores de Build

### 1. Module Not Found

**Síntomas:**
- "Module not found: Can't resolve..."
- Build falla

**Soluciones:**

```bash
# 1. Limpiar cache
rm -rf node_modules
rm package-lock.json
npm install

# 2. Verificar case sensitivity
# Windows no es case-sensitive, pero el CI sí
# Asegurarse de que los imports coincidan exactamente

# 3. Verificar alias de webpack
// webpack.config.js
module.exports = {
  resolve: {
    alias: {
      '@': path.resolve(__dirname, 'src'),
      '@components': path.resolve(__dirname, 'src/components'),
      '@utils': path.resolve(__dirname, 'src/utils')
    }
  }
};

# 4. Para TypeScript
// tsconfig.json
{
  "compilerOptions": {
    "baseUrl": ".",
    "paths": {
      "@/*": ["src/*"],
      "@components/*": ["src/components/*"]
    }
  }
}
```

### 2. Out of Memory

**Síntomas:**
- "JavaScript heap out of memory"
- Build crash

**Soluciones:**

```bash
# 1. Aumentar memoria para Node
NODE_OPTIONS=--max_old_space_size=4096 npm run build

# 2. En package.json
{
  "scripts": {
    "build": "node --max_old_space_size=4096 scripts/build.js"
  }
}

# 3. Optimizar imports de desarrollo
if (process.env.NODE_ENV === 'development') {
  // Solo importar en desarrollo
  require('./devTools');
}
```

## Problemas de Estado

### 1. Estado Perdido en Navegación

**Síntomas:**
- Estado se pierde al cambiar de ruta
- Formularios se resetean

**Soluciones:**

```javascript
// 1. Persistir estado importante
import { persistReducer } from 'redux-persist';
import storage from 'redux-persist/lib/storage';

const persistConfig = {
  key: 'root',
  storage,
  whitelist: ['auth', 'user'], // Solo persistir estos
  blacklist: ['form', 'ui'] // No persistir estos
};

const persistedReducer = persistReducer(persistConfig, rootReducer);

// 2. Usar Context para estado compartido
const FormStateContext = createContext();

export function FormStateProvider({ children }) {
  const [formState, setFormState] = useState({});
  
  return (
    <FormStateContext.Provider value={{ formState, setFormState }}>
      {children}
    </FormStateContext.Provider>
  );
}

// 3. Cache de Apollo para estado
const cache = new InMemoryCache({
  typePolicies: {
    Query: {
      fields: {
        // Campo local para UI state
        uiState: {
          read() {
            return {
              sidebarOpen: true,
              theme: 'light'
            };
          }
        }
      }
    }
  }
});
```

### 2. Race Conditions

**Síntomas:**
- Datos incorrectos después de navegación rápida
- Estados inconsistentes

**Soluciones:**

```javascript
// 1. Cancelar requests en cleanup
function useMemberData(id) {
  const [data, setData] = useState(null);
  
  useEffect(() => {
    let cancelled = false;
    
    async function fetchData() {
      try {
        const result = await api.getMember(id);
        if (!cancelled) {
          setData(result);
        }
      } catch (error) {
        if (!cancelled) {
          console.error(error);
        }
      }
    }
    
    fetchData();
    
    return () => {
      cancelled = true;
    };
  }, [id]);
  
  return data;
}

// 2. Usar AbortController
function useFetch(url) {
  const [data, setData] = useState(null);
  
  useEffect(() => {
    const controller = new AbortController();
    
    fetch(url, { signal: controller.signal })
      .then(res => res.json())
      .then(data => setData(data))
      .catch(err => {
        if (err.name !== 'AbortError') {
          console.error(err);
        }
      });
    
    return () => controller.abort();
  }, [url]);
  
  return data;
}
```

## Debugging Avanzado

### 1. Memory Leaks

**Detectar memory leaks:**

```javascript
// 1. Usar Chrome DevTools Memory Profiler
// - Tomar snapshot inicial
// - Realizar acciones
// - Tomar snapshot final
// - Comparar objetos retenidos

// 2. Detectar listeners no limpiados
class LeakDetector {
  constructor() {
    this.listeners = new Map();
  }
  
  trackListener(component, event, listener) {
    const key = `${component}-${event}`;
    if (!this.listeners.has(key)) {
      this.listeners.set(key, []);
    }
    this.listeners.get(key).push(listener);
  }
  
  removeListener(component, event, listener) {
    const key = `${component}-${event}`;
    const listeners = this.listeners.get(key);
    if (listeners) {
      const index = listeners.indexOf(listener);
      if (index > -1) {
        listeners.splice(index, 1);
      }
    }
  }
  
  checkLeaks() {
    this.listeners.forEach((listeners, key) => {
      if (listeners.length > 0) {
        console.warn(`Potential leak: ${key} has ${listeners.length} listeners`);
      }
    });
  }
}

// 3. Hook para detectar memory leaks
function useLeakDetection(componentName) {
  useEffect(() => {
    const startMemory = performance.memory?.usedJSHeapSize;
    
    return () => {
      const endMemory = performance.memory?.usedJSHeapSize;
      const diff = endMemory - startMemory;
      
      if (diff > 1000000) { // 1MB
        console.warn(`Potential memory leak in ${componentName}: ${diff} bytes`);
      }
    };
  }, [componentName]);
}
```

### 2. Performance Profiling

```javascript
// 1. React Profiler API
import { Profiler } from 'react';

function onRenderCallback(
  id, // ID del Profiler
  phase, // "mount" o "update"
  actualDuration, // Tiempo de render
  baseDuration, // Tiempo estimado sin memoización
  startTime, // Cuando React empezó
  commitTime, // Cuando React commitió
  interactions // Set de interacciones
) {
  console.log(`${id} (${phase}) took ${actualDuration}ms`);
  
  // Alertar si es muy lento
  if (actualDuration > 16) { // Más de un frame
    console.warn(`Slow render in ${id}: ${actualDuration}ms`);
  }
}

<Profiler id="MemberList" onRender={onRenderCallback}>
  <MemberList />
</Profiler>

// 2. Custom performance marks
function measureComponentPerformance(componentName) {
  const startMark = `${componentName}-start`;
  const endMark = `${componentName}-end`;
  const measureName = `${componentName}-render`;
  
  performance.mark(startMark);
  
  return () => {
    performance.mark(endMark);
    performance.measure(measureName, startMark, endMark);
    
    const measure = performance.getEntriesByName(measureName)[0];
    console.log(`${componentName} rendered in ${measure.duration}ms`);
    
    // Limpiar
    performance.clearMarks(startMark);
    performance.clearMarks(endMark);
    performance.clearMeasures(measureName);
  };
}
```

### 3. Network Debugging

```javascript
// 1. Interceptar todas las requests
if (process.env.NODE_ENV === 'development') {
  // Interceptar fetch
  const originalFetch = window.fetch;
  window.fetch = function(...args) {
    console.log('Fetch:', args[0]);
    return originalFetch.apply(this, args)
      .then(response => {
        console.log('Response:', response.status);
        return response;
      })
      .catch(error => {
        console.error('Fetch error:', error);
        throw error;
      });
  };
  
  // Interceptar XHR
  const originalXHR = window.XMLHttpRequest;
  window.XMLHttpRequest = function() {
    const xhr = new originalXHR();
    
    xhr.addEventListener('load', function() {
      console.log('XHR Load:', this.responseURL, this.status);
    });
    
    xhr.addEventListener('error', function() {
      console.error('XHR Error:', this.responseURL);
    });
    
    return xhr;
  };
}

// 2. Debug específico de GraphQL
window.debugGraphQL = {
  // Log todas las queries
  logQueries: true,
  
  // Log solo errores
  logErrors: true,
  
  // Interceptar query específica
  interceptQuery: (operationName) => {
    const link = new ApolloLink((operation, forward) => {
      if (operation.operationName === operationName) {
        console.log('Intercepted:', operation);
        debugger; // Breakpoint
      }
      return forward(operation);
    });
    
    // Añadir al client
    client.setLink(link.concat(client.link));
  }
};
```

## Herramientas de Debugging

### Browser Extensions
- React Developer Tools
- Apollo Client Devtools
- Redux DevTools
- Vue.js devtools

### VS Code Extensions
- Debugger for Chrome
- Apollo GraphQL
- Error Lens
- GitLens

### Comandos Útiles

```bash
# Limpiar todo y reinstalar
npm run clean:all

# Verificar dependencias
npm ls

# Encontrar dependencias duplicadas
npm dedupe

# Verificar vulnerabilidades
npm audit

# Actualizar dependencias
npm update

# Ver qué paquetes están desactualizados
npm outdated
```

Esta guía cubre los problemas más comunes y sus soluciones al desarrollar aplicaciones frontend con el backend de ASAM.