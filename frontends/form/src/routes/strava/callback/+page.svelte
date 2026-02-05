<script lang="ts">
  import { onMount } from "svelte";
  import { page } from "$app/stores";
  import { Button } from "$lib/components/ui/button";
  import { Check, CircleX } from "lucide-svelte";

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

<style>
  @keyframes slideIn {
    from {
      opacity: 0;
      transform: translateY(-8px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }

  @keyframes pulse {
    0%, 100% { opacity: 1; }
    50% { opacity: 0.5; }
  }

  .animate-slide-in {
    animation: slideIn 0.3s cubic-bezier(0.16, 1, 0.3, 1);
  }

  .animate-pulse-slow {
    animation: pulse 2s cubic-bezier(0.4, 0, 0.6, 1) infinite;
  }
</style>

<div class="min-h-screen bg-gradient-to-b from-slate-50 to-white dark:from-slate-950 dark:to-slate-900 flex items-center justify-center p-4">
  <div class="max-w-md w-full text-center animate-slide-in">
    {#if status === "closing"}
      <div class="mb-6 inline-flex items-center justify-center w-16 h-16 rounded-full bg-emerald-50 dark:bg-emerald-950 border border-emerald-200 dark:border-emerald-800">
        <Check class="w-8 h-8 text-emerald-600 dark:text-emerald-400" />
      </div>
      <h1 class="text-3xl sm:text-4xl font-bold tracking-tight text-slate-900 dark:text-slate-50 mb-3">
        Connected to Strava!
      </h1>
      <p class="text-lg text-slate-600 dark:text-slate-400 mb-6">
        This window will close automatically...
      </p>
      <div class="h-1 w-32 mx-auto bg-slate-200 dark:bg-slate-800 rounded-full overflow-hidden mb-6">
        <div class="h-full bg-emerald-500 animate-pulse-slow rounded-full"></div>
      </div>
      <Button variant="outline" onclick={handleClose} class="gap-2">
        <span>Close Now</span>
      </Button>
    {:else if status === "error"}
      <div class="mb-6 inline-flex items-center justify-center w-16 h-16 rounded-full bg-red-50 dark:bg-red-950 border border-red-200 dark:border-red-800">
        <CircleX class="w-8 h-8 text-red-600 dark:text-red-400" />
      </div>
      <h1 class="text-3xl sm:text-4xl font-bold tracking-tight text-slate-900 dark:text-slate-50 mb-3">
        Authorization Failed
      </h1>
      <div class="mb-6 p-4 bg-red-50 dark:bg-red-950 border-l-4 border-red-500 rounded-lg">
        <p class="text-sm text-red-900 dark:text-red-100">
          {errorMessage || "Unable to connect to Strava. Please try again."}
        </p>
      </div>
      <Button variant="outline" onclick={handleClose} class="gap-2">
        <CircleX class="h-4 w-4" />
        <span>Close</span>
      </Button>
    {:else}
      <div class="mb-6 inline-flex items-center justify-center w-16 h-16 rounded-full bg-emerald-50 dark:bg-emerald-950 border border-emerald-200 dark:border-emerald-800">
        <Check class="w-8 h-8 text-emerald-600 dark:text-emerald-400" />
      </div>
      <h1 class="text-3xl sm:text-4xl font-bold tracking-tight text-slate-900 dark:text-slate-50 mb-3">
        Authentication Successful!
      </h1>
      <p class="text-lg text-slate-600 dark:text-slate-400 mb-6">
        You can close this window now.
      </p>
      <Button variant="outline" onclick={handleClose} class="gap-2">
        <span>Close Window</span>
      </Button>
    {/if}
  </div>
</div>
