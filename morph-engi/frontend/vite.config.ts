import path from 'node:path'
import { fileURLToPath } from 'node:url'
import { defineConfig } from 'vite'
import { svelte } from '@sveltejs/vite-plugin-svelte'
import tailwindcss from '@tailwindcss/vite'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const repoRoot = path.resolve(__dirname, '../..')

const backendProxy = {
  // Legacy messaging routes — served by Morph stub (UsersPanel removed).
  '/api/messages': { target: 'http://127.0.0.1:9090', changeOrigin: true },
  '/api': { target: 'http://127.0.0.1:9096', changeOrigin: true },
} as const

export default defineConfig({
  plugins: [svelte(), tailwindcss()],
  envDir: repoRoot,
  server: { host: true, port: 5179, proxy: { ...backendProxy } },
  preview: { host: true, port: 5179, proxy: { ...backendProxy } },
})
