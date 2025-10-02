<script>
  import { Map, TileLayer, Marker, Tooltip } from "sveaflet";
  import "leaflet/dist/leaflet.css";

  import L from "leaflet";
  import RidesNotShown from "./ride/ridesNotShown.svelte";
  import LocationCards from "./locationCards.svelte";
  import Button from "$lib/components/ui/button/button.svelte";
  import RecenterIcon from "~icons/material-symbols-light/recenter-rounded";
  import { allRides, mapViewStore, TILE_URLS } from "$lib/stores";
  import { mode } from "mode-watcher";
  import { STARTING_LAT, STARTING_LON } from "$lib/config";
  import MapMarker from "./mapMarker.svelte";
  import { onDestroy } from "svelte";

  const ORIGINAL_MAP_ZOOM = 12;
  const COLLISION_THRESHOLD_PIXELS = 50;
  const SINGLE_RIDE_ZOOM = 15.5;

  export let noAddressRides = [];

  // sveaflet map logic
  let sveafletMapInstance;
  let leafletMarkers = {};
  // --- COLLISION LOGIC FUNCTION ---
  function checkLabelCollisions() {
    const map = sveafletMapInstance;

    // Convert the object values into a simple array for iteration
    const markerArray = Object.values(leafletMarkers || {});
    console.log(markerArray);

    // Safety check: Exit if the map is not ready or has too few markers
    if (!map || markerArray.length < 2) {
      // Run initial cleanup here to ensure any single marker is fully visible
      markerArray.forEach((marker) => {
        const element = marker.getElement();
        if (element) {
          element.classList.remove("marker-hidden-label");
        }
      });
      return;
    }

    // 1. Initial Visibility Cleanup (Must be run at the start to clear old hidden state)
    // We make all labels visible *before* starting the new collision check.
    markerArray.forEach((marker) => {
      const element = marker.getElement();
      // CRITICAL CHECK: Only attempt to manipulate the DOM if the element exists
      if (element) {
        element.classList.remove("marker-hidden-label");
      }
    });

    // 2. Iterate and check for overlaps (Loop runs over the temporary array)
    for (let i = 0; i < markerArray.length; i++) {
      const markerA = markerArray[i];
      const elementA = markerA.getElement();

      // CRITICAL CHECK: Skip if elementA is not attached to the DOM (due to Leaflet updates)
      if (!elementA) continue;

      // Convert Marker A's geographic location to a pixel coordinate on the screen
      const pointAPixel = map.latLngToContainerPoint(markerA.getLatLng());

      for (let j = i + 1; j < markerArray.length; j++) {
        const markerB = markerArray[j];
        const elementB = markerB.getElement();

        // CRITICAL CHECK: Skip if elementB is not attached to the DOM
        if (!elementB) continue;

        const pointBPixel = map.latLngToContainerPoint(markerB.getLatLng());

        // Calculate the distance between the two points in pixels
        const distance = pointAPixel.distanceTo(pointBPixel);

        if (distance < COLLISION_THRESHOLD_PIXELS) {
          // COLLISION FOUND: Apply "Neither Shall Speak" rule

          // 1. Hide the label of Marker A (the winner)
          elementA.classList.add("marker-hidden-label");

          // 2. Hide the label of Marker B (the loser)
          elementB.classList.add("marker-hidden-label");
        }
      }
    }
  }

  function debounce(func, delay) {
    let timeoutId;
    return function (...args) {
      clearTimeout(timeoutId);
      timeoutId = setTimeout(() => func.apply(this, args), delay);
    };
  }

  const debouncedCheckLabelCollisions = debounce(checkLabelCollisions, 50);

  function fitAllMarkers() {
    const rides = $allRides;

    // if no rides happening that day set the map zoom to the default location
    if (rides.length === 0) {
      sveafletMapInstance.setView(
        [STARTING_LAT, STARTING_LON],
        ORIGINAL_MAP_ZOOM,
        {
          animate: true,
          duration: 0.8,
        },
      );
      return;
    }

    // if there is only one ride close in on it
    if (rides.length === 1) {
      const singleLocation = rides[0];

      sveafletMapInstance.setView(
        [singleLocation.lat?.Float64, singleLocation.lon.Float64],
        SINGLE_RIDE_ZOOM,
        {
          animate: true,
          duration: 0.8,
        },
      );
      return;
    }

    // when there are many rides happening on a day create a bounds uisng location details and
    // display them all within the map
    const bounds = new L.LatLngBounds();

    rides.forEach((ride) => {
      bounds.extend(L.latLng(ride.lat?.Float64, ride.lon?.Float64));
    });

    sveafletMapInstance.fitBounds(bounds, {
      padding: [60, 60],
      animate: true,
      duration: 0.8,
    });
  }

  $: if (sveafletMapInstance && $allRides) {
    fitAllMarkers();
    debouncedCheckLabelCollisions();
  }

  const tileLayerOptions = {
    attribution:
      "Map tiles by Carto, under CC BY 3.0. Data by OpenStreetMap, under ODbL.",
  };

  function handleMarkerClick(ride) {
    console.log(ride);
    mapViewStore.setSelectedRide(ride);
    mapViewStore.showEventCards(true);
    mapViewStore.showOtherRides(false);
    sveafletMapInstance.setView(
      [ride.lat.Float64, ride.lon.Float64],
      SINGLE_RIDE_ZOOM,
      {
        animate: true,
        duration: 0.8,
      },
    );
  }

  function setMarkerInstance(rideId, markerInstance) {
    leafletMarkers[rideId] = markerInstance;
  }

  $: if (sveafletMapInstance) {
    const map = sveafletMapInstance;
    // Listen to map stop events
    map.on("zoomend", debouncedCheckLabelCollisions);
    map.on("moveend", debouncedCheckLabelCollisions);

    // Ensure cleanup (needed for non-SvelteKit apps)
    onDestroy(() => {
      map.off("zoomend", checkLabelCollisions);
      map.off("moveend", checkLabelCollisions);
    });
  }

  function handleRecenter() {
    fitAllMarkers();
    sveafletMapInstance.closePopup();
    mapViewStore.clearSelectedRides();
    mapViewStore.showEventCards(false);
    if (noAddressRides && noAddressRides.length > 1) {
      mapViewStore.showOtherRides(true);
    }
  }

  function handleCardClose() {
    fitAllMarkers();
    mapViewStore.showEventCards(false);
    mapViewStore.clearSelectedRides();
    if (noAddressRides && noAddressRides.length > 1) {
      mapViewStore.showOtherRides(true);
    }
  }

  $: if (noAddressRides && noAddressRides.length > 1) {
    mapViewStore.showOtherRides(true);
  }
