<script>
  import { Map, TileLayer, Marker, Tooltip } from "sveaflet";
  import "leaflet/dist/leaflet.css";

  import L from "leaflet";
  import RidesNotShown from "./ride/ridesNotShown.svelte";
  import LocationCards from "./locationCards.svelte";
  import Button from "$lib/components/ui/button/button.svelte";
  import RecenterIcon from "~icons/material-symbols-light/recenter-rounded";
  import { mapViewStore, TILE_URLS } from "$lib/stores";
  import { mode } from "mode-watcher";
  import { STARTING_LAT, STARTING_LON } from "$lib/config";

  const ORIGINAL_MAP_ZOOM = 12;
  const SINGLE_RIDE_ZOOM = 15.5;

  export let rides = [];
  export let noAddressRides = [];

  // sveaflet map logic
  let sveafletMapInstance;

  function fitAllMarkers() {
    const locations = rides;

    // if no rides happening that day set the map zoom to the default location
    if (locations.length === 0) {
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
    if (locations.length === 1) {
      const singleLocation = locations[0];

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

    locations.forEach((ride) => {
      bounds.extend(L.latLng(ride.lat?.Float64, ride.lon?.Float64));
    });

    sveafletMapInstance.fitBounds(bounds, {
      padding: [60, 60],
      animate: true,
      duration: 0.8,
    });
  }

  $: if (sveafletMapInstance) {
    fitAllMarkers();
  }

  const tileLayerOptions = {
    attribution:
      "Map tiles by Carto, under CC BY 3.0. Data by OpenStreetMap, under ODbL.",
  };

  function handleMarkerClick(ridesAtLocation) {
    mapViewStore.setSelectedRides(ridesAtLocation);
    mapViewStore.showEventCards(true);
    mapViewStore.showOtherRides(false);
    if (ridesAtLocation.length > 0) {
      sveafletMapInstance.setView(
        [ridesAtLocation[0].lat.Float64, ridesAtLocation[0].lon.Float64],
        SINGLE_RIDE_ZOOM,
        {
          animate: true,
          duration: 0.8,
        },
      );
    }
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
    {#each rides as ride (ride.id)}
      <Marker
        latLng={[ride.lat?.Float64, ride.lon?.Float64]}
        onclick={() => handleMarkerClick(ride.rides)}
      />
      <Tooltip
        latLng={[ride.lat?.Float64, ride.lon?.Float64]}
        options={{
          permanent: true,
          direction: "bottom",
          className: "tool-tip",
          offset: [3, 10],
        }}
      >
        <p>
          {ride.title}
        </p>
      </Tooltip>
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
</style>
