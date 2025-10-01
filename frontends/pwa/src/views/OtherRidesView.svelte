<script>
    import { filteredNoAddress } from "../lib/stores";
    import ListView from "./ListView.svelte";
</script>

<ListView rides={$filteredNoAddress} />
