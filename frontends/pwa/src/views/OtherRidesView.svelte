<script>
  import { ridesWithoutLocations } from "../lib/stores";
  import RideList from "../components/ride/rideList.svelte";
</script>

<div class="flex h-full min-h-0 flex-col">
  <div class="border-b bg-background px-5 py-3">
    <h2 class="text-base font-semibold">Rides without map locations</h2>
    <p class="mt-1 text-sm text-muted-foreground">
      These rides are listed for the selected day, but their start location could not be mapped.
    </p>
  </div>

  <div class="min-h-0 flex-1">
    <RideList
      rides={$ridesWithoutLocations}
      emptyMessage="All rides for this day have map locations."
    />
  </div>
</div>
