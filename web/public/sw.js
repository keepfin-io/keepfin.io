/* Only the public offline page is cached. Never intercept API/auth responses,
   portfolio reports, imports, exports, or navigation containing share tokens. */
const CACHE = 'keepfin-offline-v1'
self.addEventListener('install', event => { event.waitUntil(caches.open(CACHE).then(cache => cache.add('/offline.html'))) })
self.addEventListener('activate', event => { event.waitUntil(caches.keys().then(keys => Promise.all(keys.filter(key => key.startsWith('keepfin-offline-') && key !== CACHE).map(key => caches.delete(key))))) })
self.addEventListener('fetch', event => {
  const url = new URL(event.request.url)
  if (event.request.method !== 'GET' || event.request.mode !== 'navigate' || url.origin !== self.location.origin || url.pathname.startsWith('/api/')) return
  event.respondWith(fetch(event.request).catch(() => caches.match('/offline.html')))
})
