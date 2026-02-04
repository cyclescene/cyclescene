<script lang="ts">
  import { onMount, onDestroy } from "svelte";
  import { SvelteMap } from "svelte/reactivity";
  import * as Card from "$lib/components/ui/card";
  import { Progress } from "$lib/components/ui/progress";
  import { Badge } from "$lib/components/ui/badge";
  import { Button } from "$lib/components/ui/button";
  import { PUBLIC_STRAVA_DEBUG } from "$env/static/public";
  import {
    StravaImportWebSocket,
    type WebSocketState,
  } from "$lib/utils/strava-websocket";
  import type {
    EventImportConfig,
    ProgressMessage,
    ImportResult,
    ProgressStep,
  } from "$lib/types/strava";

  interface Props {
    organizerEmail: string;
    groupCode?: string; // Optional group code to associate all imported rides with
    events: EventImportConfig[];
    eventTitles: Map<string, string>; // eventId -> title
    onComplete: (results: ImportResult[]) => void;
    onError: (error: string) => void;
  }

  let {
    organizerEmail,
    groupCode = "",
    events,
    eventTitles,
    onComplete,
    onError,
  }: Props = $props();

  // WebSocket instance
  let ws: StravaImportWebSocket | null = null;

  // Progress tracking
  let wsState = $state<WebSocketState>("disconnected");
  let progress = new SvelteMap<string, ProgressMessage>();
  let completedCount = $state(0);
  let reconnectAttempts = $state({ current: 0, max: 3 });
  let canRetry = $state(false); // Show retry button when max reconnects reached
  let isRetrying = $state(false); // Disable retry button during retry

  // Steps in order
  const STEPS: ProgressStep[] = ["fetching", "coordinates", "route", "database"];

  // Debug logging
  function debugLog(message: string, data?: unknown) {
    if (PUBLIC_STRAVA_DEBUG === "true") {
      console.log(`[Strava ImportProgress] ${message}`, data ?? "");
    }
  }

  // Calculate overall progress percentage
  function getOverallProgress(): number {
    if (events.length === 0) return 0;
    return Math.round((completedCount / events.length) * 100);
  }

  // Get step status for an event
  function getStepStatus(eventId: string, step: ProgressStep): "pending" | "in_progress" | "success" | "error" {
    const msg = progress.get(eventId);
    if (!msg) return "pending";

    // If event completed (success or error), all steps are done
    if (msg.type === "complete") {
      return msg.success ? "success" : "error";
    }

    // If this is the current step
    if (msg.step === step) {
      return msg.status === "success" ? "success" : "in_progress";
    }

    // Check if this step has been passed
    const currentStepIndex = msg.step ? STEPS.indexOf(msg.step) : -1;
    const thisStepIndex = STEPS.indexOf(step);

    if (thisStepIndex < currentStepIndex) {
      return "success";
    }

    return "pending";
  }

  // Get badge variant for step status
  function getStepBadgeVariant(status: "pending" | "in_progress" | "success" | "error"): "default" | "secondary" | "destructive" | "outline" {
    switch (status) {
      case "success":
        return "default";
      case "error":
        return "destructive";
      case "in_progress":
        return "secondary";
      default:
        return "outline";
    }
  }

  // Get event status icon
  function getEventIcon(eventId: string): string {
    const msg = progress.get(eventId);
    if (!msg) return "⏳";
    if (msg.type === "complete") {
      return msg.success ? "✓" : "✗";
    }
    return "⟳";
  }

  // Get event badge variant
  function getEventBadgeVariant(eventId: string): "default" | "secondary" | "destructive" {
    const msg = progress.get(eventId);
    if (!msg) return "secondary";
    if (msg.type === "complete") {
      return msg.success ? "default" : "destructive";
    }
    return "secondary";
  }

  // Get event title
  function getEventTitle(eventId: string): string {
    return eventTitles.get(eventId) ?? progress.get(eventId)?.event_title ?? "Unknown Event";
  }

  // Handle progress updates
  function handleProgress(message: ProgressMessage) {
    debugLog("Progress update", message);

    if (message.strava_event_id) {
      progress.set(message.strava_event_id, message);

      // Track completed events
      if (message.type === "complete") {
        completedCount++;
      }
    }
  }

  // Handle completion
  function handleComplete(results: ImportResult[]) {
    debugLog("Import complete", results);
    onComplete(results);
  }

  // Handle errors
  function handleError(error: string) {
    debugLog("Import error", error);
    onError(error);
  }

  // Handle state changes
  function handleStateChange(state: WebSocketState) {
    debugLog("WebSocket state change", state);
    wsState = state;

    // Update reconnect attempts
    if (ws) {
      reconnectAttempts = ws.getReconnectAttempts();
    }

    // Show retry button if disconnected after max reconnects and not completed
    if (state === "disconnected" && reconnectAttempts.current >= reconnectAttempts.max && completedCount < events.length) {
      canRetry = true;
    }
  }

  // Handle stop button click
  function handleStop() {
    debugLog("User stopped import");
    if (ws) {
      ws.stop();
    }
    onError("Import stopped by user");
  }

  // Handle retry button click
  function handleRetry() {
    debugLog("User initiated manual retry");
    isRetrying = true;
    canRetry = false;

    if (ws) {
      ws.manualRetry();
    }

    // Re-enable retry button after a delay
    setTimeout(() => {
      isRetrying = false;
    }, 2000);
  }

  // Start import on mount
  onMount(() => {
    debugLog("Starting import", { email: organizerEmail, groupCode, eventCount: events.length });

    ws = new StravaImportWebSocket({
      onProgress: handleProgress,
      onComplete: handleComplete,
      onError: handleError,
      onStateChange: handleStateChange,
    });

    ws.connect(organizerEmail, events, groupCode || undefined);
  });

  // Cleanup on destroy
  onDestroy(() => {
    if (ws) {
      ws.close();
      ws = null;
    }
  });
