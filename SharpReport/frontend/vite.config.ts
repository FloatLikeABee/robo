import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { sveltekit } from '@sveltejs/kit/vite';
import tailwindcss from '@tailwindcss/vite';
import { defineConfig, loadEnv } from 'vite';

const repoRoot = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '../..');

function sharpReportApiOrigin(mode: string): string {
	const env = loadEnv(mode, repoRoot, '');
	const raw = (env.SHARPREPORT_PORT || '').trim();
	const port = /^\d+$/.test(raw) ? raw : '3050';
	return `http://127.0.0.1:${port}`;
}

export default defineConfig(({ mode }) => {
	const env = loadEnv(mode, repoRoot, '');
	if (env.SHARPREPORT_PORT) {
		process.env.SHARPREPORT_PORT = env.SHARPREPORT_PORT;
	}
	const dataAccessApi = sharpReportApiOrigin(mode);
	/** Legacy messaging routes (Morph stub) must be listed before `/api` (Data Access backend). */
	const backendProxy = {
		'/api/messages': {
			target: 'http://127.0.0.1:9090',
			changeOrigin: true
		},
		'/api': {
			target: dataAccessApi,
			changeOrigin: true
		},
		'/public': {
			target: dataAccessApi,
			changeOrigin: true
		},
		'/metabase': {
			target: dataAccessApi,
			changeOrigin: true,
			ws: true
		}
	};

	return {
		plugins: [tailwindcss(), sveltekit()],
		envDir: repoRoot,
		optimizeDeps: {
			include: ['marked', 'dompurify']
		},
		server: {
			host: true,
			port: 5178,
			strictPort: true,
			proxy: { ...backendProxy }
		},
		preview: {
			host: true,
			port: 5178,
			strictPort: true,
			proxy: { ...backendProxy }
		}
	};
});
