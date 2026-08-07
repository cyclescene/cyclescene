import tailwindcss from '@tailwindcss/vite';
import { defineConfig } from 'vite';
import { svelte } from '@sveltejs/vite-plugin-svelte';
import { VitePWA } from 'vite-plugin-pwa';
import Icons from "unplugin-icons/vite"
import path from "path"
import { loadEnv } from 'vite';

// City configurations for PWA manifest
const cityConfigs = {
  pdx: {
    name: "Cycle Scene - PDX",
    short_name: "CycleScenePDX",
    description: "Upcoming bike rides in Portland, Oregon",
    keywords: "bike rides, cycling events, Portland bikes, group rides, Cycle Scene, Shift2Bikes, Pedalpalooza",
    url: "https://pdx.cyclescene.cc"
  },
  slc: {
    name: "Cycle Scene - SLC",
    short_name: "CycleSceneSLC",
    description: "Upcoming bike rides in Salt Lake City, Utah",
    keywords: "bike rides, cycling events, Salt Lake City bikes, SLC cycling, group rides, Cycle Scene",
    url: "https://slc.cyclescene.cc"
  },
  la: {
    name: "Cycle Scene - LA",
    short_name: "CycleSceneLA",
    description: "Upcoming bike rides in Los Angeles, California",
    keywords: "bike rides, cycling events, Los Angeles bikes, LA cycling, group rides, Cycle Scene",
    url: "https://la.cyclescene.cc"
  }
};



// https://vite.dev/config/
export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '');

  // Get city code from environment variable, default to pdx for development
  const cityCode = env.VITE_CITY_CODE || 'pdx';
  const cityConfig = cityConfigs[cityCode] || cityConfigs.pdx;
  const htmlMetadata = {
    '%CITY_APP_NAME%': cityConfig.name,
    '%CITY_SHORT_NAME%': cityConfig.short_name,
    '%CITY_DESCRIPTION%': cityConfig.description,
    '%CITY_KEYWORDS%': cityConfig.keywords,
    '%CITY_URL%': cityConfig.url,
  };

  return {

    server: {
      allowedHosts: true

    },
    plugins: [
      {
        name: 'city-html-metadata',
        transformIndexHtml(html) {
          return Object.entries(htmlMetadata).reduce(
            (updatedHtml, [placeholder, value]) => updatedHtml.replaceAll(placeholder, value),
            html,
          );
        },
      },
      tailwindcss(),
      svelte(),
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
          "viewport-fit": "cover",
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
      })
    ],
    resolve: {
      alias: {
        $lib: path.resolve(__dirname, './src/lib')
      }
    },
    build: {
      rollupOptions: {
        output: {
          manualChunks: {
            maplibre: ['maplibre-gl', 'svelte-maplibre-gl'],
            charts: ['layerchart', 'd3-shape'],
            workbox: ['workbox-window'],
            icons: ['@lucide/svelte']
          }
        }
      }
    }
  }
});
