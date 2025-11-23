<script lang="ts">
  import { Button } from "$lib/components/ui/button";
  import cities from "$lib/data/cities.json";

  export const prerender = true;

  interface City {
    name: string;
    code: string;
    url: string;
  }

  const cityList: City[] = cities;
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
          onclick={() => window.location.href = city.url}
          size="lg"
          class="h-24 text-lg"
        >
          {city.name}
        </Button>
      {/each}
    </div>
  </div>
</div>
