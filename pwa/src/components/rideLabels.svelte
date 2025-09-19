<script>
    import { activeView, VIEW_LIST } from "$lib/stores";
    import RideLabel from "./rideLabel.svelte";

    export let ride;
</script>

<div class="flex flex-row gap-2.5">
    {#if ride.cancelled}
        <RideLabel
            color="border-red-600"
            content={ride.cancelled ? "Cancelled" : ""}
        />
    {/if}

    {#if ride.audience == "F"}
        <RideLabel color="border-green-600" content="Family Friendly" />
    {:else if ride.audience == "A"}
        <RideLabel color="border-purple-600" content="Adults Only" />
    {:else}{/if}

    {#if ride.safetyplan}
        {#if $activeView == VIEW_LIST}
            <RideLabel color="border-blue-600" content="Safety Plan" />
        {:else}
            <a
                href="https://www.shift2bikes.org/pages/public-health/#safety-plan"
                target="_blank"
                rel="noopener noreferrer"
            >
                <RideLabel color="border-blue-600" content="Safety Plan" />
            </a>
        {/if}
    {/if}
</div>
