<script>
    import { allSavedRides, savedRidesGroupedByDate } from "$lib/stores";
    import RideList from "../components/ride/rideList.svelte";

    $: {
        console.log($savedRidesGroupedByDate);
    }
</script>

<RideList rides={$allSavedRides} />
