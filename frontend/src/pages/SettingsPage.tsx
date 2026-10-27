import React, { useState, useEffect } from 'react';
import { useAuth } from '../contexts/AuthContext';

interface Reminder {
  id: number;
  user_id: number;
  enabled: boolean;
}

interface PushSubscription {
  id: number;
  user_id: number;
  endpoint: string;
  auth: string;
  p256dh: string;
}

const SettingsPage: React.FC = () => {
  const { user, updateTimezone } = useAuth();
  const [reminder, setReminder] = useState<Reminder | null>(null);
  const [pushSubscriptions, setPushSubscriptions] = useState<PushSubscription[]>([]);
  const [timezone, setTimezone] = useState(user?.timezone || Intl.DateTimeFormat().resolvedOptions().timeZone);
  const [reminderEnabled, setReminderEnabled] = useState(false);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [notificationPermission, setNotificationPermission] = useState<NotificationPermission>('default');

  const API_BASE = import.meta.env.VITE_API_BASE_URL || '/api';

  useEffect(() => {
    if (user) {
      fetchReminder();
      fetchPushSubscriptions();
      checkNotificationPermission();
    }
  }, [user]);

  useEffect(() => {
    if (user) {
      updateTimezone(timezone);
    }
  }, [timezone, user, updateTimezone]);

  const checkNotificationPermission = () => {
    if ('Notification' in window) {
      setNotificationPermission(Notification.permission);
    }
  };

  const fetchReminder = async () => {
    try {
      const response = await fetch(`${API_BASE}/reminders`, {
        credentials: 'include',
      });
      
      if (response.ok) {
        const data = await response.json();
        setReminder(data);
        setReminderEnabled(data.enabled);
      }
    } catch (err) {
      console.error('Failed to fetch reminder:', err);
    }
  };

  const fetchPushSubscriptions = async () => {
    try {
      const response = await fetch(`${API_BASE}/push/subscriptions`, {
        credentials: 'include',
      });
      
      if (response.ok) {
        const data = await response.json();
        setPushSubscriptions(data);
      }
    } catch (err) {
      console.error('Failed to fetch push subscriptions:', err);
    } finally {
      setLoading(false);
    }
  };

  const handleReminderToggle = async (enabled: boolean) => {
    try {
      setSaving(true);
      const response = await fetch(`${API_BASE}/reminders`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ enabled }),
        credentials: 'include',
      });
      
      if (response.ok) {
        setReminderEnabled(enabled);
        setReminder(prev => prev ? { ...prev, enabled } : null);
      }
    } catch (err) {
      console.error('Failed to update reminder:', err);
    } finally {
      setSaving(false);
    }
  };

  const requestNotificationPermission = async () => {
    if ('Notification' in window) {
      try {
        const permission = await Notification.requestPermission();
        setNotificationPermission(permission);
        
        if (permission === 'granted') {
          // Register service worker and subscribe to push notifications
          await subscribeToPushNotifications();
        }
      } catch (err) {
        console.error('Failed to request notification permission:', err);
      }
    }
  };

  const subscribeToPushNotifications = async () => {
    try {
      const registration = await navigator.serviceWorker.ready;
      const subscription = await registration.pushManager.subscribe({
        userVisibleOnly: true,
        applicationServerKey: import.meta.env.VITE_VAPID_PUBLIC_KEY,
      });

      // Store subscription on server
      const response = await fetch(`${API_BASE}/push/subscribe`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          endpoint: subscription.endpoint,
          keys: {
            auth: subscription.toJSON().keys.auth,
            p256dh: subscription.toJSON().keys.p256dh,
          },
        }),
        credentials: 'include',
      });

      if (response.ok) {
        await fetchPushSubscriptions();
      }
    } catch (err) {
      console.error('Failed to subscribe to push notifications:', err);
    }
  };

  const unsubscribeFromPushNotifications = async (subscriptionId: number) => {
    try {
      const response = await fetch(`${API_BASE}/push/subscriptions?id=${subscriptionId}`, {
        method: 'DELETE',
        credentials: 'include',
      });
      
      if (response.ok) {
        await fetchPushSubscriptions();
      }
    } catch (err) {
      console.error('Failed to unsubscribe from push notifications:', err);
    }
  };

  const handleTimezoneChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    setTimezone(e.target.value);
  };

  if (loading) {
    return (
      <div className="min-vh-100 d-flex align-items-center justify-content-center">
        <div className="spinner-border text-primary" role="status">
          <span className="visually-hidden">Loading...</span>
        </div>
      </div>
    );
  }

  return (
    <div className="py-4">
      <h2 className="mb-4">Settings</h2>

      <div className="row">
        {/* Profile Settings */}
        <div className="col-lg-6 mb-4">
          <div className="card shadow">
            <div className="card-header">
              <h5 className="mb-0">Profile Settings</h5>
            </div>
            <div className="card-body">
              <div className="mb-3">
                <label htmlFor="username" className="form-label">
                  Username
                </label>
                <input
                  type="text"
                  id="username"
                  className="form-control"
                  value={user?.username || ''}
                  readOnly
                />
              </div>

              <div className="mb-3">
                <label htmlFor="email" className="form-label">
                  Email
                </label>
                <input
                  type="email"
                  id="email"
                  className="form-control"
                  value={user?.email || ''}
                  readOnly
                />
              </div>

              <div className="mb-3">
                <label htmlFor="timezone" className="form-label">
                  Timezone
                </label>
                <select
                  id="timezone"
                  className="form-select"
                  value={timezone}
                  onChange={handleTimezoneChange}
                >
                  {Intl.supportedValuesOf('timeZone').map(tz => (
                    <option key={tz} value={tz}>{tz}</option>
                  ))}
                </select>
                <div className="form-text">
                  Your current timezone: {timezone}
                </div>
              </div>
            </div>
          </div>
        </div>

        {/* Notification Settings */}
        <div className="col-lg-6 mb-4">
          <div className="card shadow">
            <div className="card-header">
              <h5 className="mb-0">Notification Settings</h5>
            </div>
            <div className="card-body">
              <div className="mb-4">
                <h6>Daily Reminders</h6>
                <p className="text-muted small">
                  Get reminded to practice at 10AM and 9PM in your timezone.
                </p>
                <div className="form-check form-switch">
                  <input
                    className="form-check-input"
                    type="checkbox"
                    id="reminderToggle"
                    checked={reminderEnabled}
                    onChange={(e) => handleReminderToggle(e.target.checked)}
                    disabled={saving}
                  />
                  <label className="form-check-label" htmlFor="reminderToggle">
                    {saving ? 'Saving...' : 'Enable Daily Reminders'}
                  </label>
                </div>
              </div>

              <div className="mb-4">
                <h6>Browser Notifications</h6>
                <p className="text-muted small">
                  Receive push notifications for reminders and updates.
                </p>
                
                {'Notification' in window && (
                  <>
                    <div className="mb-3">
                      <p className="mb-2">
                        Permission: 
                        <span className={`badge ${notificationPermission === 'granted' ? 'bg-success' : notificationPermission === 'denied' ? 'bg-danger' : 'bg-secondary'}`}>
                          {notificationPermission}
                        </span>
                      </p>
                      
                      {notificationPermission === 'default' && (
                        <button
                          className="btn btn-primary btn-sm"
                          onClick={requestNotificationPermission}
                        >
                          Enable Notifications
                        </button>
                      )}
                      
                      {notificationPermission === 'denied' && (
                        <button
                          className="btn btn-outline-secondary btn-sm"
                          onClick={requestNotificationPermission}
                        >
                          Request Permission Again
                        </button>
                      )}
                    </div>

                    {pushSubscriptions.length > 0 && (
                      <div className="mt-3">
                        <h6 className="mb-2">Active Subscriptions</h6>
                        <div className="list-group list-group-flush">
                          {pushSubscriptions.map((sub) => (
                            <div key={sub.id} className="list-group-item d-flex justify-content-between align-items-center">
                              <small className="text-truncate">{sub.endpoint}</small>
                              <button
                                className="btn btn-sm btn-outline-danger"
                                onClick={() => unsubscribeFromPushNotifications(sub.id)}
                              >
                                Remove
                              </button>
                            </div>
                          ))}
                        </div>
                      </div>
                    )}
                  </>
                )}
              </div>
            </div>
          </div>
        </div>
      </div>

      <div className="card shadow">
        <div className="card-body">
          <h5 className="card-title">About</h5>
          <p className="card-text">
            This application helps you improve your French listening skills through interactive exercises.
            Your progress is tracked and you can review your past sessions at any time.
          </p>
          <p className="text-muted small mb-0">
            Version: 1.0.0 | Built with React, Go, and PostgreSQL
          </p>
        </div>
      </div>
    </div>
  );
};

export default SettingsPage;
