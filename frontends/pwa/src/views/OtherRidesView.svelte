<script>
  import { ridesWithoutLocations } from "../lib/stores";
  import ListView from "./ListView.svelte";
</script>

<ListView rides={$ridesWithoutLocations} />
