/// <reference types="svelte" />
/// <reference types="vite/client" />
/// <reference types="unplugin-icons/types/svelte" />

declare module '*.svelte' {
  import type { ComponentType } from 'svelte';

  const component: ComponentType;
  export default component;
}
