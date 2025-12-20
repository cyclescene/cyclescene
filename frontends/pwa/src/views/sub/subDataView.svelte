<script lang="ts">
  import Button from "$lib/components/ui/button/button.svelte";
  import * as Card from "$lib/components/ui/card";
  import Separator from "$lib/components/ui/separator/separator.svelte";
  import { rides, savedRidesStore, syncStatus } from "$lib/stores";

  function formatSyncTime(date: Date | null) {
    if (!date) return "Never";
    const now = Date.now();
    const diff = now - date.getTime();

    if (diff < 60000) return "Just now";
    if (diff < 3600000) return `${Math.floor(diff / 60000)}m ago`;
    if (diff < 86400000) return `${Math.floor(diff / 3600000)}h ago`;
    return `${Math.floor(diff / 86400000)}d ago`;
  }

  function getSyncStatusColor(status: string | null) {
    switch (status) {
      case "success": return "text-green-600 dark:text-green-400";
      case "error": return "text-red-600 dark:text-red-400";
      case "syncing": return "text-blue-600 dark:text-blue-400";
      default: return "text-gray-600 dark:text-gray-400";
    }
  }

  async function handleClearAndRefreshRides() {
    if (
      confirm(
        "Are you sure you want to clear all ride data and fetch the latest from the API? This will replace all cached ride data.",
      )
    ) {
      try {
        await rides.clearAndRefreshRides();
        alert("Ride data has been refreshed successfully!");
      } catch (e) {
        alert(`Failed to refresh ride data: ${e}`);
      }
    }
  }

  async function handleClearSavedRides() {
    if (
      confirm(
        "Are you sure you want to clear all saved rides? This action cannot be undone.",
      )
    ) {
      try {
        await savedRidesStore.clearAll();
        alert("All saved rides have been cleared!");
      } catch (e) {
        alert(`Failed to clear saved rides: ${e}`);
      }
    }
  }
</script>

<div class="p-5 space-y-4">
  <!-- Sync Status Section -->
  <Card.Root class="p-4">
    <Card.Header class="p-0 mb-3">
      <Card.Title class="text-lg">Background Sync Status</Card.Title>
    </Card.Header>

    <div class="space-y-2 text-sm">
      <div class="flex justify-between items-center">
        <span class="text-muted-foreground">Last sync:</span>
        <span class={getSyncStatusColor($syncStatus.lastSyncStatus)}>
          {formatSyncTime($syncStatus.lastSyncTime)}
          {#if $syncStatus.lastSyncStatus === "syncing"}
            <span class="animate-pulse ml-1">Syncing...</span>
          {/if}
        </span>
      </div>

      <div class="flex justify-between items-center">
        <span class="text-muted-foreground">Status:</span>
        <span class={getSyncStatusColor($syncStatus.lastSyncStatus)}>
          {$syncStatus.lastSyncStatus || "Unknown"}
        </span>
      </div>

      {#if $syncStatus.lastSyncError}
        <div class="flex justify-between items-start gap-2">
          <span class="text-muted-foreground">Error:</span>
          <span class="text-red-600 dark:text-red-400 text-xs text-right max-w-xs">
            {$syncStatus.lastSyncError}
          </span>
        </div>
      {/if}

      <div class="flex justify-between items-center">
        <span class="text-muted-foreground">Total syncs:</span>
        <span>{$syncStatus.syncCount}</span>
      </div>

      <div class="flex justify-between items-center">
        <span class="text-muted-foreground">Interval:</span>
        <span>Every 30 minutes</span>
      </div>
    </div>
  </Card.Root>

  <!-- Data Management Section -->
  <Card.Root class="p-2 gap-2">
    <Card.Header class=" flex p-0">
      <Card.Title class="grow text-left">
        <Button
          variant="ghost"
          class="w-full justify-start"
          disabled={false}
          onclick={handleClearAndRefreshRides}
        >
          Get latest ride data
        </Button>
      </Card.Title>
    </Card.Header>
    <Separator />
    <Card.Header class=" flex p-0">
      <Card.Title class="grow text-left">
        <Button
          variant="ghost"
          class="w-full justify-start"
          disabled={false}
          onclick={handleClearSavedRides}>Clear saved rides</Button
        >
      </Card.Title>
    </Card.Header>
  </Card.Root>
</div>
