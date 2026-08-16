import { defineConfig } from 'vite'
import { svelte } from '@sveltejs/vite-plugin-svelte'
import tailwindcss from '@tailwindcss/vite'

const backendProxy = {
  '/api/messages': { target: 'http://127.0.0.1:5001', changeOrigin: true },
  '/api': { target: 'http://127.0.0.1:9096', changeOrigin: true },
} as const

export default defineConfig({
  plugins: [svelte(), tailwindcss()],
  server: { port: 5179, proxy: { ...backendProxy } },
  preview: { port: 5179, proxy: { ...backendProxy } },
})
