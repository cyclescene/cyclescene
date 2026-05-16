<script lang="ts">
  import Button from "$lib/components/ui/button/button.svelte";
  import {
    activeView,
    dismissInstallMessage,
    installInfo,
    installMessageDismissed,
    navigateTo,
    promptInstallApp,
    SUB_VIEW_INSTALL,
  } from "$lib/stores";
  import { ENABLE_INSTALL_PROMPT_V2 } from "$lib/config";
  import IconClose from "~icons/material-symbols/close-rounded";
  import IconDownload from "~icons/material-symbols/download-rounded";

  let shouldShow = $derived(
    ENABLE_INSTALL_PROMPT_V2 &&
      !$installMessageDismissed &&
      $installInfo.shouldShowInstallEntry &&
      $activeView !== SUB_VIEW_INSTALL,
  );

  async function handleInstallClick() {
    const prompted = await promptInstallApp();

    if (!prompted) {
      navigateTo(SUB_VIEW_INSTALL);
    }
  }
</script>

{#if shouldShow}
  <div
    class="fixed left-3 right-3 bottom-[calc(var(--footer-height)_+_env(safe-area-inset-bottom)_+_12px)] z-[1100]"
  >
    <div
      class="border bg-background text-foreground shadow-lg rounded-lg p-3 flex gap-3 items-center"
    >
      <IconDownload class="w-5 h-5 shrink-0 text-yellow-600" />
      <div class="min-w-0 grow">
        <p class="font-semibold text-sm leading-tight">{$installInfo.title}</p>
        <p class="text-xs text-muted-foreground leading-snug mt-0.5">
          {$installInfo.message}
        </p>
      </div>
      <Button size="sm" onclick={handleInstallClick}>
        {$installInfo.primaryActionLabel}
      </Button>
      <Button
        aria-label="Dismiss install message"
        size="icon"
        variant="ghost"
        class="h-8 w-8 shrink-0"
        onclick={dismissInstallMessage}
      >
        <IconClose class="w-4 h-4" />
      </Button>
    </div>
  </div>
{/if}
