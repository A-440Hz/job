// Service Worker for caching collectable images
// This worker caches all images/video requests at the network layer
const CACHE_NAME = 'collectables-cache-v1';
const IMAGE_SOURCE = 'raw.githubusercontent.com';

// Install event - set up the service worker
self.addEventListener('install', (event) => {
  console.log('[Service Worker] Installing...');
  // Skip waiting to activate immediately
  self.skipWaiting();
});

// Activate event - clean up old caches
self.addEventListener('activate', (event) => {
  console.log('[Service Worker] Activating...');
  event.waitUntil(
    caches.keys().then((cacheNames) => {
      return Promise.all(
        cacheNames.map((cacheName) => {
          if (cacheName !== CACHE_NAME) {
            console.log('[Service Worker] Deleting old cache:', cacheName);
            return caches.delete(cacheName);
          }
        })
      );
    }).then(() => {
      // Take control of all clients immediately
      return self.clients.claim();
    })
  );
});

// Message event - receive list of URLs to cache from the app
self.addEventListener('message', (event) => {
  if (event.data && event.data.type === 'CACHE_COLLECTABLES') {
    const urls = event.data.urls;
    console.log('[Service Worker] Received request to cache', urls.length, 'images');

    event.waitUntil(
      caches.open(CACHE_NAME).then((cache) => {
        // Cache all URLs, but don't fail if some fail
        return Promise.allSettled(
          urls.map(url =>
            cache.add(url).catch(err => {
              console.error('[Service Worker] Failed to cache:', url, err);
              return null;
            })
          )
        ).then((results) => {
          const succeeded = results.filter(r => r.status === 'fulfilled').length;
          const failed = results.filter(r => r.status === 'rejected').length;
          console.log(`[Service Worker] Caching complete: ${succeeded} succeeded, ${failed} failed`);
        });
      })
    );
  }
});

// Fetch event - intercept network requests and serve from cache
self.addEventListener('fetch', (event) => {
  // Only handle GET requests
  if (event.request.method !== 'GET') {
    return;
  }

  // Check if it's a media request
  const url = new URL(event.request.url);
  const isMediaRequest = url.hostname === IMAGE_SOURCE ||
                         event.request.destination === 'image' ||
                         event.request.destination === 'video';

  if (!isMediaRequest) {
    return;
  }

  // For video range requests, bypass service worker and go to network
  // This allows video files to be loaded properly
  const isRangeRequest = event.request.headers.has('range');
  if (isRangeRequest) {
    console.log('[Service Worker] Bypassing cache for range request:', event.request.url);
    return; // Let the request go through to network normally
  }

  event.respondWith(
    caches.open(CACHE_NAME).then((cache) => {
      return cache.match(event.request).then((cachedResponse) => {
        if (cachedResponse) {
          console.log('[Service Worker] Serving from cache:', event.request.url);
          return cachedResponse;
        }

        // Not in cache, fetch from network
        console.log('[Service Worker] Fetching from network:', event.request.url);
        return fetch(event.request).then((response) => {
          // Only cache successful full responses (not partial content)
          if (response.status === 200) {
            const responseToCache = response.clone();
            cache.put(event.request, responseToCache);
          }

          return response;
        });
      });
    }).catch((error) => {
      console.error('[Service Worker] Fetch failed:', error);
      throw error;
    })
  );
});
