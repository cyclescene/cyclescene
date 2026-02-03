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

<Card.Root class="mx-auto max-w-2xl p-6" role="region" aria-label="Strava event import progress">
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

  <Card.Header class="px-0 pt-0">
    <Card.Title id="import-progress-heading" tabindex={-1}>
      Importing {events.length} Event{events.length !== 1 ? "s" : ""}
    </Card.Title>
    <Card.Description>
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
    <div class="space-y-2">
      <div class="flex justify-between text-sm">
        <span>Overall Progress</span>
        <span>{completedCount} of {events.length}</span>
      </div>
      <Progress value={getOverallProgress()} class="h-2" />
    </div>

    <!-- Per-Event Progress -->
    <div class="space-y-4">
      {#each events as event, i (event.strava_event_id)}
        <div class="rounded-lg border p-4">
          <div class="mb-3 flex items-center gap-2">
            <Badge variant={getEventBadgeVariant(event.strava_event_id)}>
              {getEventIcon(event.strava_event_id)}
            </Badge>
            <span class="font-medium">{getEventTitle(event.strava_event_id)}</span>
          </div>

          <!-- Step indicators -->
          <div class="flex flex-wrap gap-2">
            {#each STEPS as step}
              {@const status = getStepStatus(event.strava_event_id, step)}
              <Badge
                variant={getStepBadgeVariant(status)}
                class="text-xs capitalize {status === 'in_progress' ? 'animate-pulse' : ''}"
              >
                {#if status === "success"}✓{:else if status === "error"}✗{:else if status === "in_progress"}⟳{/if}
                {step === "coordinates" ? "Location" : step === "database" ? "Saving" : step}
              </Badge>
            {/each}
          </div>

          <!-- Error message if failed -->
          {#if progress.get(event.strava_event_id)?.type === "complete" && !progress.get(event.strava_event_id)?.success}
            <p class="mt-2 text-sm text-red-500">
              {progress.get(event.strava_event_id)?.error || "Import failed"}
            </p>
          {/if}
        </div>
      {/each}
    </div>
  </Card.Content>
</Card.Root>
