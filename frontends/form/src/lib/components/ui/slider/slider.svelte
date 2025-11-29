<script lang="ts">
  import { cn } from "$lib/utils";

  interface Props {
    value: number;
    min?: number;
    max?: number;
    step?: number;
    class?: string;
    disabled?: boolean;
    oninput?: (e: Event) => void;
  }

  let {
    value = $bindable(0),
    min = 0,
    max = 100,
    step = 1,
    class: className,
    disabled = false,
    oninput,
  }: Props = $props();
</script>

<input
  type="range"
  bind:value
  {min}
  {max}
  {step}
  {disabled}
  {oninput}
  class={cn(
    "w-full h-2 bg-secondary rounded-lg appearance-none cursor-pointer disabled:cursor-not-allowed disabled:opacity-50",
    "accent-primary hover:accent-primary/90",
    className
  )}
/>
