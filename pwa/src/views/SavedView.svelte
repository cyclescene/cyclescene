<script>
    import { savedRidesForSelectedDay } from "$lib/stores";
    import RideList from "../components/ride/rideList.svelte";
</script>

<RideList rides={$savedRidesForSelectedDay} />
