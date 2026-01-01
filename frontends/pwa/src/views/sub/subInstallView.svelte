<script lang="ts">
  import * as Card from "$lib/components/ui/card";
  import ScrollArea from "$lib/components/ui/scroll-area/scroll-area.svelte";
  import IconShare from "~icons/material-symbols/ios-share";
  import IconAdd from "~icons/material-symbols/add-box-outline";
</script>

<div class="install-view">
  <ScrollArea class="scroll-wrapper">
    <div
      class="flex flex-col gap-6 p-5 pb-[calc(var(--footer-height)_+_env(safe-area-inset-bottom)_+_10px)]"
    >
      <!-- Header -->
      <div class="pt-2">
        <h1 class="text-2xl font-bold">Install Cycle Scene</h1>
        <p class="text-muted-foreground mt-1">
          Add to your home screen for quick access
        </p>
      </div>

      <!-- iOS Instructions -->
      <div class="space-y-5">
        <!-- Step 1 -->
        <div class="flex gap-4 items-center">
          <div
            class="flex-shrink-0 w-10 h-10 rounded-full bg-yellow-500 flex items-center justify-center text-white font-bold text-lg"
          >
            1
          </div>
          <p class="font-semibold text-base flex items-center gap-2">
            <IconShare class="w-5 h-5" />
            Click the Share button
          </p>
        </div>

        <!-- Step 2 -->
        <div class="flex gap-4 items-center">
          <div
            class="flex-shrink-0 w-10 h-10 rounded-full bg-yellow-500 flex items-center justify-center text-white font-bold text-lg"
          >
            2
          </div>
          <p class="font-semibold text-base">
            Scroll and click "Add to Home Screen"
          </p>
        </div>

        <!-- Step 3 -->
        <div class="flex gap-4 items-center">
          <div
            class="flex-shrink-0 w-10 h-10 rounded-full bg-yellow-500 flex items-center justify-center text-white font-bold text-lg"
          >
            3
          </div>
          <p class="font-semibold text-base flex items-center gap-2">
            <IconAdd class="w-5 h-5" />
            Click Add
          </p>
        </div>
      </div>

      <!-- Benefits Card -->
      <Card.Root
        class="p-4 bg-green-50 dark:bg-green-950 border-green-200 dark:border-green-800 mt-4"
      >
        <Card.Header class="p-0 mb-3">
          <Card.Title class="text-base">Benefits</Card.Title>
        </Card.Header>
        <ul class="text-sm space-y-2 text-foreground">
          <li>✓ One-tap access from home screen</li>
          <li>✓ Works offline</li>
          <li>✓ App-like experience</li>
        </ul>
      </Card.Root>
    </div>
  </ScrollArea>
</div>

<style>
  .install-view {
    height: 100%;
    width: 100%;
    display: flex;
    flex-direction: column;
    overflow: hidden;
  }

  :global(.scroll-wrapper) {
    height: 100%;
    width: 100%;
  }
</style>
