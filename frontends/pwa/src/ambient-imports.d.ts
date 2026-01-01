declare module '*.svelte' {
  export { SvelteComponent as default } from 'svelte';
}

// Type definition for PWA appinstalled event
declare global {
  interface WindowEventMap {
    appinstalled: Event;
  }
}

import type { SvelteWindowAttributes } from 'svelte/elements';

declare module 'svelte/elements' {
  interface SvelteWindowAttributes {
    onappinstalled?: ((event: Event) => any) | null;
  }
}
