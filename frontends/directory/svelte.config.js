import adapter from '@sveltejs/adapter-static';
import { vitePreprocess } from '@sveltejs/vite-plugin-svelte';

/** @type {import('@sveltejs/kit').Config} */
const config = {
  // Consult https://svelte.dev/docs/kit/integrations
  // for more information about preprocessors
  preprocess: vitePreprocess(),
  kit: {
    adapter: adapter({ strict: false }),
    prerender: {
      entries: ['/']
    },
    alias: { '@': './src/lib/*' },
    env: {
      publicPrefix: 'PUBLIC_'
    }
  }
};

export default config;
