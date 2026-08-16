import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'

// https://vite.dev/config/
const backendProxy = {
  // UsersPanel messaging — must be listed before `/api` (FormsX backend).
  '/api/messages': { target: 'http://127.0.0.1:5001', changeOrigin: true },
  '/api': { target: 'http://localhost:29909', changeOrigin: true },
  '/uploads': { target: 'http://localhost:29909', changeOrigin: true },
} as const;

export default defineConfig({
  plugins: [react(), tailwindcss()],
  server: {
    port: 19909,
    proxy: { ...backendProxy },
  },
  /** Same-origin `/api` during `vite preview`; without this, login POSTs hit the preview server and return 404. */
  preview: {
    port: 19909,
    proxy: { ...backendProxy },
  },
})
