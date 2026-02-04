<script lang="ts">
  import * as Card from "$lib/components/ui/card";
  import * as Alert from "$lib/components/ui/alert";
  import { Badge } from "$lib/components/ui/badge";
  import { Button } from "$lib/components/ui/button";
  import type { ImportResult } from "$lib/types/strava";

  interface Props {
    results: ImportResult[];
    organizerEmail: string;
    onImportMore: () => void;
    onDone: () => void;
  }

  let { results, organizerEmail, onImportMore, onDone }: Props = $props();

  // Count successes and failures
  let successCount = $derived(results.filter((r) => r.success).length);
  let failureCount = $derived(results.filter((r) => !r.success).length);

  // Check if all succeeded
  let allSucceeded = $derived(failureCount === 0);

  // Button state
  let isImportingMore = $state(false);

  function handleImportMoreClick() {
    // Prevent double-click
    if (isImportingMore) return;

    isImportingMore = true;
    onImportMore();

    // Re-enable after brief delay (button will unmount anyway when navigating)
    setTimeout(() => {
      isImportingMore = false;
    }, 1000);
  }
</script>

<div class="mx-auto max-w-2xl space-y-6">
  <!-- Summary Alert -->
  {#if allSucceeded}
    <div class="rounded-xl border-2 border-green-500 bg-green-50 p-6 shadow-sm animate-in fade-in duration-300" role="alert">
      <div class="flex items-start gap-4">
        <!-- Success icon -->
        <div class="flex h-12 w-12 flex-shrink-0 items-center justify-center rounded-full bg-green-500">
          <svg
            class="h-6 w-6 text-white"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
            aria-hidden="true"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2.5"
              d="M5 13l4 4L19 7"
            />
          </svg>
        </div>

        <!-- Content -->
        <div class="flex-1 space-y-2">
          <Alert.Title id="results-heading" class="text-xl font-bold text-green-800" tabindex={-1}>
            Import Complete
          </Alert.Title>
          <Alert.Description class="text-sm text-green-700">
            Successfully imported {successCount} event{successCount !== 1 ? "s" : ""} from Strava. Confirmation email sent to <strong>{organizerEmail}</strong>
          </Alert.Description>
        </div>
      </div>
    </div>
  {:else if successCount > 0}
    <div class="rounded-xl border-2 border-yellow-400 bg-yellow-50 p-6 shadow-sm animate-in fade-in duration-300" role="alert">
      <div class="flex items-start gap-4">
        <div class="flex h-12 w-12 flex-shrink-0 items-center justify-center rounded-full bg-yellow-500">
          <svg class="h-6 w-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
          </svg>
        </div>
        <div class="flex-1 space-y-2">
          <Alert.Title id="results-heading" class="text-xl font-bold text-yellow-900" tabindex={-1}>
            Partial Success
          </Alert.Title>
          <Alert.Description class="text-sm text-yellow-800">
            {successCount} event{successCount !== 1 ? "s" : ""} imported successfully, {failureCount} failed. Confirmation email sent to {organizerEmail}
          </Alert.Description>
        </div>
      </div>
    </div>
  {:else}
    <div class="rounded-xl border-2 border-red-500 bg-red-50 p-6 shadow-sm animate-in fade-in duration-300" role="alert">
      <div class="flex items-start gap-4">
        <div class="flex h-12 w-12 flex-shrink-0 items-center justify-center rounded-full bg-red-500">
          <svg class="h-6 w-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </div>
        <div class="flex-1 space-y-2">
          <Alert.Title id="results-heading" class="text-xl font-bold text-red-900" tabindex={-1}>Import Failed</Alert.Title>
          <Alert.Description class="text-sm text-red-800">
            All {failureCount} event{failureCount !== 1 ? "s" : ""} failed to import. Please check the errors below and try again.
          </Alert.Description>
        </div>
      </div>
    </div>
  {/if}

  <!-- Results List -->
  <Card.Root class="border-2 shadow-md">
    <div class="p-6">
      <h3 class="text-lg font-bold mb-4 flex items-center gap-2">
        <svg class="h-5 w-5 text-[#FC5200]" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-3 7h3m-3 4h3m-6-4h.01M9 16h.01" />
        </svg>
        Event Summary
      </h3>
      <div class="space-y-3">
        {#each results as result, i (result.strava_event_id)}
          <div class="flex items-center justify-between rounded-lg border-2 p-4 transition-all duration-200 {result.success ? 'border-green-200 bg-green-50/50 hover:bg-green-50' : 'border-red-200 bg-red-50/50'} animate-in fade-in slide-in-from-bottom-2" style="animation-delay: {i * 80}ms">
            <div class="flex items-center gap-3 flex-1 min-w-0">
              <!-- Status badge -->
              <div class="flex h-9 w-9 flex-shrink-0 items-center justify-center rounded-lg {result.success ? 'bg-green-500' : 'bg-red-500'} shadow-sm">
                {#if result.success}
                  <svg class="h-5 w-5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="3" d="M5 13l4 4L19 7" />
                  </svg>
                {:else}
                  <svg class="h-5 w-5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="3" d="M6 18L18 6M6 6l12 12" />
                  </svg>
                {/if}
              </div>
              <div class="min-w-0 flex-1">
                <span class="font-semibold text-sm block truncate {result.success ? 'text-green-900' : 'text-red-900'}">{result.title}</span>
                {#if !result.success && result.error}
                  <p class="text-xs text-red-700 mt-1 font-medium">{result.error}</p>
                {/if}
              </div>
            </div>

            {#if result.success && result.edit_url}
              <Button
                variant="outline"
                size="sm"
                href={result.edit_url}
                target="_blank"
                rel="noopener noreferrer"
                class="flex-shrink-0 ml-3 border-green-300 text-green-700 hover:bg-green-100 hover:border-green-400"
              >
                Edit
                <svg class="h-3.5 w-3.5 ml-1.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14" />
                </svg>
              </Button>
            {/if}
          </div>
        {/each}
      </div>
    </div>
  </Card.Root>

  <!-- Footer note -->
  <div class="flex items-start gap-3 rounded-lg border bg-muted/30 p-4">
    <svg class="h-5 w-5 flex-shrink-0 mt-0.5 text-[#FC5200]" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
    </svg>
    <p class="text-sm text-muted-foreground flex-1">
      <strong class="font-semibold text-foreground">Pro tip:</strong> You can update event details anytime using the magic edit links sent to your email.
    </p>
  </div>

  <!-- Action Buttons -->
  <div class="flex flex-col sm:flex-row justify-center gap-3">
    <Button
      variant="outline"
      size="lg"
      onclick={handleImportMoreClick}
      disabled={isImportingMore}
      aria-label="Import additional events from Strava"
      class="group border-2 w-full sm:w-auto"
    >
      <svg class="h-5 w-5 mr-2 group-hover:-translate-x-0.5 transition-transform" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
      </svg>
      Import More Events
    </Button>
    <Button
      size="lg"
      onclick={onDone}
      aria-label="Close import and return to form"
      class="bg-[#FC5200] hover:bg-[#E04A00] text-white shadow-sm hover:shadow-md transition-all duration-200 group w-full sm:w-auto"
    >
      Done
      <svg class="h-5 w-5 ml-2 group-hover:translate-x-0.5 transition-transform" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
      </svg>
    </Button>
  </div>
</div>
