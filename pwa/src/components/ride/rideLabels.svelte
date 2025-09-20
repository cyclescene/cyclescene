<script>
    import { activeView, VIEW_LIST } from "$lib/stores";
    import RideLabel from "./rideLabel.svelte";

    import FamilyIcon from "~icons/ic/round-family-restroom";
    import CancelledIcon from "~icons/gridicons/cross-circle";
    import AdultsOnlyIcon from "~icons/uil/21-plus";
    import SafetyPlanIcons from "~icons/f7/facemask-fill";

    export let ride;
</script>

<div class="flex flex-row gap-2.5">
    {#if ride.cancelled}
        <RideLabel class="border-red-500 text-red-500">
            <svelte:component this={CancelledIcon} />
            <p class="text-white">Cancelled</p>
        </RideLabel>
    {/if}

    {#if ride.audience == "F"}
        <RideLabel class="border-green-500 text-green-500">
            <svelte:component this={FamilyIcon} />
            <p class="text-white">Family Friendly</p>
        </RideLabel>
    {:else if ride.audience == "A"}
        <RideLabel class="border-purple-500 text-purple-500">
            <svelte:component this={AdultsOnlyIcon} />
            <p class="text-white">Adults Only</p>
        </RideLabel>
    {:else}{/if}

    {#if ride.safetyplan}
        {#if $activeView == VIEW_LIST}
            <RideLabel class="border-blue-500 text-blue-500">
                <svelte:component this={SafetyPlanIcons} />
                <p class="text-white">Safety Plan</p>
            </RideLabel>
        {:else}
            <a
                href="https://www.shift2bikes.org/pages/public-health/#safety-plan"
                target="_blank"
                rel="noopener noreferrer"
            >
                <RideLabel class="border-blue-500 text-blue-500">
                    <svelte:component this={SafetyPlanIcons} />
                    <p class="text-white">Safety Plan</p>
                </RideLabel>
            </a>
        {/if}
    {/if}
</div>
