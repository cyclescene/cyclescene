<script lang="ts">
  import { onMount } from "svelte";
  import { page } from "$app/stores";
  import { Button } from "$lib/components/ui/button";

  let status: "success" | "error" | "closing" = "success";
  let errorMessage = "";

  onMount(() => {
    // Check for error in URL params (from backend redirect)
    const error = $page.url.searchParams.get("error");

    if (error) {
      status = "error";
      errorMessage = decodeURIComponent(error);

      // Notify parent of error
      if (window.opener) {
        window.opener.postMessage(
          { type: "strava-auth-error", error: errorMessage },
          "*"
        );
      }
    } else {
      status = "closing";

      // Notify parent of success
      if (window.opener) {
        window.opener.postMessage({ type: "strava-auth-complete" }, "*");

        // Auto-close after short delay
        setTimeout(() => {
          window.close();
        }, 1500);
      }
    }
  });

  function handleClose() {
    window.close();
  }
</script>

<svelte:head>
  <title>Strava Authorization</title>
</svelte:head>

<div class="flex min-h-screen items-center justify-center p-4">
  <div class="max-w-md text-center">
    {#if status === "closing"}
      <div class="mb-4">
        <svg
          class="mx-auto h-16 w-16 text-green-500"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M5 13l4 4L19 7"
          />
        </svg>
      </div>
      <h1 class="mb-2 text-2xl font-semibold">Connected to Strava!</h1>
      <p class="text-muted-foreground mb-4">
        This window will close automatically...
      </p>
      <Button variant="outline" onclick={handleClose}>Close Now</Button>
    {:else if status === "error"}
      <div class="mb-4">
        <svg
          class="mx-auto h-16 w-16 text-red-500"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M6 18L18 6M6 6l12 12"
          />
        </svg>
      </div>
      <h1 class="mb-2 text-2xl font-semibold">Authorization Failed</h1>
      <p class="text-muted-foreground mb-4">
        {errorMessage || "Unable to connect to Strava. Please try again."}
      </p>
      <Button variant="outline" onclick={handleClose}>Close</Button>
    {:else}
      <div class="mb-4">
        <svg
          class="mx-auto h-16 w-16 text-green-500"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M5 13l4 4L19 7"
          />
        </svg>
      </div>
      <h1 class="mb-2 text-2xl font-semibold">Authentication Successful!</h1>
      <p class="text-muted-foreground mb-4">
        You can close this window now.
      </p>
      <Button variant="outline" onclick={handleClose}>Close Window</Button>
    {/if}
  </div>
</div>
