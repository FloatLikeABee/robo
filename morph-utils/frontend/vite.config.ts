import path from 'node:path'
import { fileURLToPath } from 'node:url'
import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const repoRoot = path.resolve(__dirname, '../..')

const morphProxy = {
  '/api': { target: 'http://127.0.0.1:9090', changeOrigin: true },
} as const;

export default defineConfig({
  plugins: [react()],
  envDir: repoRoot,
  server: {
    port: 3040,
    host: true,
    proxy: { ...morphProxy },
  },
  preview: {
    port: 3040,
    proxy: { ...morphProxy },
  },
});
