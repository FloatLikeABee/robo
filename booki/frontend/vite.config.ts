import path from 'node:path'
import { fileURLToPath } from 'node:url'
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const morphBroadcastRoot = path.resolve(__dirname, '../../morph-broadcast')

// https://vite.dev/config/
const backendProxy = {
  // UsersPanel messaging — must be listed before `/api` (Booki backend).
  '/api/messages': { target: 'http://127.0.0.1:5001', changeOrigin: true },
  // Morph broadcast (contacts, ComposerX import, text-assist) — Morph API.
  '/api/tran': { target: 'http://localhost:9090', changeOrigin: true },
  '/api/composerx': { target: 'http://localhost:9090', changeOrigin: true },
  '/api': { target: 'http://127.0.0.1:9095', changeOrigin: true },
} as const;

export default defineConfig({
  plugins: [react(), tailwindcss()],
  resolve: {
    alias: {
      '@robo/morph-broadcast': morphBroadcastRoot,
    },
  },
  server: {
    port: 5174,
    proxy: { ...backendProxy },
    fs: {
      allow: [path.resolve(__dirname, '../..')],
    },
  },
  preview: {
    port: 5174,
    proxy: { ...backendProxy },
  },
})
