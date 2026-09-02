import path from 'node:path'
import { fileURLToPath } from 'node:url'
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const repoRoot = path.resolve(__dirname, '../..')

// https://vite.dev/config/
const backendProxy = {
  // Legacy messaging routes — served by Morph stub (UsersPanel removed). Must precede `/api` (FormsX backend).
  '/api/messages': { target: 'http://127.0.0.1:9090', changeOrigin: true },
  '/api': { target: 'http://localhost:29909', changeOrigin: true },
  '/uploads': { target: 'http://localhost:29909', changeOrigin: true },
} as const;

export default defineConfig({
  plugins: [react(), tailwindcss()],
  envDir: repoRoot,
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
