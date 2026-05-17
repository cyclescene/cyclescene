<script lang="ts">
  import SearchIcon from "@lucide/svelte/icons/search";
  import RideList from "../components/ride/rideList.svelte";
  import { ENABLE_RIDE_SEARCH } from "../lib/config";
  import { navigateTo, rideSearchQuery, SUB_VIEW_SEARCH_RESULTS, todaysRides } from "../lib/stores";

  function openSearch() {
    rideSearchQuery.set("");
    navigateTo(SUB_VIEW_SEARCH_RESULTS);
  }
</script>

<div class="flex h-full min-h-0 flex-col">
  {#if ENABLE_RIDE_SEARCH}
    <div class="border-b bg-background px-5 py-3">
      <button
        type="button"
        class="relative h-11 w-full rounded-md border bg-background px-10 text-left text-base text-muted-foreground outline-none transition-colors hover:bg-muted/40 focus-visible:border-yellow-400 focus-visible:ring-2 focus-visible:ring-yellow-400/30"
        onclick={openSearch}
      >
        <SearchIcon
          class="pointer-events-none absolute left-3 top-1/2 size-5 -translate-y-1/2 text-muted-foreground"
          aria-hidden="true"
        />
        Search upcoming rides
      </button>
    </div>
  {/if}

  <div class="min-h-0 flex-1">
    <RideList rides={$todaysRides} />
  </div>
</div>
