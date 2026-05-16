<script>
  import {
    currentRide,
    currentRideStore,
    goBackInHistory,
    savedRidesStore,
  } from "$lib/stores";
  import { toast } from "svelte-sonner";

  import SavedIcon from "~icons/material-symbols/bookmark-sharp";
  import UnsavedIcon from "~icons/material-symbols/bookmark-outline";
  import ShareIcon from "~icons/material-symbols/battery-android-share-outline";
  import BackIcon from "~icons/ic/baseline-keyboard-backspace";
  import Button from "$lib/components/ui/button/button.svelte";

  let ride = $derived($currentRide);
  let rideExists = $state(false);
  let loading = $state(true);

  function handleGoBack() {
    goBackInHistory();
    currentRideStore.clearRide();
  }

  async function copyToClipboard() {
    if (!ride) return false;

    try {
      await navigator.clipboard.writeText(ride.shareable);
      return true;
    } catch (err) {
      return false;
    }
  }

  function handleShareRide() {
    if (navigator.share && ride) {
      navigator.share({
        title: ride.title,
        url: ride.shareable,
      });
    } else if (navigator.clipboard && ride) {
      copyToClipboard(ride.shareable)
        .then(() => {
          toast("Link copied to clipboard!");
        })
        .catch(() => {
          toast("Could not copy link. Please manually copy the URL.");
        });
    }
  }

  $effect(() => {
    (async () => {
      if (typeof window !== "undefined" && $currentRide?.id) {
        loading = true;
        rideExists = await savedRidesStore.isRideSaved($currentRide.id);
        loading = false;
      } else {
        rideExists = false;
        loading = false;
      }
    })();
  });

  async function handleSavedRide() {
    if (!ride || loading) return;

    loading = true;

    try {
      if (rideExists) {
        await savedRidesStore.deleteRide(ride.id);
        rideExists = false;
        toast.success("Ride removed!");
      } else {
        await savedRidesStore.saveRide(ride);
        rideExists = true;
        toast.success("Ride saved!");
      }
    } catch (error) {
      if (rideExists) {
        toast.error("Unable to remove from saved");
      } else {
        toast.error("Unable to save ride");
      }
    } finally {
      loading = false;
    }
  }
</script>

<div class="flex justify-center items-center p-2.5 z-[1000]">
  <Button
    variant="ghost"
    disabled={false}
    class="h-10 w-10"
    onclick={handleGoBack}
  >
    <BackIcon />
  </Button>

  <div class="grow ml-10 font-bold py-2 px-2.5 text-center text-xl">
    Ride Details
  </div>

  <div>
    <Button
      variant="ghost"
      disabled={false}
      class="h-10 w-10"
      onclick={handleShareRide}
    >
      <ShareIcon />
    </Button>
    <Button
      variant="ghost"
      disabled={loading || !ride}
      class={`h-10 w-10`}
      onclick={handleSavedRide}
    >
      {#if rideExists}
        <SavedIcon />
      {:else}
        <UnsavedIcon />
      {/if}
    </Button>
  </div>
</div>
