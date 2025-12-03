<script lang="ts">
  import Button from "$lib/components/ui/button/button.svelte";
  import * as Card from "$lib/components/ui/card";
  import Separator from "$lib/components/ui/separator/separator.svelte";
  import { rides, savedRidesStore } from "$lib/stores";

  async function handleClearAndRefreshRides() {
    if (
      confirm(
        "Are you sure you want to clear all ride data and fetch the latest from the API? This will replace all cached ride data.",
      )
    ) {
      try {
        await rides.clearAndRefreshRides();
        alert("Ride data has been refreshed successfully!");
      } catch (e) {
        alert(`Failed to refresh ride data: ${e}`);
      }
    }
  }

  async function handleClearSavedRides() {
    if (
      confirm(
        "Are you sure you want to clear all saved rides? This action cannot be undone.",
      )
    ) {
      try {
        await savedRidesStore.clearAll();
        alert("All saved rides have been cleared!");
      } catch (e) {
        alert(`Failed to clear saved rides: ${e}`);
      }
    }
  }
</script>

<div class="p-5">
  <Card.Root class="p-2 gap-2">
    <Card.Header class=" flex p-0">
      <Card.Title class="grow text-left">
        <Button
          variant="ghost"
          class="w-full justify-start"
          disabled={false}
          onclick={handleClearAndRefreshRides}
        >
          Get latest ride data
        </Button>
      </Card.Title>
    </Card.Header>
    <Separator />
    <Card.Header class=" flex p-0">
      <Card.Title class="grow text-left">
        <Button
          variant="ghost"
          class="w-full justify-start"
          disabled={false}
          onclick={handleClearSavedRides}>Clear saved rides</Button
        >
      </Card.Title>
    </Card.Header>
  </Card.Root>
</div>
