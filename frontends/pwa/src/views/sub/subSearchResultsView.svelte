<script lang="ts">
  import SearchIcon from "@lucide/svelte/icons/search";
  import XIcon from "@lucide/svelte/icons/x";
  import { getLocalTimeZone, parseDate, today } from "@internationalized/date";
  import { onDestroy, tick } from "svelte";
  import Card from "../../components/card.svelte";
  import { searchUpcomingRides } from "../../lib/api";
  import { rideSearchQuery } from "../../lib/stores";
  import type { RideData } from "../../lib/types";
  import { formatDate } from "../../lib/utils";

  let searchInput: HTMLInputElement;
  let searchResults: RideData[] = [];
  let searchLoading = false;
  let searchError = "";
  let searchTimer: ReturnType<typeof setTimeout> | undefined;
  let searchRequestID = 0;
  let lastQueuedQuery = "";

  $: trimmedSearchQuery = $rideSearchQuery.trim();
  $: resultCountLabel =
    searchResults.length === 1 ? "1 result" : `${searchResults.length} results`;
  $: emptyMessage =
    trimmedSearchQuery.length === 0
      ? "Search upcoming rides"
      : trimmedSearchQuery.length < 2
        ? "Keep typing to search upcoming rides"
        : searchLoading
          ? "Searching upcoming rides..."
          : searchError
            ? searchError
            : "No upcoming rides match your search";
  $: if (trimmedSearchQuery !== lastQueuedQuery) {
    lastQueuedQuery = trimmedSearchQuery;
    queueSearch(trimmedSearchQuery);
  }

  tick().then(() => searchInput?.focus());

  onDestroy(() => {
    if (searchTimer) {
      clearTimeout(searchTimer);
    }
  });

  function queueSearch(query: string) {
    if (searchTimer) {
      clearTimeout(searchTimer);
    }

    searchError = "";

    if (query.length === 0 || query.length < 2) {
      searchResults = [];
      searchLoading = false;
      return;
    }

    searchLoading = true;
    const requestID = ++searchRequestID;

    searchTimer = setTimeout(async () => {
      try {
        const results = await searchUpcomingRides(query);
        if (requestID === searchRequestID) {
          searchResults = results;
          searchError = "";
        }
      } catch (error) {
        if (requestID === searchRequestID) {
          searchResults = [];
          searchError = "Search failed. Try again.";
        }
      } finally {
        if (requestID === searchRequestID) {
          searchLoading = false;
        }
      }
    }, 250);
  }

  function getDaysOutLabel(dateString: string) {
    const timezone = getLocalTimeZone();
    const rideDate = parseDate(dateString);
    const todayDate = today(timezone);
    const dayInMs = 24 * 60 * 60 * 1000;
    const daysOut = Math.round(
      (rideDate.toDate(timezone).getTime() - todayDate.toDate(timezone).getTime()) / dayInMs,
    );

    if (daysOut === 0) {
      return "Today";
    }
    if (daysOut === 1) {
      return "Tomorrow";
    }
    return `In ${daysOut} days`;
  }

  function clearSearch() {
    rideSearchQuery.set("");
    searchInput?.focus();
  }
</script>

<div class="flex h-full min-h-0 flex-col">
  <div class="border-b bg-background px-5 py-3">
    <div class="relative">
      <SearchIcon
        class="pointer-events-none absolute left-3 top-1/2 size-5 -translate-y-1/2 text-muted-foreground"
        aria-hidden="true"
      />
      <input
        bind:this={searchInput}
        bind:value={$rideSearchQuery}
        type="search"
        placeholder="Search upcoming rides"
        class="search-input h-11 w-full rounded-md border bg-background px-10 text-base outline-none transition-colors placeholder:text-muted-foreground focus:border-yellow-400 focus:ring-2 focus:ring-yellow-400/30"
      />
      {#if $rideSearchQuery}
        <button
          type="button"
          class="absolute right-2 top-1/2 flex size-8 -translate-y-1/2 items-center justify-center rounded-md text-muted-foreground hover:bg-muted hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-yellow-400/40"
          aria-label="Clear search"
          onclick={clearSearch}
        >
          <XIcon class="size-5" aria-hidden="true" />
        </button>
      {/if}
    </div>
  </div>

  <div class="min-h-0 flex-1 overflow-y-auto p-5 pb-[calc(var(--footer-height)_+_env(safe-area-inset-bottom)_+_10px)]">
    {#if trimmedSearchQuery.length >= 2 && !searchLoading && !searchError}
      <p class="mb-3 text-sm text-muted-foreground">{resultCountLabel}</p>
    {/if}

    <div class="flex flex-col gap-2">
      {#if searchResults.length > 0}
        {#each searchResults as ride (ride.id)}
          <div class="mt-2 flex items-center justify-between text-sm text-muted-foreground">
            <span>{formatDate(ride.date)}</span>
            <span>{getDaysOutLabel(ride.date)}</span>
          </div>
          <Card {ride} />
        {/each}
      {:else}
        <div class="pt-8 text-center text-muted-foreground">{emptyMessage}</div>
      {/if}
    </div>
  </div>
</div>

<style>
  .search-input::-webkit-search-decoration,
  .search-input::-webkit-search-cancel-button,
  .search-input::-webkit-search-results-button,
  .search-input::-webkit-search-results-decoration {
    display: none;
  }
</style>
