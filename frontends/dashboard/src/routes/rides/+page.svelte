<script lang="ts">
  import { onMount } from "svelte";
  import * as Card from "$lib/components/ui/card";
  import { Button } from "$lib/components/ui/button";
  import { Input } from "$lib/components/ui/input";

  interface Ride {
    id: number;
    title: string;
    description: string;
    city: string;
    venue_name: string;
    organizer_name: string;
    organizer_email: string;
    image_url: string;
    image_uuid: string;
    is_published: boolean;
    is_loop_ride: boolean;
    created_at: string;
  }

  let rides: Ride[] = [];
  let loading = true;
  let error = "";
  let selectedRideId: number | null = null;
  let publishingId: number | null = null;
  let adminToken = "";
  let showApiKeyForm = false;
  let apiKeyInput = "";

  onMount(async () => {
    adminToken = localStorage.getItem("adminToken") || "";
    if (!adminToken) {
      showApiKeyForm = true;
      loading = false;
      return;
    }
    await loadRides();
  });

  function setApiKey() {
    if (!apiKeyInput.trim()) {
      error = "Please enter an API key";
      return;
    }
    adminToken = apiKeyInput.trim();
    localStorage.setItem("adminToken", adminToken);
    showApiKeyForm = false;
    apiKeyInput = "";
    error = "";
    loadRides();
  }

  function clearApiKey() {
    localStorage.removeItem("adminToken");
    adminToken = "";
    showApiKeyForm = true;
    rides = [];
  }

  async function loadRides() {
    try {
      loading = true;
      error = "";
      const response = await fetch("http://localhost:8080/v1/rides/admin/pending", {
        headers: {
          "X-Admin-Token": adminToken,
        },
      });

      if (!response.ok) {
        throw new Error(`Failed to fetch rides: ${response.statusText}`);
      }

      rides = await response.json();
    } catch (err) {
      error = err instanceof Error ? err.message : "Failed to load rides";
    } finally {
      loading = false;
    }
  }

  async function publishRide(rideId: number) {
    try {
      publishingId = rideId;
      error = "";

      const response = await fetch(`http://localhost:8080/v1/rides/admin/${rideId}/publish`, {
        method: "PATCH",
        headers: {
          "X-Admin-Token": adminToken,
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          moderation_notes: "",
        }),
      });

      if (!response.ok) {
        throw new Error(`Failed to publish ride: ${response.statusText}`);
      }

      // Remove published ride from list
      rides = rides.filter((r) => r.id !== rideId);
      selectedRideId = null;
    } catch (err) {
      error = err instanceof Error ? err.message : "Failed to publish ride";
    } finally {
      publishingId = null;
    }
  }
</script>

<div class="container max-w-6xl mx-auto py-8 px-4">
  <div class="mb-8 flex items-center justify-between">
    <div>
      <h1 class="text-4xl font-bold tracking-tight">Rides</h1>
      <p class="text-muted-foreground mt-2">Publish pending rides</p>
    </div>
    {#if adminToken && !showApiKeyForm}
      <Button variant="outline" onclick={clearApiKey} size="sm">
        Change API Key
      </Button>
    {/if}
  </div>

  {#if error}
    <div class="mb-4 p-4 border border-destructive bg-destructive/10 rounded-lg">
      <p class="text-sm text-destructive">{error}</p>
    </div>
  {/if}

  {#if showApiKeyForm}
    <div class="max-w-md mx-auto py-12">
      <Card.Root>
        <Card.Header>
          <Card.Title>Enter API Key</Card.Title>
          <Card.Description>
            Your API key provides access to the admin dashboard
          </Card.Description>
        </Card.Header>
        <Card.Content class="space-y-4">
          <Input
            type="password"
            placeholder="Paste your API key..."
            bind:value={apiKeyInput}
            onkeydown={(e) => {
              if (e.key === "Enter") setApiKey();
            }}
          />
          <div class="flex gap-2">
            <Button onclick={setApiKey} class="flex-1">
              Continue
            </Button>
            <Button variant="outline" onclick={() => (apiKeyInput = "")}>
              Clear
            </Button>
          </div>
        </Card.Content>
      </Card.Root>
    </div>
  {:else if loading}
    <div class="flex items-center justify-center py-12">
      <p class="text-muted-foreground">Loading...</p>
    </div>
  {:else if rides.length === 0}
    <div class="flex items-center justify-center py-12">
      <p class="text-muted-foreground">No pending rides</p>
    </div>
  {:else}
    <div class="space-y-4">
      {#each rides as ride}
        <button
          onclick={() => (selectedRideId = selectedRideId === ride.id ? null : ride.id)}
          class="w-full text-left p-4 border rounded-lg hover:bg-accent/50 transition-colors"
        >
          <div class="flex items-center justify-between">
            <div class="flex-1">
              <p class="font-medium">{ride.title}</p>
              <p class="text-sm text-muted-foreground">
                {ride.city} • {ride.venue_name} • {ride.organizer_name}
              </p>
            </div>
            <Button
              onclick={(e) => {
                e.stopPropagation();
                publishRide(ride.id);
              }}
              disabled={publishingId === ride.id}
              size="sm"
              class="ml-4"
            >
              {publishingId === ride.id ? "Publishing..." : "Publish"}
            </Button>
          </div>
        </button>

        <!-- Ride Details (expand on click) -->
        {#if selectedRideId === ride.id}
          <Card.Root class="ml-4 mb-4">
            <Card.Content class="pt-6 space-y-4">
              {#if ride.image_url}
                <div>
                  <img
                    src={ride.image_url}
                    alt={ride.title}
                    class="w-full h-48 object-cover rounded-lg border"
                  />
                </div>
              {/if}

              <div class="grid grid-cols-2 gap-4">
                <div>
                  <p class="text-xs font-medium text-muted-foreground">Venue</p>
                  <p class="text-sm mt-1">{ride.venue_name}</p>
                </div>
                <div>
                  <p class="text-xs font-medium text-muted-foreground">Type</p>
                  <p class="text-sm mt-1">
                    {ride.is_loop_ride ? "Loop Ride" : "Point-to-Point"}
                  </p>
                </div>
                <div>
                  <p class="text-xs font-medium text-muted-foreground">Organizer</p>
                  <p class="text-sm mt-1">{ride.organizer_name}</p>
                </div>
                <div>
                  <p class="text-xs font-medium text-muted-foreground">Email</p>
                  <p class="text-sm mt-1 break-all">{ride.organizer_email}</p>
                </div>
              </div>

              <div>
                <p class="text-xs font-medium text-muted-foreground">Description</p>
                <p class="text-sm mt-2 whitespace-pre-wrap">{ride.description}</p>
              </div>

              {#if ride.image_uuid}
                <div class="pt-2 border-t">
                  <p class="text-xs font-medium text-muted-foreground">Image UUID</p>
                  <p class="text-xs mt-1 break-all text-muted-foreground">{ride.image_uuid}</p>
                </div>
              {/if}
            </Card.Content>
          </Card.Root>
        {/if}
      {/each}
    </div>
  {/if}
</div>
