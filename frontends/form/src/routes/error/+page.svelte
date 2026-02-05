<script lang="ts">
  import { page } from "$app/state";
  import { Button } from "$lib/components/ui/button";
  import * as Card from "$lib/components/ui/card";
  import { CircleX, AlertTriangle, MapPin } from "lucide-svelte";

  const message = page.url.searchParams.get("message") || "An error occurred";

  const href = "https://cyclescene.cc";
</script>

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

  @keyframes fadeIn {
    from { opacity: 0; }
    to { opacity: 1; }
  }

  .animate-slide-in {
    animation: slideIn 0.3s cubic-bezier(0.16, 1, 0.3, 1);
  }

  .animate-fade-in {
    animation: fadeIn 0.2s ease-out;
  }
</style>

<div class="min-h-screen bg-gradient-to-b from-slate-50 to-white dark:from-slate-950 dark:to-slate-900">
  <div class="max-w-5xl mx-auto px-4 sm:px-6 lg:px-8 py-8 sm:py-12">

    <!-- Header -->
    <header class="mb-10 animate-slide-in">
      <div class="inline-flex items-center gap-2 text-sm font-medium text-slate-500 mb-3">
        <MapPin class="h-4 w-4" />
        <span class="uppercase tracking-wide">CycleScene</span>
      </div>

      <div class="text-center sm:text-left mb-8">
        <div class="inline-flex items-center justify-center w-16 h-16 rounded-full bg-red-50 dark:bg-red-950 border border-red-200 dark:border-red-800 mb-6">
          <CircleX class="w-8 h-8 text-red-600 dark:text-red-400" />
        </div>
        <h1 class="text-4xl sm:text-5xl font-bold tracking-tight text-slate-900 dark:text-slate-50 mb-4 leading-tight">
          Oops! Something Went Wrong
        </h1>
        <p class="text-lg text-slate-600 dark:text-slate-400">
          We couldn't process your request
        </p>
      </div>
    </header>

    <!-- Error Message -->
    <div class="mb-6 p-4 bg-red-50 dark:bg-red-950 border-l-4 border-red-500 rounded-lg shadow-sm animate-slide-in">
      <div class="flex items-center gap-3">
        <CircleX class="h-5 w-5 text-red-600 dark:text-red-400 flex-shrink-0" />
        <p class="text-sm font-medium text-red-900 dark:text-red-100">{message}</p>
      </div>
    </div>

    <!-- Main Content Card -->
    <div class="bg-white dark:bg-slate-900 rounded-2xl border border-slate-200 dark:border-slate-800 shadow-sm overflow-hidden animate-fade-in">
      <div class="px-6 py-5 border-b border-slate-200 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-900/50">
        <h2 class="text-xl font-semibold text-slate-900 dark:text-slate-50">What you can do</h2>
        <p class="text-sm text-slate-500 dark:text-slate-400 mt-0.5">Try these troubleshooting steps</p>
      </div>

      <div class="px-6 py-6">
        <ul class="space-y-4">
          <li class="flex items-start gap-3">
            <div class="h-6 w-6 rounded-full bg-slate-100 dark:bg-slate-800 flex items-center justify-center flex-shrink-0 mt-0.5">
              <span class="text-xs font-semibold text-slate-600 dark:text-slate-400">1</span>
            </div>
            <p class="text-sm text-slate-700 dark:text-slate-300 leading-relaxed">
              Make sure you clicked the "Host a Ride" button from the CycleScene app
            </p>
          </li>
          <li class="flex items-start gap-3">
            <div class="h-6 w-6 rounded-full bg-slate-100 dark:bg-slate-800 flex items-center justify-center flex-shrink-0 mt-0.5">
              <span class="text-xs font-semibold text-slate-600 dark:text-slate-400">2</span>
            </div>
            <p class="text-sm text-slate-700 dark:text-slate-300 leading-relaxed">
              Check that your link hasn't expired (tokens are valid for 30 minutes)
            </p>
          </li>
          <li class="flex items-start gap-3">
            <div class="h-6 w-6 rounded-full bg-slate-100 dark:bg-slate-800 flex items-center justify-center flex-shrink-0 mt-0.5">
              <span class="text-xs font-semibold text-slate-600 dark:text-slate-400">3</span>
            </div>
            <p class="text-sm text-slate-700 dark:text-slate-300 leading-relaxed">
              Try requesting a new submission link from the app
            </p>
          </li>
          <li class="flex items-start gap-3">
            <div class="h-6 w-6 rounded-full bg-slate-100 dark:bg-slate-800 flex items-center justify-center flex-shrink-0 mt-0.5">
              <span class="text-xs font-semibold text-slate-600 dark:text-slate-400">4</span>
            </div>
            <p class="text-sm text-slate-700 dark:text-slate-300 leading-relaxed">
              Contact support if the problem persists
            </p>
          </li>
        </ul>
      </div>
    </div>

    <!-- Footer Actions -->
    <div class="flex justify-center pt-8">
      <Button class="gap-2 bg-slate-900 hover:bg-slate-800 dark:bg-slate-50 dark:hover:bg-slate-200">
        <a {href} class="flex items-center gap-2">
          <span>← Back to CycleScene</span>
        </a>
      </Button>
    </div>

  </div>
</div>
