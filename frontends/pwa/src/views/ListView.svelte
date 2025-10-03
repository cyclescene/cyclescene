<script>
  import RideList from "../components/ride/rideList.svelte";
  import { todaysRides } from "../lib/stores";
</script>

<RideList rides={$todaysRides} />
