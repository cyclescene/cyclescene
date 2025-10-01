<script>
    import RideList from "../components/ride/rideList.svelte";
    import { allRides } from "../lib/stores";
</script>

<RideList rides={$allRides} />
