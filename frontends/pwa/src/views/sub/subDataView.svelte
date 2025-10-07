<script lang="ts">
  import Button from "$lib/components/ui/button/button.svelte";
  import { triggerForegroundSync } from "$lib/stores";

  // your script goes here
</script>

<Button disabled={false} onclick={triggerForegroundSync}>FORCE SYNC</Button>
