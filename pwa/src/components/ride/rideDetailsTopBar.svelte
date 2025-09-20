<script>
    import {
        allSavedRides,
        currentRideStore,
        goBackInHistory,
        rides,
        savedRides,
    } from "$lib/stores";

    import SaveIcon from "~icons/material-symbols/save-rounded";
    import SharteIcon from "~icons/material-symbols/battery-android-share-outline";
    import BackIcon from "~icons/ic/baseline-keyboard-backspace";
    import Button from "$lib/components/ui/button/button.svelte";

    function handleGoBack() {
        goBackInHistory();
        currentRideStore.clearRide();
    }

    async function saveRide() {
        const ride = currentRideStore.getRide();
        try {
            await savedRides.saveRide(ride);
        } catch (e) {
            console.error(`unable to save ride ${e}`);
        }
    }

    $: {
        console.log($allSavedRides);
    }
</script>

<div class="flex justify-center items-center p-2.5 bg-black z-[500] text-white">
    <Button class="bg-black h-10 w-10" onclick={handleGoBack}>
        <BackIcon />
    </Button>

    <div class="grow ml-10 font-bold py-2 px-2.5 text-center text-xl">
        Ride Details
    </div>

    <div>
        <Button class="bg-black h-10 w-10" onclick={saveRide}>
            <SaveIcon />
        </Button>
        <Button class="bg-black h-10 w-10">
            <SharteIcon />
        </Button>
    </div>
</div>
