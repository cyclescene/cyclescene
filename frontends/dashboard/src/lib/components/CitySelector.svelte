<script lang="ts">
  import * as Select from "$lib/components/ui/select";
  import * as Sheet from "$lib/components/ui/sheet";
  import { Button } from "$lib/components/ui/button";
  import { CITIES, type CityCode } from "$lib/config/cities";
  import { MapPin } from "@lucide/svelte";

  interface Props {
    value: CityCode;
    onValueChange: (value: CityCode) => void;
    variant?: "desktop" | "mobile";
  }

  let { value = $bindable("all"), onValueChange, variant = "desktop" }: Props = $props();

  let mobileSheetOpen = $state(false);
  let previousValue = $state(value);

  // Watch for value changes and trigger callback
  $effect(() => {
    if (value !== previousValue) {
      previousValue = value;
      onValueChange(value);
    }
  });

  function handleMobileCitySelect(city: CityCode) {
    value = city;
    mobileSheetOpen = false;
  }

  const selectedCity = $derived(
    CITIES.find((c) => c.code === value) || CITIES[0]
  );
</script>

{#if variant === "desktop"}
  <Select.Root bind:value type="single">
    <Select.Trigger class="w-[200px]">
      {selectedCity.name}
      {#if selectedCity.state}
        <span class="text-muted-foreground ml-1">({selectedCity.state})</span>
      {/if}
    </Select.Trigger>
    <Select.Content>
      {#each CITIES as city}
        <Select.Item value={city.code}>
          {city.name}
          {#if city.state}
            <span class="text-muted-foreground ml-1">({city.state})</span>
          {/if}
        </Select.Item>
      {/each}
    </Select.Content>
  </Select.Root>
{:else}
  <Sheet.Root bind:open={mobileSheetOpen}>
    <Sheet.Trigger>
      {#snippet child({ props })}
        <Button {...props} variant="outline" size="sm" class="gap-2">
          <MapPin class="h-4 w-4" />
          {selectedCity.name}
        </Button>
      {/snippet}
    </Sheet.Trigger>
    <Sheet.Content side="bottom" class="max-h-[80vh]">
      <Sheet.Header>
        <Sheet.Title>Select City</Sheet.Title>
        <Sheet.Description>
          Filter rides by city or view all cities
        </Sheet.Description>
      </Sheet.Header>
      <div class="mt-4 space-y-2">
        {#each CITIES as city}
          <button
            onclick={() => handleMobileCitySelect(city.code)}
            class="w-full text-left p-4 rounded-lg border transition-colors {value ===
            city.code
              ? 'bg-accent border-primary'
              : 'hover:bg-accent/50'}"
          >
            <p class="font-medium">{city.name}</p>
            {#if city.state}
              <p class="text-sm text-muted-foreground">{city.state}</p>
            {/if}
          </button>
        {/each}
      </div>
    </Sheet.Content>
  </Sheet.Root>
{/if}
