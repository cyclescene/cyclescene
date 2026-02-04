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

<div class="flex flex-col items-stretch sm:items-end gap-2 w-full sm:w-auto">
  <!--
    Strava Branding Guidelines: https://developers.strava.com/guidelines/
    - Official Strava orange: #FC5200
    - Button links to OAuth flow
    - Compatible with Strava
  -->
  <Button
    variant="outline"
    size="lg"
    class="group relative overflow-hidden border-2 border-[#FC5200] bg-white text-[#FC5200] font-semibold shadow-sm hover:shadow-md hover:border-[#FC5200] hover:-translate-y-0.5 active:translate-y-0 transition-all duration-200 disabled:hover:translate-y-0 disabled:opacity-60 w-full sm:w-auto"
    onclick={handleClick}
    disabled={isLoading || disabled}
    aria-busy={isLoading}
    aria-label="Import cycling events from your Strava clubs"
  >
    <!-- Animated background on hover -->
    <span class="absolute inset-0 bg-[#FC5200] transform scale-x-0 group-hover:scale-x-100 transition-transform duration-300 origin-left" aria-hidden="true"></span>

    <span class="relative flex items-center gap-2 group-hover:text-white transition-colors duration-300">
      {#if isLoading}
        <svg
          class="h-5 w-5 animate-spin"
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
            stroke-width="3"
          />
          <path
            class="opacity-75"
            fill="currentColor"
            d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
          />
        </svg>
        <span class="sr-only">Connecting to Strava</span>
        <span>Connecting...</span>
      {:else}
        <!-- Strava wordmark logo with subtle pulse -->
        <svg
          class="h-5 w-5 group-hover:scale-110 transition-transform duration-300"
          viewBox="0 0 24 24"
          fill="currentColor"
          aria-hidden="true"
        >
          <path d="M15.387 17.944l-2.089-4.116h-3.065L15.387 24l5.15-10.172h-3.066m-7.008-5.599l2.836 5.598h4.172L10.463 0l-7 13.828h4.169" />
        </svg>
        <span>Import from Strava</span>
        <!-- Animated arrow -->
        <svg
          class="h-4 w-4 group-hover:translate-x-1 transition-transform duration-300"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
          aria-hidden="true"
        >
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M13 7l5 5m0 0l-5 5m5-5H6" />
        </svg>
      {/if}
    </span>
  </Button>

  {#if error}
    <p class="text-sm text-red-600 font-medium animate-in fade-in slide-in-from-top-1 duration-300">{error}</p>
  {/if}
</div>
