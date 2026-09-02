import path from 'node:path'
import { fileURLToPath } from 'node:url'
import { defineConfig } from 'vite'
import { svelte } from '@sveltejs/vite-plugin-svelte'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const repoRoot = path.resolve(__dirname, '../..')

// https://vite.dev/config/
export default defineConfig({
  plugins: [svelte()],
  envDir: repoRoot,
  server: {
    port: 8044,
    // Same-origin dev: set VITE_API_BASE="" in .env so requests hit the Vite server and proxy to the API (no CORS).
    proxy: {
      '/auth': { target: 'http://127.0.0.1:8043', changeOrigin: true },
      '/health': { target: 'http://127.0.0.1:8043', changeOrigin: true },
      '/emails': { target: 'http://127.0.0.1:8043', changeOrigin: true },
      '/contacts': { target: 'http://127.0.0.1:8043', changeOrigin: true },
      '/templates': { target: 'http://127.0.0.1:8043', changeOrigin: true },
      '/merge-data': { target: 'http://127.0.0.1:8043', changeOrigin: true },
      '/email': { target: 'http://127.0.0.1:8043', changeOrigin: true },
      '/publishes': { target: 'http://127.0.0.1:8043', changeOrigin: true },
      '/publish-drafts': { target: 'http://127.0.0.1:8043', changeOrigin: true },
      '/public': { target: 'http://127.0.0.1:8043', changeOrigin: true },
      '/ai': {
        target: 'http://127.0.0.1:8043',
        changeOrigin: true,
        timeout: 600_000,
        proxyTimeout: 600_000,
      },
      '/reports': { target: 'http://127.0.0.1:8043', changeOrigin: true },
    },
  },
  preview: {
    port: 8044,
  },
})
