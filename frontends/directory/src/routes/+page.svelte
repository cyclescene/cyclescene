<script lang="ts">
  import { Button } from "$lib/components/ui/button";
  import cities from "$lib/data/cities.json";
  import { onMount } from "svelte";

  interface City {
    name: string;
    code: string;
    url: string;
  }

  const cityList: City[] = cities;
  let analyticsId: number | null = null;

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

<div class="container max-w-4xl mx-auto py-16 px-4">
  <div class="text-center space-y-12">
    <div class="space-y-3">
      <h1 class="text-6xl font-bold tracking-tight">Cycle Scene</h1>
      <p class="text-xl text-muted-foreground">Discover bike rides in your city</p>
    </div>

    <div class="grid grid-cols-1 md:grid-cols-2 gap-6 pt-8">
      {#each cityList as city}
        <Button
          onclick={() => trackCityClick(city)}
          size="lg"
          class="h-24 text-lg"
        >
          {city.name}
        </Button>
      {/each}
    </div>
  </div>
</div>
