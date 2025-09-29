<script>
  import {
    activeView,
    allUpcomingAdultOnlyRides,
    allUpcomingCovidSafetyRides,
    allUpcomingFamilyFriendlyRides,
    SUB_VIEW_ADULT_ONLY_RIDES,
    SUB_VIEW_COVID_SAFETY_RIDES,
    SUB_VIEW_FAMILY_FRIENDLY_RIDES,
  } from "$lib/stores";
  import RideList from "../../components/ride/rideList.svelte";
</script>

{#if $activeView === SUB_VIEW_ADULT_ONLY_RIDES}
  <RideList rides={$allUpcomingAdultOnlyRides} />
{:else if $activeView === SUB_VIEW_FAMILY_FRIENDLY_RIDES}
  <RideList rides={$allUpcomingFamilyFriendlyRides} />
{:else if $activeView === SUB_VIEW_COVID_SAFETY_RIDES}
  <RideList rides={$allUpcomingCovidSafetyRides} />
{:else}
  Hello
{/if}
