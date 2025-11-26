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
    height: 100%;
    overflow-y: auto;
  }
</style>
