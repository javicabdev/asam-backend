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

#### Desarrollo Local (Backend Local)
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

#### Desarrollo con Backend de Producción
```bash
# .env.development.prod
REACT_APP_API_URL=https://asam-backend-jtpswzdxuq-ew.a.run.app
REACT_APP_GRAPHQL_URL=https://asam-backend-jtpswzdxuq-ew.a.run.app/graphql
REACT_APP_WS_URL=wss://asam-backend-jtpswzdxuq-ew.a.run.app/ws
REACT_APP_ENVIRONMENT=development

# Features flags
REACT_APP_ENABLE_ANALYTICS=false
REACT_APP_ENABLE_SENTRY=false
REACT_APP_ENABLE_MOCK_DATA=false

# Development tools
REACT_APP_ENABLE_DEVTOOLS=true
REACT_APP_APOLLO_DEVTOOLS=true
```

### 3. Configuración del Backend

#### Opción A: Backend Local
```bash
# En otra terminal, levantar el backend
cd ../asam-backend
make dev-setup  # Configuración completa con Docker

# O solo Docker
docker-compose up -d

# Verificar que esté funcionando
curl http://localhost:8080/health
# Output: {"status":"healthy"}
```

#### Opción B: Backend de Producción
```bash
# Verificar conexión al backend de producción
curl https://asam-backend-jtpswzdxuq-ew.a.run.app/health
# Output: {"status":"healthy"}

# Usar el script para cambiar entre entornos
npm run start:prod  # Usa backend de producción
npm run start:local # Usa backend local
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
├── .env.development.prod
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
    "start:local": "REACT_APP_ENV_FILE=.env.development npm start",
    "start:prod": "REACT_APP_ENV_FILE=.env.development.prod npm start",
    "start:mock": "REACT_APP_USE_MOCK=true npm start",
    "build": "react-scripts build",
    "build:staging": "REACT_APP_ENV_FILE=.env.staging npm run build",
    "build:prod": "REACT_APP_ENV_FILE=.env.production npm run build",
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

// Determinar el endpoint basado en el entorno
const getGraphQLEndpoint = () => {
  const envFile = process.env.REACT_APP_ENV_FILE;
  if (envFile === '.env.development.prod') {
    return 'https://asam-backend-jtpswzdxuq-ew.a.run.app/graphql';
  }
  return process.env.REACT_APP_GRAPHQL_URL || 'http://localhost:8080/graphql';
};

// Development-specific link for logging
const loggingLink = new ApolloLink((operation, forward) => {
  const endpoint = getGraphQLEndpoint();
  console.group(`GraphQL ${operation.operationName} → ${endpoint}`);
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
      
      // Mostrar notificación en desarrollo
      if (window.showDevNotification) {
        window.showDevNotification({
          type: 'error',
          title: 'GraphQL Error',
          message: message,
          details: extensions
        });
      }
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
  uri: getGraphQLEndpoint(),
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
      Query: {
        fields: {
          listMembers: {
            keyArgs: ["filter"],
            merge(existing, incoming) {
              return incoming;
            }
          },
          listUsers: {
            keyArgs: ["page", "pageSize"],
            merge(existing, incoming) {
              return incoming;
            }
          }
        }
      }
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

// Mostrar banner de desarrollo
if (process.env.NODE_ENV === 'development') {
  const endpoint = process.env.REACT_APP_GRAPHQL_URL;
  console.log(`
    %c🚀 ASAM Frontend Development Mode
    %c📡 Backend: ${endpoint}
    %c🔧 DevTools: Enabled
    %c📚 Docs: http://localhost:3000/docs
  `,
    'color: #007acc; font-size: 16px; font-weight: bold;',
    'color: #28a745; font-size: 12px;',
    'color: #ffc107; font-size: 12px;',
    'color: #6c757d; font-size: 12px;'
  );
}

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
├── .env.development.prod
├── vite.config.js
└── package.json
```

### 2. Configuración de Vite

