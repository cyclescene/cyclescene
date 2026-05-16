<script lang="ts">
  import Button from "$lib/components/ui/button/button.svelte";
  import * as Card from "$lib/components/ui/card";
  import ConfirmActionDialog from "../../components/settings/confirmActionDialog.svelte";
  import Separator from "$lib/components/ui/separator/separator.svelte";
  import { rides, savedRidesStore, syncStatus } from "$lib/stores";
  import { toast } from "svelte-sonner";

  let refreshDialogOpen = $state(false);
  let clearSavedDialogOpen = $state(false);

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

  async function refreshRideData() {
    try {
      await rides.clearAndRefreshRides();
      toast.success("Ride data refreshed");
    } catch (e) {
      toast.error(`Failed to refresh ride data: ${e}`);
    }
  }

  async function clearSavedRides() {
    try {
      await savedRidesStore.clearAll();
      toast.success("Saved rides cleared");
    } catch (e) {
      toast.error(`Failed to clear saved rides: ${e}`);
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
          onclick={() => (refreshDialogOpen = true)}
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
          onclick={() => (clearSavedDialogOpen = true)}>Clear saved rides</Button
        >
      </Card.Title>
    </Card.Header>
  </Card.Root>
</div>

<ConfirmActionDialog
  bind:open={refreshDialogOpen}
  title="Refresh ride data?"
  description="This clears cached ride data and fetches the latest rides from the API."
  confirmLabel="Refresh"
  onConfirm={refreshRideData}
/>

<ConfirmActionDialog
  bind:open={clearSavedDialogOpen}
  title="Clear saved rides?"
  description="This removes every saved ride from this device. This action cannot be undone."
  confirmLabel="Clear"
  destructive
  onConfirm={clearSavedRides}
/>
