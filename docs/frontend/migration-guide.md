# Guía de Migración y Actualización

Esta guía proporciona instrucciones para migrar y actualizar aplicaciones frontend que consumen el backend de ASAM.

## Tabla de Contenidos
1. [Estrategia de Versionado](#estrategia-de-versionado)
2. [Migración de Versiones](#migración-de-versiones)
3. [Actualización de Dependencias](#actualización-de-dependencias)
4. [Breaking Changes](#breaking-changes)
5. [Migración de Datos](#migración-de-datos)
6. [Testing de Migraciones](#testing-de-migraciones)
7. [Rollback Plan](#rollback-plan)
8. [Checklist de Migración](#checklist-de-migración)

## Estrategia de Versionado

### 1. Semantic Versioning

```
MAJOR.MINOR.PATCH

MAJOR: Cambios incompatibles con versiones anteriores
MINOR: Nueva funcionalidad compatible con versiones anteriores
PATCH: Correcciones de bugs compatibles
```

### 2. Compatibilidad Backend-Frontend

```javascript
// utils/versionCheck.js
export async function checkBackendCompatibility() {
  try {
    const response = await fetch('/api/version');
    const { version, minClientVersion } = await response.json();
    
    const currentVersion = process.env.REACT_APP_VERSION;
    
    if (compareVersions(currentVersion, minClientVersion) < 0) {
      // Version del cliente muy antigua
      return {
        compatible: false,
        message: 'Por favor actualiza la aplicación',
        updateUrl: '/update'
      };
    }
    
    return { compatible: true };
  } catch (error) {
    console.error('Version check failed:', error);
    return { compatible: true }; // Asumir compatible si falla
  }
}

function compareVersions(v1, v2) {
  const parts1 = v1.split('.').map(Number);
  const parts2 = v2.split('.').map(Number);
  
  for (let i = 0; i < 3; i++) {
    if (parts1[i] > parts2[i]) return 1;
    if (parts1[i] < parts2[i]) return -1;
  }
  
  return 0;
}
```

## Migración de Versiones

### 1. De v1.x a v2.0

#### Breaking Changes

```javascript
// ❌ v1.x - Query antigua
const OLD_MEMBER_QUERY = gql`
  query GetMember($id: ID!) {
    member(id: $id) {  // Campo renombrado
      id              // Campo renombrado
      memberNumber    // Campo renombrado
      fullName        // Campo eliminado
    }
  }
`;

// ✅ v2.0 - Query nueva
const NEW_MEMBER_QUERY = gql`
  query GetMember($id: ID!) {
    getMember(id: $id) {    // Nuevo nombre
      miembro_id            // Nuevo nombre
      numero_socio          // Nuevo nombre
      nombre                // Campos separados
      apellidos             // Campos separados
    }
  }
`;

// Capa de compatibilidad
export function useMemberCompat(id) {
  const { data, ...rest } = useQuery(NEW_MEMBER_QUERY, {
    variables: { id }
  });
  
  // Transformar datos al formato antiguo si es necesario
  const compatData = data ? {
    member: {
      id: data.getMember.miembro_id,
      memberNumber: data.getMember.numero_socio,
      fullName: `${data.getMember.nombre} ${data.getMember.apellidos}`
    }
  } : null;
  
  return { data: compatData, ...rest };
}
```

#### Script de Migración

```javascript
// scripts/migrate-v2.js
const fs = require('fs');
const path = require('path');
const { parse, print, visit } = require('graphql');

function migrateGraphQLQueries(filePath) {
  const content = fs.readFileSync(filePath, 'utf8');
  
  // Buscar queries GraphQL
  const gqlRegex = /gql`([\s\S]*?)`/g;
  let match;
  let newContent = content;
  
  while ((match = gqlRegex.exec(content)) !== null) {
    const query = match[1];
    const ast = parse(query);
    
    // Transformar AST
    const newAst = visit(ast, {
      Field(node) {
        // Renombrar campos
        if (node.name.value === 'member') {
          return { ...node, name: { ...node.name, value: 'getMember' } };
        }
        if (node.name.value === 'id') {
          return { ...node, name: { ...node.name, value: 'miembro_id' } };
        }
        if (node.name.value === 'memberNumber') {
          return { ...node, name: { ...node.name, value: 'numero_socio' } };
        }
      }
    });
    
    const newQuery = print(newAst);
    newContent = newContent.replace(match[0], `gql\`${newQuery}\``);
  }
  
  // Backup del archivo original
  fs.writeFileSync(`${filePath}.backup`, content);
  
  // Escribir archivo migrado
  fs.writeFileSync(filePath, newContent);
  
  console.log(`✅ Migrated: ${filePath}`);
}

// Ejecutar migración
function runMigration() {
  const srcDir = path.join(__dirname, '../src');
  const files = getAllFiles(srcDir, '.js', '.jsx', '.ts', '.tsx');
  
  files.forEach(file => {
    try {
      migrateGraphQLQueries(file);
    } catch (error) {
      console.error(`❌ Error migrating ${file}:`, error);
    }
  });
}

runMigration();
```

### 2. Migración Gradual

```javascript
// config/featureFlags.js
export const featureFlags = {
  useNewMemberAPI: process.env.REACT_APP_USE_NEW_MEMBER_API === 'true',
  useNewPaymentFlow: process.env.REACT_APP_USE_NEW_PAYMENT_FLOW === 'true',
  enableWebSockets: process.env.REACT_APP_ENABLE_WEBSOCKETS === 'true'
};

// hooks/useMemberData.js
import { featureFlags } from '../config/featureFlags';

export function useMemberData(id) {
  // Usar API nueva o antigua según feature flag
  if (featureFlags.useNewMemberAPI) {
    return useNewMemberAPI(id);
  } else {
    return useOldMemberAPI(id);
  }
}

// Componente con migración gradual
function MemberProfile({ id }) {
  const { data, loading } = useMemberData(id);
  
  if (loading) return <Spinner />;
  
  // Renderizar según versión de API
  if (featureFlags.useNewMemberAPI) {
    return <NewMemberProfile data={data} />;
  } else {
    return <LegacyMemberProfile data={data} />;
  }
}
```

## Actualización de Dependencias

### 1. Proceso de Actualización Segura

```bash
# 1. Crear rama de actualización
git checkout -b update-dependencies

# 2. Actualizar dependencias menores
npm update

# 3. Ver dependencias desactualizadas
npm outdated

# 4. Actualizar dependencias mayores una por una
npm install react@latest
npm install @apollo/client@latest

# 5. Ejecutar tests después de cada actualización
npm test

# 6. Verificar build
npm run build
```

### 2. Herramienta de Actualización Automatizada

```javascript
// scripts/update-dependencies.js
const { execSync } = require('child_process');
const fs = require('fs');

const criticalDependencies = [
  'react',
  'react-dom',
  '@apollo/client',
  'graphql'
];

const testAfterUpdate = async (dep) => {
  console.log(`Testing after updating ${dep}...`);
  
  try {
    execSync('npm test -- --watchAll=false', { stdio: 'inherit' });
    execSync('npm run build', { stdio: 'inherit' });
    return true;
  } catch (error) {
    console.error(`Tests failed after updating ${dep}`);
    return false;
  }
};

const updateDependency = async (dep) => {
  console.log(`Updating ${dep}...`);
  
  // Guardar versión actual
  const packageJson = JSON.parse(fs.readFileSync('package.json', 'utf8'));
  const currentVersion = packageJson.dependencies[dep] || packageJson.devDependencies[dep];
  
  try {
    // Actualizar a latest
    execSync(`npm install ${dep}@latest`, { stdio: 'inherit' });
    
    // Test
    if (await testAfterUpdate(dep)) {
      console.log(`✅ Successfully updated ${dep}`);
      return true;
    } else {
      // Rollback
      console.log(`Rolling back ${dep} to ${currentVersion}`);
      execSync(`npm install ${dep}@${currentVersion}`, { stdio: 'inherit' });
      return false;
    }
  } catch (error) {
    console.error(`Failed to update ${dep}:`, error);
    return false;
  }
};

// Ejecutar actualizaciones
async function runUpdates() {
  const results = {};
  
  for (const dep of criticalDependencies) {
    results[dep] = await updateDependency(dep);
  }
  
  console.log('\n📊 Update Summary:');
  Object.entries(results).forEach(([dep, success]) => {
    console.log(`${success ? '✅' : '❌'} ${dep}`);
  });
}

runUpdates();
```

### 3. Migración de Apollo Client v2 a v3

```javascript
// Cambios principales en Apollo Client v3

// ❌ v2
import { ApolloClient } from 'apollo-client';
import { InMemoryCache } from 'apollo-cache-inmemory';
import { HttpLink } from 'apollo-link-http';

const client = new ApolloClient({
  link: new HttpLink({ uri: '/graphql' }),
  cache: new InMemoryCache()
});

// ✅ v3
import { ApolloClient, InMemoryCache } from '@apollo/client';

const client = new ApolloClient({
  uri: '/graphql',
  cache: new InMemoryCache()
});

// Hook changes
// ❌ v2
import { Query, Mutation } from 'react-apollo';

<Query query={GET_MEMBERS}>
  {({ data, loading, error }) => {
    // render
  }}
</Query>

// ✅ v3
import { useQuery, useMutation } from '@apollo/client';

function Component() {
  const { data, loading, error } = useQuery(GET_MEMBERS);
  // render
}
```

## Breaking Changes

### 1. Documentación de Breaking Changes

```markdown
# Breaking Changes v2.0

## GraphQL Schema

### Queries
- `member` → `getMember`
- `members` → `listMembers`
- `payment` → `getPayment`

### Fields
- `Member.id` → `Member.miembro_id`
- `Member.memberNumber` → `Member.numero_socio`
- `Member.fullName` → Separado en `Member.nombre` y `Member.apellidos`

### Mutations
- `createMemberMutation` → `createMember`
- Input types ahora terminan en `Input`

## API REST (Deprecada)
- Todos los endpoints REST han sido eliminados
- Usar GraphQL exclusivamente

## Autenticación
- Token format cambió de JWT a formato custom
- Refresh token ahora es obligatorio
- Cookie `session` → Header `Authorization`
```

### 2. Herramienta de Detección

```javascript
// scripts/detect-breaking-changes.js
const { buildSchema, findBreakingChanges } = require('graphql');
const fs = require('fs');

function detectBreakingChanges() {
  // Cargar schemas
  const oldSchema = buildSchema(fs.readFileSync('schema-v1.graphql', 'utf8'));
  const newSchema = buildSchema(fs.readFileSync('schema-v2.graphql', 'utf8'));
  
  // Detectar cambios
  const breakingChanges = findBreakingChanges(oldSchema, newSchema);
  
  if (breakingChanges.length > 0) {
    console.log('⚠️  Breaking Changes Detected:');
    breakingChanges.forEach(change => {
      console.log(`- ${change.type}: ${change.description}`);
    });
    
    // Generar reporte
    const report = {
      date: new Date().toISOString(),
      version: 'v2.0',
      breakingChanges: breakingChanges.map(change => ({
        type: change.type,
        description: change.description,
        path: change.path
      }))
    };
    
    fs.writeFileSync(
      'breaking-changes-report.json',
      JSON.stringify(report, null, 2)
    );
  } else {
    console.log('✅ No breaking changes detected');
  }
}

detectBreakingChanges();
```

## Migración de Datos

### 1. Migración de LocalStorage

```javascript
// utils/storageMigration.js
export function migrateLocalStorage() {
  const version = localStorage.getItem('storage_version');
  
  if (!version || version < '2.0') {
    console.log('Migrating localStorage to v2.0...');
    
    // Migrar tokens
    const oldToken = localStorage.getItem('authToken');
    if (oldToken) {
      localStorage.setItem('accessToken', oldToken);
      localStorage.removeItem('authToken');
    }
    
    // Migrar preferencias
    const oldPrefs = localStorage.getItem('userPreferences');
    if (oldPrefs) {
      try {
        const prefs = JSON.parse(oldPrefs);
        const newPrefs = {
          theme: prefs.darkMode ? 'dark' : 'light',
          language: prefs.lang || 'es',
          notifications: {
            email: prefs.emailNotifications ?? true,
            push: prefs.pushNotifications ?? false
          }
        };
        localStorage.setItem('preferences', JSON.stringify(newPrefs));
        localStorage.removeItem('userPreferences');
      } catch (error) {
        console.error('Failed to migrate preferences:', error);
      }
    }
    
    // Marcar como migrado
    localStorage.setItem('storage_version', '2.0');
    console.log('✅ LocalStorage migration complete');
  }
}

// Ejecutar al inicio de la app
migrateLocalStorage();
```

### 2. Migración de Cache de Apollo

```javascript
// utils/cacheMigration.js
import { InMemoryCache } from '@apollo/client';

export function migrateCacheData(oldCache, newCache) {
  const oldData = oldCache.extract();
  const migratedData = {};
  
  // Migrar datos del cache
  Object.entries(oldData).forEach(([key, value]) => {
    if (key.startsWith('Member:')) {
      // Migrar entidades Member
      const oldId = key.split(':')[1];
      const newKey = `Member:${value.id || oldId}`;
      
      migratedData[newKey] = {
        __typename: 'Member',
        miembro_id: value.id,
        numero_socio: value.memberNumber,
        nombre: value.firstName || value.fullName?.split(' ')[0],
        apellidos: value.lastName || value.fullName?.split(' ').slice(1).join(' '),
        ...value
      };
    } else {
      // Copiar otros datos sin cambios
      migratedData[key] = value;
    }
  });
  
  // Restaurar en nuevo cache
  newCache.restore(migratedData);
  
  return newCache;
}
```

## Testing de Migraciones

### 1. Tests de Compatibilidad

```javascript
// __tests__/migration.test.js
import { renderHook } from '@testing-library/react-hooks';
import { MockedProvider } from '@apollo/client/testing';
import { useMemberCompat } from '../hooks/useMemberCompat';

describe('Migration Compatibility', () => {
  it('should transform new API response to old format', async () => {
    const mocks = [{
      request: {
        query: NEW_MEMBER_QUERY,
        variables: { id: '1' }
      },
      result: {
        data: {
          getMember: {
            miembro_id: '1',
            numero_socio: '2024-001',
            nombre: 'Juan',
            apellidos: 'Pérez'
          }
        }
      }
    }];
    
    const { result, waitForNextUpdate } = renderHook(
      () => useMemberCompat('1'),
      {
        wrapper: ({ children }) => (
          <MockedProvider mocks={mocks}>
            {children}
          </MockedProvider>
        )
      }
    );
    
    await waitForNextUpdate();
    
    // Verificar formato compatible
    expect(result.current.data).toEqual({
      member: {
        id: '1',
        memberNumber: '2024-001',
        fullName: 'Juan Pérez'
      }
    });
  });
});
```

### 2. Tests E2E de Migración

```javascript
// e2e/migration.spec.js
describe('Migration E2E', () => {
  beforeEach(() => {
    // Setup datos v1
    cy.task('seedDatabase', { version: 'v1' });
    cy.task('setupLocalStorage', { version: 'v1' });
  });
  
  it('should migrate user data correctly', () => {
    // Visitar app v2
    cy.visit('/');
    
    // Verificar migración de localStorage
    cy.window().then(win => {
      expect(win.localStorage.getItem('storage_version')).to.equal('2.0');
      expect(win.localStorage.getItem('authToken')).to.be.null;
      expect(win.localStorage.getItem('accessToken')).to.exist;
    });
    
    // Verificar que datos se muestran correctamente
    cy.contains('Juan Pérez').should('be.visible');
    cy.contains('2024-001').should('be.visible');
  });
  
  it('should handle missing data gracefully', () => {
    // Limpiar algunos datos
    cy.task('clearLocalStorage', ['userPreferences']);
    
    cy.visit('/');
    
    // No debe crashear
    cy.contains('Dashboard').should('be.visible');
  });
});
```

## Rollback Plan

### 1. Estrategia de Rollback

```javascript
// deployment/rollback.js
const rollbackPlan = {
  version: '2.0',
  steps: [
    {
      name: 'Verificar health check',
      command: 'curl -f https://app.asam.org/health',
      timeout: 30,
      critical: true
    },
    {
      name: 'Verificar métricas críticas',
      check: async () => {
        const metrics = await getMetrics();
        return metrics.errorRate < 0.05; // < 5% error rate
      },
      critical: true
    },
    {
      name: 'Backup de datos',
      command: 'npm run backup:production',
      timeout: 300
    },
    {
      name: 'Rollback deployment',
      command: 'npm run deploy:rollback',
      timeout: 600
    },
    {
      name: 'Verificar rollback',
      command: 'npm run test:production',
      timeout: 300
    }
  ]
};

async function executeRollback() {
  console.log('🔄 Starting rollback...');
  
  for (const step of rollbackPlan.steps) {
    console.log(`Executing: ${step.name}`);
    
    try {
      if (step.command) {
        await execCommand(step.command, step.timeout);
      } else if (step.check) {
        const success = await step.check();
        if (!success && step.critical) {
          throw new Error(`Critical check failed: ${step.name}`);
        }
      }
      
      console.log(`✅ ${step.name} completed`);
    } catch (error) {
      console.error(`❌ ${step.name} failed:`, error);
      
      if (step.critical) {
        throw error;
      }
    }
  }
  
  console.log('✅ Rollback completed successfully');
}
```

### 2. Feature Flag Rollback

```javascript
// config/featureFlags.js
export const featureFlags = {
  // Flags con fechas de sunset
  useNewMemberAPI: {
    enabled: process.env.REACT_APP_USE_NEW_MEMBER_API === 'true',
    sunset: '2024-06-01',
    fallback: true
  },
  
  // Kill switch remoto
  async getRemoteFlags() {
    try {
      const response = await fetch('/api/feature-flags');
      const flags = await response.json();
      
      // Merge con flags locales
      Object.assign(this, flags);
    } catch (error) {
      console.error('Failed to fetch remote flags:', error);
      // Usar flags locales como fallback
    }
  }
};

// Hook para feature flags
export function useFeatureFlag(flagName) {
  const [enabled, setEnabled] = useState(false);
  
  useEffect(() => {
    const flag = featureFlags[flagName];
    
    if (typeof flag === 'object') {
      // Verificar sunset date
      if (flag.sunset && new Date() > new Date(flag.sunset)) {
        setEnabled(flag.fallback);
        console.warn(`Feature flag ${flagName} has passed sunset date`);
      } else {
        setEnabled(flag.enabled);
      }
    } else {
      setEnabled(!!flag);
    }
  }, [flagName]);
  
  return enabled;
}
```

## Checklist de Migración

### Pre-Migración
- [ ] Backup completo de producción
- [ ] Documentar todos los breaking changes
- [ ] Preparar scripts de migración
- [ ] Tests de migración en staging
- [ ] Plan de rollback documentado
- [ ] Comunicación a usuarios

### Durante la Migración
- [ ] Modo mantenimiento activado
- [ ] Ejecutar scripts de migración de datos
- [ ] Verificar migraciones de base de datos
- [ ] Deploy de nueva versión
- [ ] Health checks pasando
- [ ] Smoke tests básicos

### Post-Migración
- [ ] Monitoreo intensivo (primeras 24h)
- [ ] Verificar métricas clave
- [ ] Recolectar feedback de usuarios
- [ ] Documentar issues encontrados
- [ ] Actualizar documentación
- [ ] Cleanup de código legacy (después de periodo de gracia)

### Criterios de Éxito
- [ ] Error rate < 1%
- [ ] Performance metrics estables
- [ ] No pérdida de datos
- [ ] Funcionalidades críticas operativas
- [ ] Usuario puede completar flujos principales
- [ ] Rollback no fue necesario

Esta guía proporciona un framework completo para gestionar migraciones y actualizaciones de manera segura y controlada.