import adapter from '@sveltejs/adapter-auto';
import { sveltekit } from '@sveltejs/kit/vite';
import tailwindcss from '@tailwindcss/vite';
import { defineConfig } from 'vite';

// Remote dev (Coder, Codespaces): the browser reaches vite through a proxy host,
// which vite rejects unless listed. Comma-separated; a leading dot matches subdomains.
const allowedHosts = (process.env.VITE_ALLOWED_HOSTS ?? '')
	.split(',')
	.map((host) => host.trim())
	.filter(Boolean);

const apiTarget = process.env.OPSYBOT_API_URL ?? 'http://127.0.0.1:8099';

export default defineConfig({
	plugins: [
		tailwindcss(),
		sveltekit({
			compilerOptions: {
				// Force runes for project code, not node_modules; removable in svelte 6
				runes: ({ filename }) =>
					filename.split(/[/\\]/).includes('node_modules') ? undefined : true
			},

			adapter: adapter()
		})
	],
	server: {
		host: true,
		allowedHosts,
		// Keeps the API on the same origin as the app, matching the deployed topology
		proxy: { '/v1': { target: apiTarget, changeOrigin: false } }
	}
});
