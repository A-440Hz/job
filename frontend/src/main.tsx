// this file bootstraps the React app and connects it to the HTML page

import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import App from './App'
import { TrackerDataProvider } from './JobAppTrackerDataContext';
import './index.css'
import Topbar from './Topbar';

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <TrackerDataProvider>
      <Topbar />
      <App />
    </TrackerDataProvider>
  </StrictMode>,
);
