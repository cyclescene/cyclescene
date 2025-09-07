<script>
    import { onMount } from "svelte";
    import { rides, filteredRides, filteredNoAddress } from "../lib/stores.js";
    import Map from "../components/map.svelte";

    onMount(() => {
        rides.init();
        rides.fetchUpcoming();
    });
</script>

<Map rides={$filteredRides} noAddressRides={$filteredNoAddress} />
