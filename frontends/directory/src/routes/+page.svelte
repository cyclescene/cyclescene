<script lang="ts">
  import { Button } from "$lib/components/ui/button";
  import EmojiBanner from "$lib/components/EmojiBanner.svelte";
  import cities from "$lib/data/cities.json";
  import { onMount } from "svelte";

  interface City {
    name: string;
    code: string;
    url: string;
  }

  const cityList: City[] = cities;
  let analyticsId: string | null = null;

  onMount(async () => {
    // Extract source parameter from URL
    const searchParams = new URLSearchParams(window.location.search);
    const source = searchParams.get("source");

    // Track initial page visit
    try {
      const response = await fetch("/api/analytics", {
        method: "POST",
        headers: {
          "Content-Type": "application/json"
        },
        body: JSON.stringify({ source })
      });

      if (response.ok) {
        const data = await response.json();
        analyticsId = data.analyticsId;
      }
    } catch (error) {
      console.error("Failed to track analytics:", error);
    }
  });

  async function trackCityClick(city: City) {
    // Track CTA click if we have an analytics ID
    if (analyticsId) {
      try {
        await fetch("/api/analytics", {
          method: "PATCH",
          headers: {
            "Content-Type": "application/json"
          },
          body: JSON.stringify({
            id: analyticsId,
            pwa_clicked: city.code
          })
        });
      } catch (error) {
        console.error("Failed to update analytics:", error);
      }
    }

    // Navigate to city PWA (don't wait for analytics to complete)
    window.location.href = city.url;
  }
</script>

<div class="flex flex-col h-screen overflow-hidden" style="padding-bottom: env(safe-area-inset-bottom)">
  <div class="flex-1 container max-w-4xl mx-auto py-8 px-4 flex flex-col justify-center">
    <div class="text-center space-y-8">
      <div class="space-y-2">
        <h1 class="text-5xl font-bold tracking-tight">Cycle Scene</h1>
        <p class="text-lg text-muted-foreground">Discover bike rides in your city</p>
      </div>

      <div class="grid grid-cols-1 md:grid-cols-2 gap-4 pt-4 max-w-2xl mx-auto">
        {#each cityList as city}
          <Button
            onclick={() => trackCityClick(city)}
            size="lg"
            class="h-16 text-base"
          >
            {city.name}
          </Button>
        {/each}
      </div>
    </div>
  </div>

  <div class="flex-shrink-0">
    <EmojiBanner />
  </div>
</div>
