<script lang="ts">
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
    Official Strava "Connect with Strava" Button
    Per Strava Brand Guidelines: https://developers.strava.com/guidelines/
    - Must not be modified, stretched, recolored, or animated
    - Button height: 48px @1x (this is the @2x version at 96px, scaled to 48px)
    - Must link to OAuth flow
  -->
  {#if isLoading}
    <!-- Loading state: show spinner overlay on button -->
    <button
      type="button"
      disabled
      class="relative w-full sm:w-auto opacity-60 cursor-not-allowed"
      aria-busy="true"
      aria-label="Connecting to Strava"
    >
      <img
        src="/btn_strava_connect_with_orange_x2.png"
        alt="Connect with Strava"
        class="h-12 w-auto"
      />
      <div class="absolute inset-0 flex items-center justify-center">
        <svg
          class="h-6 w-6 animate-spin text-white"
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
      </div>
      <span class="sr-only">Connecting to Strava...</span>
    </button>
  {:else}
    <!-- Standard state: official Strava button -->
    <button
      type="button"
      onclick={handleClick}
      disabled={disabled}
      class="w-full sm:w-auto transition-opacity hover:opacity-90 active:opacity-100 disabled:opacity-50 disabled:cursor-not-allowed"
      aria-label="Connect with Strava to import cycling events from your clubs"
    >
      <img
        src="/btn_strava_connect_with_orange_x2.png"
        alt="Connect with Strava"
        class="h-12 w-auto"
      />
    </button>
  {/if}

  {#if error}
    <p class="text-sm text-red-600 font-medium animate-in fade-in slide-in-from-top-1 duration-300">{error}</p>
  {/if}
</div>
