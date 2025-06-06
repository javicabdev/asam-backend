# Guía del Sistema de Notificaciones

Esta guía documenta cómo integrar y utilizar el sistema de notificaciones del backend de ASAM en aplicaciones frontend.

## Tabla de Contenidos
1. [Visión General](#visión-general)
2. [Tipos de Notificaciones](#tipos-de-notificaciones)
3. [Integración Frontend](#integración-frontend)
4. [Notificaciones en Tiempo Real](#notificaciones-en-tiempo-real)
5. [Gestión de Preferencias](#gestión-de-preferencias)
6. [Mejores Prácticas](#mejores-prácticas)

## Visión General

El sistema de notificaciones de ASAM permite enviar comunicaciones a los usuarios a través de diferentes canales:
- Email
- SMS (futuro)
- Notificaciones in-app
- Push notifications (futuro)

### Casos de Uso Principales
- Recordatorios de pago de cuotas
- Confirmaciones de pagos recibidos
- Avisos de cambios en el estado de membresía
- Notificaciones de eventos y actividades
- Comunicados importantes de la asociación

## Tipos de Notificaciones

### 1. Notificaciones Transaccionales

```javascript
// Tipos de notificaciones transaccionales
const TRANSACTION_NOTIFICATIONS = {
  // Pagos
  PAYMENT_RECEIVED: 'payment_received',
  PAYMENT_REMINDER: 'payment_reminder',
  PAYMENT_OVERDUE: 'payment_overdue',
  
  // Membresía
  MEMBERSHIP_ACTIVATED: 'membership_activated',
  MEMBERSHIP_SUSPENDED: 'membership_suspended',
  MEMBERSHIP_RENEWED: 'membership_renewed',
  
  // Cuenta
  WELCOME_EMAIL: 'welcome_email',
  PASSWORD_RESET: 'password_reset',
  EMAIL_VERIFICATION: 'email_verification'
};
```

### 2. Notificaciones Masivas

```javascript
// Tipos de notificaciones masivas
const BROADCAST_NOTIFICATIONS = {
  ANNOUNCEMENT: 'announcement',
  EVENT_INVITATION: 'event_invitation',
  NEWSLETTER: 'newsletter',
  URGENT_NOTICE: 'urgent_notice'
};
```

## Integración Frontend

### 1. Hook para Notificaciones In-App

```javascript
// hooks/useNotifications.js
import { useState, useEffect, useCallback } from 'react';
import { useQuery, useMutation } from '@apollo/client';
import { toast } from 'react-toastify';

const GET_NOTIFICATIONS = gql`
  query GetNotifications($filter: NotificationFilter) {
    getNotifications(filter: $filter) {
      nodes {
        id
        type
        title
        message
        read
        createdAt
        priority
        data
      }
      pageInfo {
        hasNextPage
        totalCount
        unreadCount
      }
    }
  }
`;

const MARK_AS_READ = gql`
  mutation MarkNotificationAsRead($id: ID!) {
    markNotificationAsRead(id: $id) {
      id
      read
    }
  }
`;

const MARK_ALL_AS_READ = gql`
  mutation MarkAllNotificationsAsRead {
    markAllNotificationsAsRead {
      success
      message
    }
  }
`;

export function useNotifications() {
  const [notifications, setNotifications] = useState([]);
  const [unreadCount, setUnreadCount] = useState(0);
  
  // Query notifications
  const { data, loading, error, refetch } = useQuery(GET_NOTIFICATIONS, {
    variables: {
      filter: {
        read: false,
        pagination: { page: 1, pageSize: 20 }
      }
    },
    pollInterval: 30000 // Poll every 30 seconds
  });
  
  // Mutations
  const [markAsRead] = useMutation(MARK_AS_READ);
  const [markAllAsRead] = useMutation(MARK_ALL_AS_READ);
  
  // Update state when data changes
  useEffect(() => {
    if (data?.getNotifications) {
      setNotifications(data.getNotifications.nodes);
      setUnreadCount(data.getNotifications.pageInfo.unreadCount);
    }
  }, [data]);
  
  // Mark single notification as read
  const handleMarkAsRead = useCallback(async (notificationId) => {
    try {
      await markAsRead({
        variables: { id: notificationId },
        optimisticResponse: {
          markNotificationAsRead: {
            id: notificationId,
            read: true,
            __typename: 'Notification'
          }
        },
        update: (cache) => {
          // Update unread count
          cache.modify({
            fields: {
              getNotifications(existingData) {
                return {
                  ...existingData,
                  pageInfo: {
                    ...existingData.pageInfo,
                    unreadCount: Math.max(0, existingData.pageInfo.unreadCount - 1)
                  }
                };
              }
            }
          });
        }
      });
    } catch (error) {
      console.error('Error marking notification as read:', error);
    }
  }, [markAsRead]);
  
  // Mark all notifications as read
  const handleMarkAllAsRead = useCallback(async () => {
    try {
      await markAllAsRead({
        update: (cache) => {
          // Update all notifications in cache
          cache.modify({
            fields: {
              getNotifications(existingData) {
                return {
                  ...existingData,
                  nodes: existingData.nodes.map(node => ({
                    ...node,
                    read: true
                  })),
                  pageInfo: {
                    ...existingData.pageInfo,
                    unreadCount: 0
                  }
                };
              }
            }
          });
        }
      });
      
      toast.success('Todas las notificaciones marcadas como leídas');
    } catch (error) {
      toast.error('Error al marcar las notificaciones');
    }
  }, [markAllAsRead]);
  
  // Show toast for high priority notifications
  useEffect(() => {
    notifications
      .filter(n => !n.read && n.priority === 'HIGH')
      .forEach(notification => {
        toast.info(notification.message, {
          autoClose: 5000,
          onClick: () => handleMarkAsRead(notification.id)
        });
      });
  }, [notifications, handleMarkAsRead]);
  
  return {
    notifications,
    unreadCount,
    loading,
    error,
    markAsRead: handleMarkAsRead,
    markAllAsRead: handleMarkAllAsRead,
    refetch
  };
}
```

### 2. Componente de Notificaciones

```javascript
// components/NotificationCenter.jsx
import { useState } from 'react';
import { Bell, Check, CheckCheck, AlertCircle } from 'lucide-react';
import { useNotifications } from '../hooks/useNotifications';
import { formatDistanceToNow } from 'date-fns';
import { es } from 'date-fns/locale';

export function NotificationCenter() {
  const [isOpen, setIsOpen] = useState(false);
  const { 
    notifications, 
    unreadCount, 
    loading, 
    markAsRead, 
    markAllAsRead 
  } = useNotifications();
  
  const handleNotificationClick = (notification) => {
    if (!notification.read) {
      markAsRead(notification.id);
    }
    
    // Handle navigation based on notification type
    switch (notification.type) {
      case 'payment_received':
        navigate(`/payments/${notification.data.paymentId}`);
        break;
      case 'membership_suspended':
        navigate('/membership');
        break;
      default:
        // Just mark as read
    }
    
    setIsOpen(false);
  };
  
  const getPriorityIcon = (priority) => {
    switch (priority) {
      case 'HIGH':
        return <AlertCircle className="text-red-500" size={16} />;
      case 'MEDIUM':
        return <AlertCircle className="text-yellow-500" size={16} />;
      default:
        return null;
    }
  };
  
  return (
    <div className="notification-center">
      <button 
        className="notification-trigger"
        onClick={() => setIsOpen(!isOpen)}
      >
        <Bell size={24} />
        {unreadCount > 0 && (
          <span className="notification-badge">{unreadCount}</span>
        )}
      </button>
      
      {isOpen && (
        <div className="notification-dropdown">
          <div className="notification-header">
            <h3>Notificaciones</h3>
            {unreadCount > 0 && (
              <button 
                onClick={markAllAsRead}
                className="mark-all-read"
              >
                <CheckCheck size={16} />
                Marcar todas como leídas
              </button>
            )}
          </div>
          
          <div className="notification-list">
            {loading && <div className="loading">Cargando...</div>}
            
            {!loading && notifications.length === 0 && (
              <div className="empty-state">
                No tienes notificaciones nuevas
              </div>
            )}
            
            {notifications.map(notification => (
              <div
                key={notification.id}
                className={`notification-item ${notification.read ? 'read' : 'unread'}`}
                onClick={() => handleNotificationClick(notification)}
              >
                <div className="notification-content">
                  <div className="notification-title">
                    {getPriorityIcon(notification.priority)}
                    <span>{notification.title}</span>
                    {!notification.read && <span className="unread-dot" />}
                  </div>
                  <p className="notification-message">{notification.message}</p>
                  <span className="notification-time">
                    {formatDistanceToNow(new Date(notification.createdAt), {
                      addSuffix: true,
                      locale: es
                    })}
                  </span>
                </div>
              </div>
            ))}
          </div>
          
          <div className="notification-footer">
            <Link to="/notifications" onClick={() => setIsOpen(false)}>
              Ver todas las notificaciones
            </Link>
          </div>
        </div>
      )}
    </div>
  );
}
```

### 3. Sistema de Toast Notifications

```javascript
// components/ToastContainer.jsx
import { ToastContainer as ReactToastContainer } from 'react-toastify';
import 'react-toastify/dist/ReactToastify.css';

export function ToastContainer() {
  return (
    <ReactToastContainer
      position="top-right"
      autoClose={5000}
      hideProgressBar={false}
      newestOnTop
      closeOnClick
      rtl={false}
      pauseOnFocusLoss
      draggable
      pauseOnHover
      theme="light"
      toastClassName="custom-toast"
      bodyClassName="custom-toast-body"
    />
  );
}

// utils/notifications.js
import { toast } from 'react-toastify';

export const notify = {
  success: (message, options = {}) => {
    toast.success(message, {
      icon: '✅',
      ...options
    });
  },
  
  error: (message, options = {}) => {
    toast.error(message, {
      icon: '❌',
      ...options
    });
  },
  
  info: (message, options = {}) => {
    toast.info(message, {
      icon: 'ℹ️',
      ...options
    });
  },
  
  warning: (message, options = {}) => {
    toast.warning(message, {
      icon: '⚠️',
      ...options
    });
  },
  
  payment: (amount, memberName) => {
    toast.success(
      `Pago de €${amount} recibido de ${memberName}`,
      {
        icon: '💰',
        autoClose: 7000
      }
    );
  },
  
  reminder: (message, actionLabel, onAction) => {
    toast.info(
      <div>
        <p>{message}</p>
        <button 
          onClick={onAction}
          className="toast-action-button"
        >
          {actionLabel}
        </button>
      </div>,
      {
        autoClose: false,
        closeButton: true
      }
    );
  }
};
```

## Notificaciones en Tiempo Real

### 1. WebSocket Integration (Futuro)

```javascript
// services/realtimeService.js
import { io } from 'socket.io-client';

class RealtimeService {
  constructor() {
    this.socket = null;
    this.listeners = new Map();
  }
  
  connect(token) {
    this.socket = io(process.env.REACT_APP_WS_URL, {
      auth: { token },
      reconnection: true,
      reconnectionDelay: 1000,
      reconnectionAttempts: 5
    });
    
    this.socket.on('connect', () => {
      console.log('Connected to realtime service');
    });
    
    this.socket.on('notification', (data) => {
      this.emit('notification', data);
    });
    
    this.socket.on('disconnect', () => {
      console.log('Disconnected from realtime service');
    });
  }
  
  disconnect() {
    if (this.socket) {
      this.socket.disconnect();
      this.socket = null;
    }
  }
  
  on(event, callback) {
    if (!this.listeners.has(event)) {
      this.listeners.set(event, new Set());
    }
    this.listeners.get(event).add(callback);
  }
  
  off(event, callback) {
    if (this.listeners.has(event)) {
      this.listeners.get(event).delete(callback);
    }
  }
  
  emit(event, data) {
    if (this.listeners.has(event)) {
      this.listeners.get(event).forEach(callback => {
        callback(data);
      });
    }
  }
}

export const realtimeService = new RealtimeService();
```

### 2. Hook para Notificaciones en Tiempo Real

```javascript
// hooks/useRealtimeNotifications.js
import { useEffect } from 'react';
import { useApolloClient } from '@apollo/client';
import { realtimeService } from '../services/realtimeService';
import { notify } from '../utils/notifications';

export function useRealtimeNotifications() {
  const client = useApolloClient();
  
  useEffect(() => {
    const handleNotification = (notification) => {
      // Mostrar toast
      notify.info(notification.message);
      
      // Actualizar cache de Apollo
      client.cache.modify({
        fields: {
          getNotifications(existingData = { nodes: [] }) {
            return {
              ...existingData,
              nodes: [notification, ...existingData.nodes],
              pageInfo: {
                ...existingData.pageInfo,
                unreadCount: (existingData.pageInfo?.unreadCount || 0) + 1
              }
            };
          }
        }
      });
      
      // Reproducir sonido para notificaciones importantes
      if (notification.priority === 'HIGH') {
        playNotificationSound();
      }
    };
    
    realtimeService.on('notification', handleNotification);
    
    return () => {
      realtimeService.off('notification', handleNotification);
    };
  }, [client]);
}

function playNotificationSound() {
  const audio = new Audio('/sounds/notification.mp3');
  audio.volume = 0.5;
  audio.play().catch(e => console.log('Audio play failed:', e));
}
```

## Gestión de Preferencias

### 1. Componente de Preferencias

```javascript
// components/NotificationPreferences.jsx
import { useState, useEffect } from 'react';
import { useMutation, useQuery } from '@apollo/client';

const GET_PREFERENCES = gql`
  query GetNotificationPreferences {
    getNotificationPreferences {
      email {
        payments
        reminders
        announcements
        newsletters
      }
      inApp {
        payments
        reminders
        announcements
      }
      frequency {
        reminders
        newsletters
      }
    }
  }
`;

const UPDATE_PREFERENCES = gql`
  mutation UpdateNotificationPreferences($input: NotificationPreferencesInput!) {
    updateNotificationPreferences(input: $input) {
      email {
        payments
        reminders
        announcements
        newsletters
      }
      inApp {
        payments
        reminders
        announcements
      }
      frequency {
        reminders
        newsletters
      }
    }
  }
`;

export function NotificationPreferences() {
  const { data, loading } = useQuery(GET_PREFERENCES);
  const [updatePreferences] = useMutation(UPDATE_PREFERENCES);
  
  const [preferences, setPreferences] = useState({
    email: {
      payments: true,
      reminders: true,
      announcements: true,
      newsletters: true
    },
    inApp: {
      payments: true,
      reminders: true,
      announcements: true
    },
    frequency: {
      reminders: 'WEEKLY',
      newsletters: 'MONTHLY'
    }
  });
  
  useEffect(() => {
    if (data?.getNotificationPreferences) {
      setPreferences(data.getNotificationPreferences);
    }
  }, [data]);
  
  const handleToggle = (channel, type) => {
    setPreferences(prev => ({
      ...prev,
      [channel]: {
        ...prev[channel],
        [type]: !prev[channel][type]
      }
    }));
  };
  
  const handleFrequencyChange = (type, value) => {
    setPreferences(prev => ({
      ...prev,
      frequency: {
        ...prev.frequency,
        [type]: value
      }
    }));
  };
  
  const handleSave = async () => {
    try {
      await updatePreferences({
        variables: { input: preferences }
      });
      notify.success('Preferencias actualizadas correctamente');
    } catch (error) {
      notify.error('Error al actualizar preferencias');
    }
  };
  
  if (loading) return <div>Cargando...</div>;
  
  return (
    <div className="notification-preferences">
      <h2>Preferencias de Notificaciones</h2>
      
      <section className="preference-section">
        <h3>Notificaciones por Email</h3>
        
        <label className="preference-item">
          <input
            type="checkbox"
            checked={preferences.email.payments}
            onChange={() => handleToggle('email', 'payments')}
          />
          <div>
            <span className="preference-title">Pagos</span>
            <span className="preference-description">
              Recibe confirmaciones de pagos realizados
            </span>
          </div>
        </label>
        
        <label className="preference-item">
          <input
            type="checkbox"
            checked={preferences.email.reminders}
            onChange={() => handleToggle('email', 'reminders')}
          />
          <div>
            <span className="preference-title">Recordatorios</span>
            <span className="preference-description">
              Recordatorios de pagos pendientes
            </span>
          </div>
        </label>
        
        <label className="preference-item">
          <input
            type="checkbox"
            checked={preferences.email.announcements}
            onChange={() => handleToggle('email', 'announcements')}
          />
          <div>
            <span className="preference-title">Anuncios</span>
            <span className="preference-description">
              Comunicados importantes de la asociación
            </span>
          </div>
        </label>
        
        <label className="preference-item">
          <input
            type="checkbox"
            checked={preferences.email.newsletters}
            onChange={() => handleToggle('email', 'newsletters')}
          />
          <div>
            <span className="preference-title">Boletines</span>
            <span className="preference-description">
              Boletín informativo de la asociación
            </span>
          </div>
        </label>
      </section>
      
      <section className="preference-section">
        <h3>Notificaciones en la Aplicación</h3>
        
        <label className="preference-item">
          <input
            type="checkbox"
            checked={preferences.inApp.payments}
            onChange={() => handleToggle('inApp', 'payments')}
          />
          <div>
            <span className="preference-title">Pagos</span>
          </div>
        </label>
        
        <label className="preference-item">
          <input
            type="checkbox"
            checked={preferences.inApp.reminders}
            onChange={() => handleToggle('inApp', 'reminders')}
          />
          <div>
            <span className="preference-title">Recordatorios</span>
          </div>
        </label>
        
        <label className="preference-item">
          <input
            type="checkbox"
            checked={preferences.inApp.announcements}
            onChange={() => handleToggle('inApp', 'announcements')}
          />
          <div>
            <span className="preference-title">Anuncios</span>
          </div>
        </label>
      </section>
      
      <section className="preference-section">
        <h3>Frecuencia</h3>
        
        <div className="preference-item">
          <label>
            <span className="preference-title">Recordatorios de pago</span>
            <select
              value={preferences.frequency.reminders}
              onChange={(e) => handleFrequencyChange('reminders', e.target.value)}
            >
              <option value="DAILY">Diario</option>
              <option value="WEEKLY">Semanal</option>
              <option value="MONTHLY">Mensual</option>
              <option value="NEVER">Nunca</option>
            </select>
          </label>
        </div>
        
        <div className="preference-item">
          <label>
            <span className="preference-title">Boletines</span>
            <select
              value={preferences.frequency.newsletters}
              onChange={(e) => handleFrequencyChange('newsletters', e.target.value)}
            >
              <option value="WEEKLY">Semanal</option>
              <option value="MONTHLY">Mensual</option>
              <option value="QUARTERLY">Trimestral</option>
              <option value="NEVER">Nunca</option>
            </select>
          </label>
        </div>
      </section>
      
      <button onClick={handleSave} className="save-button">
        Guardar Preferencias
      </button>
    </div>
  );
}
```

## Mejores Prácticas

### 1. Frecuencia y Timing

```javascript
// utils/notificationRules.js
export const notificationRules = {
  // Evitar notificaciones en horarios no apropiados
  isAppropriateTime: () => {
    const hour = new Date().getHours();
    return hour >= 8 && hour <= 22; // 8 AM - 10 PM
  },
  
  // Rate limiting por tipo
  rateLimit: {
    payment_reminder: {
      max: 1,
      window: 24 * 60 * 60 * 1000 // 1 por día
    },
    announcement: {
      max: 3,
      window: 24 * 60 * 60 * 1000 // 3 por día
    }
  },
  
  // Prioridad por tipo
  priority: {
    payment_overdue: 'HIGH',
    payment_received: 'MEDIUM',
    announcement: 'LOW',
    newsletter: 'LOW'
  }
};
```

### 2. Agrupación de Notificaciones

```javascript
// utils/notificationGrouping.js
export function groupNotifications(notifications) {
  const groups = {
    today: [],
    yesterday: [],
    thisWeek: [],
    older: []
  };
  
  const now = new Date();
  const today = new Date(now.getFullYear(), now.getMonth(), now.getDate());
  const yesterday = new Date(today);
  yesterday.setDate(yesterday.getDate() - 1);
  const weekAgo = new Date(today);
  weekAgo.setDate(weekAgo.getDate() - 7);
  
  notifications.forEach(notification => {
    const date = new Date(notification.createdAt);
    
    if (date >= today) {
      groups.today.push(notification);
    } else if (date >= yesterday) {
      groups.yesterday.push(notification);
    } else if (date >= weekAgo) {
      groups.thisWeek.push(notification);
    } else {
      groups.older.push(notification);
    }
  });
  
  return groups;
}
```

### 3. Templates de Notificación

```javascript
// templates/notificationTemplates.js
export const notificationTemplates = {
  paymentReceived: (data) => ({
    title: 'Pago Recibido',
    message: `Se ha recibido tu pago de €${data.amount}. ¡Gracias por tu contribución!`,
    priority: 'MEDIUM',
    type: 'payment_received',
    data: {
      paymentId: data.paymentId,
      amount: data.amount
    }
  }),
  
  paymentReminder: (data) => ({
    title: 'Recordatorio de Pago',
    message: `Tu cuota de €${data.amount} vence el ${formatDate(data.dueDate)}`,
    priority: data.daysUntilDue <= 3 ? 'HIGH' : 'MEDIUM',
    type: 'payment_reminder',
    data: {
      memberId: data.memberId,
      amount: data.amount,
      dueDate: data.dueDate
    }
  }),
  
  membershipSuspended: (data) => ({
    title: 'Membresía Suspendida',
    message: 'Tu membresía ha sido suspendida por falta de pago. Regulariza tu situación para reactivarla.',
    priority: 'HIGH',
    type: 'membership_suspended',
    data: {
      memberId: data.memberId,
      suspendedDate: data.suspendedDate
    }
  }),
  
  announcement: (data) => ({
    title: data.title,
    message: data.message,
    priority: data.urgent ? 'HIGH' : 'LOW',
    type: 'announcement',
    data: {
      announcementId: data.id
    }
  })
};
```

## Testing de Notificaciones

```javascript
// __tests__/notifications.test.js
import { renderHook, act } from '@testing-library/react-hooks';
import { MockedProvider } from '@apollo/client/testing';
import { useNotifications } from '../hooks/useNotifications';

describe('Notifications', () => {
  const mockNotifications = [
    {
      id: '1',
      type: 'payment_received',
      title: 'Pago Recibido',
      message: 'Se ha recibido tu pago de €50',
      read: false,
      createdAt: new Date().toISOString(),
      priority: 'MEDIUM',
      data: { paymentId: '123' }
    }
  ];
  
  const mocks = [
    {
      request: {
        query: GET_NOTIFICATIONS,
        variables: {
          filter: {
            read: false,
            pagination: { page: 1, pageSize: 20 }
          }
        }
      },
      result: {
        data: {
          getNotifications: {
            nodes: mockNotifications,
            pageInfo: {
              hasNextPage: false,
              totalCount: 1,
              unreadCount: 1
            }
          }
        }
      }
    }
  ];
  
  it('should fetch and display notifications', async () => {
    const { result, waitForNextUpdate } = renderHook(
      () => useNotifications(),
      {
        wrapper: ({ children }) => (
          <MockedProvider mocks={mocks}>
            {children}
          </MockedProvider>
        )
      }
    );
    
    await waitForNextUpdate();
    
    expect(result.current.notifications).toHaveLength(1);
    expect(result.current.unreadCount).toBe(1);
  });
});
```

Esta guía proporciona toda la información necesaria para implementar y gestionar notificaciones en aplicaciones frontend de ASAM.