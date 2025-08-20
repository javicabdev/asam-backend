# Guía de Testing para Frontend

Esta guía proporciona estrategias y ejemplos para testear aplicaciones frontend que consumen el backend de ASAM.

## Tabla de Contenidos
1. [Configuración del Entorno de Testing](#configuración-del-entorno-de-testing)
2. [Testing de Queries y Mutations](#testing-de-queries-y-mutations)
3. [Mocking de Datos](#mocking-de-datos)
4. [Testing de Componentes](#testing-de-componentes)
5. [Testing de Integración](#testing-de-integración)
6. [Testing E2E](#testing-e2e)
7. [Mejores Prácticas](#mejores-prácticas)

## Configuración del Entorno de Testing

### React con Jest y Testing Library

```bash
npm install --save-dev @testing-library/react @testing-library/jest-dom @testing-library/user-event
npm install --save-dev @apollo/client/testing
npm install --save-dev msw
```

```javascript
// jest.config.js
module.exports = {
  testEnvironment: 'jsdom',
  setupFilesAfterEnv: ['<rootDir>/src/setupTests.js'],
  moduleNameMapper: {
    '^@/(.*)$': '<rootDir>/src/$1',
    '\\.(css|less|scss|sass)$': 'identity-obj-proxy',
  },
  transform: {
    '^.+\\.(js|jsx|ts|tsx)$': ['babel-jest', { presets: ['@babel/preset-react'] }]
  },
  collectCoverageFrom: [
    'src/**/*.{js,jsx}',
    '!src/index.js',
    '!src/serviceWorker.js',
  ],
  coverageThreshold: {
    global: {
      branches: 80,
      functions: 80,
      lines: 80,
      statements: 80
    }
  }
};
```

```javascript
// src/setupTests.js
import '@testing-library/jest-dom';
import { server } from './mocks/server';

// Establecer API mocking antes de todos los tests
beforeAll(() => server.listen());

// Reset handlers después de cada test
afterEach(() => server.resetHandlers());

// Limpiar después de todos los tests
afterAll(() => server.close());

// Mock de localStorage
const localStorageMock = {
  getItem: jest.fn(),
  setItem: jest.fn(),
  removeItem: jest.fn(),
  clear: jest.fn(),
};
global.localStorage = localStorageMock;

// Mock de window.matchMedia
Object.defineProperty(window, 'matchMedia', {
  writable: true,
  value: jest.fn().mockImplementation(query => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: jest.fn(),
    removeListener: jest.fn(),
    addEventListener: jest.fn(),
    removeEventListener: jest.fn(),
    dispatchEvent: jest.fn(),
  })),
});
```

### Vue con Vitest

```bash
npm install --save-dev @vue/test-utils vitest @vitest/ui happy-dom
npm install --save-dev @graphql-tools/mock @graphql-tools/schema
```

```javascript
// vitest.config.js
import { defineConfig } from 'vitest/config';
import vue from '@vitejs/plugin-vue';

export default defineConfig({
  plugins: [vue()],
  test: {
    globals: true,
    environment: 'happy-dom',
    setupFiles: './src/tests/setup.js',
    coverage: {
      reporter: ['text', 'json', 'html'],
      exclude: [
        'node_modules/',
        'src/tests/',
      ],
    },
  },
  resolve: {
    alias: {
      '@': '/src',
    },
  },
});
```

## Testing de Queries y Mutations

### Mocks de Apollo Client

```javascript
// tests/utils/apolloMocks.js
import { InMemoryCache } from '@apollo/client';

export const createMockClient = () => {
  return {
    cache: new InMemoryCache(),
    defaultOptions: {
      watchQuery: {
        errorPolicy: 'all',
      },
      query: {
        errorPolicy: 'all',
      },
    },
  };
};

export const createMockResponse = (query, data, variables = {}) => ({
  request: {
    query,
    variables,
  },
  result: {
    data,
  },
});

export const createErrorResponse = (query, error, variables = {}) => ({
  request: {
    query,
    variables,
  },
  error,
});

// Mock builders para tipos comunes
export const mockMember = (overrides = {}) => ({
  __typename: 'Member',
  miembro_id: '1',
  numero_socio: '2024-001',
  nombre: 'Test',
  apellidos: 'User',
  estado: 'ACTIVE',
  tipo_membresia: 'INDIVIDUAL',
  correo_electronico: 'test@example.com',
  fecha_alta: '2024-01-01T00:00:00Z',
  ...overrides,
});

export const mockPayment = (overrides = {}) => ({
  __typename: 'Payment',
  id: '1',
  amount: 50.0,
  payment_date: '2024-01-15T00:00:00Z',
  status: 'PAID',
  payment_method: 'TRANSFERENCIA',
  ...overrides,
});

export const mockPageInfo = (overrides = {}) => ({
  __typename: 'PageInfo',
  hasNextPage: false,
  hasPreviousPage: false,
  totalCount: 0,
  ...overrides,
});
```

### Testing de Hooks con Queries

```javascript
// hooks/__tests__/useMembers.test.js
import { renderHook, waitFor } from '@testing-library/react';
import { MockedProvider } from '@apollo/client/testing';
import { useMembers } from '../useMembers';
import { LIST_MEMBERS_QUERY } from '@/graphql/queries';
import { mockMember, mockPageInfo } from '@/tests/utils/apolloMocks';

describe('useMembers', () => {
  const wrapper = ({ children, mocks = [] }) => (
    <MockedProvider mocks={mocks} addTypename={true}>
      {children}
    </MockedProvider>
  );

  it('should fetch members successfully', async () => {
    const mockData = {
      listMembers: {
        nodes: [
          mockMember({ miembro_id: '1', nombre: 'Juan' }),
          mockMember({ miembro_id: '2', nombre: 'María' }),
        ],
        pageInfo: mockPageInfo({ totalCount: 2 }),
      },
    };

    const mocks = [
      {
        request: {
          query: LIST_MEMBERS_QUERY,
          variables: { filter: { pagination: { page: 1, pageSize: 20 } } },
        },
        result: { data: mockData },
      },
    ];

    const { result } = renderHook(() => useMembers(), {
      wrapper: (props) => wrapper({ ...props, mocks }),
    });

    // Estado inicial
    expect(result.current.loading).toBe(true);
    expect(result.current.members).toEqual([]);

    // Esperar a que se resuelva
    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    // Verificar datos
    expect(result.current.members).toHaveLength(2);
    expect(result.current.members[0].nombre).toBe('Juan');
    expect(result.current.totalCount).toBe(2);
  });

  it('should handle errors', async () => {
    const mocks = [
      {
        request: {
          query: LIST_MEMBERS_QUERY,
          variables: { filter: { pagination: { page: 1, pageSize: 20 } } },
        },
        error: new Error('Network error'),
      },
    ];

    const { result } = renderHook(() => useMembers(), {
      wrapper: (props) => wrapper({ ...props, mocks }),
    });

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.error).toBeDefined();
    expect(result.current.error.message).toBe('Network error');
  });

  it('should refetch with new filters', async () => {
    const initialMock = {
      request: {
        query: LIST_MEMBERS_QUERY,
        variables: { filter: { pagination: { page: 1, pageSize: 20 } } },
      },
      result: {
        data: {
          listMembers: {
            nodes: [mockMember()],
            pageInfo: mockPageInfo({ totalCount: 1 }),
          },
        },
      },
    };

    const filteredMock = {
      request: {
        query: LIST_MEMBERS_QUERY,
        variables: {
          filter: {
            estado: 'ACTIVE',
            pagination: { page: 1, pageSize: 20 },
          },
        },
      },
      result: {
        data: {
          listMembers: {
            nodes: [mockMember({ estado: 'ACTIVE' })],
            pageInfo: mockPageInfo({ totalCount: 1 }),
          },
        },
      },
    };

    const { result, rerender } = renderHook(
      ({ filters }) => useMembers(filters),
      {
        initialProps: { filters: {} },
        wrapper: (props) => wrapper({ ...props, mocks: [initialMock, filteredMock] }),
      }
    );

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    // Cambiar filtros
    rerender({ filters: { estado: 'ACTIVE' } });

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.members[0].estado).toBe('ACTIVE');
  });
});
```

### Testing de Mutations

```javascript
// components/__tests__/CreateMemberForm.test.js
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MockedProvider } from '@apollo/client/testing';
import { CreateMemberForm } from '../CreateMemberForm';
import { CREATE_MEMBER_MUTATION } from '@/graphql/mutations';
import { mockMember } from '@/tests/utils/apolloMocks';

describe('CreateMemberForm', () => {
  const mockOnSuccess = jest.fn();
  const user = userEvent.setup();

  const renderComponent = (mocks = []) => {
    return render(
      <MockedProvider mocks={mocks} addTypename={false}>
        <CreateMemberForm onSuccess={mockOnSuccess} />
      </MockedProvider>
    );
  };

  beforeEach(() => {
    mockOnSuccess.mockClear();
  });

  it('should create member successfully', async () => {
    const newMember = mockMember({
      miembro_id: '123',
      nombre: 'Juan',
      apellidos: 'Pérez',
    });

    const mocks = [
      {
        request: {
          query: CREATE_MEMBER_MUTATION,
          variables: {
            input: {
              numero_socio: '2024-001',
              nombre: 'Juan',
              apellidos: 'Pérez',
              tipo_membresia: 'INDIVIDUAL',
              calle_numero_piso: 'Calle Test 123',
              codigo_postal: '07001',
              poblacion: 'Palma',
            },
          },
        },
        result: {
          data: {
            createMember: newMember,
          },
        },
      },
    ];

    renderComponent(mocks);

    // Llenar formulario
    await user.type(screen.getByLabelText(/número de socio/i), '2024-001');
    await user.type(screen.getByLabelText(/nombre/i), 'Juan');
    await user.type(screen.getByLabelText(/apellidos/i), 'Pérez');
    await user.selectOptions(screen.getByLabelText(/tipo de membresía/i), 'INDIVIDUAL');
    await user.type(screen.getByLabelText(/calle/i), 'Calle Test 123');
    await user.type(screen.getByLabelText(/código postal/i), '07001');
    await user.type(screen.getByLabelText(/población/i), 'Palma');

    // Enviar formulario
    await user.click(screen.getByRole('button', { name: /crear miembro/i }));

    // Verificar que se llamó la mutation y el callback
    await waitFor(() => {
      expect(mockOnSuccess).toHaveBeenCalledWith(newMember);
    });
  });

  it('should show validation errors', async () => {
    const mocks = [
      {
        request: {
          query: CREATE_MEMBER_MUTATION,
          variables: expect.any(Object),
        },
        error: {
          graphQLErrors: [
            {
              message: 'Validation failed',
              extensions: {
                code: 'VALIDATION_ERROR',
                details: {
                  numero_socio: 'El número de socio ya existe',
                  correo_electronico: 'Email inválido',
                },
              },
            },
          ],
        },
      },
    ];

    renderComponent(mocks);

    // Llenar parcialmente el formulario
    await user.type(screen.getByLabelText(/número de socio/i), '2024-001');
    await user.type(screen.getByLabelText(/nombre/i), 'Test');

    // Enviar
    await user.click(screen.getByRole('button', { name: /crear miembro/i }));

    // Verificar errores de validación
    await waitFor(() => {
      expect(screen.getByText(/el número de socio ya existe/i)).toBeInTheDocument();
      expect(screen.getByText(/email inválido/i)).toBeInTheDocument();
    });

    expect(mockOnSuccess).not.toHaveBeenCalled();
  });

  it('should handle network errors', async () => {
    const mocks = [
      {
        request: {
          query: CREATE_MEMBER_MUTATION,
          variables: expect.any(Object),
        },
        error: new Error('Network error'),
      },
    ];

    renderComponent(mocks);

    // Llenar formulario mínimo
    await user.type(screen.getByLabelText(/nombre/i), 'Test');

    // Enviar
    await user.click(screen.getByRole('button', { name: /crear miembro/i }));

    // Verificar mensaje de error
    await waitFor(() => {
      expect(screen.getByText(/error de conexión/i)).toBeInTheDocument();
    });
  });
});
```

## Mocking de Datos

### Mock Service Worker (MSW)

```javascript
// mocks/handlers.js
import { graphql } from 'msw';
import { mockMember, mockPageInfo } from '../tests/utils/apolloMocks';

export const handlers = [
  // Login
  graphql.mutation('Login', (req, res, ctx) => {
    const { username, password } = req.variables.input;

    if (username === 'admin@example.com' && password === 'password') {
      return res(
        ctx.data({
          login: {
            user: {
              id: '1',
              username: 'admin@example.com',
              role: 'ADMIN',
              isActive: true,
            },
            accessToken: 'mock-access-token',
            refreshToken: 'mock-refresh-token',
            expiresAt: new Date(Date.now() + 900000).toISOString(),
          },
        })
      );
    }

    return res(
      ctx.errors([
        {
          message: 'Invalid credentials',
          extensions: {
            code: 'INVALID_CREDENTIALS',
          },
        },
      ])
    );
  }),

  // List Members
  graphql.query('ListMembers', (req, res, ctx) => {
    const { filter } = req.variables;
    const page = filter?.pagination?.page || 1;
    const pageSize = filter?.pagination?.pageSize || 20;

    // Simular paginación
    const totalMembers = 50;
    const members = Array.from({ length: pageSize }, (_, i) =>
      mockMember({
        miembro_id: `${(page - 1) * pageSize + i + 1}`,
        numero_socio: `2024-${String((page - 1) * pageSize + i + 1).padStart(3, '0')}`,
      })
    );

    return res(
      ctx.data({
        listMembers: {
          nodes: members,
          pageInfo: mockPageInfo({
            hasNextPage: page * pageSize < totalMembers,
            hasPreviousPage: page > 1,
            totalCount: totalMembers,
          }),
        },
      })
    );
  }),

  // Get Member
  graphql.query('GetMember', (req, res, ctx) => {
    const { id } = req.variables;

    if (id === '999') {
      return res(
        ctx.errors([
          {
            message: 'Member not found',
            extensions: {
              code: 'MEMBER_NOT_FOUND',
            },
          },
        ])
      );
    }

    return res(
      ctx.data({
        getMember: mockMember({ miembro_id: id }),
      })
    );
  }),

  // Create Member
  graphql.mutation('CreateMember', (req, res, ctx) => {
    const { input } = req.variables;

    // Simular validación
    if (!input.nombre || !input.apellidos) {
      return res(
        ctx.errors([
          {
            message: 'Validation failed',
            extensions: {
              code: 'VALIDATION_ERROR',
              details: {
                nombre: !input.nombre ? 'Nombre es requerido' : null,
                apellidos: !input.apellidos ? 'Apellidos son requeridos' : null,
              },
            },
          },
        ])
      );
    }

    return res(
      ctx.data({
        createMember: mockMember({
          ...input,
          miembro_id: String(Date.now()),
          fecha_alta: new Date().toISOString(),
        }),
      })
    );
  }),
];

// mocks/server.js
import { setupServer } from 'msw/node';
import { handlers } from './handlers';

export const server = setupServer(...handlers);

// mocks/browser.js
import { setupWorker } from 'msw';
import { handlers } from './handlers';

export const worker = setupWorker(...handlers);
```

### Factory Functions para Datos de Test

```javascript
// tests/factories/index.js
import { faker } from '@faker-js/faker/locale/es';

export const memberFactory = {
  build: (overrides = {}) => ({
    miembro_id: faker.string.uuid(),
    numero_socio: `2024-${faker.string.numeric(3)}`,
    tipo_membresia: faker.helpers.arrayElement(['INDIVIDUAL', 'FAMILY']),
    nombre: faker.person.firstName(),
    apellidos: `${faker.person.lastName()} ${faker.person.lastName()}`,
    calle_numero_piso: faker.location.streetAddress(),
    codigo_postal: faker.location.zipCode(),
    poblacion: faker.location.city(),
    provincia: faker.location.state(),
    pais: 'España',
    estado: faker.helpers.arrayElement(['ACTIVE', 'INACTIVE']),
    fecha_alta: faker.date.past().toISOString(),
    fecha_nacimiento: faker.date.birthdate({ min: 18, max: 80 }).toISOString(),
    documento_identidad: generateDNI(),
    correo_electronico: faker.internet.email(),
    profesion: faker.person.jobTitle(),
    nacionalidad: 'Española',
    observaciones: faker.lorem.sentence(),
    ...overrides,
  }),

  buildList: (count, overrides = {}) => {
    return Array.from({ length: count }, () => memberFactory.build(overrides));
  },
};

export const paymentFactory = {
  build: (overrides = {}) => ({
    id: faker.string.uuid(),
    amount: parseFloat(faker.finance.amount(10, 200)),
    payment_date: faker.date.recent().toISOString(),
    status: faker.helpers.arrayElement(['PENDING', 'PAID', 'CANCELLED']),
    payment_method: faker.helpers.arrayElement(['EFECTIVO', 'TRANSFERENCIA', 'TARJETA']),
    notes: faker.lorem.sentence(),
    member: overrides.member || memberFactory.build(),
    ...overrides,
  }),
};

function generateDNI() {
  const number = faker.string.numeric(8);
  const letters = 'TRWAGMYFPDXBNJZSQVHLCKE';
  const letter = letters[parseInt(number) % 23];
  return `${number}${letter}`;
}
```

## Testing de Componentes

### Testing de Formularios

```javascript
// components/__tests__/MemberForm.test.js
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemberForm } from '../MemberForm';

describe('MemberForm', () => {
  const user = userEvent.setup();
  const mockOnSubmit = jest.fn();

  beforeEach(() => {
    mockOnSubmit.mockClear();
  });

  it('should validate required fields', async () => {
    render(<MemberForm onSubmit={mockOnSubmit} />);

    // Intentar enviar sin llenar campos requeridos
    const submitButton = screen.getByRole('button', { name: /guardar/i });
    await user.click(submitButton);

    // Verificar mensajes de error
    expect(screen.getByText(/el nombre es requerido/i)).toBeInTheDocument();
    expect(screen.getByText(/los apellidos son requeridos/i)).toBeInTheDocument();
    expect(mockOnSubmit).not.toHaveBeenCalled();
  });

  it('should validate email format', async () => {
    render(<MemberForm onSubmit={mockOnSubmit} />);

    const emailInput = screen.getByLabelText(/email/i);
    await user.type(emailInput, 'invalid-email');
    await user.tab(); // Trigger blur

    await waitFor(() => {
      expect(screen.getByText(/email inválido/i)).toBeInTheDocument();
    });
  });

  it('should validate DNI format', async () => {
    render(<MemberForm onSubmit={mockOnSubmit} />);

    const dniInput = screen.getByLabelText(/dni/i);
    
    // DNI inválido
    await user.type(dniInput, '12345678');
    await user.tab();

    await waitFor(() => {
      expect(screen.getByText(/formato dni.*inválido/i)).toBeInTheDocument();
    });

    // DNI válido
    await user.clear(dniInput);
    await user.type(dniInput, '12345678Z');
    await user.tab();

    await waitFor(() => {
      expect(screen.queryByText(/formato dni.*inválido/i)).not.toBeInTheDocument();
    });
  });

  it('should submit valid form data', async () => {
    render(<MemberForm onSubmit={mockOnSubmit} />);

    // Llenar formulario
    await user.type(screen.getByLabelText(/número de socio/i), '2024-001');
    await user.type(screen.getByLabelText(/nombre/i), 'Juan');
    await user.type(screen.getByLabelText(/apellidos/i), 'Pérez García');
    await user.type(screen.getByLabelText(/calle/i), 'Calle Principal 123');
    await user.type(screen.getByLabelText(/código postal/i), '07001');
    await user.type(screen.getByLabelText(/población/i), 'Palma');

    // Enviar
    await user.click(screen.getByRole('button', { name: /guardar/i }));

    await waitFor(() => {
      expect(mockOnSubmit).toHaveBeenCalledWith({
        numero_socio: '2024-001',
        nombre: 'Juan',
        apellidos: 'Pérez García',
        calle_numero_piso: 'Calle Principal 123',
        codigo_postal: '07001',
        poblacion: 'Palma',
        tipo_membresia: 'INDIVIDUAL', // Valor por defecto
        pais: 'España', // Valor por defecto
      });
    });
  });
});
```

### Testing de Listas y Tablas

```javascript
// components/__tests__/MemberTable.test.js
import { render, screen, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemberTable } from '../MemberTable';
import { memberFactory } from '@/tests/factories';

describe('MemberTable', () => {
  const user = userEvent.setup();
  const mockOnSort = jest.fn();
  const mockOnStatusChange = jest.fn();

  const defaultProps = {
    members: memberFactory.buildList(5),
    loading: false,
    onSort: mockOnSort,
    onStatusChange: mockOnStatusChange,
    sortConfig: { field: 'nombre', direction: 'ASC' },
  };

  it('should render members correctly', () => {
    render(<MemberTable {...defaultProps} />);

    // Verificar headers
    expect(screen.getByText(/número socio/i)).toBeInTheDocument();
    expect(screen.getByText(/nombre/i)).toBeInTheDocument();
    expect(screen.getByText(/estado/i)).toBeInTheDocument();

    // Verificar que se renderizan todos los miembros
    defaultProps.members.forEach((member) => {
      expect(screen.getByText(member.numero_socio)).toBeInTheDocument();
      expect(screen.getByText(member.nombre)).toBeInTheDocument();
    });
  });

  it('should handle sorting', async () => {
    render(<MemberTable {...defaultProps} />);

    const nombreHeader = screen.getByRole('columnheader', { name: /nombre/i });
    await user.click(nombreHeader);

    expect(mockOnSort).toHaveBeenCalledWith('nombre');
  });

  it('should show sort indicators', () => {
    render(<MemberTable {...defaultProps} />);

    const nombreHeader = screen.getByRole('columnheader', { name: /nombre/i });
    const sortIcon = within(nombreHeader).getByTestId('sort-icon-asc');

    expect(sortIcon).toBeInTheDocument();
  });

  it('should handle status change', async () => {
    const members = [
      memberFactory.build({ miembro_id: '1', estado: 'ACTIVE' }),
    ];

    render(<MemberTable {...defaultProps} members={members} />);

    const statusButton = screen.getByRole('button', { name: /desactivar/i });
    await user.click(statusButton);

    expect(mockOnStatusChange).toHaveBeenCalledWith('1', 'INACTIVE');
  });

  it('should show loading state', () => {
    render(<MemberTable {...defaultProps} loading={true} />);

    expect(screen.getByTestId('loading-spinner')).toBeInTheDocument();
  });

  it('should show empty state', () => {
    render(<MemberTable {...defaultProps} members={[]} />);

    expect(screen.getByText(/no hay miembros para mostrar/i)).toBeInTheDocument();
  });
});
```

## Testing de Integración

### Testing de Flujos Completos

```javascript
// integration/__tests__/memberManagement.test.js
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { MockedProvider } from '@apollo/client/testing';
import { MemberListPage } from '@/pages/MemberListPage';
import { MemberDetailPage } from '@/pages/MemberDetailPage';
import { CreateMemberPage } from '@/pages/CreateMemberPage';
import { server } from '@/mocks/server';
import { graphql } from 'msw';

describe('Member Management Flow', () => {
  const user = userEvent.setup();

  const renderWithRouter = (initialRoute = '/members') => {
    return render(
      <MockedProvider>
        <MemoryRouter initialEntries={[initialRoute]}>
          <Routes>
            <Route path="/members" element={<MemberListPage />} />
            <Route path="/members/new" element={<CreateMemberPage />} />
            <Route path="/members/:id" element={<MemberDetailPage />} />
          </Routes>
        </MemoryRouter>
      </MockedProvider>
    );
  };

  it('should complete full member creation flow', async () => {
    renderWithRouter('/members');

    // 1. Click en nuevo miembro
    const newButton = await screen.findByRole('button', { name: /nuevo miembro/i });
    await user.click(newButton);

    // 2. Verificar que estamos en el formulario
    expect(screen.getByText(/crear nuevo miembro/i)).toBeInTheDocument();

    // 3. Llenar formulario
    await user.type(screen.getByLabelText(/nombre/i), 'Juan');
    await user.type(screen.getByLabelText(/apellidos/i), 'Pérez');
    await user.type(screen.getByLabelText(/número de socio/i), '2024-100');
    await user.type(screen.getByLabelText(/calle/i), 'Calle Test 123');
    await user.type(screen.getByLabelText(/código postal/i), '07001');
    await user.type(screen.getByLabelText(/población/i), 'Palma');

    // 4. Enviar formulario
    const submitButton = screen.getByRole('button', { name: /crear miembro/i });
    await user.click(submitButton);

    // 5. Verificar redirección a lista
    await waitFor(() => {
      expect(screen.getByText(/gestión de miembros/i)).toBeInTheDocument();
    });

    // 6. Verificar que el nuevo miembro aparece en la lista
    expect(await screen.findByText('2024-100')).toBeInTheDocument();
    expect(screen.getByText('Juan')).toBeInTheDocument();
  });

  it('should handle search and filter', async () => {
    // Mock para búsqueda
    server.use(
      graphql.query('ListMembers', (req, res, ctx) => {
        const searchTerm = req.variables.filter?.search_term;
        
        if (searchTerm === 'Juan') {
          return res(
            ctx.data({
              listMembers: {
                nodes: [
                  {
                    miembro_id: '1',
                    numero_socio: '2024-001',
                    nombre: 'Juan',
                    apellidos: 'Pérez',
                    estado: 'ACTIVE',
                    tipo_membresia: 'INDIVIDUAL',
                  },
                ],
                pageInfo: {
                  hasNextPage: false,
                  hasPreviousPage: false,
                  totalCount: 1,
                },
              },
            })
          );
        }

        return res(ctx.data({ listMembers: { nodes: [], pageInfo: {} } }));
      })
    );

    renderWithRouter('/members');

    // Buscar
    const searchInput = await screen.findByPlaceholderText(/buscar/i);
    await user.type(searchInput, 'Juan');

    // Esperar debounce y verificar resultados
    await waitFor(() => {
      expect(screen.getByText('Juan')).toBeInTheDocument();
    }, { timeout: 1000 });

    // Filtrar por estado
    const statusFilter = screen.getByLabelText(/estado/i);
    await user.selectOptions(statusFilter, 'ACTIVE');

    // Verificar que se mantienen los resultados
    expect(screen.getByText('Juan')).toBeInTheDocument();
  });
});
```

## Testing E2E

### Cypress para React

```javascript
// cypress/e2e/memberManagement.cy.js
describe('Member Management E2E', () => {
  beforeEach(() => {
    // Login
    cy.visit('/login');
    cy.get('[data-cy=username]').type('admin@example.com');
    cy.get('[data-cy=password]').type('password123');
    cy.get('[data-cy=login-button]').click();
    
    // Esperar redirección
    cy.url().should('include', '/dashboard');
  });

  it('should create and view a member', () => {
    // Navegar a miembros
    cy.get('[data-cy=nav-members]').click();
    cy.url().should('include', '/members');

    // Click en nuevo
    cy.get('[data-cy=new-member-button]').click();

    // Llenar formulario
    cy.get('[data-cy=numero-socio]').type('2024-E2E');
    cy.get('[data-cy=nombre]').type('Test');
    cy.get('[data-cy=apellidos]').type('E2E User');
    cy.get('[data-cy=tipo-membresia]').select('INDIVIDUAL');
    cy.get('[data-cy=calle]').type('Calle E2E 123');
    cy.get('[data-cy=codigo-postal]').type('07001');
    cy.get('[data-cy=poblacion]').type('Palma');
    cy.get('[data-cy=email]').type('e2e@test.com');

    // Enviar
    cy.get('[data-cy=submit-button]').click();

    // Verificar notificación
    cy.get('[data-cy=notification]').should('contain', 'Miembro creado');

    // Verificar que aparece en la lista
    cy.get('[data-cy=member-table]').should('contain', '2024-E2E');
    cy.get('[data-cy=member-table]').should('contain', 'Test E2E User');

    // Click para ver detalle
    cy.get('[data-cy=member-row-2024-E2E]').click();

    // Verificar detalle
    cy.url().should('match', /\/members\/\d+/);
    cy.get('[data-cy=member-detail]').should('contain', 'Test E2E User');
    cy.get('[data-cy=member-detail]').should('contain', 'e2e@test.com');
  });

  it('should handle pagination', () => {
    cy.visit('/members');

    // Verificar paginación
    cy.get('[data-cy=pagination-info]').should('contain', 'Página 1');
    
    // Ir a siguiente página
    cy.get('[data-cy=next-page]').click();
    cy.get('[data-cy=pagination-info]').should('contain', 'Página 2');

    // Cambiar tamaño de página
    cy.get('[data-cy=page-size]').select('50');
    cy.get('[data-cy=member-table] tbody tr').should('have.length.at.least', 30);
  });

  it('should export data', () => {
    cy.visit('/members');

    // Click en exportar
    cy.get('[data-cy=export-button]').click();

    // Verificar descarga
    cy.readFile('cypress/downloads/miembros_export.csv').should('exist');
  });
});
```

### Playwright para Vue

```javascript
// tests/e2e/memberManagement.spec.js
import { test, expect } from '@playwright/test';

test.describe('Member Management', () => {
  test.beforeEach(async ({ page }) => {
    // Login
    await page.goto('/login');
    await page.fill('[data-testid=username]', 'admin@example.com');
    await page.fill('[data-testid=password]', 'password123');
    await page.click('[data-testid=login-button]');
    
    // Esperar navegación
    await page.waitForURL('**/dashboard');
  });

  test('should perform CRUD operations', async ({ page }) => {
    // Ir a miembros
    await page.click('[data-testid=nav-members]');
    await expect(page).toHaveURL(/.*\/members/);

    // Crear miembro
    await page.click('[data-testid=new-member]');
    
    // Llenar formulario
    await page.fill('[data-testid=numero-socio]', '2024-PW');
    await page.fill('[data-testid=nombre]', 'Playwright');
    await page.fill('[data-testid=apellidos]', 'Test User');
    await page.selectOption('[data-testid=tipo-membresia]', 'INDIVIDUAL');
    await page.fill('[data-testid=calle]', 'Calle Test 456');
    await page.fill('[data-testid=codigo-postal]', '07002');
    await page.fill('[data-testid=poblacion]', 'Palma');

    // Enviar
    await page.click('[data-testid=submit]');

    // Verificar notificación
    await expect(page.locator('[data-testid=notification]')).toContainText('creado exitosamente');

    // Buscar el miembro creado
    await page.fill('[data-testid=search]', '2024-PW');
    await page.waitForTimeout(500); // Esperar debounce

    // Verificar que aparece
    const memberRow = page.locator('[data-testid=member-2024-PW]');
    await expect(memberRow).toBeVisible();

    // Editar
    await memberRow.locator('[data-testid=edit]').click();
    await page.fill('[data-testid=email]', 'updated@test.com');
    await page.click('[data-testid=save]');

    // Verificar actualización
    await expect(page.locator('[data-testid=notification]')).toContainText('actualizado');

    // Eliminar
    await memberRow.locator('[data-testid=delete]').click();
    await page.click('[data-testid=confirm-delete]');

    // Verificar eliminación
    await expect(memberRow).not.toBeVisible();
  });

  test('should handle errors gracefully', async ({ page }) => {
    // Simular error de red
    await page.route('**/graphql', route => route.abort());

    await page.goto('/members');

    // Verificar mensaje de error
    await expect(page.locator('[data-testid=error-message]')).toContainText('Error de conexión');

    // Verificar botón de reintentar
    await expect(page.locator('[data-testid=retry-button]')).toBeVisible();
  });
});
```

## Mejores Prácticas

### 1. Estructura de Tests

```javascript
// Organización recomendada
src/
├── components/
│   ├── MemberForm/
│   │   ├── MemberForm.jsx
│   │   ├── MemberForm.test.jsx
│   │   └── MemberForm.stories.jsx
│   └── __tests__/
│       └── integration/
├── hooks/
│   ├── useMembers.js
│   └── __tests__/
│       └── useMembers.test.js
├── tests/
│   ├── setup.js
│   ├── utils/
│   ├── factories/
│   └── fixtures/
└── e2e/
    ├── cypress/
    └── playwright/
```

### 2. Patrón AAA (Arrange, Act, Assert)

```javascript
it('should update member status', async () => {
  // Arrange
  const member = memberFactory.build({ estado: 'ACTIVE' });
  const onStatusChange = jest.fn();
  render(<MemberCard member={member} onStatusChange={onStatusChange} />);

  // Act
  await user.click(screen.getByRole('button', { name: /desactivar/i }));

  // Assert
  expect(onStatusChange).toHaveBeenCalledWith(member.miembro_id, 'INACTIVE');
});
```

### 3. Custom Render Functions

```javascript
// tests/utils/test-utils.js
import { render } from '@testing-library/react';
import { MockedProvider } from '@apollo/client/testing';
import { MemoryRouter } from 'react-router-dom';
import { AuthProvider } from '@/contexts/AuthContext';
import { NotificationProvider } from '@/contexts/NotificationContext';

export function renderWithProviders(
  ui,
  {
    mocks = [],
    initialRoute = '/',
    authState = { isAuthenticated: true, user: { role: 'ADMIN' } },
    ...renderOptions
  } = {}
) {
  function Wrapper({ children }) {
    return (
      <MockedProvider mocks={mocks} addTypename={false}>
        <MemoryRouter initialEntries={[initialRoute]}>
          <AuthProvider initialState={authState}>
            <NotificationProvider>
              {children}
            </NotificationProvider>
          </AuthProvider>
        </MemoryRouter>
      </MockedProvider>
    );
  }

  return {
    ...render(ui, { wrapper: Wrapper, ...renderOptions }),
  };
}

// Re-export todo
export * from '@testing-library/react';
export { renderWithProviders as render };
```

### 4. Test Data Builders

```javascript
// tests/builders/memberBuilder.js
export class MemberBuilder {
  constructor() {
    this.member = {
      miembro_id: '1',
      numero_socio: '2024-001',
      nombre: 'Test',
      apellidos: 'User',
      estado: 'ACTIVE',
      tipo_membresia: 'INDIVIDUAL',
    };
  }

  withId(id) {
    this.member.miembro_id = id;
    return this;
  }

  withName(nombre, apellidos) {
    this.member.nombre = nombre;
    this.member.apellidos = apellidos;
    return this;
  }

  inactive() {
    this.member.estado = 'INACTIVE';
    return this;
  }

  familyMember() {
    this.member.tipo_membresia = 'FAMILY';
    return this;
  }

  build() {
    return { ...this.member };
  }
}

// Uso
const member = new MemberBuilder()
  .withName('Juan', 'Pérez')
  .inactive()
  .build();
```

### 5. Snapshot Testing Selectivo

```javascript
// Solo para componentes estables
it('should match snapshot for member card', () => {
  const member = memberFactory.build();
  const { container } = render(<MemberCard member={member} />);
  
  // Snapshot solo del contenido relevante
  const card = container.querySelector('[data-testid=member-card]');
  expect(card).toMatchSnapshot();
});

// Evitar snapshots de:
// - Listas grandes
// - Componentes que cambian frecuentemente
// - Datos con timestamps
```

### 6. Coverage Guidelines

```javascript
// jest.config.js
module.exports = {
  collectCoverageFrom: [
    'src/**/*.{js,jsx}',
    '!src/**/*.stories.js',
    '!src/tests/**',
    '!src/mocks/**',
  ],
  coverageThreshold: {
    global: {
      branches: 80,
      functions: 80,
      lines: 80,
      statements: 80,
    },
    // Umbrales específicos para código crítico
    'src/hooks/useAuth.js': {
      branches: 95,
      functions: 95,
      lines: 95,
      statements: 95,
    },
  },
};
```

Esta guía proporciona una base sólida para implementar testing completo en aplicaciones frontend que consumen el backend de ASAM.
