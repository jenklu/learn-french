// Service Worker for French Learning App
// Handles caching, offline support, and push notifications

const CACHE_NAME = 'learn-french-v1';
const ASSETS_TO_CACHE = [
  '/',
  '/index.html',
  '/manifest.json',
  '/vite.svg',
];

// Install service worker
self.addEventListener('install', (event) => {
  event.waitUntil(
    caches.open(CACHE_NAME)
      .then((cache) => {
        console.log('Service Worker: Caching assets');
        return cache.addAll(ASSETS_TO_CACHE);
      })
      .then(() => {
        console.log('Service Worker: Assets cached successfully');
        return self.skipWaiting();
      })
      .catch((error) => {
        console.error('Service Worker: Failed to cache assets:', error);
      })
  );
});

// Activate service worker
self.addEventListener('activate', (event) => {
  event.waitUntil(
    caches.keys().then((cacheNames) => {
      return Promise.all(
        cacheNames.map((cacheName) => {
          if (cacheName !== CACHE_NAME) {
            console.log('Service Worker: Removing old cache:', cacheName);
            return caches.delete(cacheName);
          }
        })
      );
    })
    .then(() => {
      console.log('Service Worker: Activated');
      return self.clients.claim();
    })
  );
});

// Fetch assets from cache or network
self.addEventListener('fetch', (event) => {
  event.respondWith(
    caches.match(event.request)
      .then((response) => {
        if (response) {
          console.log('Service Worker: Serving from cache:', event.request.url);
          return response;
        }
        console.log('Service Worker: Fetching from network:', event.request.url);
        return fetch(event.request);
      })
  );
});

// Handle push notifications
self.addEventListener('push', (event) => {
  console.log('Service Worker: Push notification received');
  
  if (!self.Notification) {
    console.log('Service Worker: Notifications not supported');
    return;
  }

  // Parse the push notification data
  let notificationData;
  try {
    notificationData = event.data?.json();
  } catch (error) {
    console.error('Service Worker: Failed to parse push data:', error);
    notificationData = {
      title: 'Learn French',
      body: 'New notification',
    };
  }

  // Show notification
  event.waitUntil(
    self.registration.showNotification(notificationData.title || 'Learn French', {
      body: notificationData.body || 'You have a new notification',
      icon: '/icon-192x192.png',
      badge: '/icon-192x192.png',
      data: notificationData.data || {},
      actions: notificationData.actions || [],
    })
  );
});

// Handle notification click
self.addEventListener('notificationclick', (event) => {
  console.log('Service Worker: Notification clicked');
  
  event.notification.close();
  
  // Focus or open the app
  event.waitUntil(
    self.clients.matchAll({
      type: 'window',
      includeUncontrolled: true,
    })
      .then((clients) => {
        if (clients.length > 0) {
          // Focus existing window
          clients[0].focus();
          // Navigate to the URL from the notification
          if (event.notification.data && event.notification.data.url) {
            clients[0].navigate(event.notification.data.url);
          }
        } else {
          // Open new window
          return self.clients.openWindow(event.notification.data?.url || '/');
        }
      })
  );
});

// Handle notification close
self.addEventListener('notificationclose', (event) => {
  console.log('Service Worker: Notification closed');
});

// Handle push subscription change
self.addEventListener('pushsubscriptionchange', (event) => {
  console.log('Service Worker: Push subscription changed');
  
  event.waitUntil(
    // In a real app, you would send the new subscription to the server
    // For now, we'll just log it
    Promise.resolve().then(() => {
      console.log('Service Worker: Push subscription updated');
    })
  );
});

console.log('Service Worker: Loaded');
