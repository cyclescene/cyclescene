import tailwindcss from '@tailwindcss/vite';
import { defineConfig } from 'vite';
import { svelte } from '@sveltejs/vite-plugin-svelte';
import { VitePWA } from 'vite-plugin-pwa';
import Icons from "unplugin-icons/vite"
import path from "path"
import { ngrok } from 'vite-plugin-ngrok';

// https://vite.dev/config/
export default defineConfig({

  server: {
    allowedHosts: true

  },
  plugins: [
    tailwindcss(),
    svelte(),
    ngrok('1hwZtwwW6brY5mU9EirSqlr0WtI_x5NNvoe8PcWkKffzmgRU'),
    Icons({
      compiler: "svelte",
      autoInstall: true,
      prefix: "i"
    }),
    VitePWA({
      registerType: 'autoUpdate',
      manifest: {
        name: "Cycle Scene - PDX",
        short_name: "CycleScenePDX",
        description: "Upcoming bike rides in Portland, Oregon",
        theme_color: "#000000",
        icons: [
          {
            src: "public/icons/manifest-icon-192.maskable.png",
            sizes: "192x192",
            type: "image/png",
            purpose: "any"
          },
          {
            src: "public/icons/manifest-icon-192.maskable.png",
            sizes: "192x192",
            type: "image/png",
            purpose: "maskable"
          },
          {
            src: "public/icons/manifest-icon-512.maskable.png",
            sizes: "512x512",
            type: "image/png",
            purpose: "any"
          },
          {
            src: "public/icons/manifest-icon-512.maskable.png",
            sizes: "512x512",
            type: "image/png",
            purpose: "maskable"
          }
        ]
      },
      workbox: {
        globPatterns: ['**/*.{js,css,html,ico,png,svg}'],
        runtimeCaching: [
          {
            urlPattern: /^https:\/\/(\w+\.)?basemaps\.cartocdn\.com\/light_all\/.*\.png$/,
            handler: 'CacheFirst',
            options: {
              cacheName: 'cartodb-light-tiles-cache',
              expiration: {
                maxEntries: 500,
                maxAgeSeconds: 60 * 60 * 24 * 30
              },
              cacheableResponse: { statuses: [0, 200] }
            }
          },
          {
            urlPattern: /^https:\/\/(\w+\.)?basemaps\.cartocdn\.com\/dark_all\/.*\.png$/,
            handler: 'CacheFirst',
            options: {
              cacheName: 'cartodb-dark-tiles-cache',
              expiration: {
                maxEntries: 500,
                maxAgeSeconds: 60 * 60 * 24 * 30
              },
              cacheableResponse: { statuses: [0, 200] }
            }
          },
          {
            urlPattern: /^https:\/\/faas-sfo3-7872a1dd\.doserverless\.co\/api\/v1\/web\/fn-69328def-615c-4bce-88c0-dc912d5f1d84\/api\/(upcoming|past)$/,
            handler: 'NetworkFirst',
            options: {
              cacheName: 'bike-bae-api-cache',
              cacheableResponse: { statuses: [0, 200] }
            }
          }
        ]
      }
    })
  ],
  resolve: {
    alias: {
      $lib: path.resolve(__dirname, './src/lib')
    }
  }
});