```javascript
// vite.config.js
import { defineConfig, loadEnv } from 'vite';
import vue from '@vitejs/plugin-vue';
import path from 'path';

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '');
  const isProductionBackend = env.VITE_USE_PRODUCTION_BACKEND === 'true';
  
  return {
    plugins: [vue()],
    resolve: {
      alias: {
        '@': path.resolve(__dirname, './src'),
      },
    },
    server: {
      port: 3000,
      proxy: isProductionBackend ? {} : {
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
      __BACKEND_URL__: JSON.stringify(
        isProductionBackend 
          ? 'https://asam-backend-jtpswzdxuq-ew.a.run.app'
          : 'http://localhost:8080'
      ),
    },
  };
});
```

### 3. Composables para Desarrollo

```javascript
// src/composables/useDevTools.js
import { ref, onMounted } from 'vue';

export function useDevTools() {
  const isDevToolsOpen = ref(false);
  const backendUrl = ref(window.__BACKEND_URL__ || 'http://localhost:8080');
  
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
    if (import.meta.env.DEV) {
      console.group(`🔍 ${label}`);
      console.log(data);
      console.log('Backend:', backendUrl.value);
      console.trace();
      console.groupEnd();
    }
  };
  
  const time = (label) => {
    if (import.meta.env.DEV) {
      console.time(label);
    }
  };
  
  const timeEnd = (label) => {
    if (import.meta.env.DEV) {
      console.timeEnd(label);
    }
  };
  
  const switchBackend = (useProduction) => {
    if (import.meta.env.DEV) {
      localStorage.setItem('useProductionBackend', useProduction);
      window.location.reload();
    }
  };
  
  return {
    isDevToolsOpen,
    backendUrl,
    log,
    time,
    timeEnd,
    switchBackend
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

### 2. Backend Switcher Component

```javascript
// src/components/DevTools/BackendSwitcher.jsx
import { useState, useEffect } from 'react';

