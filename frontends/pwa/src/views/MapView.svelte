<script>
  import { filteredRides, filteredNoAddress } from "../lib/stores.js";
  import MapComponent from "../components/map/MapComponent.svelte";
</script>

<MapComponent rides={$filteredRides} otherRides={$filteredNoAddress} />
