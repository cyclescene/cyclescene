import { mount } from 'svelte'
import 'maplibre-gl/dist/maplibre-gl.css'
import './app.css'
import App from './App.svelte'
import { registerSW } from 'virtual:pwa-register'
import { installPromptEvent } from './lib/stores.js'

const app = mount(App, {
  target: document.getElementById('app'),
})

// Register service worker for PWA
registerSW({
  immediate: true,
  onNeedRefresh() {
    console.log('PWA: Update ready')
  },
  onOfflineReady() {
    console.log('PWA: Ready for offline use')
  },
  onRegisterError(error) {
    console.error('PWA register error:', error)
  }
})

// Listen for the beforeinstallprompt event
window.addEventListener('beforeinstallprompt', (event) => {
  console.log('[main.js] beforeinstallprompt fired!')
  event.preventDefault()
  installPromptEvent.set(event)
})

export default app
