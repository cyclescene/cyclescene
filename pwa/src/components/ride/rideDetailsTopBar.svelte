<script>
  import {
    currentRide,
    currentRideStore,
    goBackInHistory,
    savedRidesStore,
  } from "$lib/stores";
  import { toast, Toaster } from "svelte-sonner";

  import SaveIcon from "~icons/material-symbols/save-rounded";
  import ShareIcon from "~icons/material-symbols/battery-android-share-outline";
  import BackIcon from "~icons/ic/baseline-keyboard-backspace";
  import Button from "$lib/components/ui/button/button.svelte";

  function handleGoBack() {
    goBackInHistory();
    currentRideStore.clearRide();
  }

  let rideExists = false;
  let loading = true;

  $: {
    if (typeof window !== "undefined") {
      (async () => {
        loading = true;
        rideExists = await savedRidesStore.isRideSaved($currentRide?.id);
        loading = false;
      })();
    }
  }

  async function saveRide() {
    const ride = currentRideStore.getRide();
    toast.promise(savedRidesStore.saveRide(ride), {
      loading: "Saving...",
      success: "Ride saved!",
      error: "Unable to save ride",
    });
  }
</script>

<div class="flex justify-center items-center p-2.5 z-[500] text-white">
  <Toaster position="top-center" />
  <Button
    variant="secondary"
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
      variant="secondary"
      disabled={rideExists ? true : false}
      class={`${rideExists ? "bg-green-500" : ""} h-10 w-10`}
      onclick={saveRide}
    >
      <SaveIcon />
    </Button>
    <Button variant="secondary" disabled={false} class="h-10 w-10">
      <ShareIcon />
    </Button>
  </div>
</div>
