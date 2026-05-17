<script lang="ts">
  import * as Card from "$lib/components/ui/card";
  import { Button } from "$lib/components/ui/button";
  import { ListOrdered, TrendingUp, MapPin, RefreshCw } from "@lucide/svelte";

  const API_URL = import.meta.env.PUBLIC_API_URL || "https://api.cyclescene.cc";

  let triggeringSync = $state(false);
  let syncMessage = $state("");
  let syncError = $state("");

  async function triggerShift2BikesSync() {
    const adminToken = localStorage.getItem("adminToken") || "";
    if (!adminToken) {
      syncError = "Missing admin API key.";
      return;
    }

    try {
      triggeringSync = true;
      syncMessage = "";
      syncError = "";

      const response = await fetch(`${API_URL}/v1/admin/sync/shift2bikes`, {
        method: "POST",
        headers: {
          "X-Admin-Token": adminToken,
        },
      });

      const text = await response.text();
      let payload: { message?: string; status?: string } = {};
      try {
        payload = text ? JSON.parse(text) : {};
      } catch {
        payload = { message: text };
      }

      if (!response.ok) {
        throw new Error(payload.message || `Sync trigger failed with ${response.status}`);
      }

      syncMessage = payload.message || "Shift2Bikes sync job started.";
    } catch (err) {
      syncError = err instanceof Error ? err.message : "Failed to trigger Shift2Bikes sync.";
    } finally {
      triggeringSync = false;
    }
  }
</script>

<div class="container max-w-6xl mx-auto py-8 md:py-16 px-4">
  <!-- Hero Section -->
  <div class="text-center space-y-4 mb-12 md:mb-16">
    <div
      class="inline-block px-4 py-1.5 rounded-full bg-primary/10 text-primary text-sm font-medium mb-4"
    >
      Admin Dashboard
    </div>
    <h1 class="text-4xl md:text-6xl font-bold tracking-tight text-foreground">
      CycleScene Dashboard
    </h1>
    <p class="text-lg md:text-xl text-muted-foreground max-w-2xl mx-auto">
      Review and publish community-submitted rides across Portland and Salt Lake
      City
    </p>
  </div>

  <!-- Cards Grid -->
  <div class="grid grid-cols-1 md:grid-cols-2 gap-6 max-w-4xl mx-auto">
    <!-- Manage Rides Card -->
    <Card.Root
      class="group cursor-pointer transition-colors hover:bg-muted/50 hover:border-primary/50"
      onclick={() => (window.location.href = "/rides")}
    >
      <Card.Header>
        <div
          class="h-12 w-12 rounded-lg bg-primary/10 flex items-center justify-center mb-4 group-hover:bg-primary/20 transition-colors"
        >
          <ListOrdered class="h-6 w-6 text-primary" />
        </div>
        <Card.Title class="text-2xl">Manage Rides</Card.Title>
        <Card.Description class="text-base">
          Review pending ride submissions and publish them to the community
        </Card.Description>
      </Card.Header>
      <Card.Content>
        <Button
          variant="ghost"
          class="w-full justify-between group-hover:bg-accent"
        >
          Go to Rides
          <span class="group-hover:translate-x-1 transition-transform"
            >&rarr;</span
          >
        </Button>
      </Card.Content>
    </Card.Root>

    <!-- Quick Stats Card -->
    <Card.Root>
      <Card.Header>
        <div
          class="h-12 w-12 rounded-lg bg-muted flex items-center justify-center mb-4"
        >
          <TrendingUp class="h-6 w-6 text-muted-foreground" />
        </div>
        <Card.Title class="text-2xl">Quick Stats</Card.Title>
        <Card.Description class="text-base">
          Platform overview and insights
        </Card.Description>
      </Card.Header>
      <Card.Content class="space-y-4">
        <div class="flex items-center gap-3">
          <MapPin class="h-5 w-5 text-muted-foreground" />
          <div>
            <p class="text-sm font-medium">Active Cities</p>
            <p class="text-2xl font-bold">2</p>
          </div>
        </div>
        <div class="pt-2 border-t">
          <p class="text-xs text-muted-foreground">
            Portland, OR • Salt Lake City, UT
          </p>
        </div>
      </Card.Content>
    </Card.Root>
  </div>

  <div class="max-w-4xl mx-auto mt-6">
    <Card.Root>
      <Card.Header>
        <div
          class="h-12 w-12 rounded-lg bg-muted flex items-center justify-center mb-4"
        >
          <RefreshCw class="h-6 w-6 text-muted-foreground" />
        </div>
        <Card.Title class="text-2xl">Shift2Bikes Sync</Card.Title>
        <Card.Description class="text-base">
          Trigger the scraper now instead of waiting for the scheduled refresh.
        </Card.Description>
      </Card.Header>
      <Card.Content class="space-y-4">
        {#if syncMessage}
          <div class="rounded-lg border border-green-500/30 bg-green-500/10 p-3 text-sm text-green-700 dark:text-green-300">
            {syncMessage}
          </div>
        {/if}
        {#if syncError}
          <div class="rounded-lg border border-destructive bg-destructive/10 p-3 text-sm text-destructive">
            {syncError}
          </div>
        {/if}
        <Button
          onclick={triggerShift2BikesSync}
          disabled={triggeringSync}
          class="gap-2"
        >
          <RefreshCw class={`h-4 w-4 ${triggeringSync ? "animate-spin" : ""}`} />
          {triggeringSync ? "Starting sync..." : "Refresh Shift2Bikes Now"}
        </Button>
      </Card.Content>
    </Card.Root>
  </div>

  <!-- Footer Info -->
  <div class="mt-12 md:mt-16 text-center">
    <p class="text-sm text-muted-foreground">
      Need help? Contact support or check the documentation
    </p>
  </div>
</div>