export function BackendSwitcher() {
  const [currentBackend, setCurrentBackend] = useState('local');
  
  useEffect(() => {
    const endpoint = process.env.REACT_APP_GRAPHQL_URL;
    if (endpoint.includes('asam-backend-jtpswzdxuq-ew.a.run.app')) {
      setCurrentBackend('production');
    }
  }, []);
  
  if (process.env.NODE_ENV !== 'development') {
    return null;
  }
  
  const switchBackend = (backend) => {
    if (backend === 'production') {
      localStorage.setItem('REACT_APP_ENV_FILE', '.env.development.prod');
    } else {
      localStorage.setItem('REACT_APP_ENV_FILE', '.env.development');
    }
    window.location.reload();
  };
  
  return (
    <div className="backend-switcher">
      <span>Backend:</span>
      <button 
        className={currentBackend === 'local' ? 'active' : ''}
        onClick={() => switchBackend('local')}
      >
        Local
      </button>
      <button 
        className={currentBackend === 'production' ? 'active' : ''}
        onClick={() => switchBackend('production')}
      >
        Production
      </button>
      
      <style jsx>{`
        .backend-switcher {
          position: fixed;
          bottom: 20px;
          right: 20px;
          background: #333;
          color: white;
          padding: 10px;
          border-radius: 8px;
          font-size: 12px;
          z-index: 9999;
        }
        
        .backend-switcher button {
          margin-left: 10px;
          padding: 4px 8px;
          border: none;
          background: #555;
          color: white;
          cursor: pointer;
          border-radius: 4px;
        }
        
        .backend-switcher button.active {
          background: #007acc;
        }
        
        .backend-switcher button:hover {
          opacity: 0.8;
        }
      `}</style>
    </div>
  );
}
```

### 3. Redux DevTools

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

### 4. Apollo DevTools Helper

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
      },
      
      // Cambiar backend
      switchBackend: (useProduction) => {
        localStorage.setItem('useProductionBackend', useProduction);
        window.location.reload();
      },
      
      // Estado actual
      status: () => {
        const endpoint = process.env.REACT_APP_GRAPHQL_URL;
        console.log('Current endpoint:', endpoint);
        console.log('Environment:', process.env.NODE_ENV);
        console.log('Cache size:', Object.keys(client.cache.extract()).length);
      }
    };
    
    console.log('🚀 Apollo DevTools ready. Use window.apolloDevTools');
  }
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
  "editor.suggestSelection": "first",
  
  // Configuración específica para ASAM
  "workbench.colorCustomizations": {
    "activityBar.background": "#1a1a2e",
    "titleBar.activeBackground": "#16213e",
    "titleBar.activeForeground": "#e7e7e7"
  },
  
  // Variables de entorno
  "terminal.integrated.env.windows": {
    "REACT_APP_GRAPHQL_URL": "http://localhost:8080/graphql"
  },
  "terminal.integrated.env.linux": {
    "REACT_APP_GRAPHQL_URL": "http://localhost:8080/graphql"
  },
  "terminal.integrated.env.osx": {
    "REACT_APP_GRAPHQL_URL": "http://localhost:8080/graphql"
  }
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
// .vscode/asam.code-snippets
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
  },
  "ASAM Component": {
    "prefix": "asamcomp",
    "body": [
      "import React from 'react';",
      "import { useQuery } from '@apollo/client';",
      "import { ${1:QUERY_NAME} } from '@/graphql/queries/${2:queryFile}';",
      "import { useAuth } from '@/hooks/useAuth';",
      "import { useNotification } from '@/hooks/useNotification';",
      "",
      "export function ${3:ComponentName}() {",
      "  const { isAdmin } = useAuth();",
      "  const { showError } = useNotification();",
      "  ",
      "  const { data, loading, error } = useQuery(${1:QUERY_NAME});",
      "  ",
      "  if (loading) return <div>Cargando...</div>;",
      "  if (error) {",
      "    showError('Error al cargar datos');",
      "    return <div>Error: {error.message}</div>;",
      "  }",
      "  ",
      "  return (",
      "    <div className=\"${4:component-class}\">",
      "      ${5:content}",
      "    </div>",
      "  );",
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
      "name": "Launch Chrome (Local Backend)",
      "url": "http://localhost:3000",
      "webRoot": "${workspaceFolder}/src",
      "sourceMaps": true,
      "sourceMapPathOverrides": {
        "webpack:///src/*": "${webRoot}/*"
      },
      "env": {
        "REACT_APP_GRAPHQL_URL": "http://localhost:8080/graphql"
      }
    },
    {
      "type": "chrome",
      "request": "launch",
      "name": "Launch Chrome (Production Backend)",
      "url": "http://localhost:3000",
      "webRoot": "${workspaceFolder}/src",
      "sourceMaps": true,
      "sourceMapPathOverrides": {
        "webpack:///src/*": "${webRoot}/*"
      },
      "env": {
        "REACT_APP_GRAPHQL_URL": "https://asam-backend-jtpswzdxuq-ew.a.run.app/graphql"
      }
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
      const backend = process.env.REACT_APP_GRAPHQL_URL;
      console.log(`%c[DEBUG] ${message} (Backend: ${backend})`, style, data);
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
      const backend = process.env.REACT_APP_GRAPHQL_URL;
      console.group(`🔄 GraphQL ${operation} → ${backend}`);
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
  },
  
  // Backend status
  checkBackend: async () => {
    const backend = process.env.REACT_APP_GRAPHQL_URL;
    try {
      const response = await fetch(backend.replace('/graphql', '/health'));
      const data = await response.json();
      console.log(`✅ Backend ${backend} is ${data.status}`);
      return data;
    } catch (error) {
      console.error(`❌ Backend ${backend} is unavailable`, error);
      return null;
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

### 3. Network Request Logger

```javascript
// src/utils/networkLogger.js
export function setupNetworkLogger() {
  if (process.env.NODE_ENV === 'development') {
    const originalFetch = window.fetch;
    
    window.fetch = async (...args) => {
      const [url, options] = args;
      const isGraphQL = url.includes('/graphql');
      const backend = new URL(url).hostname;
      
      console.group(`🌐 ${options?.method || 'GET'} ${url}`);
      console.log('Backend:', backend);
      console.log('Options:', options);
      
      if (isGraphQL && options?.body) {
        try {
          const body = JSON.parse(options.body);
          console.log('GraphQL Operation:', body.operationName);
          console.log('Variables:', body.variables);
        } catch (e) {}
      }
      
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
      
      // Ctrl+Shift+B: Switch backend
      if (e.ctrlKey && e.shiftKey && e.key === 'B') {
        const currentBackend = process.env.REACT_APP_GRAPHQL_URL;
        const isProduction = currentBackend.includes('asam-backend-jtpswzdxuq-ew.a.run.app');
        
        if (isProduction) {
          localStorage.setItem('REACT_APP_ENV_FILE', '.env.development');
        } else {
          localStorage.setItem('REACT_APP_ENV_FILE', '.env.development.prod');
        }
        
        window.location.reload();
      }
      
      // Ctrl+Shift+L: Clear localStorage
      if (e.ctrlKey && e.shiftKey && e.key === 'L') {
        if (confirm('Clear localStorage?')) {
          localStorage.clear();
          window.location.reload();
        }
      }
      
      // Ctrl+Shift+H: Check backend health
      if (e.ctrlKey && e.shiftKey && e.key === 'H') {
        debug.checkBackend();
      }
    });
    
    console.log(`
      Dev Shortcuts:
      Ctrl+Shift+D: Toggle debug mode
      Ctrl+Shift+M: Toggle mock data
      Ctrl+Shift+B: Switch backend (local/production)
      Ctrl+Shift+L: Clear localStorage
      Ctrl+Shift+H: Check backend health
    `);
  }
}
```

### 2. Environment Indicator

```javascript
// src/components/DevTools/EnvironmentIndicator.jsx
import { useState, useEffect } from 'react';

