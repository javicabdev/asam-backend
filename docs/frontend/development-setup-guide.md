# Guía de Configuración del Entorno de Desarrollo

Esta guía proporciona instrucciones detalladas para configurar el entorno de desarrollo local para trabajar con el backend de ASAM.

## Tabla de Contenidos
1. [Requisitos Previos](#requisitos-previos)
2. [Configuración Inicial](#configuración-inicial)
3. [Desarrollo con React](#desarrollo-con-react)
4. [Desarrollo con Vue](#desarrollo-con-vue)
5. [Herramientas de Desarrollo](#herramientas-de-desarrollo)
6. [Configuración de VS Code](#configuración-de-vs-code)
7. [Debugging](#debugging)
8. [Tips y Trucos](#tips-y-trucos)

## Requisitos Previos

### Software Necesario
- **Node.js**: v18.0.0 o superior
- **npm**: v8.0.0 o superior (o yarn/pnpm)
- **Git**: v2.30.0 o superior
- **Editor**: VS Code recomendado

### Verificar Instalaciones
```bash
# Verificar Node.js
node --version
# Output: v18.x.x

# Verificar npm
npm --version
# Output: 8.x.x

# Verificar Git
git --version
# Output: git version 2.x.x
```

## Configuración Inicial

### 1. Clonar el Proyecto

```bash
# Clonar el repositorio
git clone https://github.com/javicabdev/asam-frontend.git
cd asam-frontend

# Instalar dependencias
npm install

# Copiar archivo de configuración
cp .env.example .env.development
```

### 2. Configuración de Variables de Entorno

```bash
# .env.development
REACT_APP_API_URL=http://localhost:8080
REACT_APP_GRAPHQL_URL=http://localhost:8080/graphql
REACT_APP_WS_URL=ws://localhost:8080/ws
REACT_APP_ENVIRONMENT=development

# Features flags
REACT_APP_ENABLE_ANALYTICS=false
REACT_APP_ENABLE_SENTRY=false
REACT_APP_ENABLE_MOCK_DATA=true

# Development tools
REACT_APP_ENABLE_DEVTOOLS=true
REACT_APP_APOLLO_DEVTOOLS=true
REACT_APP_REDUX_DEVTOOLS=true
```

### 3. Configuración del Backend Local

```bash
# En otra terminal, levantar el backend
cd ../asam-backend
docker-compose up -d

# Verificar que esté funcionando
curl http://localhost:8080/health
# Output: {"status":"healthy"}
```

## Desarrollo con React

### 1. Estructura del Proyecto React

```
asam-frontend/
├── public/
│   ├── index.html
│   └── assets/
├── src/
│   ├── components/
│   │   ├── common/
│   │   ├── features/
│   │   └── layouts/
│   ├── hooks/
│   ├── pages/
│   ├── services/
│   ├── utils/
│   ├── graphql/
│   │   ├── queries/
│   │   ├── mutations/
│   │   └── fragments/
│   ├── styles/
│   ├── App.js
│   └── index.js
├── .env.development
├── .eslintrc.js
├── .prettierrc
└── package.json
```

### 2. Scripts de Desarrollo

```json
// package.json
{
  "scripts": {
    "start": "react-scripts start",
    "start:mock": "REACT_APP_USE_MOCK=true npm start",
    "build": "react-scripts build",
    "test": "react-scripts test",
    "test:coverage": "npm test -- --coverage --watchAll=false",
    "lint": "eslint src --ext .js,.jsx",
    "lint:fix": "npm run lint -- --fix",
    "format": "prettier --write \"src/**/*.{js,jsx,json,css,md}\"",
    "analyze": "source-map-explorer 'build/static/js/*.js'",
    "storybook": "start-storybook -p 6006",
    "build-storybook": "build-storybook"
  }
}
```

### 3. Configuración de Apollo Client para Desarrollo

```javascript
// src/config/apollo.dev.js
import { ApolloClient, InMemoryCache } from '@apollo/client';
import { createHttpLink } from '@apollo/client/link/http';
import { setContext } from '@apollo/client/link/context';
import { onError } from '@apollo/client/link/error';
import { ApolloLink } from '@apollo/client/link/core';

// Development-specific link for logging
const loggingLink = new ApolloLink((operation, forward) => {
  console.group(`GraphQL ${operation.operationName}`);
  console.log('Query:', operation.query.loc?.source.body);
  console.log('Variables:', operation.variables);
  console.groupEnd();
  
  const start = Date.now();
  
  return forward(operation).map(response => {
    const duration = Date.now() - start;
    console.group(`GraphQL ${operation.operationName} Response (${duration}ms)`);
    console.log('Data:', response.data);
    if (response.errors) {
      console.error('Errors:', response.errors);
    }
    console.groupEnd();
    
    return response;
  });
});

// Error handling for development
const errorLink = onError(({ graphQLErrors, networkError, operation, forward }) => {
  if (graphQLErrors) {
    graphQLErrors.forEach(({ message, locations, path, extensions }) => {
      console.error(
        `GraphQL error: Message: ${message}, Location: ${locations}, Path: ${path}`,
        extensions
      );
    });
  }
  
  if (networkError) {
    console.error(`Network error: ${networkError}`);
    
    // Retry logic for development
    if (networkError.statusCode === 500) {
      return forward(operation);
    }
  }
});

// HTTP Link
const httpLink = createHttpLink({
  uri: process.env.REACT_APP_GRAPHQL_URL,
  credentials: 'include'
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

// Apollo Client for development
export const apolloClient = new ApolloClient({
  link: ApolloLink.from([
    loggingLink,
    errorLink,
    authLink,
    httpLink
  ]),
  cache: new InMemoryCache({
    addTypename: true,
    typePolicies: {
      // Development-specific type policies
    }
  }),
  defaultOptions: {
    watchQuery: {
      fetchPolicy: 'cache-and-network',
    },
  },
  connectToDevTools: true,
  name: 'asam-frontend-dev',
  version: '1.0.0'
});
```

### 4. Hot Module Replacement

```javascript
// src/index.js
import React from 'react';
import ReactDOM from 'react-dom/client';
import App from './App';

const root = ReactDOM.createRoot(document.getElementById('root'));

function render() {
  root.render(
    <React.StrictMode>
      <App />
    </React.StrictMode>
  );
}

render();

// Hot Module Replacement
if (module.hot) {
  module.hot.accept('./App', () => {
    render();
  });
}
```

## Desarrollo con Vue

### 1. Estructura del Proyecto Vue

```
asam-frontend-vue/
├── public/
│   └── index.html
├── src/
│   ├── assets/
│   ├── components/
│   │   ├── common/
│   │   └── features/
│   ├── composables/
│   ├── router/
│   ├── stores/
│   ├── views/
│   ├── graphql/
│   ├── utils/
│   ├── App.vue
│   └── main.js
├── .env.development
├── vite.config.js
└── package.json
```

### 2. Configuración de Vite

```javascript
// vite.config.js
import { defineConfig } from 'vite';
import vue from '@vitejs/plugin-vue';
import path from 'path';

export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  server: {
    port: 3000,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
        secure: false,
      },
      '/graphql': {
        target: 'http://localhost:8080',
        changeOrigin: true,
        secure: false,
        ws: true,
      },
    },
  },
  define: {
    __VUE_OPTIONS_API__: true,
    __VUE_PROD_DEVTOOLS__: false,
  },
});
```

### 3. Composables para Desarrollo

```javascript
// src/composables/useDevTools.js
import { ref, onMounted } from 'vue';

export function useDevTools() {
  const isDevToolsOpen = ref(false);
  
  onMounted(() => {
    // Detectar si DevTools está abierto
    const devtools = { open: false, orientation: null };
    const threshold = 160;
    
    setInterval(() => {
      if (window.outerHeight - window.innerHeight > threshold ||
          window.outerWidth - window.innerWidth > threshold) {
        if (!devtools.open) {
          devtools.open = true;
          isDevToolsOpen.value = true;
          console.log('DevTools opened');
        }
      } else {
        if (devtools.open) {
          devtools.open = false;
          isDevToolsOpen.value = false;
          console.log('DevTools closed');
        }
      }
    }, 500);
  });
  
  const log = (label, data) => {
    if (process.env.NODE_ENV === 'development') {
      console.group(`🔍 ${label}`);
      console.log(data);
      console.trace();
      console.groupEnd();
    }
  };
  
  const time = (label) => {
    if (process.env.NODE_ENV === 'development') {
      console.time(label);
    }
  };
  
  const timeEnd = (label) => {
    if (process.env.NODE_ENV === 'development') {
      console.timeEnd(label);
    }
  };
  
  return {
    isDevToolsOpen,
    log,
    time,
    timeEnd
  };
}
```

## Herramientas de Desarrollo

### 1. Mock Service Worker

```javascript
// src/mocks/setup.js
import { setupWorker } from 'msw';
import { handlers } from './handlers';

// Configurar MSW para desarrollo
export const worker = setupWorker(...handlers);

// Opciones de inicio
const workerOptions = {
  onUnhandledRequest: 'warn',
  serviceWorker: {
    url: '/mockServiceWorker.js',
    options: {
      scope: '/'
    }
  }
};

// Iniciar solo en desarrollo
if (process.env.NODE_ENV === 'development' && 
    process.env.REACT_APP_ENABLE_MOCK_DATA === 'true') {
  worker.start(workerOptions).then(() => {
    console.log('🔧 Mock Service Worker started');
  });
}
```

### 2. Redux DevTools

```javascript
// src/store/index.js
import { configureStore } from '@reduxjs/toolkit';
import { composeWithDevTools } from '@redux-devtools/extension';

const composeEnhancers = composeWithDevTools({
  // Opciones específicas para desarrollo
  trace: true,
  traceLimit: 25,
  features: {
    pause: true,
    lock: true,
    persist: true,
    export: true,
    import: 'custom',
    jump: true,
    skip: true,
    reorder: true,
    dispatch: true,
    test: true
  }
});

export const store = configureStore({
  reducer: rootReducer,
  middleware: (getDefaultMiddleware) =>
    getDefaultMiddleware({
      serializableCheck: {
        // Ignore these action types
        ignoredActions: ['persist/PERSIST'],
      },
    }),
  devTools: process.env.NODE_ENV === 'development' && {
    name: 'ASAM Frontend',
    trace: true,
    traceLimit: 25,
  }
});
```

### 3. React Query DevTools

```javascript
// src/App.js
import { QueryClient, QueryClientProvider } from 'react-query';
import { ReactQueryDevtools } from 'react-query/devtools';

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      refetchOnWindowFocus: false,
      retry: process.env.NODE_ENV === 'production' ? 3 : 0,
      staleTime: 5 * 60 * 1000, // 5 minutos
    },
  },
});

function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <Router>
        {/* Tu aplicación */}
      </Router>
      {process.env.NODE_ENV === 'development' && (
        <ReactQueryDevtools 
          initialIsOpen={false}
          position="bottom-right"
        />
      )}
    </QueryClientProvider>
  );
}
```

## Configuración de VS Code

### 1. Configuración del Workspace

```json
// .vscode/settings.json
{
  "editor.formatOnSave": true,
  "editor.codeActionsOnSave": {
    "source.fixAll.eslint": true
  },
  "eslint.validate": [
    "javascript",
    "javascriptreact",
    "typescript",
    "typescriptreact"
  ],
  "editor.defaultFormatter": "esbenp.prettier-vscode",
  "[javascript]": {
    "editor.defaultFormatter": "esbenp.prettier-vscode"
  },
  "[javascriptreact]": {
    "editor.defaultFormatter": "esbenp.prettier-vscode"
  },
  "emmet.includeLanguages": {
    "javascript": "javascriptreact"
  },
  "files.associations": {
    "*.js": "javascriptreact"
  },
  "javascript.updateImportsOnFileMove.enabled": "always",
  "typescript.updateImportsOnFileMove.enabled": "always",
  "editor.snippetSuggestions": "top",
  "editor.suggestSelection": "first"
}
```

### 2. Extensiones Recomendadas

```json
// .vscode/extensions.json
{
  "recommendations": [
    "dbaeumer.vscode-eslint",
    "esbenp.prettier-vscode",
    "dsznajder.es7-react-js-snippets",
    "burkeholland.simple-react-snippets",
    "apollographql.vscode-apollo",
    "prisma.vscode-graphql",
    "formulahendry.auto-rename-tag",
    "formulahendry.auto-close-tag",
    "christian-kohler.path-intellisense",
    "christian-kohler.npm-intellisense",
    "eg2.vscode-npm-script",
    "mikestead.dotenv",
    "usernamehw.errorlens",
    "wix.vscode-import-cost",
    "chakrounanas.turbo-console-log",
    "naumovs.color-highlight",
    "vincaslt.highlight-matching-tag",
    "firefox-devtools.vscode-firefox-debug",
    "msjsdiag.debugger-for-chrome"
  ]
}
```

### 3. Snippets Personalizados

```json
// .vscode/react.code-snippets
{
  "GraphQL Query Hook": {
    "prefix": "gqlquery",
    "body": [
      "const ${1:QUERY_NAME} = gql`",
      "  query ${2:QueryName}(${3:$id: ID!}) {",
      "    ${4:queryField}(${5:id: $id}) {",
      "      ${6:fields}",
      "    }",
      "  }",
      "`;",
      "",
      "export function use${2:QueryName}(${7:id}) {",
      "  const { data, loading, error } = useQuery(${1:QUERY_NAME}, {",
      "    variables: { ${8:id} },",
      "    skip: !${8:id}",
      "  });",
      "",
      "  return {",
      "    ${9:dataName}: data?.${4:queryField},",
      "    loading,",
      "    error",
      "  };",
      "}"
    ]
  },
  "GraphQL Mutation Hook": {
    "prefix": "gqlmutation",
    "body": [
      "const ${1:MUTATION_NAME} = gql`",
      "  mutation ${2:MutationName}($input: ${3:InputType}!) {",
      "    ${4:mutationField}(input: $input) {",
      "      ${5:fields}",
      "    }",
      "  }",
      "`;",
      "",
      "export function use${2:MutationName}() {",
      "  const [${6:mutate}, { data, loading, error }] = useMutation(${1:MUTATION_NAME});",
      "",
      "  const ${7:executeMutation} = async (input) => {",
      "    try {",
      "      const result = await ${6:mutate}({",
      "        variables: { input }",
      "      });",
      "      return result.data.${4:mutationField};",
      "    } catch (error) {",
      "      console.error('Error in ${2:MutationName}:', error);",
      "      throw error;",
      "    }",
      "  };",
      "",
      "  return {",
      "    ${7:executeMutation},",
      "    loading,",
      "    error",
      "  };",
      "}"
    ]
  }
}
```

## Debugging

### 1. Configuración de Chrome DevTools

```json
// .vscode/launch.json
{
  "version": "0.2.0",
  "configurations": [
    {
      "type": "chrome",
      "request": "launch",
      "name": "Launch Chrome against localhost",
      "url": "http://localhost:3000",
      "webRoot": "${workspaceFolder}/src",
      "sourceMaps": true,
      "sourceMapPathOverrides": {
        "webpack:///src/*": "${webRoot}/*"
      }
    },
    {
      "type": "chrome",
      "request": "attach",
      "name": "Attach to Chrome",
      "port": 9222,
      "webRoot": "${workspaceFolder}/src",
      "sourceMaps": true
    },
    {
      "type": "node",
      "request": "launch",
      "name": "Debug Tests",
      "runtimeExecutable": "${workspaceFolder}/node_modules/.bin/react-scripts",
      "args": ["test", "--runInBand", "--no-cache", "--watchAll=false"],
      "cwd": "${workspaceFolder}",
      "protocol": "inspector",
      "console": "integratedTerminal",
      "internalConsoleOptions": "neverOpen",
      "env": {
        "CI": "true"
      }
    }
  ]
}
```

### 2. Debug Utilities

```javascript
// src/utils/debug.js
export const debug = {
  // Log con estilo
  log: (message, data, style = 'color: #007acc; font-weight: bold;') => {
    if (process.env.NODE_ENV === 'development') {
      console.log(`%c[DEBUG] ${message}`, style, data);
    }
  },
  
  // Tabla de datos
  table: (data, columns) => {
    if (process.env.NODE_ENV === 'development') {
      console.table(data, columns);
    }
  },
  
  // Performance timing
  performance: {
    marks: {},
    
    start: (label) => {
      if (process.env.NODE_ENV === 'development') {
        debug.performance.marks[label] = performance.now();
      }
    },
    
    end: (label) => {
      if (process.env.NODE_ENV === 'development' && debug.performance.marks[label]) {
        const duration = performance.now() - debug.performance.marks[label];
        console.log(`⏱️ ${label}: ${duration.toFixed(2)}ms`);
        delete debug.performance.marks[label];
      }
    }
  },
  
  // GraphQL query logger
  graphql: (operation, variables, result) => {
    if (process.env.NODE_ENV === 'development') {
      console.group(`🔄 GraphQL ${operation}`);
      console.log('Variables:', variables);
      console.log('Result:', result);
      console.groupEnd();
    }
  },
  
  // Component render tracker
  renders: new Map(),
  
  trackRender: (componentName) => {
    if (process.env.NODE_ENV === 'development') {
      const count = (debug.renders.get(componentName) || 0) + 1;
      debug.renders.set(componentName, count);
      console.log(`🔄 ${componentName} rendered ${count} times`);
    }
  }
};

// React hook para debugging
export function useDebug(componentName) {
  useEffect(() => {
    debug.trackRender(componentName);
  });
  
  return debug;
}
```

### 3. Apollo Client DevTools Helper

```javascript
// src/utils/apolloDevTools.js
export function setupApolloDevTools(client) {
  if (process.env.NODE_ENV === 'development') {
    // Exponer cliente para debugging
    window.__APOLLO_CLIENT__ = client;
    
    // Helper functions
    window.apolloDevTools = {
      // Ver cache
      cache: () => client.cache.extract(),
      
      // Ejecutar query
      query: async (query, variables) => {
        const result = await client.query({ query, variables });
        console.log('Query result:', result);
        return result;
      },
      
      // Ejecutar mutation
      mutate: async (mutation, variables) => {
        const result = await client.mutate({ mutation, variables });
        console.log('Mutation result:', result);
        return result;
      },
      
      // Limpiar cache
      clearCache: () => {
        client.clearStore();
        console.log('Cache cleared');
      },
      
      // Ver queries activas
      activeQueries: () => {
        const queries = client.getObservableQueries();
        console.log('Active queries:', queries);
        return queries;
      }
    };
    
    console.log('🚀 Apollo DevTools ready. Use window.apolloDevTools');
  }
}
```

## Tips y Trucos

### 1. Atajos Útiles

```javascript
// src/utils/devShortcuts.js
export function setupDevShortcuts() {
  if (process.env.NODE_ENV === 'development') {
    document.addEventListener('keydown', (e) => {
      // Ctrl+Shift+D: Toggle debug mode
      if (e.ctrlKey && e.shiftKey && e.key === 'D') {
        const debugMode = localStorage.getItem('debugMode') === 'true';
        localStorage.setItem('debugMode', !debugMode);
        window.location.reload();
      }
      
      // Ctrl+Shift+M: Toggle mock data
      if (e.ctrlKey && e.shiftKey && e.key === 'M') {
        const useMock = localStorage.getItem('useMockData') === 'true';
        localStorage.setItem('useMockData', !useMock);
        window.location.reload();
      }
      
      // Ctrl+Shift+L: Clear localStorage
      if (e.ctrlKey && e.shiftKey && e.key === 'L') {
        if (confirm('Clear localStorage?')) {
          localStorage.clear();
          window.location.reload();
        }
      }
    });
  }
}
```

### 2. Component Inspector

```javascript
// src/components/DevTools/ComponentInspector.jsx
import { useState } from 'react';

export function ComponentInspector({ children, name }) {
  const [showInfo, setShowInfo] = useState(false);
  const [renderCount, setRenderCount] = useState(0);
  
  useEffect(() => {
    setRenderCount(prev => prev + 1);
  });
  
  if (process.env.NODE_ENV !== 'development') {
    return children;
  }
  
  return (
    <div className="component-inspector" data-component={name}>
      <button 
        className="inspector-toggle"
        onClick={() => setShowInfo(!showInfo)}
      >
        🔍
      </button>
      
      {showInfo && (
        <div className="inspector-info">
          <h4>{name}</h4>
          <p>Renders: {renderCount}</p>
          <p>Props: {Object.keys(children.props).length}</p>
        </div>
      )}
      
      {children}
    </div>
  );
}
```

### 3. Network Request Logger

```javascript
// src/utils/networkLogger.js
export function setupNetworkLogger() {
  if (process.env.NODE_ENV === 'development') {
    const originalFetch = window.fetch;
    
    window.fetch = async (...args) => {
      const [url, options] = args;
      console.group(`🌐 Fetch: ${options?.method || 'GET'} ${url}`);
      console.log('Options:', options);
      
      const startTime = performance.now();
      
      try {
        const response = await originalFetch(...args);
        const duration = performance.now() - startTime;
        
        console.log(`✅ Status: ${response.status} (${duration.toFixed(2)}ms)`);
        console.groupEnd();
        
        return response;
      } catch (error) {
        const duration = performance.now() - startTime;
        console.error(`❌ Error (${duration.toFixed(2)}ms):`, error);
        console.groupEnd();
        throw error;
      }
    };
  }
}
```

## Comandos Útiles

```bash
# Instalar dependencias
npm install

# Desarrollo
npm start                    # Iniciar servidor de desarrollo
npm run start:mock          # Iniciar con datos mock
npm run storybook           # Iniciar Storybook

# Testing
npm test                    # Ejecutar tests en modo watch
npm run test:coverage       # Generar reporte de cobertura
npm run test:debug          # Debug tests en VS Code

# Linting y Formato
npm run lint               # Verificar linting
npm run lint:fix           # Corregir errores de linting
npm run format             # Formatear código

# Build y Análisis
npm run build              # Build de producción
npm run analyze            # Analizar bundle size
npm run build:stats        # Generar stats para webpack-bundle-analyzer

# Utilidades
npm run clean              # Limpiar cache y node_modules
npm run update:deps        # Actualizar dependencias
npm run check:security     # Verificar vulnerabilidades
```

Esta guía proporciona todo lo necesario para configurar y optimizar el entorno de desarrollo local para trabajar con el backend de ASAM.