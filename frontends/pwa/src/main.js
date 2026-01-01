import { mount } from 'svelte'
import './app.css'
import App from './App.svelte'
import { registerSW } from 'virtual:pwa-register'

const app = mount(App, {
  target: document.getElementById('app'),
})

// Register service worker for PWA
registerSW({
  onNeedRefresh() {
    console.log('PWA: Update ready')
  },
  onOfflineReady() {
    console.log('PWA: Ready for offline use')
  },
})

export default app