export function EnvironmentIndicator() {
  const [backendStatus, setBackendStatus] = useState('checking');
  
  useEffect(() => {
    checkBackendStatus();
    const interval = setInterval(checkBackendStatus, 30000); // Check every 30s
    return () => clearInterval(interval);
  }, []);
  
  const checkBackendStatus = async () => {
    try {
      const url = process.env.REACT_APP_GRAPHQL_URL.replace('/graphql', '/health');
      const response = await fetch(url);
      if (response.ok) {
        setBackendStatus('healthy');
      } else {
        setBackendStatus('unhealthy');
      }
    } catch (error) {
      setBackendStatus('offline');
    }
  };
  
  if (process.env.NODE_ENV !== 'development') {
    return null;
  }
  
  const backend = process.env.REACT_APP_GRAPHQL_URL;
  const isProduction = backend.includes('asam-backend-jtpswzdxuq-ew.a.run.app');
  
  const statusColor = {
    healthy: '#28a745',
    unhealthy: '#ffc107',
    offline: '#dc3545',
    checking: '#6c757d'
  }[backendStatus];
  
  return (
    <div className="environment-indicator">
      <div className="status" style={{ backgroundColor: statusColor }} />
      <span className="backend-type">
        {isProduction ? 'PROD' : 'LOCAL'}
      </span>
      <span className="backend-url">
        {backend.replace('http://', '').replace('https://', '').split('/')[0]}
      </span>
      
      <style jsx>{`
        .environment-indicator {
          position: fixed;
          top: 0;
          left: 50%;
          transform: translateX(-50%);
          background: #1a1a1a;
          color: white;
          padding: 4px 12px;
          font-size: 11px;
          border-bottom-left-radius: 4px;
          border-bottom-right-radius: 4px;
          display: flex;
          align-items: center;
          gap: 8px;
          z-index: 10000;
          font-family: monospace;
        }
        
        .status {
          width: 8px;
          height: 8px;
          border-radius: 50%;
        }
        
        .backend-type {
          font-weight: bold;
          color: ${isProduction ? '#dc3545' : '#28a745'};
        }
        
        .backend-url {
          opacity: 0.7;
          font-size: 10px;
        }
      `}</style>
    </div>
  );
}
```

## Comandos Útiles

```bash
# Instalar dependencias
npm install

# Desarrollo con backend local
npm start                    # Iniciar con backend local
npm run start:local         # Explícitamente con backend local

# Desarrollo con backend de producción
npm run start:prod          # Iniciar con backend de producción

# Mock y Storybook
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
npm run build:staging      # Build para staging
npm run analyze            # Analizar bundle size
npm run build:stats        # Generar stats para webpack-bundle-analyzer

# Utilidades
npm run clean              # Limpiar cache y node_modules
npm run update:deps        # Actualizar dependencias
npm run check:security     # Verificar vulnerabilidades

# Desarrollo específico
npm run dev:check-backend  # Verificar estado del backend
npm run dev:switch-backend # Cambiar entre local/producción
```

Esta guía proporciona todo lo necesario para configurar y optimizar el entorno de desarrollo local para trabajar con el backend de ASAM, tanto en modo local como conectándose al backend de producción.
