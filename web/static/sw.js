const CACHE = 'wt-v1';
const STATIC = [
  '/manifest.json',
  '/icons/icon-192.png',
  '/icons/icon-512.png',
];

self.addEventListener('install', function(e) {
  e.waitUntil(
    caches.open(CACHE).then(function(c) { return c.addAll(STATIC); })
  );
  self.skipWaiting();
});

self.addEventListener('activate', function(e) {
  e.waitUntil(
    caches.keys().then(function(keys) {
      return Promise.all(
        keys.filter(function(k) { return k !== CACHE; }).map(function(k) { return caches.delete(k); })
      );
    })
  );
  self.clients.claim();
});

// Network-first for all navigation; cache-first for static assets
self.addEventListener('fetch', function(e) {
  var url = new URL(e.request.url);

  // Only handle same-origin requests
  if (url.origin !== self.location.origin) return;

  // Static assets: cache-first
  if (url.pathname.startsWith('/icons/') || url.pathname === '/manifest.json') {
    e.respondWith(
      caches.match(e.request).then(function(cached) {
        return cached || fetch(e.request);
      })
    );
    return;
  }

  // Everything else: network-first (app is server-rendered, always needs fresh data)
  e.respondWith(
    fetch(e.request).catch(function() {
      return caches.match(e.request);
    })
  );
});
