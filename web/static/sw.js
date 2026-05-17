const CACHE_STATIC = 'wt-static-v2';
const CACHE_PAGES  = 'wt-pages-v2';
const CACHE_MEDIA  = 'wt-media-v2';

const STATIC_ASSETS = [
  '/manifest.json',
  '/icons/icon-192.png',
  '/icons/icon-512.png',
  '/icons/favicon-32.png',
  '/offline',
];

// Install: pre-cache static assets
self.addEventListener('install', function(e) {
  e.waitUntil(
    caches.open(CACHE_STATIC).then(function(c) {
      return c.addAll(STATIC_ASSETS);
    }).catch(function() {})
  );
  self.skipWaiting();
});

// Activate: delete old caches
self.addEventListener('activate', function(e) {
  var keep = [CACHE_STATIC, CACHE_PAGES, CACHE_MEDIA];
  e.waitUntil(
    caches.keys().then(function(keys) {
      return Promise.all(
        keys.filter(function(k) { return keep.indexOf(k) === -1; })
            .map(function(k) { return caches.delete(k); })
      );
    })
  );
  self.clients.claim();
});

self.addEventListener('fetch', function(e) {
  var req = e.request;
  var url = new URL(req.url);

  // Only handle same-origin
  if (url.origin !== self.location.origin) return;

  // Never cache POST/DELETE/etc.
  if (req.method !== 'GET') return;

  // Static icons & manifest — cache-first
  if (url.pathname.startsWith('/icons/') || url.pathname === '/manifest.json' || url.pathname === '/sw.js') {
    e.respondWith(
      caches.match(req).then(function(cached) {
        return cached || fetch(req).then(function(resp) {
          return caches.open(CACHE_STATIC).then(function(c) { c.put(req, resp.clone()); return resp; });
        });
      })
    );
    return;
  }

  // Media files — cache-first, then network
  if (url.pathname.startsWith('/media/')) {
    e.respondWith(
      caches.match(req).then(function(cached) {
        if (cached) return cached;
        return fetch(req).then(function(resp) {
          if (resp.ok) {
            caches.open(CACHE_MEDIA).then(function(c) { c.put(req, resp.clone()); });
          }
          return resp;
        });
      })
    );
    return;
  }

  // HTMX partials — network-only (don't cache partial HTML)
  if (url.pathname.startsWith('/workouts/htmx/') || url.pathname.startsWith('/analytics/')) {
    return;
  }

  // App pages — network-first, cache on success, offline fallback
  e.respondWith(
    fetch(req).then(function(resp) {
      if (resp.ok && req.headers.get('accept') && req.headers.get('accept').indexOf('text/html') !== -1) {
        caches.open(CACHE_PAGES).then(function(c) { c.put(req, resp.clone()); });
      }
      return resp;
    }).catch(function() {
      return caches.match(req).then(function(cached) {
        return cached || caches.match('/offline');
      });
    })
  );
});
