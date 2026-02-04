<script lang="ts">
  import { onMount } from "svelte";
  import * as Card from "$lib/components/ui/card";
  import { Button } from "$lib/components/ui/button";
  import { Input } from "$lib/components/ui/input";
  import { Skeleton } from "$lib/components/ui/skeleton";
  import RideCard from "$lib/components/RideCard.svelte";
  import { CITIES, type CityCode } from "$lib/config/cities";

  const API_URL = import.meta.env.PUBLIC_API_URL || "https://api.cyclescene.cc";

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

  let rides = $state<Ride[]>([]);
  let loading = $state(true);
  let error = $state("");
  let publishingId = $state<number | null>(null);
  let adminToken = $state("");
  let showApiKeyForm = $state(false);
  let apiKeyInput = $state("");
  let selectedCity = $state<CityCode>("all");

  // Derived state for filtered rides
  let filteredRides = $derived(
    selectedCity === "all"
      ? rides
      : rides.filter((r) => r.city === selectedCity)
  );

  // Derived state for city info
  let selectedCityInfo = $derived(
    CITIES.find((c) => c.code === selectedCity) || CITIES[0]
  );

  onMount(() => {
    adminToken = localStorage.getItem("adminToken") || "";
    if (!adminToken) {
      showApiKeyForm = true;
      loading = false;
      return;
    }

    // Load selected city from localStorage
    const savedCity = localStorage.getItem("selectedCity") as CityCode | null;
    if (savedCity) {
      selectedCity = savedCity;
    }

    // Listen for city changes from header
    window.addEventListener("citychange", handleCityChange);
    window.addEventListener("changeapikey", clearApiKey);

    loadRides();

    return () => {
      window.removeEventListener("citychange", handleCityChange);
      window.removeEventListener("changeapikey", clearApiKey);
    };
  });

  function handleCityChange(event: Event) {
    const customEvent = event as CustomEvent<CityCode>;
    selectedCity = customEvent.detail;
  }

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
    // Force page reload to update header
    window.location.reload();
  }

  async function loadRides() {
    try {
      loading = true;
      error = "";
      const response = await fetch(`${API_URL}/v1/rides/admin/pending`, {
        headers: {
          "X-Admin-Token": adminToken,
        },
      });

      if (!response.ok) {
        throw new Error(`Failed to fetch rides: ${response.statusText}`);
      }

      const data = await response.json();
      rides = Array.isArray(data) ? data : [];
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

      const response = await fetch(
        `${API_URL}/v1/rides/admin/${rideId}/publish`,
        {
          method: "PATCH",
          headers: {
            "X-Admin-Token": adminToken,
            "Content-Type": "application/json",
          },
          body: JSON.stringify({
            moderation_notes: "",
          }),
        },
      );

      if (!response.ok) {
        throw new Error(`Failed to publish ride: ${response.statusText}`);
      }

      // Remove published ride from list
      rides = rides.filter((r) => r.id !== rideId);
    } catch (err) {
      error = err instanceof Error ? err.message : "Failed to publish ride";
    } finally {
      publishingId = null;
    }
  }
</script>

<div class="container max-w-6xl mx-auto py-4 md:py-8 px-4">
  <div class="mb-6 md:mb-8">
    <h1 class="text-3xl md:text-4xl font-bold tracking-tight">Rides</h1>
    <p class="text-sm md:text-base text-muted-foreground mt-2">
      {#if selectedCity === "all"}
        Showing {filteredRides.length} {filteredRides.length === 1
          ? "ride"
          : "rides"} across all cities
      {:else}
        Showing {filteredRides.length} {filteredRides.length === 1
          ? "ride"
          : "rides"} in {selectedCityInfo.name}
      {/if}
    </p>
  </div>

  {#if error}
    <div
      class="mb-4 p-4 border border-destructive bg-destructive/10 rounded-lg"
    >
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
            <Button onclick={setApiKey} class="flex-1">Continue</Button>
            <Button variant="outline" onclick={() => (apiKeyInput = "")}>
              Clear
            </Button>
          </div>
        </Card.Content>
      </Card.Root>
    </div>
  {:else if loading}
    <div class="space-y-4">
      {#each Array(3) as _}
        <div class="p-4 border rounded-lg">
          <div class="flex items-center justify-between gap-4">
            <div class="flex-1 space-y-2">
              <Skeleton class="h-5 w-3/4" />
              <Skeleton class="h-4 w-1/2" />
            </div>
            <Skeleton class="h-9 w-20" />
          </div>
        </div>
      {/each}
    </div>
  {:else if rides.length === 0}
    <div class="flex items-center justify-center py-12">
      <Card.Root class="max-w-xl w-full">
        <Card.Header class="text-center">
          <Card.Title>No Pending Rides</Card.Title>
          <Card.Description>
            All submitted rides have been reviewed and published. Check back
            later for new submissions.
          </Card.Description>
        </Card.Header>
        <Card.Content class="flex justify-center">
          <Button variant="outline" onclick={loadRides}>Refresh</Button>
        </Card.Content>
      </Card.Root>
    </div>
  {:else if filteredRides.length === 0}
    <div class="flex items-center justify-center py-12">
      <Card.Root class="max-w-xl w-full">
        <Card.Header class="text-center">
          <Card.Title>No Rides in {selectedCityInfo.name}</Card.Title>
          <Card.Description>
            There are no pending rides for this city. Try selecting a different
            city or view all cities.
          </Card.Description>
        </Card.Header>
        <Card.Content class="flex justify-center gap-2">
          <Button
            variant="outline"
            onclick={() => {
              selectedCity = "all";
              localStorage.setItem("selectedCity", "all");
              window.dispatchEvent(
                new CustomEvent("citychange", { detail: "all" })
              );
            }}
          >
            View All Cities
          </Button>
          <Button variant="outline" onclick={loadRides}>Refresh</Button>
        </Card.Content>
      </Card.Root>
    </div>
  {:else}
    <div class="grid grid-cols-1 gap-4">
      {#each filteredRides as ride (ride.id)}
        <RideCard
          {ride}
          onPublish={publishRide}
          isPublishing={publishingId === ride.id}
          showCityBadge={selectedCity === "all"}
        />
      {/each}
    </div>
  {/if}
</div>
