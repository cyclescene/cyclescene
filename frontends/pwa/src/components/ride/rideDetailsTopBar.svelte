<script>
  import {
    currentRide,
    currentRideStore,
    goBackInHistory,
    savedRidesStore,
  } from "$lib/stores";
  import { toast, Toaster } from "svelte-sonner";

  import SavedIcon from "~icons/material-symbols/bookmark-sharp";
  import UnsavedIcon from "~icons/material-symbols/bookmark-outline";
  import ShareIcon from "~icons/material-symbols/battery-android-share-outline";
  import BackIcon from "~icons/ic/baseline-keyboard-backspace";
  import Button from "$lib/components/ui/button/button.svelte";
  import { mode } from "mode-watcher";

  let ride = $derived(currentRideStore.getRide());
  let rideExists = $state(false);
  let loading = $state(true);

  function handleGoBack() {
    goBackInHistory();
    currentRideStore.clearRide();
  }

  async function copyToClipboard() {
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
      if (typeof window !== "undefined") {
        loading = true;
        rideExists = await savedRidesStore.isRideSaved($currentRide?.id);
        loading = false;
      }
    })();
  });

  async function handleSavedRide() {
    if (rideExists) {
      toast.promise(savedRidesStore.deleteRide(ride.id), {
        loading: "Removing...",
        success: "Ride removed!",
        error: "Unable to remove from saved",
      });
      rideExists = false;
    } else {
      toast.promise(savedRidesStore.saveRide(ride), {
        loading: "Saving...",
        success: "Ride saved!",
        error: "Unable to save ride",
      });
      rideExists = true;
    }
  }
</script>

<div class="flex justify-center items-center p-2.5 z-[1000]">
  <Toaster
    position="top-center"
    theme={mode.current}
    duration={1000}
    visibleToasts={1}
  />
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
      disabled={false}
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
