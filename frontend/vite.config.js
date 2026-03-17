import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';

export default defineConfig({
	plugins: [sveltekit()],
	server: {
		host: true,
		port: 5173,
		watch: {
			usePolling: true,
			interval: 500
		},
		proxy: {
			'/api': {
				target: 'http://api:8080',
				changeOrigin: true
			}
		}
	}
});
