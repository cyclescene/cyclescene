import tailwindcss from '@tailwindcss/vite';
import { defineConfig } from 'vite';
import { svelte } from '@sveltejs/vite-plugin-svelte';
import { VitePWA } from 'vite-plugin-pwa';
import Icons from "unplugin-icons/vite"
import path from "path"
import { ngrok } from 'vite-plugin-ngrok';

// City configurations for PWA manifest
const cityConfigs = {
  pdx: {
    name: "Cycle Scene - PDX",
    short_name: "CycleScenePDX",
    description: "Upcoming bike rides in Portland, Oregon"
  },
  slc: {
    name: "Cycle Scene - SLC",
    short_name: "CycleSceneSLC",
    description: "Upcoming bike rides in Salt Lake City, Utah"
  }
};

// Get city code from environment variable, default to pdx for development
const cityCode = process.env.VITE_CITY_CODE || 'pdx';
const cityConfig = cityConfigs[cityCode] || cityConfigs.pdx;

// Get ngrok token from environment variable (only for local development)
const ngrokToken = process.env.NGROK_TOKEN;

// https://vite.dev/config/
export default defineConfig({

  server: {
    allowedHosts: true

  },
  plugins: [
    tailwindcss(),
    svelte(),
    ...(ngrokToken ? [ngrok(ngrokToken)] : []),
    Icons({
      compiler: "svelte",
      autoInstall: true,
      prefix: "i"
    }),
    VitePWA({
      registerType: 'prompt',
      strategies: "injectManifest",
      srcDir: "src/lib/",
      filename: "sw.ts",
      includeManifestIcons: false,
      manifest: {
        name: cityConfig.name,
        short_name: cityConfig.short_name,
        description: cityConfig.description,
        theme_color: "#000000",
        background_color: "#000000",
        display: "standalone",
        start_url: "/",
        scope: "/",
        orientation: "portrait-primary",
        categories: ["lifestyle", "sports"],
        icons: [
          {
            src: "/icons/manifest-icon-192.maskable.png",
            sizes: "192x192",
            type: "image/png",
            purpose: "any"
          },
          {
            src: "/icons/manifest-icon-192.maskable.png",
            sizes: "192x192",
            type: "image/png",
            purpose: "maskable"
          },
          {
            src: "/icons/manifest-icon-512.maskable.png",
            sizes: "512x512",
            type: "image/png",
            purpose: "any"
          },
          {
            src: "/icons/manifest-icon-512.maskable.png",
            sizes: "512x512",
            type: "image/png",
            purpose: "maskable"
          },
          {
            src: "/icons/favicon-196.png",
            sizes: "196x196",
            type: "image/png",
            purpose: "any"
          },
          {
            src: "/cyclescene_temp.png",
            sizes: "any",
            type: "image/png",
            purpose: "any"
          }
        ]
      },
      // workbox: {
      //   globPatterns: ['**/*.{js,css,html,ico,png,svg}'],
      //   runtimeCaching: [
      //     {
      //       urlPattern: /^https:\/\/(\w+\.)?basemaps\.cartocdn\.com\/light_all\/.*\.png$/,
      //       handler: 'CacheFirst',
      //       options: {
      //         cacheName: 'cartodb-light-tiles-cache',
      //         expiration: {
      //           maxEntries: 500,
      //           maxAgeSeconds: 60 * 60 * 24 * 30
      //         },
      //         cacheableResponse: { statuses: [0, 200] }
      //       }
      //     },
      //     {
      //       urlPattern: /^https:\/\/(\w+\.)?basemaps\.cartocdn\.com\/dark_all\/.*\.png$/,
      //       handler: 'CacheFirst',
      //       options: {
      //         cacheName: 'cartodb-dark-tiles-cache',
      //         expiration: {
      //           maxEntries: 500,
      //           maxAgeSeconds: 60 * 60 * 24 * 30
      //         },
      //         cacheableResponse: { statuses: [0, 200] }
      //       }
      //     },
      //     {
      //       urlPattern: /^https:\/\/faas-sfo3-7872a1dd\.doserverless\.co\/api\/v1\/web\/fn-69328def-615c-4bce-88c0-dc912d5f1d84\/api\/(upcoming|past)$/,
      //       handler: 'NetworkFirst',
      //       options: {
      //         cacheName: 'cycle-scene-api-cache',
      //         expiration: 60 * 60 * 6,
      //         cacheableResponse: { statuses: [0, 200] }
      //       }
      //     }
      //   ]
      // }
    })
  ],
  resolve: {
    alias: {
      $lib: path.resolve(__dirname, './src/lib')
    }
  }
});
