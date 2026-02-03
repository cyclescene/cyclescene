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
    <Alert.Root class="border-green-200 bg-green-50" role="alert">
      <svg
        class="h-5 w-5 text-green-600"
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
      <Alert.Title id="results-heading" class="text-green-800" tabindex={-1}>
        Import Complete!
      </Alert.Title>
      <Alert.Description class="text-green-700">
        {successCount} event{successCount !== 1 ? "s" : ""} imported successfully.
        A confirmation email has been sent to {organizerEmail}.
      </Alert.Description>
    </Alert.Root>
  {:else if successCount > 0}
    <Alert.Root class="border-yellow-200 bg-yellow-50" role="alert">
      <svg
        class="h-5 w-5 text-yellow-600"
        fill="none"
        stroke="currentColor"
        viewBox="0 0 24 24"
      >
        <path
          stroke-linecap="round"
          stroke-linejoin="round"
          stroke-width="2"
          d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"
        />
      </svg>
      <Alert.Title id="results-heading" class="text-yellow-800" tabindex={-1}>
        Partial Success
      </Alert.Title>
      <Alert.Description class="text-yellow-700">
        {successCount} event{successCount !== 1 ? "s" : ""} imported, {failureCount} failed.
        A confirmation email has been sent to {organizerEmail}.
      </Alert.Description>
    </Alert.Root>
  {:else}
    <Alert.Root variant="destructive" role="alert">
      <svg
        class="h-5 w-5"
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
      <Alert.Title id="results-heading" tabindex={-1}>Import Failed</Alert.Title>
      <Alert.Description>
        All {failureCount} event{failureCount !== 1 ? "s" : ""} failed to import.
        Please check the errors below and try again.
      </Alert.Description>
    </Alert.Root>
  {/if}

  <!-- Results List -->
  <Card.Root class="p-4">
    <div class="divide-y">
      {#each results as result (result.strava_event_id)}
        <div class="flex items-center justify-between py-3 first:pt-0 last:pb-0">
          <div class="flex items-center gap-3">
            <Badge variant={result.success ? "default" : "destructive"}>
              {result.success ? "✓" : "✗"}
            </Badge>
            <div>
              <span class="font-medium">{result.title}</span>
              {#if !result.success && result.error}
                <p class="text-sm text-red-500">{result.error}</p>
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
            >
              Edit →
            </Button>
          {/if}
        </div>
      {/each}
    </div>
  </Card.Root>

  <!-- Footer note -->
  <p class="text-muted-foreground text-center text-sm">
    You can update event details anytime using the edit links in your confirmation email.
  </p>

  <!-- Action Buttons -->
  <div class="flex justify-center gap-3">
    <Button
      variant="outline"
      onclick={handleImportMoreClick}
      disabled={isImportingMore}
      aria-label="Import additional events from Strava"
    >
      Import More Events
    </Button>
    <Button onclick={onDone} aria-label="Close import and return to form">
      Done
    </Button>
  </div>
</div>
