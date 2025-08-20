# Guía de Monitoreo y Analytics

Esta guía proporciona estrategias para implementar monitoreo y analytics en aplicaciones frontend que consumen el backend de ASAM.

## Tabla de Contenidos
1. [Visión General](#visión-general)
2. [Monitoreo de Errores](#monitoreo-de-errores)
3. [Analytics de Usuario](#analytics-de-usuario)
4. [Performance Monitoring](#performance-monitoring)
5. [Business Metrics](#business-metrics)
6. [Dashboards y Reportes](#dashboards-y-reportes)
7. [Privacidad y Compliance](#privacidad-y-compliance)

## Visión General

### Stack de Monitoreo Recomendado
- **Errores**: Sentry
- **Analytics**: Google Analytics 4 / Plausible
- **Performance**: Web Vitals + Custom Metrics
- **APM**: DataDog / New Relic
- **Logs**: LogRocket / FullStory

### Métricas Clave
- Errores y crashes
- Performance (Core Web Vitals)
- Engagement de usuarios
- Conversión y flujos
- Métricas de negocio

## Monitoreo de Errores

### 1. Configuración de Sentry

```javascript
// services/sentry.js
import * as Sentry from '@sentry/react';
import { BrowserTracing } from '@sentry/tracing';
import { CaptureConsole } from '@sentry/integrations';

export function initSentry() {
  if (process.env.NODE_ENV === 'production') {
    Sentry.init({
      dsn: process.env.REACT_APP_SENTRY_DSN,
      environment: process.env.REACT_APP_ENVIRONMENT,
      release: process.env.REACT_APP_VERSION,
      
      integrations: [
        new BrowserTracing({
          // Configurar rutas a trackear
          routingInstrumentation: Sentry.reactRouterV6Instrumentation(
            React.useEffect,
            useLocation,
            useNavigationType,
            createRoutesFromChildren,
            matchRoutes
          ),
          
          // Tracking de GraphQL
          tracingOrigins: [
            'localhost',
            process.env.REACT_APP_API_URL,
            /^\//
          ],
          
          // Configurar sampling
          tracesSampleRate: 0.1,
        }),
        
        new CaptureConsole({
          levels: ['error']
        })
      ],
      
      // Sampling
      sampleRate: 0.9,
      tracesSampleRate: 0.1,
      
      // Filtros
      beforeSend(event, hint) {
        // Filtrar errores conocidos o no críticos
        if (event.exception) {
          const error = hint.originalException;
          
          // Ignorar errores de red esperados
          if (error?.message?.includes('Network request failed')) {
            return null;
          }
          
          // Ignorar errores de extensiones del navegador
          if (error?.stack?.includes('chrome-extension://')) {
            return null;
          }
        }
        
        // Añadir contexto adicional
        event.contexts = {
          ...event.contexts,
          app: {
            version: process.env.REACT_APP_VERSION,
            build: process.env.REACT_APP_BUILD_ID
          }
        };
        
        return event;
      },
      
      // Configurar user context
      initialScope: {
        tags: { 
          component: 'frontend',
          version: process.env.REACT_APP_VERSION
        },
        user: {
          id: authService.getCurrentUser()?.id
        }
      }
    });
  }
}

// Error Boundary con Sentry
export const SentryErrorBoundary = Sentry.ErrorBoundary;

// Profiler
export const SentryProfiler = Sentry.Profiler;
```

### 2. Custom Error Tracking

```javascript
// services/errorTracking.js
import * as Sentry from '@sentry/react';

class ErrorTracker {
  /**
   * Track error con contexto adicional
   */
  trackError(error, context = {}) {
    const enrichedContext = {
      ...context,
      timestamp: new Date().toISOString(),
      userAgent: navigator.userAgent,
      url: window.location.href,
      viewport: {
        width: window.innerWidth,
        height: window.innerHeight
      }
    };
    
    if (process.env.NODE_ENV === 'production') {
      Sentry.captureException(error, {
        contexts: {
          custom: enrichedContext
        }
      });
    } else {
      console.error('Error tracked:', error, enrichedContext);
    }
  }
  
  /**
   * Track error de GraphQL
   */
  trackGraphQLError(error, operation) {
    this.trackError(error, {
      type: 'graphql',
      operation: operation.operationName,
      variables: operation.variables,
      query: operation.query.loc?.source.body
    });
  }
  
  /**
   * Track error de validación
   */
  trackValidationError(errors, formName) {
    Sentry.captureMessage('Validation Error', {
      level: 'warning',
      contexts: {
        validation: {
          formName,
          errors,
          timestamp: new Date().toISOString()
        }
      }
    });
  }
  
  /**
   * Track error crítico de negocio
   */
  trackBusinessError(error, context) {
    Sentry.captureException(new Error(error), {
      level: 'error',
      tags: {
        type: 'business_logic',
        ...context
      }
    });
  }
}

export const errorTracker = new ErrorTracker();
```

## Analytics de Usuario

### 1. Google Analytics 4 Setup

```javascript
// services/analytics.js
import ReactGA from 'react-ga4';

class AnalyticsService {
  constructor() {
    this.initialized = false;
  }
  
  init() {
    if (this.initialized) return;
    
    ReactGA.initialize(process.env.REACT_APP_GA_TRACKING_ID, {
      gaOptions: {
        anonymizeIp: true,
        cookieFlags: 'SameSite=None;Secure'
      }
    });
    
    this.initialized = true;
    
    // Set user properties
    const user = authService.getCurrentUser();
    if (user) {
      this.setUser(user.id);
      this.setUserProperties({
        role: user.role,
        membershipType: user.membershipType
      });
    }
  }
  
  // Page views
  trackPageView(path, title) {
    if (!this.initialized) return;
    
    ReactGA.send({
      hitType: 'pageview',
      page: path,
      title
    });
  }
  
  // Events
  trackEvent(category, action, label, value) {
    if (!this.initialized) return;
    
    ReactGA.event({
      category,
      action,
      label,
      value
    });
  }
  
  // User properties
  setUser(userId) {
    if (!this.initialized) return;
    
    ReactGA.set({ userId });
  }
  
  setUserProperties(properties) {
    if (!this.initialized) return;
    
    ReactGA.set(properties);
  }
  
  // E-commerce
  trackPurchase(transactionData) {
    if (!this.initialized) return;
    
    ReactGA.event({
      category: 'ecommerce',
      action: 'purchase',
      value: transactionData.value,
      currency: 'EUR',
      items: transactionData.items
    });
  }
  
  // Custom dimensions
  setCustomDimension(index, value) {
    if (!this.initialized) return;
    
    ReactGA.set({ [`dimension${index}`]: value });
  }
}

export const analytics = new AnalyticsService();
```

### 2. Event Tracking Hooks

```javascript
// hooks/useAnalytics.js
import { useEffect } from 'react';
import { useLocation } from 'react-router-dom';
import { analytics } from '../services/analytics';

export function usePageTracking() {
  const location = useLocation();
  
  useEffect(() => {
    analytics.trackPageView(location.pathname, document.title);
  }, [location]);
}

export function useEventTracking() {
  const trackEvent = (category, action, label, value) => {
    analytics.trackEvent(category, action, label, value);
  };
  
  const trackFormSubmit = (formName, success = true) => {
    trackEvent('Form', success ? 'Submit' : 'Error', formName);
  };
  
  const trackButtonClick = (buttonName, context) => {
    trackEvent('UI', 'Button Click', buttonName, context);
  };
  
  const trackSearch = (searchTerm, resultsCount) => {
    trackEvent('Search', 'Query', searchTerm, resultsCount);
  };
  
  const trackPayment = (amount, method) => {
    analytics.trackPurchase({
      value: amount,
      items: [{
        item_name: 'Membership Fee',
        price: amount,
        quantity: 1,
        currency: 'EUR'
      }]
    });
  };
  
  return {
    trackEvent,
    trackFormSubmit,
    trackButtonClick,
    trackSearch,
    trackPayment
  };
}
```

### 3. Componente Analytics Wrapper

```javascript
// components/AnalyticsProvider.jsx
import { createContext, useContext, useEffect } from 'react';
import { useLocation } from 'react-router-dom';
import { analytics } from '../services/analytics';

const AnalyticsContext = createContext();

export function AnalyticsProvider({ children }) {
  const location = useLocation();
  
  useEffect(() => {
    analytics.init();
  }, []);
  
  useEffect(() => {
    // Track route changes
    analytics.trackPageView(location.pathname, document.title);
    
    // Track time on page
    const startTime = Date.now();
    
    return () => {
      const timeOnPage = Math.round((Date.now() - startTime) / 1000);
      analytics.trackEvent('Engagement', 'Time on Page', location.pathname, timeOnPage);
    };
  }, [location]);
  
  return (
    <AnalyticsContext.Provider value={analytics}>
      {children}
    </AnalyticsContext.Provider>
  );
}

export const useAnalytics = () => useContext(AnalyticsContext);
```

## Performance Monitoring

### 1. Web Vitals Tracking

```javascript
// services/performanceMonitoring.js
import { getCLS, getFID, getFCP, getLCP, getTTFB } from 'web-vitals';

class PerformanceMonitor {
  constructor() {
    this.metrics = {};
    this.observers = new Map();
  }
  
  init() {
    // Core Web Vitals
    this.trackWebVitals();
    
    // Custom metrics
    this.trackCustomMetrics();
    
    // Resource timing
    this.trackResourceTiming();
    
    // Long tasks
    this.trackLongTasks();
  }
  
  trackWebVitals() {
    const sendToAnalytics = (metric) => {
      // Send to Google Analytics
      if (window.gtag) {
        window.gtag('event', metric.name, {
          value: Math.round(metric.value),
          event_category: 'Web Vitals',
          event_label: metric.id,
          non_interaction: true
        });
      }
      
      // Send to custom endpoint
      this.sendMetric({
        type: 'web-vital',
        name: metric.name,
        value: metric.value,
        id: metric.id,
        rating: metric.rating
      });
      
      // Log in development
      if (process.env.NODE_ENV === 'development') {
        console.log(`${metric.name}: ${metric.value} (${metric.rating})`);
      }
    };
    
    getCLS(sendToAnalytics);
    getFID(sendToAnalytics);
    getFCP(sendToAnalytics);
    getLCP(sendToAnalytics);
    getTTFB(sendToAnalytics);
  }
  
  trackCustomMetrics() {
    // Time to Interactive
    if ('PerformanceObserver' in window) {
      const observer = new PerformanceObserver((list) => {
        for (const entry of list.getEntries()) {
          if (entry.name === 'first-contentful-paint') {
            this.metrics.tti = entry.startTime;
            this.sendMetric({
              type: 'custom',
              name: 'TTI',
              value: entry.startTime
            });
          }
        }
      });
      
      observer.observe({ entryTypes: ['paint'] });
      this.observers.set('paint', observer);
    }
    
    // JavaScript bundle load time
    window.addEventListener('load', () => {
      const perfData = performance.getEntriesByType('navigation')[0];
      this.sendMetric({
        type: 'custom',
        name: 'page_load_time',
        value: perfData.loadEventEnd - perfData.fetchStart
      });
    });
  }
  
  trackResourceTiming() {
    if ('PerformanceObserver' in window) {
      const observer = new PerformanceObserver((list) => {
        for (const entry of list.getEntries()) {
          // Track slow resources
          if (entry.duration > 1000) {
            this.sendMetric({
              type: 'slow-resource',
              name: entry.name,
              value: entry.duration,
              size: entry.transferSize
            });
          }
        }
      });
      
      observer.observe({ entryTypes: ['resource'] });
      this.observers.set('resource', observer);
    }
  }
  
  trackLongTasks() {
    if ('PerformanceLongTaskTiming' in window) {
      const observer = new PerformanceObserver((list) => {
        for (const entry of list.getEntries()) {
          this.sendMetric({
            type: 'long-task',
            name: 'long_task',
            value: entry.duration,
            startTime: entry.startTime
          });
        }
      });
      
      observer.observe({ entryTypes: ['longtask'] });
      this.observers.set('longtask', observer);
    }
  }
  
  // Tracking manual para operaciones específicas
  startMeasure(name) {
    performance.mark(`${name}-start`);
  }
  
  endMeasure(name) {
    performance.mark(`${name}-end`);
    performance.measure(name, `${name}-start`, `${name}-end`);
    
    const measure = performance.getEntriesByName(name)[0];
    this.sendMetric({
      type: 'measure',
      name,
      value: measure.duration
    });
    
    // Limpiar marks
    performance.clearMarks(`${name}-start`);
    performance.clearMarks(`${name}-end`);
    performance.clearMeasures(name);
  }
  
  sendMetric(metric) {
    // Batch metrics
    if (!this.metricQueue) {
      this.metricQueue = [];
    }
    
    this.metricQueue.push({
      ...metric,
      timestamp: new Date().toISOString(),
      url: window.location.href,
      userAgent: navigator.userAgent
    });
    
    // Send batch every 5 seconds
    if (!this.batchTimer) {
      this.batchTimer = setTimeout(() => {
        this.flushMetrics();
      }, 5000);
    }
  }
  
  async flushMetrics() {
    if (!this.metricQueue || this.metricQueue.length === 0) return;
    
    const metrics = [...this.metricQueue];
    this.metricQueue = [];
    this.batchTimer = null;
    
    try {
      await fetch('/api/metrics', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({ metrics })
      });
    } catch (error) {
      console.error('Failed to send metrics:', error);
    }
  }
  
  cleanup() {
    this.observers.forEach(observer => observer.disconnect());
    this.observers.clear();
    
    if (this.batchTimer) {
      clearTimeout(this.batchTimer);
      this.flushMetrics();
    }
  }
}

export const performanceMonitor = new PerformanceMonitor();
```

### 2. Component Performance Tracking

```javascript
// hooks/usePerformanceTracking.js
import { useEffect } from 'react';
import { performanceMonitor } from '../services/performanceMonitoring';

export function useComponentPerformance(componentName) {
  useEffect(() => {
    // Track mount time
    performanceMonitor.startMeasure(`${componentName}-mount`);
    
    return () => {
      performanceMonitor.endMeasure(`${componentName}-mount`);
    };
  }, [componentName]);
  
  const trackOperation = (operationName, operation) => {
    performanceMonitor.startMeasure(`${componentName}-${operationName}`);
    
    const result = operation();
    
    if (result instanceof Promise) {
      return result.finally(() => {
        performanceMonitor.endMeasure(`${componentName}-${operationName}`);
      });
    } else {
      performanceMonitor.endMeasure(`${componentName}-${operationName}`);
      return result;
    }
  };
  
  return { trackOperation };
}
```

## Business Metrics

### 1. Métricas de Negocio

```javascript
// services/businessMetrics.js
class BusinessMetrics {
  constructor() {
    this.queue = [];
  }
  
  // Métricas de miembros
  trackMemberActivity(action, data) {
    this.track('member_activity', {
      action,
      memberId: data.memberId,
      memberType: data.memberType,
      ...data
    });
  }
  
  // Métricas de pagos
  trackPayment(data) {
    this.track('payment', {
      amount: data.amount,
      method: data.method,
      type: data.type,
      memberId: data.memberId,
      success: data.success
    });
    
    // También enviar a analytics
    if (data.success) {
      analytics.trackPurchase({
        value: data.amount,
        items: [{
          item_name: data.type,
          price: data.amount,
          quantity: 1
        }]
      });
    }
  }
  
  // Métricas de engagement
  trackEngagement(action, data) {
    this.track('engagement', {
      action,
      duration: data.duration,
      page: data.page,
      ...data
    });
  }
  
  // Métricas de conversión
  trackConversion(funnel, step, data) {
    this.track('conversion', {
      funnel,
      step,
      completed: data.completed,
      abandonReason: data.abandonReason,
      ...data
    });
  }
  
  // Tracking genérico
  track(metric, data) {
    const event = {
      metric,
      data,
      timestamp: new Date().toISOString(),
      sessionId: this.getSessionId(),
      userId: authService.getCurrentUser()?.id
    };
    
    this.queue.push(event);
    
    // Send immediately for important metrics
    if (this.isImportantMetric(metric)) {
      this.flush();
    }
  }
  
  isImportantMetric(metric) {
    return ['payment', 'conversion'].includes(metric);
  }
  
  async flush() {
    if (this.queue.length === 0) return;
    
    const events = [...this.queue];
    this.queue = [];
    
    try {
      await fetch('/api/metrics/business', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${tokenService.getAccessToken()}`
        },
        body: JSON.stringify({ events })
      });
    } catch (error) {
      // Re-queue on failure
      this.queue.unshift(...events);
    }
  }
  
  getSessionId() {
    let sessionId = sessionStorage.getItem('metrics_session_id');
    if (!sessionId) {
      sessionId = `session_${Date.now()}_${Math.random().toString(36)}`;
      sessionStorage.setItem('metrics_session_id', sessionId);
    }
    return sessionId;
  }
}

export const businessMetrics = new BusinessMetrics();
```

### 2. Funnel Tracking

```javascript
// hooks/useFunnelTracking.js
import { useState, useEffect } from 'react';
import { businessMetrics } from '../services/businessMetrics';

export function useFunnelTracking(funnelName) {
  const [currentStep, setCurrentStep] = useState(0);
  const [startTime, setStartTime] = useState(Date.now());
  const [stepTimes, setStepTimes] = useState({});
  
  const trackStep = (stepName, data = {}) => {
    const now = Date.now();
    const stepDuration = now - (stepTimes[currentStep] || startTime);
    
    businessMetrics.trackConversion(funnelName, stepName, {
      ...data,
      stepNumber: currentStep + 1,
      duration: stepDuration,
      completed: true
    });
    
    setCurrentStep(currentStep + 1);
    setStepTimes({ ...stepTimes, [currentStep + 1]: now });
  };
  
  const abandonFunnel = (reason) => {
    const totalDuration = Date.now() - startTime;
    
    businessMetrics.trackConversion(funnelName, 'abandoned', {
      stepNumber: currentStep,
      duration: totalDuration,
      completed: false,
      abandonReason: reason
    });
  };
  
  useEffect(() => {
    return () => {
      // Track abandonment if component unmounts without completion
      if (currentStep < 5) { // Assuming 5 steps in funnel
        abandonFunnel('navigation');
      }
    };
  }, [currentStep]);
  
  return { trackStep, abandonFunnel, currentStep };
}
```

## Dashboards y Reportes

### 1. Dashboard de Métricas

```javascript
// components/MetricsDashboard.jsx
import { useState, useEffect } from 'react';
import { Line, Bar, Doughnut } from 'react-chartjs-2';
import { useQuery } from '@apollo/client';

const GET_METRICS = gql`
  query GetMetrics($filter: MetricsFilter!) {
    getMetrics(filter: $filter) {
      payments {
        date
        total
        count
        averageAmount
      }
      members {
        active
        inactive
        new
        churnRate
      }
      engagement {
        dailyActiveUsers
        averageSessionDuration
        pageViews
        bounceRate
      }
    }
  }
`;

export function MetricsDashboard() {
  const [dateRange, setDateRange] = useState('last30days');
  const { data, loading, error } = useQuery(GET_METRICS, {
    variables: {
      filter: { dateRange }
    }
  });
  
  if (loading) return <LoadingSpinner />;
  if (error) return <ErrorMessage error={error} />;
  
  const { payments, members, engagement } = data.getMetrics;
  
  // Chart configurations
  const paymentChartData = {
    labels: payments.map(p => formatDate(p.date)),
    datasets: [{
      label: 'Ingresos Totales',
      data: payments.map(p => p.total),
      borderColor: 'rgb(75, 192, 192)',
      tension: 0.1
    }]
  };
  
  const memberChartData = {
    labels: ['Activos', 'Inactivos', 'Nuevos'],
    datasets: [{
      data: [members.active, members.inactive, members.new],
      backgroundColor: ['#36A2EB', '#FF6384', '#FFCE56']
    }]
  };
  
  return (
    <div className="metrics-dashboard">
      <h1>Dashboard de Métricas</h1>
      
      <div className="date-range-selector">
        <select value={dateRange} onChange={(e) => setDateRange(e.target.value)}>
          <option value="last7days">Últimos 7 días</option>
          <option value="last30days">Últimos 30 días</option>
          <option value="last90days">Últimos 90 días</option>
          <option value="lastyear">Último año</option>
        </select>
      </div>
      
      <div className="metrics-grid">
        <div className="metric-card">
          <h3>KPIs Principales</h3>
          <div className="kpi-grid">
            <div className="kpi">
              <span className="kpi-value">{members.active}</span>
              <span className="kpi-label">Miembros Activos</span>
              <span className="kpi-change">+5.2%</span>
            </div>
            <div className="kpi">
              <span className="kpi-value">€{calculateTotal(payments)}</span>
              <span className="kpi-label">Ingresos Totales</span>
              <span className="kpi-change">+12.3%</span>
            </div>
            <div className="kpi">
              <span className="kpi-value">{members.churnRate}%</span>
              <span className="kpi-label">Tasa de Abandono</span>
              <span className="kpi-change">-2.1%</span>
            </div>
            <div className="kpi">
              <span className="kpi-value">{engagement.dailyActiveUsers}</span>
              <span className="kpi-label">Usuarios Activos Diarios</span>
              <span className="kpi-change">+8.7%</span>
            </div>
          </div>
        </div>
        
        <div className="metric-card">
          <h3>Evolución de Pagos</h3>
          <Line data={paymentChartData} options={chartOptions} />
        </div>
        
        <div className="metric-card">
          <h3>Distribución de Miembros</h3>
          <Doughnut data={memberChartData} />
        </div>
        
        <div className="metric-card">
          <h3>Métricas de Engagement</h3>
          <div className="engagement-metrics">
            <div className="metric">
              <label>Duración Media de Sesión</label>
              <span>{formatDuration(engagement.averageSessionDuration)}</span>
            </div>
            <div className="metric">
              <label>Páginas por Sesión</label>
              <span>{engagement.pageViews}</span>
            </div>
            <div className="metric">
              <label>Tasa de Rebote</label>
              <span>{engagement.bounceRate}%</span>
            </div>
          </div>
        </div>
      </div>
      
      <div className="export-actions">
        <button onClick={exportToCSV}>Exportar CSV</button>
        <button onClick={exportToPDF}>Exportar PDF</button>
        <button onClick={scheduleReport}>Programar Reporte</button>
      </div>
    </div>
  );
}
```

### 2. Real-time Dashboard

```javascript
// components/RealTimeDashboard.jsx
import { useState, useEffect } from 'react';
import { useSubscription } from '@apollo/client';

const METRICS_SUBSCRIPTION = gql`
  subscription OnMetricsUpdate {
    metricsUpdate {
      activeUsers
      recentPayments {
        id
        amount
        memberName
        timestamp
      }
      currentActivity {
        page
        users
      }
    }
  }
`;

export function RealTimeDashboard() {
  const [metrics, setMetrics] = useState({
    activeUsers: 0,
    recentPayments: [],
    currentActivity: []
  });
  
  const { data } = useSubscription(METRICS_SUBSCRIPTION);
  
  useEffect(() => {
    if (data?.metricsUpdate) {
      setMetrics(data.metricsUpdate);
    }
  }, [data]);
  
  return (
    <div className="realtime-dashboard">
      <h2>Actividad en Tiempo Real</h2>
      
      <div className="realtime-grid">
        <div className="active-users">
          <h3>Usuarios Activos</h3>
          <div className="big-number">{metrics.activeUsers}</div>
          <div className="sparkline">
            {/* Mini gráfico de actividad */}
          </div>
        </div>
        
        <div className="recent-payments">
          <h3>Pagos Recientes</h3>
          <div className="payment-feed">
            {metrics.recentPayments.map(payment => (
              <div key={payment.id} className="payment-item animate-in">
                <span className="amount">€{payment.amount}</span>
                <span className="member">{payment.memberName}</span>
                <span className="time">{formatTime(payment.timestamp)}</span>
              </div>
            ))}
          </div>
        </div>
        
        <div className="activity-map">
          <h3>Actividad por Página</h3>
          <div className="activity-list">
            {metrics.currentActivity.map((activity, idx) => (
              <div key={idx} className="activity-item">
                <span className="page">{activity.page}</span>
                <div className="activity-bar">
                  <div 
                    className="activity-fill"
                    style={{ width: `${(activity.users / metrics.activeUsers) * 100}%` }}
                  />
                </div>
                <span className="users">{activity.users}</span>
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
}
```

## Privacidad y Compliance

### 1. Consent Management

```javascript
// components/CookieConsent.jsx
import { useState, useEffect } from 'react';
import { analytics } from '../services/analytics';

export function CookieConsent() {
  const [showBanner, setShowBanner] = useState(false);
  const [preferences, setPreferences] = useState({
    necessary: true,
    analytics: false,
    marketing: false
  });
  
  useEffect(() => {
    const consent = localStorage.getItem('cookie_consent');
    if (!consent) {
      setShowBanner(true);
    } else {
      const savedPreferences = JSON.parse(consent);
      setPreferences(savedPreferences);
      applyPreferences(savedPreferences);
    }
  }, []);
  
  const applyPreferences = (prefs) => {
    if (prefs.analytics) {
      analytics.init();
      performanceMonitor.init();
    }
    
    if (prefs.marketing) {
      // Initialize marketing tools
    }
    
    // Update GTM consent
    if (window.gtag) {
      window.gtag('consent', 'update', {
        'analytics_storage': prefs.analytics ? 'granted' : 'denied',
        'ad_storage': prefs.marketing ? 'granted' : 'denied'
      });
    }
  };
  
  const handleAcceptAll = () => {
    const newPreferences = {
      necessary: true,
      analytics: true,
      marketing: true
    };
    
    savePreferences(newPreferences);
  };
  
  const handleAcceptSelected = () => {
    savePreferences(preferences);
  };
  
  const handleRejectAll = () => {
    const newPreferences = {
      necessary: true,
      analytics: false,
      marketing: false
    };
    
    savePreferences(newPreferences);
  };
  
  const savePreferences = (prefs) => {
    localStorage.setItem('cookie_consent', JSON.stringify(prefs));
    localStorage.setItem('cookie_consent_date', new Date().toISOString());
    applyPreferences(prefs);
    setShowBanner(false);
  };
  
  if (!showBanner) return null;
  
  return (
    <div className="cookie-consent-banner">
      <div className="consent-content">
        <h3>Configuración de Cookies</h3>
        <p>
          Utilizamos cookies para mejorar tu experiencia. 
          Puedes elegir qué tipos de cookies aceptar.
        </p>
        
        <div className="cookie-categories">
          <label className="cookie-category">
            <input
              type="checkbox"
              checked={preferences.necessary}
              disabled
            />
            <span>
              <strong>Necesarias</strong>
              <p>Esenciales para el funcionamiento del sitio</p>
            </span>
          </label>
          
          <label className="cookie-category">
            <input
              type="checkbox"
              checked={preferences.analytics}
              onChange={(e) => setPreferences({
                ...preferences,
                analytics: e.target.checked
              })}
            />
            <span>
              <strong>Analíticas</strong>
              <p>Nos ayudan a mejorar el sitio</p>
            </span>
          </label>
          
          <label className="cookie-category">
            <input
              type="checkbox"
              checked={preferences.marketing}
              onChange={(e) => setPreferences({
                ...preferences,
                marketing: e.target.checked
              })}
            />
            <span>
              <strong>Marketing</strong>
              <p>Para mostrarte contenido relevante</p>
            </span>
          </label>
        </div>
        
        <div className="consent-actions">
          <button onClick={handleRejectAll} className="reject-button">
            Rechazar todas
          </button>
          <button onClick={handleAcceptSelected} className="accept-selected-button">
            Aceptar seleccionadas
          </button>
          <button onClick={handleAcceptAll} className="accept-all-button">
            Aceptar todas
          </button>
        </div>
      </div>
    </div>
  );
}
```

### 2. Data Privacy Utils

```javascript
// utils/privacy.js
export const privacy = {
  // Anonimizar datos sensibles
  anonymize: {
    email: (email) => {
      const [user, domain] = email.split('@');
      const anonymizedUser = user.substring(0, 2) + '***';
      return `${anonymizedUser}@${domain}`;
    },
    
    phone: (phone) => {
      return phone.replace(/\d(?=\d{4})/g, '*');
    },
    
    ip: (ip) => {
      const parts = ip.split('.');
      return `${parts[0]}.${parts[1]}.xxx.xxx`;
    },
    
    userId: (id) => {
      // Hash para analytics pero mantener consistencia
      return btoa(id).substring(0, 16);
    }
  },
  
  // Verificar consentimiento
  hasConsent: (type) => {
    const consent = localStorage.getItem('cookie_consent');
    if (!consent) return false;
    
    const preferences = JSON.parse(consent);
    return preferences[type] || false;
  },
  
  // Limpiar datos al retirar consentimiento
  clearAnalyticsData: () => {
    // Clear cookies
    document.cookie.split(";").forEach((c) => {
      document.cookie = c
        .replace(/^ +/, "")
        .replace(/=.*/, "=;expires=" + new Date().toUTCString() + ";path=/");
    });
    
    // Clear localStorage analytics
    const keysToRemove = [];
    for (let i = 0; i < localStorage.length; i++) {
      const key = localStorage.key(i);
      if (key.includes('ga_') || key.includes('_ga')) {
        keysToRemove.push(key);
      }
    }
    keysToRemove.forEach(key => localStorage.removeItem(key));
  }
};
```

## Checklist de Implementación

### Setup Inicial
- [ ] Configurar Sentry con DSN
- [ ] Implementar Google Analytics 4
- [ ] Configurar performance monitoring
- [ ] Implementar consent management
- [ ] Setup de dashboards

### Tracking de Eventos
- [ ] Page views automáticos
- [ ] Eventos de UI principales
- [ ] Conversiones y funnels
- [ ] Errores y excepciones
- [ ] Performance metrics

### Business Metrics
- [ ] Métricas de miembros
- [ ] Tracking de pagos
- [ ] Engagement metrics
- [ ] Conversion funnels
- [ ] Custom KPIs

### Compliance
- [ ] Cookie consent banner
- [ ] Política de privacidad
- [ ] Anonimización de datos
- [ ] Derecho al olvido
- [ ] Audit trail

### Monitoreo
- [ ] Alertas configuradas
- [ ] Dashboards en tiempo real
- [ ] Reportes automatizados
- [ ] Health checks
- [ ] SLAs definidos

Esta guía proporciona un framework completo para implementar monitoreo y analytics en aplicaciones frontend de ASAM, asegurando visibilidad completa del comportamiento y rendimiento de la aplicación.