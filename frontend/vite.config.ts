import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react(), tailwindcss()],
  // Copy service worker to build output without processing it
  publicDir: './src/assets',
  base: './',
  server: {
    allowedHosts: ['.localhost', "frontend-staging-60b5.up.railway.app", "jobapptracker.up.railway.app", "haotianswebsite.com"],
    host: '0.0.0.0',
  },
  build: {
    minify: 'esbuild', // Ensures production builds are minimized
  },
})
