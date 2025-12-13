// this file bootstraps the React app and connects it to the HTML page

import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import App from './App'
import { TrackerDataProvider } from './JobAppTrackerDataContext';
import { CollectablesDataProvider } from './CollectablesDataContext';
import './index.css'
import { queryClient } from './queryClient'
import { QueryClientProvider } from '@tanstack/react-query';

// Register service worker for cacheing
if ('serviceWorker' in navigator) {
  window.addEventListener('load', () => {
    navigator.serviceWorker
      .register('/sw.js', { scope: '/' })
      .then((registration) => {
        console.log('[App] Service Worker registered successfully:', registration.scope);
      })
      .catch((error) => {
        console.error('[App] Service Worker registration failed:', error);
      });
  });
}

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <QueryClientProvider client={queryClient}>
      <TrackerDataProvider>
        <CollectablesDataProvider>
          <App />
        </CollectablesDataProvider>
      </TrackerDataProvider>
    </QueryClientProvider>
  </StrictMode>,
);
