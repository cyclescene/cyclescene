<script lang="ts">
  import * as Card from "$lib/components/ui/card";
  import ScrollArea from "$lib/components/ui/scroll-area/scroll-area.svelte";
  import IconShare from "~icons/material-symbols/ios-share";
  import IconAdd from "~icons/material-symbols/add-box-outline";
</script>

<div class="install-view">
  <ScrollArea class="scroll-wrapper">
    <div
      class="flex flex-col gap-4 p-5 pb-[calc(var(--footer-height)_+_env(safe-area-inset-bottom)_+_10px)]"
    >
      <!-- Header Card -->
      <Card.Root class="p-4">
        <Card.Header class="p-0">
          <Card.Title>Install Cycle Scene</Card.Title>
          <Card.Description
            >Add Cycle Scene to your home screen for quick access</Card.Description
          >
        </Card.Header>
      </Card.Root>

      <!-- iOS Instructions Card -->
      <Card.Root class="p-4">
        <Card.Header class="p-0 mb-4">
          <Card.Title class="text-base">For iPhone & iPad</Card.Title>
        </Card.Header>

        <div class="space-y-4">
          <!-- Step 1 -->
          <div class="flex gap-3 items-start">
            <div
              class="flex-shrink-0 w-8 h-8 rounded-full bg-yellow-500 flex items-center justify-center text-white font-bold text-sm"
            >
              1
            </div>
            <div>
              <p class="font-semibold text-sm">Tap the Share button</p>
              <p
                class="text-xs text-muted-foreground flex items-center gap-1 mt-1"
              >
                Look for the share icon
                <IconShare class="inline w-4 h-4" />
                at the bottom of Safari
              </p>
            </div>
          </div>

          <!-- Step 2 -->
          <div class="flex gap-3 items-start">
            <div
              class="flex-shrink-0 w-8 h-8 rounded-full bg-yellow-500 flex items-center justify-center text-white font-bold text-sm"
            >
              2
            </div>
            <div>
              <p class="font-semibold text-sm">
                Scroll and select "Add to Home Screen"
              </p>
              <p class="text-xs text-muted-foreground mt-1">
                Swipe through the action menu to find this option
              </p>
            </div>
          </div>

          <!-- Step 3 -->
          <div class="flex gap-3 items-start">
            <div
              class="flex-shrink-0 w-8 h-8 rounded-full bg-yellow-500 flex items-center justify-center text-white font-bold text-sm"
            >
              3
            </div>
            <div>
              <p class="font-semibold text-sm">Confirm and save</p>
              <p
                class="text-xs text-muted-foreground flex items-center gap-1 mt-1"
              >
                Tap
                <IconAdd class="inline w-4 h-4" />
                "Add" to add Cycle Scene to your home screen
              </p>
            </div>
          </div>
        </div>
      </Card.Root>

      <!-- Benefits Card -->
      <Card.Root
        class="p-4 bg-green-50 dark:bg-green-950 border-green-200 dark:border-green-800"
      >
        <Card.Header class="p-0 mb-2">
          <Card.Title class="text-base">Benefits of Installing</Card.Title>
        </Card.Header>
        <ul class="text-sm space-y-2 text-foreground">
          <li>✓ One-tap access from your home screen</li>
          <li>✓ Works offline with saved ride data</li>
          <li>✓ Faster loading than the browser</li>
          <li>✓ Native app-like experience</li>
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
