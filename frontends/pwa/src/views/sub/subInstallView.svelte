<script lang="ts">
  import Button from "$lib/components/ui/button/button.svelte";
  import * as Card from "$lib/components/ui/card";
  import ScrollArea from "$lib/components/ui/scroll-area/scroll-area.svelte";
  import { ENABLE_INSTALL_PROMPT_V2 } from "$lib/config";
  import { installInfo, promptInstallApp } from "$lib/stores";
  import IconShare from "~icons/material-symbols/ios-share";
  import IconAdd from "~icons/material-symbols/add-box-outline";
  import IconDownload from "~icons/material-symbols/download-rounded";
  import IconMore from "~icons/material-symbols/more-vert";

  async function handleInstall() {
    await promptInstallApp();
  }
</script>

<div class="install-view">
  <ScrollArea class="scroll-wrapper">
    <div
      class="flex flex-col gap-6 p-5 pb-[calc(var(--footer-height)_+_env(safe-area-inset-bottom)_+_10px)]"
    >
      {#if ENABLE_INSTALL_PROMPT_V2}
        <div class="pt-2">
          <h1 class="text-2xl font-bold">{$installInfo.title}</h1>
          <p class="text-muted-foreground mt-1">{$installInfo.message}</p>
        </div>

        {#if $installInfo.canUseNativePrompt}
          <Button class="w-full" size="lg" onclick={handleInstall}>
            <IconDownload class="w-5 h-5" />
            {$installInfo.primaryActionLabel}
          </Button>
        {:else if $installInfo.platform === "installed"}
          <Card.Root class="p-4 bg-green-50 dark:bg-green-950 border-green-200 dark:border-green-800">
            <Card.Title class="text-base">Installed</Card.Title>
          </Card.Root>
        {:else if $installInfo.platform === "ios"}
          <div class="space-y-5">
            <div class="flex gap-4 items-center">
              <div
                class="flex-shrink-0 w-10 h-10 rounded-full bg-yellow-500 flex items-center justify-center text-white font-bold text-lg"
              >
                1
              </div>
              <p class="font-semibold text-base flex items-center gap-2">
                <IconShare class="w-5 h-5" />
                Tap Share
              </p>
            </div>

            <div class="flex gap-4 items-center">
              <div
                class="flex-shrink-0 w-10 h-10 rounded-full bg-yellow-500 flex items-center justify-center text-white font-bold text-lg"
              >
                2
              </div>
              <p class="font-semibold text-base flex items-center gap-2">
                <IconAdd class="w-5 h-5" />
                Tap Add to Home Screen
              </p>
            </div>

            <div class="flex gap-4 items-center">
              <div
                class="flex-shrink-0 w-10 h-10 rounded-full bg-yellow-500 flex items-center justify-center text-white font-bold text-lg"
              >
                3
              </div>
              <p class="font-semibold text-base">
                Tap <span class="font-bold text-blue-400">Add</span>
              </p>
            </div>
          </div>
        {:else}
          <div class="space-y-5">
            <div class="flex gap-4 items-center">
              <div
                class="flex-shrink-0 w-10 h-10 rounded-full bg-yellow-500 flex items-center justify-center text-white font-bold text-lg"
              >
                1
              </div>
              <p class="font-semibold text-base flex items-center gap-2">
                <IconMore class="w-5 h-5" />
                Open your browser menu
              </p>
            </div>

            <div class="flex gap-4 items-center">
              <div
                class="flex-shrink-0 w-10 h-10 rounded-full bg-yellow-500 flex items-center justify-center text-white font-bold text-lg"
              >
                2
              </div>
              <p class="font-semibold text-base flex items-center gap-2">
                <IconDownload class="w-5 h-5" />
                Choose Install app or Add to Home Screen
              </p>
            </div>
          </div>
        {/if}

        <Card.Root
          class="p-4 bg-green-50 dark:bg-green-950 border-green-200 dark:border-green-800 mt-4"
        >
          <Card.Header class="p-0 mb-0.5">
            <Card.Title class="text-base">Benefits</Card.Title>
            <ul class="text-sm space-y-2 text-foreground">
              <li>One-tap access from home screen</li>
              <li>Works offline</li>
              <li>App-like experience</li>
            </ul>
          </Card.Header>
        </Card.Root>
      {:else}
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
            <p class="font-semibold text-base flex items-center gap-2">
              <IconAdd class="w-5 h-5" />
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
            <p class="font-semibold text-base">
              Click <span class="font-bold text-blue-400">Add</span>
            </p>
          </div>
        </div>

        <!-- Benefits Card -->
        <Card.Root
          class="p-4 bg-green-50 dark:bg-green-950 border-green-200 dark:border-green-800 mt-4"
        >
          <Card.Header class="p-0 mb-0.5">
            <Card.Title class="text-base">Benefits</Card.Title>
            <ul class="text-sm space-y-2 text-foreground">
              <li>✓ One-tap access from home screen</li>
              <li>✓ Works offline</li>
              <li>✓ App-like experience</li>
            </ul>
          </Card.Header>
        </Card.Root>
      {/if}
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