</script>

<Card.Root class="mx-auto max-w-2xl border-2 shadow-lg" role="region" aria-label="Strava event import progress">
  <!-- Animated gradient header -->
  <div class="h-2 bg-gradient-to-r from-[#FC5200] via-orange-400 to-[#FC5200] animate-[gradient_2s_ease-in-out_infinite] bg-[length:200%_100%]" aria-hidden="true"></div>

  <!-- Screen reader live region for progress announcements -->
  <div class="sr-only" aria-live="polite" aria-atomic="true">
    {#if wsState === "connecting"}
      {#if reconnectAttempts.current > 0}
        Reconnecting, attempt {reconnectAttempts.current} of {reconnectAttempts.max}
      {:else}
        Connecting to import service
      {/if}
    {:else if wsState === "error"}
      Connection error, attempting to reconnect
    {:else if completedCount > 0 && completedCount < events.length}
      Imported {completedCount} of {events.length} events
    {:else if completedCount === events.length}
      Import complete, {completedCount} events imported successfully
    {/if}
  </div>

  <div class="p-6">
    <Card.Header class="px-0 pt-0 pb-6">
      <div class="flex items-center gap-3 mb-4">
        <div class="flex h-12 w-12 items-center justify-center rounded-xl bg-[#FC5200] shadow-sm">
          <svg class="h-6 w-6 text-white animate-spin" fill="none" viewBox="0 0 24 24" aria-hidden="true">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="3"></circle>
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
          </svg>
        </div>
        <div>
          <Card.Title id="import-progress-heading" tabindex={-1} class="text-xl mb-1">
            Importing {events.length} Event{events.length !== 1 ? "s" : ""}
          </Card.Title>
          <Card.Description class="text-sm font-medium">
      {#if wsState === "connecting"}
        {#if reconnectAttempts.current > 0}
          Reconnecting (attempt {reconnectAttempts.current}/{reconnectAttempts.max})...
        {:else}
          Connecting to server...
        {/if}
      {:else if wsState === "connected"}
        Import in progress. Please don't close this window.
      {:else if wsState === "error"}
        Connection error. Retrying...
      {:else if canRetry}
        Connection lost after {reconnectAttempts.max} attempts.
            {:else}
              Finishing up...
            {/if}
          </Card.Description>
        </div>
      </div>
    </Card.Header>

    <Card.Content class="space-y-6 px-0">
    <!-- Action Buttons -->
    <div class="flex gap-2">
      {#if wsState === "connected" || wsState === "connecting"}
        <Button
          variant="outline"
          size="sm"
          onclick={handleStop}
          aria-label="Stop current import and return to event selection"
        >
          Stop Import
        </Button>
      {/if}
      {#if canRetry}
        <Button
          variant="default"
          size="sm"
          onclick={handleRetry}
          disabled={isRetrying}
          aria-label="Retry failed import from where it stopped"
        >
          {isRetrying ? "Retrying..." : "Retry Import"}
        </Button>
      {/if}
    </div>

      <!-- Overall Progress -->
      <div class="space-y-3 rounded-xl border-2 bg-gradient-to-br from-background to-muted/30 p-4 sm:p-5">
        <div class="flex justify-between items-center gap-4">
          <span class="text-sm font-semibold">Overall Progress</span>
          <span class="text-xl sm:text-2xl font-bold text-[#FC5200]">{getOverallProgress()}%</span>
        </div>
        <div class="relative h-3 sm:h-3 overflow-hidden rounded-full bg-muted">
          <div
            class="h-full bg-gradient-to-r from-[#FC5200] to-orange-400 transition-all duration-500 ease-out relative"
            style="width: {getOverallProgress()}%"
            role="progressbar"
            aria-valuenow={getOverallProgress()}
            aria-valuemin="0"
            aria-valuemax="100"
          >
            <!-- Animated shine effect -->
            <div class="absolute inset-0 bg-gradient-to-r from-transparent via-white/30 to-transparent animate-[shimmer_2s_ease-in-out_infinite]" aria-hidden="true"></div>
          </div>
        </div>
        <div class="flex justify-between text-xs sm:text-xs text-muted-foreground font-medium">
          <span>{completedCount} completed</span>
          <span>{events.length - completedCount} remaining</span>
        </div>
      </div>

      <!-- Per-Event Progress -->
      <div class="space-y-3">
        <h3 class="text-sm font-semibold text-muted-foreground uppercase tracking-wide">Event Details</h3>
        {#each events as event, i (event.strava_event_id)}
          {@const msg = progress.get(event.strava_event_id)}
          {@const isComplete = msg?.type === "complete"}
          {@const isSuccess = isComplete && msg?.success}
          {@const isError = isComplete && !msg?.success}
          <div class="rounded-xl border-2 p-3 sm:p-4 transition-all duration-300 {isSuccess ? 'border-green-500 bg-green-50' : isError ? 'border-red-500 bg-red-50' : 'border-border bg-background'} animate-in fade-in slide-in-from-bottom-2" style="animation-delay: {i * 50}ms">
            <div class="mb-3 flex items-start gap-2.5">
              <!-- Status icon -->
              <div class="flex h-8 w-8 flex-shrink-0 items-center justify-center rounded-lg {isSuccess ? 'bg-green-500' : isError ? 'bg-red-500' : 'bg-[#FC5200]'}">
                {#if isSuccess}
                  <svg class="h-5 w-5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="3" d="M5 13l4 4L19 7" />
                  </svg>
                {:else if isError}
                  <svg class="h-5 w-5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="3" d="M6 18L18 6M6 6l12 12" />
                  </svg>
                {:else}
                  <svg class="h-5 w-5 text-white animate-spin" fill="none" viewBox="0 0 24 24" aria-hidden="true">
                    <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="3"></circle>
                    <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                  </svg>
                {/if}
              </div>
              <span class="font-semibold text-sm flex-1 min-w-0 break-words leading-tight {isSuccess ? 'text-green-700' : isError ? 'text-red-700' : ''}">{getEventTitle(event.strava_event_id)}</span>
            </div>

            <!-- Step indicators -->
            <div class="flex flex-wrap gap-2">
              {#each STEPS as step}
                {@const status = getStepStatus(event.strava_event_id, step)}
                <span class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-md text-xs font-medium transition-all duration-200
                  {status === 'success' ? 'bg-green-100 text-green-700 border border-green-200' :
                   status === 'error' ? 'bg-red-100 text-red-700 border border-red-200' :
                   status === 'in_progress' ? 'bg-[#FC5200]/10 text-[#FC5200] border border-[#FC5200]/20 animate-pulse' :
                   'bg-muted text-muted-foreground border border-border'}">
                  {#if status === "success"}
                    <svg class="h-3 w-3" fill="currentColor" viewBox="0 0 20 20" aria-hidden="true">
                      <path fill-rule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clip-rule="evenodd" />
                    </svg>
                  {:else if status === "error"}
                    <svg class="h-3 w-3" fill="currentColor" viewBox="0 0 20 20" aria-hidden="true">
                      <path fill-rule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clip-rule="evenodd" />
                    </svg>
                  {:else if status === "in_progress"}
                    <svg class="h-3 w-3 animate-spin" fill="none" viewBox="0 0 24 24" aria-hidden="true">
                      <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                      <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"></path>
                    </svg>
                  {:else}
                    <svg class="h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                    </svg>
                  {/if}
                  <span class="capitalize">{step === "coordinates" ? "Location" : step === "database" ? "Saving" : step}</span>
                </span>
              {/each}
            </div>

            <!-- Error message if failed -->
            {#if isError}
              <p class="mt-3 text-sm text-red-700 font-medium flex items-start gap-2 animate-in fade-in slide-in-from-top-1 duration-300">
                <svg class="h-4 w-4 flex-shrink-0 mt-0.5" fill="currentColor" viewBox="0 0 20 20" aria-hidden="true">
                  <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clip-rule="evenodd" />
                </svg>
                {msg?.error || "Import failed"}
              </p>
            {/if}
          </div>
        {/each}
      </div>
    </Card.Content>
  </div>
</Card.Root>
