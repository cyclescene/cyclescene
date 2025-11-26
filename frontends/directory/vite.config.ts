import tailwindcss from '@tailwindcss/vite';
import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';
import { loadEnv } from 'vite';

export default defineConfig(({ command }) => {
	// Load .env variables for dev server
	if (command === 'serve') {
		const env = loadEnv('development', process.cwd(), '');
		Object.assign(process.env, env);
	}

	return {
		plugins: [tailwindcss(), sveltekit()]
	};
});
