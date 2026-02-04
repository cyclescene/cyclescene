import adapter from '@sveltejs/adapter-vercel';
import { vitePreprocess } from '@sveltejs/vite-plugin-svelte';

/** @type {import('@sveltejs/kit').Config} */
const config = {
  // Consult https://svelte.dev/docs/kit/integrations
  // for more information about preprocessors
  preprocess: vitePreprocess(),

  kit: {
    // Using Vercel adapter for optimized Vercel deployments
    adapter: adapter(),
    alias: {
      "$lib/*": "./src/lib/*"
    },
    env: {
      privatePrefix: "API_",
      publicPrefix: "PUBLIC_"
    }
  }
};

export default config;
