<script>
  import { savedRidesForSelectedDay } from "$lib/stores";
  import RideList from "../components/ride/rideList.svelte";
  import RideTab from "../components/saved/rideTab.svelte";
</script>

<div class="saved-view">
  <RideTab />
</div>

<style>
  .saved-view {
    height: calc(100vh - var(--header-height) - (var(--footer-height)));
  }
</style>
