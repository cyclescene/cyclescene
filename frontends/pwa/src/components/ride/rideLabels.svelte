<script lang="ts">
  import { activeView, VIEW_LIST } from "$lib/stores";
  import RideLabel from "./rideLabel.svelte";

  import FamilyIcon from "~icons/ic/round-family-restroom";
  import CancelledIcon from "~icons/gridicons/cross-circle";
  import AdultsOnlyIcon from "~icons/uil/21-plus";
  import SafetyPlanIcon from "~icons/f7/facemask-fill";
  import LoopIcon from "~icons/qlementine-icons/loop-16";
  import type { RideData } from "$lib/types";

  let { ride }: { ride: RideData } = $props();
</script>

<div class="mt-2 flex flex-row gap-2.5">
  {#if ride.cancelled}
    <RideLabel class="border-red-500 text-red-500">
      <CancelledIcon />
      <p class="text-card-foreground">Cancelled</p>
    </RideLabel>
  {/if}

  {#if ride.audience == "F"}
    <RideLabel class="border-green-500 text-green-500">
      <FamilyIcon />
      <p class="text-card-foreground">Family Friendly</p>
    </RideLabel>
  {:else if ride.audience == "A"}
    <RideLabel class="border-purple-500 text-purple-500">
      <AdultsOnlyIcon />
      <p class="text-card-foreground">Adults Only</p>
    </RideLabel>
  {:else}{/if}

  {#if ride.safetyplan}
    {#if $activeView == VIEW_LIST}
      <RideLabel class="border-blue-500 text-blue-500">
        <SafetyPlanIcon />
        <p class="text-card-foreground">Safety Plan</p>
      </RideLabel>
    {:else}
      <a
        href="https://www.shift2bikes.org/pages/public-health/#safety-plan"
        target="_blank"
        rel="noopener noreferrer"
      >
        <RideLabel class="border-blue-500 text-blue-500">
          <SafetyPlanIcon />
          <p class="text-card-foreground">Safety Plan</p>
        </RideLabel>
      </a>
    {/if}
  {/if}

  {#if ride.loopride}
    <RideLabel class="border-orange-500 text-orange-500">
      <LoopIcon style="transform: rotate(90deg);" />
      <p class="text-card-foreground">Loop Ride</p>
    </RideLabel>
  {:else}{/if}
</div>
