# Limpieza de Deuda Técnica - Fase 3 (Propuesta)

## Resumen de Cambios Propuestos

Esta es la tercera fase de limpieza de deuda técnica, enfocada en elementos de prioridad baja.

### Cambios Propuestos

#### 1. Reorganizar scripts por categorías

**Estructura propuesta**:
```
scripts/
├── db/                    # Scripts relacionados con base de datos
│   ├── migrate.ps1
│   ├── seed.ps1
│   └── setup_and_migrate.ps1
├── dev/                   # Utilidades de desarrollo
│   ├── cleanup-tech-debt-phase*.ps1/sh
│   ├── install-hooks.bat/sh
│   └── generate-gql.ps1
├── ops/                   # Scripts operacionales
│   ├── Set-CloudRunEnv.ps1
│   └── set-cloudrun-env.sh
├── user-management/       # (mantener como está)
└── verification/          # Scripts de verificación de email
    ├── check-email-verification-status.ps1
    ├── check-token-info.ps1
    ├── cleanup-verification-tokens.ps1
    └── manually-verify-user.ps1
```

#### 2. Scripts a eliminar o consolidar

**Candidatos a revisión**:
- Scripts duplicados con misma funcionalidad
- Scripts obsoletos que ya no se usan
- Scripts temporales de migración completada

#### 3. Documentación de scripts

**Acción**: Crear un `scripts/README.md` que documente:
- Propósito de cada carpeta
- Descripción breve de cada script
- Orden de ejecución cuando aplique
- Dependencias entre scripts

### Impacto Esperado

- **Mejor organización**: Scripts agrupados por propósito
- **Fácil navegación**: Estructura lógica e intuitiva
- **Documentación clara**: README dedicado para scripts
- **Menos confusión**: Eliminación de scripts obsoletos

### Análisis Necesario

Antes de implementar esta fase, se requiere:
1. Auditoría completa de todos los scripts existentes
2. Identificar scripts obsoletos o duplicados
3. Verificar dependencias entre scripts
4. Determinar la mejor categorización

### Beneficios

- Onboarding más rápido para nuevos desarrolladores
- Menor tiempo buscando el script correcto
- Reducción de errores por usar scripts incorrectos
- Mantenimiento más sencillo

---

**Nota**: Esta es una propuesta. La implementación específica puede ajustarse según el análisis detallado de los scripts existentes.
