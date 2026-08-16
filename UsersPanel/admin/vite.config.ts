import { defineConfig } from 'vite'
import { svelte } from '@sveltejs/vite-plugin-svelte'

// https://vite.dev/config/
export default defineConfig({
  plugins: [svelte()],
  server: {
    proxy: {
      '/api': {
        target: 'http://127.0.0.1:5001',
        changeOrigin: true,
      },
      // OpenAPI (same backend as /api) — e.g. http://localhost:5173/swagger-ui
      '/swagger-ui': { target: 'http://127.0.0.1:5001', changeOrigin: true },
      '/openapi.json': { target: 'http://127.0.0.1:5001', changeOrigin: true },
    },
  },
})
