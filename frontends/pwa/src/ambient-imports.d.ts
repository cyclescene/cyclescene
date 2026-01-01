import type { ComponentType } from 'svelte';

declare module '*.svelte' {
  const component: ComponentType;
  export default component;
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
