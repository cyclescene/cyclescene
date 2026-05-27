<script lang="ts">
  import { mapStore, navigateTo, ridesWithoutLocations, VIEW_OTHER_RIDES } from "$lib/stores";
  import Button from "$lib/components/ui/button/button.svelte";
  import LocationOffIcon from "~icons/material-symbols/location-off-outline-rounded";

  function onShowNotShown() {
    navigateTo(VIEW_OTHER_RIDES);
  }
</script>

{#if !$mapStore.showCurrentRide && $ridesWithoutLocations.length > 0}
  <Button
    class="absolute right-5 bottom-4 z-[1000] h-11 border !bg-card px-3 !text-card-foreground shadow-lg active:scale-[0.98] hover:!bg-accent hover:!text-accent-foreground"
    type="button"
    variant="outline"
    aria-label={`Show ${$ridesWithoutLocations.length} ride${$ridesWithoutLocations.length === 1 ? "" : "s"} without map locations`}
    onclick={onShowNotShown}
  >
    <LocationOffIcon class="size-5 shrink-0 text-yellow-600 dark:text-yellow-400" aria-hidden="true" />
    <span class="text-sm font-semibold">
      {$ridesWithoutLocations.length} unmapped
    </span>
  </Button>
{/if}
