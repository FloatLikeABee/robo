import { sveltekit } from '@sveltejs/kit/vite';
import tailwindcss from '@tailwindcss/vite';
import { defineConfig } from 'vite';

/** UsersPanel messaging must be listed before `/api` (DataX backend). */
const backendProxy = {
	'/api/messages': {
		target: 'http://127.0.0.1:5001',
		changeOrigin: true
	},
	'/api': {
		target: 'http://127.0.0.1:3050',
		changeOrigin: true
	},
	'/public': {
		target: 'http://127.0.0.1:3050',
		changeOrigin: true
	},
	'/metabase': {
		target: 'http://127.0.0.1:3050',
		changeOrigin: true,
		ws: true
	}
} as const;

export default defineConfig({
	plugins: [tailwindcss(), sveltekit()],
	server: {
		port: 5178,
		strictPort: true,
		proxy: { ...backendProxy }
	},
	preview: {
		port: 5178,
		strictPort: true,
		proxy: { ...backendProxy }
	}
});
