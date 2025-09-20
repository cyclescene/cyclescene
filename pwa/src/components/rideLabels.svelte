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
        <RideLabel
            color="red-600"
            icon={CancelledIcon}
            content={ride.cancelled ? "Cancelled" : ""}
        />
    {/if}

    {#if ride.audience == "F"}
        <RideLabel
            color="green-600"
            icon={FamilyIcon}
            content="Family Friendly"
        />
    {:else if ride.audience == "A"}
        <RideLabel
            color="purple-600"
            content="Adults Only"
            icon={AdultsOnlyIcon}
        />
    {:else}{/if}

    {#if ride.safetyplan}
        {#if $activeView == VIEW_LIST}
            <RideLabel
                color="blue-600"
                content="Safety Plan"
                icon={SafetyPlanIcons}
            />
        {:else}
            <a
                href="https://www.shift2bikes.org/pages/public-health/#safety-plan"
                target="_blank"
                rel="noopener noreferrer"
            >
                <RideLabel
                    color="blue-600"
                    content="Safety Plan"
                    icon={SafetyPlanIcons}
                />
            </a>
        {/if}
    {/if}
</div>
