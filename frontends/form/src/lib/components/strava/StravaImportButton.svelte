<script lang="ts">
  import { Button } from "$lib/components/ui/button";
  import { initiateAuth } from "$lib/api/strava";

  interface Props {
    city: string;
    onAuthComplete: () => void;
    disabled?: boolean;
  }

  let { city, onAuthComplete, disabled = false }: Props = $props();

  let isLoading = $state(false);
  let error = $state<string | null>(null);

  async function handleClick() {
    if (isLoading || disabled) return;

    isLoading = true;
    error = null;

    try {
      await initiateAuth(city);
      onAuthComplete();
    } catch (err) {
      error = err instanceof Error ? err.message : "Failed to connect to Strava";
    } finally {
      isLoading = false;
    }
  }
</script>

<div class="inline-flex flex-col items-end gap-1">
  <!--
    Strava Branding Guidelines: https://developers.strava.com/guidelines/
    - Official Strava orange: #FC5200
    - Button links to OAuth flow
    - Compatible with Strava
  -->
  <Button
    variant="outline"
    class="border-[#FC5200] text-[#FC5200] hover:bg-[#FC5200] hover:text-white"
    onclick={handleClick}
    disabled={isLoading || disabled}
    aria-busy={isLoading}
    aria-label="Import cycling events from your Strava clubs"
  >
    {#if isLoading}
      <svg
        class="mr-2 h-4 w-4 animate-spin"
        fill="none"
        viewBox="0 0 24 24"
        aria-hidden="true"
      >
        <circle
          class="opacity-25"
          cx="12"
          cy="12"
          r="10"
          stroke="currentColor"
          stroke-width="4"
        />
        <path
          class="opacity-75"
          fill="currentColor"
          d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
        />
      </svg>
      <span class="sr-only">Connecting to Strava</span>
      Connecting...
    {:else}
      <!-- Strava wordmark logo -->
      <svg
        class="mr-2 h-4 w-4"
        viewBox="0 0 24 24"
        fill="currentColor"
        aria-hidden="true"
      >
        <path d="M15.387 17.944l-2.089-4.116h-3.065L15.387 24l5.15-10.172h-3.066m-7.008-5.599l2.836 5.598h4.172L10.463 0l-7 13.828h4.169" />
      </svg>
      Import from Strava
    {/if}
  </Button>

  {#if error}
    <p class="text-sm text-red-500">{error}</p>
  {/if}
</div>
