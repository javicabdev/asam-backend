# Guía de Deployment para Frontend

Esta guía proporciona instrucciones detalladas para desplegar aplicaciones frontend que consumen el backend de ASAM.

## Tabla de Contenidos
1. [Preparación del Build](#preparación-del-build)
2. [Configuración de Entornos](#configuración-de-entornos)
3. [Deployment en Diferentes Plataformas](#deployment-en-diferentes-plataformas)
4. [CI/CD Pipeline](#cicd-pipeline)
5. [Monitoreo Post-Deployment](#monitoreo-post-deployment)
6. [Rollback Strategy](#rollback-strategy)

## Preparación del Build

### 1. Optimización del Build

```javascript
// webpack.prod.js
const path = require('path');
const HtmlWebpackPlugin = require('html-webpack-plugin');
const MiniCssExtractPlugin = require('mini-css-extract-plugin');
const CssMinimizerPlugin = require('css-minimizer-webpack-plugin');
const TerserPlugin = require('terser-webpack-plugin');
const CompressionPlugin = require('compression-webpack-plugin');
const { BundleAnalyzerPlugin } = require('webpack-bundle-analyzer');

module.exports = {
  mode: 'production',
  entry: './src/index.js',
  output: {
    path: path.resolve(__dirname, 'dist'),
    filename: '[name].[contenthash].js',
    chunkFilename: '[name].[contenthash].chunk.js',
    clean: true
  },
  
  optimization: {
    minimize: true,
    minimizer: [
      new TerserPlugin({
        terserOptions: {
          compress: {
            drop_console: true,
            drop_debugger: true
          },
          mangle: true
        }
      }),
      new CssMinimizerPlugin()
    ],
    
    splitChunks: {
      chunks: 'all',
      cacheGroups: {
        vendor: {
          test: /[\\/]node_modules[\\/]/,
          name: 'vendors',
          priority: 10
        },
        common: {
          minChunks: 2,
          priority: 5,
          reuseExistingChunk: true
        }
      }
    },
    
    runtimeChunk: 'single'
  },
  
  plugins: [
    new HtmlWebpackPlugin({
      template: './public/index.html',
      minify: {
        removeComments: true,
        collapseWhitespace: true,
        removeAttributeQuotes: true
      }
    }),
    
    new MiniCssExtractPlugin({
      filename: '[name].[contenthash].css',
      chunkFilename: '[name].[contenthash].chunk.css'
    }),
    
    new CompressionPlugin({
      algorithm: 'gzip',
      test: /\.(js|css|html|svg)$/,
      threshold: 8192,
      minRatio: 0.8
    }),
    
    new BundleAnalyzerPlugin({
      analyzerMode: 'static',
      openAnalyzer: false,
      reportFilename: 'bundle-report.html'
    })
  ]
};
```

### 2. Variables de Entorno

```javascript
// config/env.js
const fs = require('fs');
const path = require('path');
const dotenv = require('dotenv');

function getClientEnvironment(nodeEnv) {
  // Cargar archivos .env en orden de prioridad
  const dotenvFiles = [
    `.env.${nodeEnv}.local`,
    `.env.${nodeEnv}`,
    '.env.local',
    '.env'
  ].filter(Boolean);

  // Cargar variables de entorno
  dotenvFiles.forEach(dotenvFile => {
    if (fs.existsSync(dotenvFile)) {
      dotenv.config({ path: dotenvFile });
    }
  });

  // Variables que serán expuestas al cliente
  const raw = Object.keys(process.env)
    .filter(key => /^REACT_APP_/i.test(key))
    .reduce((env, key) => {
      env[key] = process.env[key];
      return env;
    }, {
      NODE_ENV: nodeEnv,
      PUBLIC_URL: process.env.PUBLIC_URL || ''
    });

  // Stringify para webpack.DefinePlugin
  const stringified = {
    'process.env': Object.keys(raw).reduce((env, key) => {
      env[key] = JSON.stringify(raw[key]);
      return env;
    }, {})
  };

  return { raw, stringified };
}

module.exports = getClientEnvironment;
```

### 3. Archivos de Configuración por Entorno

```bash
# .env.production
REACT_APP_API_URL=https://api.asam.org
REACT_APP_GRAPHQL_URL=https://api.asam.org/graphql
REACT_APP_SENTRY_DSN=https://your-sentry-dsn
REACT_APP_GA_TRACKING_ID=UA-XXXXXXXXX-X
REACT_APP_ENVIRONMENT=production

# .env.staging
REACT_APP_API_URL=https://staging-api.asam.org
REACT_APP_GRAPHQL_URL=https://staging-api.asam.org/graphql
REACT_APP_SENTRY_DSN=https://your-staging-sentry-dsn
REACT_APP_GA_TRACKING_ID=UA-STAGING-ID
REACT_APP_ENVIRONMENT=staging
```

## Configuración de Entornos

### 1. Build Scripts

```json
// package.json
{
  "scripts": {
    "build:dev": "cross-env NODE_ENV=development webpack --config webpack.dev.js",
    "build:staging": "cross-env NODE_ENV=staging webpack --config webpack.prod.js",
    "build:prod": "cross-env NODE_ENV=production webpack --config webpack.prod.js",
    "analyze": "cross-env NODE_ENV=production webpack --config webpack.prod.js --analyze",
    "test:build": "npm run build:prod && serve -s dist -l 3000"
  }
}
```

### 2. Configuración de Runtime

```javascript
// src/config/runtime.js
class RuntimeConfig {
  constructor() {
    this.config = null;
  }

  async load() {
    try {
      // Cargar configuración del servidor en runtime
      const response = await fetch('/config.json');
      this.config = await response.json();
    } catch (error) {
      console.warn('Using build-time config');
      this.config = {
        apiUrl: process.env.REACT_APP_API_URL,
        graphqlUrl: process.env.REACT_APP_GRAPHQL_URL,
        environment: process.env.REACT_APP_ENVIRONMENT
      };
    }
  }

  get(key) {
    return this.config?.[key];
  }
}

export const runtimeConfig = new RuntimeConfig();

// En index.js
import { runtimeConfig } from './config/runtime';

async function initApp() {
  await runtimeConfig.load();
  
  const root = ReactDOM.createRoot(document.getElementById('root'));
  root.render(
    <React.StrictMode>
      <App />
    </React.StrictMode>
  );
}

initApp();
```

## Deployment en Diferentes Plataformas

### 1. Vercel

```json
// vercel.json
{
  "version": 2,
  "builds": [
    {
      "src": "package.json",
      "use": "@vercel/static-build",
      "config": {
        "distDir": "dist"
      }
    }
  ],
  "routes": [
    {
      "src": "/static/(.*)",
      "headers": {
        "Cache-Control": "public, max-age=31536000, immutable"
      }
    },
    {
      "src": "/(.*)\\.(js|css|ico|png|jpg|jpeg|gif|svg|woff|woff2)$",
      "headers": {
        "Cache-Control": "public, max-age=86400"
      }
    },
    {
      "src": "/(.*)",
      "dest": "/index.html"
    }
  ],
  "headers": [
    {
      "source": "/(.*)",
      "headers": [
        {
          "key": "X-Frame-Options",
          "value": "DENY"
        },
        {
          "key": "X-Content-Type-Options",
          "value": "nosniff"
        },
        {
          "key": "X-XSS-Protection",
          "value": "1; mode=block"
        },
        {
          "key": "Referrer-Policy",
          "value": "strict-origin-when-cross-origin"
        }
      ]
    }
  ]
}
```

### 2. Netlify

```toml
# netlify.toml
[build]
  command = "npm run build:prod"
  publish = "dist"

[build.environment]
  NODE_VERSION = "18"

[[redirects]]
  from = "/*"
  to = "/index.html"
  status = 200

[[headers]]
  for = "/*"
  [headers.values]
    X-Frame-Options = "DENY"
    X-XSS-Protection = "1; mode=block"
    X-Content-Type-Options = "nosniff"
    Referrer-Policy = "strict-origin-when-cross-origin"

[[headers]]
  for = "/static/*"
  [headers.values]
    Cache-Control = "public, max-age=31536000, immutable"

# Configuración de funciones para proxy
[[redirects]]
  from = "/api/*"
  to = "https://api.asam.org/:splat"
  status = 200
  force = true
```

### 3. AWS S3 + CloudFront

```bash
#!/bin/bash
# deploy-to-s3.sh

# Variables
BUCKET_NAME="asam-frontend-prod"
DISTRIBUTION_ID="E1234567890ABC"
BUILD_DIR="./dist"

# Build
echo "Building application..."
npm run build:prod

# Sync to S3
echo "Uploading to S3..."
aws s3 sync $BUILD_DIR s3://$BUCKET_NAME \
  --delete \
  --cache-control max-age=31536000,public \
  --exclude index.html

# Upload index.html with no-cache
aws s3 cp $BUILD_DIR/index.html s3://$BUCKET_NAME/index.html \
  --cache-control max-age=0,no-cache,no-store,must-revalidate

# Invalidate CloudFront
echo "Invalidating CloudFront cache..."
aws cloudfront create-invalidation \
  --distribution-id $DISTRIBUTION_ID \
  --paths "/*"

echo "Deployment complete!"
```

### 4. Docker

```dockerfile
# Dockerfile
FROM node:18-alpine as builder

WORKDIR /app

# Cache dependencies
COPY package*.json ./
RUN npm ci --only=production

# Copy source
COPY . .

# Build
ARG NODE_ENV=production
ENV NODE_ENV=$NODE_ENV
RUN npm run build:prod

# Production image
FROM nginx:alpine

# Copy custom nginx config
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Copy built app
COPY --from=builder /app/dist /usr/share/nginx/html

# Add runtime config script
COPY docker-entrypoint.sh /
RUN chmod +x /docker-entrypoint.sh

EXPOSE 80

ENTRYPOINT ["/docker-entrypoint.sh"]
CMD ["nginx", "-g", "daemon off;"]
```

```nginx
# nginx.conf
server {
    listen 80;
    server_name localhost;
    root /usr/share/nginx/html;
    index index.html;

    # Gzip
    gzip on;
    gzip_vary on;
    gzip_min_length 1024;
    gzip_types text/plain text/css text/xml text/javascript application/javascript application/xml+rss application/json;

    # Security headers
    add_header X-Frame-Options "DENY" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header Referrer-Policy "strict-origin-when-cross-origin" always;
    add_header Content-Security-Policy "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; connect-src 'self' https://api.asam.org" always;

    # Cache static assets
    location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg|woff|woff2)$ {
        expires 1y;
        add_header Cache-Control "public, immutable";
    }

    # Don't cache index.html
    location = /index.html {
        expires -1;
        add_header Cache-Control "no-cache, no-store, must-revalidate";
    }

    # SPA fallback
    location / {
        try_files $uri $uri/ /index.html;
    }
}
```

## CI/CD Pipeline

### 1. GitHub Actions

```yaml
# .github/workflows/deploy.yml
name: Deploy to Production

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

env:
  NODE_VERSION: '18'

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Setup Node.js
        uses: actions/setup-node@v3
        with:
          node-version: ${{ env.NODE_VERSION }}
          cache: 'npm'
      
      - name: Install dependencies
        run: npm ci
      
      - name: Run tests
        run: npm test -- --coverage --watchAll=false
      
      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          file: ./coverage/lcov.info

  build:
    needs: test
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Setup Node.js
        uses: actions/setup-node@v3
        with:
          node-version: ${{ env.NODE_VERSION }}
          cache: 'npm'
      
      - name: Install dependencies
        run: npm ci
      
      - name: Build application
        run: npm run build:prod
        env:
          REACT_APP_API_URL: ${{ secrets.REACT_APP_API_URL }}
          REACT_APP_GRAPHQL_URL: ${{ secrets.REACT_APP_GRAPHQL_URL }}
          REACT_APP_SENTRY_DSN: ${{ secrets.REACT_APP_SENTRY_DSN }}
      
      - name: Upload build artifacts
        uses: actions/upload-artifact@v3
        with:
          name: dist
          path: dist/

  deploy:
    needs: build
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/main'
    
    steps:
      - uses: actions/checkout@v3
      
      - name: Download build artifacts
        uses: actions/download-artifact@v3
        with:
          name: dist
          path: dist/
      
      - name: Deploy to S3
        env:
          AWS_ACCESS_KEY_ID: ${{ secrets.AWS_ACCESS_KEY_ID }}
          AWS_SECRET_ACCESS_KEY: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
          AWS_REGION: eu-west-1
        run: |
          aws s3 sync dist/ s3://${{ secrets.S3_BUCKET }} \
            --delete \
            --cache-control max-age=31536000,public \
            --exclude index.html
          
          aws s3 cp dist/index.html s3://${{ secrets.S3_BUCKET }}/index.html \
            --cache-control max-age=0,no-cache,no-store,must-revalidate
      
      - name: Invalidate CloudFront
        env:
          AWS_ACCESS_KEY_ID: ${{ secrets.AWS_ACCESS_KEY_ID }}
          AWS_SECRET_ACCESS_KEY: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
          AWS_REGION: eu-west-1
        run: |
          aws cloudfront create-invalidation \
            --distribution-id ${{ secrets.CLOUDFRONT_DISTRIBUTION_ID }} \
            --paths "/*"
      
      - name: Notify Slack
        uses: 8398a7/action-slack@v3
        with:
          status: ${{ job.status }}
          text: 'Deployment to production completed'
          webhook_url: ${{ secrets.SLACK_WEBHOOK }}
        if: always()
```

### 2. GitLab CI

```yaml
# .gitlab-ci.yml
stages:
  - test
  - build
  - deploy

variables:
  NODE_VERSION: "18"

cache:
  paths:
    - node_modules/

test:
  stage: test
  image: node:${NODE_VERSION}
  script:
    - npm ci
    - npm test -- --coverage --watchAll=false
  coverage: '/Lines\s*:\s*(\d+\.\d+)%/'
  artifacts:
    reports:
      coverage_report:
        coverage_format: cobertura
        path: coverage/cobertura-coverage.xml

build:
  stage: build
  image: node:${NODE_VERSION}
  script:
    - npm ci
    - npm run build:prod
  artifacts:
    paths:
      - dist/
    expire_in: 1 week
  only:
    - main
    - staging

deploy_staging:
  stage: deploy
  image: alpine:latest
  before_script:
    - apk add --no-cache curl
  script:
    - curl -X POST -H "Authorization: Bearer $VERCEL_TOKEN" 
      -H "Content-Type: application/json" 
      -d '{"name":"asam-frontend","gitSource":{"ref":"staging"}}' 
      https://api.vercel.com/v13/deployments
  environment:
    name: staging
    url: https://staging.asam.org
  only:
    - staging

deploy_production:
  stage: deploy
  image: python:3.9
  before_script:
    - pip install awscli
  script:
    - aws s3 sync dist/ s3://$S3_BUCKET --delete
    - aws cloudfront create-invalidation --distribution-id $CF_DISTRIBUTION_ID --paths "/*"
  environment:
    name: production
    url: https://app.asam.org
  only:
    - main
  when: manual
```

## Monitoreo Post-Deployment

### 1. Health Check

```javascript
// public/health.json
{
  "status": "healthy",
  "version": "1.0.0",
  "buildTime": "2024-01-15T10:00:00Z"
}

// src/utils/healthCheck.js
export async function performHealthCheck() {
  const checks = {
    app: false,
    api: false,
    auth: false
  };

  try {
    // Check app
    const appResponse = await fetch('/health.json');
    checks.app = appResponse.ok;

    // Check API
    const apiResponse = await fetch(`${process.env.REACT_APP_API_URL}/health`);
    checks.api = apiResponse.ok;

    // Check auth service
    const authResponse = await fetch(`${process.env.REACT_APP_API_URL}/auth/health`);
    checks.auth = authResponse.ok;

    return {
      healthy: Object.values(checks).every(check => check),
      checks
    };
  } catch (error) {
    return {
      healthy: false,
      checks,
      error: error.message
    };
  }
}
```

### 2. Monitoring Script

```javascript
// scripts/monitor-deployment.js
const axios = require('axios');

const SITES = [
  { name: 'Production', url: 'https://app.asam.org' },
  { name: 'Staging', url: 'https://staging.asam.org' }
];

const CHECKS = [
  { name: 'Homepage', path: '/' },
  { name: 'API Health', path: '/api/health' },
  { name: 'Static Assets', path: '/static/js/main.js' }
];

async function checkSite(site) {
  console.log(`Checking ${site.name}...`);
  
  for (const check of CHECKS) {
    try {
      const response = await axios.get(`${site.url}${check.path}`, {
        timeout: 5000,
        validateStatus: () => true
      });
      
      const status = response.status === 200 ? '✅' : '❌';
      console.log(`  ${status} ${check.name}: ${response.status}`);
      
      if (response.status !== 200) {
        // Send alert
        await sendAlert({
          site: site.name,
          check: check.name,
          status: response.status,
          url: `${site.url}${check.path}`
        });
      }
    } catch (error) {
      console.log(`  ❌ ${check.name}: ${error.message}`);
      await sendAlert({
        site: site.name,
        check: check.name,
        error: error.message
      });
    }
  }
}

async function sendAlert(data) {
  // Implementar notificaciones (Slack, email, etc.)
  console.error('ALERT:', data);
}

async function monitor() {
  for (const site of SITES) {
    await checkSite(site);
  }
}

// Run checks
monitor().catch(console.error);
```

## Rollback Strategy

### 1. Versioned Deployments

```bash
#!/bin/bash
# deploy-with-version.sh

VERSION=$(git rev-parse --short HEAD)
TIMESTAMP=$(date +%Y%m%d%H%M%S)
DEPLOYMENT_ID="${VERSION}-${TIMESTAMP}"

echo "Deploying version: $DEPLOYMENT_ID"

# Build with version
npm run build:prod

# Create versioned directory in S3
aws s3 sync dist/ s3://$BUCKET_NAME/versions/$DEPLOYMENT_ID/

# Update current symlink
echo $DEPLOYMENT_ID > dist/VERSION
aws s3 cp dist/VERSION s3://$BUCKET_NAME/current/VERSION

# Copy files to current
aws s3 sync s3://$BUCKET_NAME/versions/$DEPLOYMENT_ID/ s3://$BUCKET_NAME/current/

# Keep last 5 versions
aws s3 ls s3://$BUCKET_NAME/versions/ | \
  sort -r | \
  awk 'NR>5 {print $4}' | \
  xargs -I {} aws s3 rm --recursive s3://$BUCKET_NAME/versions/{}
```

### 2. Rollback Script

```bash
#!/bin/bash
# rollback.sh

# Get previous version
CURRENT_VERSION=$(aws s3 cp s3://$BUCKET_NAME/current/VERSION -)
PREVIOUS_VERSION=$(aws s3 ls s3://$BUCKET_NAME/versions/ | \
  grep -v $CURRENT_VERSION | \
  sort -r | \
  head -1 | \
  awk '{print $4}' | \
  sed 's/\///')

if [ -z "$PREVIOUS_VERSION" ]; then
  echo "No previous version found!"
  exit 1
fi

echo "Rolling back from $CURRENT_VERSION to $PREVIOUS_VERSION"

# Copy previous version to current
aws s3 sync s3://$BUCKET_NAME/versions/$PREVIOUS_VERSION/ s3://$BUCKET_NAME/current/ --delete

# Update version file
echo $PREVIOUS_VERSION | aws s3 cp - s3://$BUCKET_NAME/current/VERSION

# Invalidate CloudFront
aws cloudfront create-invalidation \
  --distribution-id $DISTRIBUTION_ID \
  --paths "/*"

echo "Rollback complete!"
```

## Checklist de Deployment

### Pre-Deployment
- [ ] Tests pasando al 100%
- [ ] Build de producción sin errores
- [ ] Bundle size dentro de límites
- [ ] Variables de entorno configuradas
- [ ] Backup de versión actual

### Deployment
- [ ] Build optimizado generado
- [ ] Assets subidos con cache headers correctos
- [ ] index.html sin cache
- [ ] CloudFront/CDN invalidado
- [ ] Health checks pasando

### Post-Deployment
- [ ] Verificar funcionalidad crítica
- [ ] Monitorear errores en Sentry
- [ ] Verificar métricas de performance
- [ ] Notificar al equipo
- [ ] Documentar versión desplegada

### Rollback (si necesario)
- [ ] Identificar versión estable anterior
- [ ] Ejecutar rollback script
- [ ] Verificar restauración
- [ ] Investigar causa del fallo
- [ ] Documentar incidente

Esta guía proporciona un proceso completo y robusto para desplegar aplicaciones frontend de ASAM en producción.