</script>

<div class="map-container">
  <Map
    bind:instance={sveafletMapInstance}
    options={{
      zoomControl: false,
    }}
    onclick={handleCardClose}
  >
    {#if mode.current === "dark"}
      <TileLayer url={TILE_URLS.dark} options={tileLayerOptions} />
    {:else}
      <TileLayer url={TILE_URLS.light} options={tileLayerOptions} />
    {/if}
    {#each $allRides as ride (ride.id)}
      <MapMarker
        mapInstance={sveafletMapInstance}
        {ride}
        onMarkerClick={handleMarkerClick}
        onSetInstance={(instance) => setMarkerInstance(ride.id, instance)}
      />
    {/each}
  </Map>
  <Button
    disabled={false}
    class={`absolute top-[85px] h-10 w-10 z-[1000] right-2.5`}
    variant="secondary"
    onclick={handleRecenter}
  >
    <RecenterIcon style="width: 30px; height: 30px;" />
  </Button>
</div>

<LocationCards on:close={handleCardClose} />

<RidesNotShown />

<style>
  .map-container {
    height: calc(100% - 115px);
    width: 100%;
    margin-top: 60px;
    margin-bottom: 50px;
  }

  /* 1. Target the Outer Tooltip (.tool-tip) */
  :global(.map-container .leaflet-tooltip-pane .leaflet-tooltip.tool-tip) {
    /* Set the maximum width the ENTIRE container is allowed to grow to */
    white-space: normal !important;
    max-width: 200px !important; /* Adjusted slightly lower for better mobile fit */
    width: auto !important;

    /* Transparency and box removal */
    background-color: transparent !important;
    border: none !important;
    padding: 0 !important;
    margin: 0 !important;
    box-shadow: none !important;
  }

  /* 2. Target the Leaflet Content Wrapper (.leaflet-tooltip-content) */
  /* This is the inner Leaflet-generated div that often holds the hidden 'nowrap' */
  :global(.map-container .leaflet-tooltip-bottom) {
    white-space: normal !important; /* CRITICAL: Must be normal */
    max-width: 150px !important; /* This is the desired wrapping width */
    width: 100% !important;
    padding: 0 !important;
    margin: 0 !important;
    text-align: center !important; /* Center the text block itself */
  }

  /* 3. Target your <p> Tag (The actual content) */
  /* This is the final container before the text */
  :global(.map-container .leaflet-tooltip-pane .leaflet-tooltip.tool-tip p) {
    white-space: normal !important;
    width: 200px !important;
    margin: 0 !important;
    padding: 0 !important;

    /* Text appearance styles */
    color: var(--color-foreground);
    font-weight: 700; /* Add bold styling here since <strong> was removed */
    line-height: 1.2;
    font-size: 10px;
  }

  /* 4. Hide the Arrow */
  :global(
      .map-container
        .leaflet-tooltip-pane
        .leaflet-tooltip-bottom.tool-tip::before
    ) {
    content: none !important;
    border-width: 0 !important;
  }

  :global(.marker-hidden-label .marker-label) {
    opacity: 0 !important;
  }

  /* Ensure the pin remains visible when the label is hidden */
  :global(.marker-hidden-label .marker-icon-pin) {
    opacity: 1 !important;
  }
</style>
