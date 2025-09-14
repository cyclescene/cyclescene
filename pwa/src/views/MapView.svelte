<script>
    import { onMount } from "svelte";
    import { rides, filteredRides, filteredNoAddress } from "../lib/stores.js";
    import Map from "../components/map.svelte";
    import DatePicker from "../components/datePicker.svelte";
    import NavigationBar from "../components/navigationBar.svelte";

    onMount(() => {
        rides.init();
        rides.fetchUpcoming();
    });
</script>

<Map rides={$filteredRides} noAddressRides={$filteredNoAddress} />
