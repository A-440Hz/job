import { getImageURL } from '../api/collectable';

// Must match the cache name in sw.js
const CACHE_NAME = 'collectables-cache-v1';

/**
 * Sends a list of collectable image URLs to the service worker for caching
 * Only caches URLs that aren't already in the cache (checked via Cache API)
 * @param all_collectables Array of collectables with Filename property
 */
export async function cacheCollectables(all_collectables: any[]): Promise<void> {
  // Check if service workers are supported
  if (!('serviceWorker' in navigator)) {
    console.warn('[Cache] Service workers not supported in this browser');
    return;
  }

  // Check if Cache API is supported
  if (!('caches' in window)) {
    console.warn('[Cache] Cache API not supported in this browser');
    return;
  }

  // Check if service worker is active
  if (!navigator.serviceWorker.controller) {
    console.warn('[Cache] No active service worker controller yet');
    return;
  }

  try {
    // Convert filenames to full URLs using getImageURL
    const urls = all_collectables
      .filter(c => c.Filename) // Filter out any collectables without filenames
      .map(c => getImageURL(c.Filename));

    console.log('[Cache] Checking cache status for', urls.length, 'images');

    // Check which URLs aren't cached yet
    const cache = await caches.open(CACHE_NAME);
    const uncachedUrls: string[] = [];

    for (const url of urls) {
      const cached = await cache.match(url);
      if (!cached) {
        uncachedUrls.push(url);
      }
    }

    const alreadyCached = urls.length - uncachedUrls.length;
    if (alreadyCached > 0) {
      console.log('[Cache]', alreadyCached, 'images already cached');
    }

    if (uncachedUrls.length === 0) {
      console.log('[Cache] All images already cached, skipping');
      return;
    }

    console.log('[Cache] Requesting service worker to cache', uncachedUrls.length, 'new images');

    // Send only uncached URLs to service worker
    navigator.serviceWorker.controller.postMessage({
      type: 'CACHE_COLLECTABLES',
      urls: uncachedUrls
    });
  } catch (error) {
    console.error('[Cache] Failed to check cache or send request to service worker:', error);
  }
}
