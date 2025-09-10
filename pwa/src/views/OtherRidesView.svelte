<script>
    import OtherRides from "../components/otherRides.svelte";
    import { filteredNoAddress } from "../lib/stores";
</script>

<OtherRides rides={$filteredNoAddress} />
