<script>
    import RideList from "../components/rideList.svelte";
    import { allRides } from "../lib/stores";
</script>

<RideList rides={$allRides} />
