<script>
    import { Map, TileLayer, Marker, Popup } from "sveaflet";
    import "leaflet/dist/leaflet.css";

    import { SvelteMap } from "svelte/reactivity";

    import L from "leaflet";
    import RidesNotShown from "./ride/ridesNotShown.svelte";
    import LocationCards from "./locationCards.svelte";
    import Button from "$lib/components/ui/button/button.svelte";
    import RecenterIcon from "~icons/material-symbols-light/recenter-rounded";

    const ORIGINAL_MAP_CENTER = [45.52, -122.65];
    const ORIGINAL_MAP_ZOOM = 12;

    let mapCenter = ORIGINAL_MAP_CENTER;
    let mapZoom = ORIGINAL_MAP_ZOOM;

    export let rides = [];
    export let noAddressRides = [];

    // group locations logic

    let groupedLocations = [];

    $: {
        let ridesByLocation = new SvelteMap();
        rides.forEach((ride) => {
            const key = `${ride.lat.Float64},${ride.lon.Float64}`;
            if (!ridesByLocation.has(key)) {
                ridesByLocation.set(key, {
                    venue: ride.venue.String,
                    lat: ride.lat.Float64,
                    lng: ride.lon.Float64,
                    rides: [],
                });
            }
            ridesByLocation.get(key).rides.push(ride);
        });

        groupedLocations = Array.from(ridesByLocation.values());
    }

    // sveaflet map logic
    let sveafletMapInstance;

    function fitAllMarkers() {
        if (sveafletMapInstance && groupedLocations.length > 1) {
            if (groupedLocations.length == 1) {
                const singleLocation = groupedLocations[0];
                sveafletMapInstance.setView(
                    L.LatLngBounds(singleLocation.lat, singleLocation.lng),
                    ORIGINAL_MAP_ZOOM,
                    { animate: true, duration: 0.8 },
                );
            } else {
                const bounds = new L.LatLngBounds();

                groupedLocations.forEach((group) => {
                    bounds.extend(L.latLng(group.lat, group.lng));
                });

                sveafletMapInstance.fitBounds(bounds, {
                    padding: [60, 60],
                    animate: true,
                    duration: 0.8,
                });
            }
        } else if (sveafletMapInstance && groupedLocations.length == 1) {
            mapCenter = [...ORIGINAL_MAP_CENTER];
            mapZoom = ORIGINAL_MAP_ZOOM;
        }
    }

    $: if (groupedLocations) {
        fitAllMarkers();
    }

    // Add the tile url as a store that can be changed by the user
    // light url - https://{s}.basemaps.cartocdn.com/voyager_labels_under/{z}/{x}/{y}{r}.png
    // dark url - https://{s}.basemaps.cartocdn.com/dark_all/{z}/{x}/{y}{r}.png

    const tileURL =
        "https://{s}.basemaps.cartocdn.com/dark_all/{z}/{x}/{y}{r}.png";
    const tileLayerOptions = {
        attribution:
            "Map tiles by Carto, under CC BY 3.0. Data by OpenStreetMap, under ODbL.",
    };

    function handleMarkerClick(ridesAtLocation) {
        selectedEvents = ridesAtLocation;
        showEventCard = true;
        showNotShown = false;
        if (ridesAtLocation.length > 0) {
            mapCenter = [
                ridesAtLocation[0].lat.Float64,
                ridesAtLocation[0].lon.Float64,
            ];
        }
    }

    let selectedEvents = null;
    let showEventCard = false;

    function handleRecenter() {
        fitAllMarkers();
        sveafletMapInstance.closePopup();
        selectedEvents = null;
        showEventCard = false;
        if (noAddressRides && noAddressRides.length > 1) {
            showNotShown = true;
        }
    }

    function handleCardClose() {
        selectedEvents = null;
        showEventCard = false;
        if (noAddressRides && noAddressRides.length > 1) {
            showNotShown = true;
        }
    }

    let showNotShown = false;

    $: if (noAddressRides && noAddressRides.length > 1) {
        showNotShown = true;
    }
</script>

<div class="map-container">
    <Map
        bind:instance={sveafletMapInstance}
        options={{
            center: mapCenter,
            zoom: mapZoom,
            zoomControl: false,
        }}
        onclick={handleCardClose}
    >
        <TileLayer url={tileURL} options={tileLayerOptions} />
        {#each groupedLocations as group (group.lat + "," + group.lng)}
            <Marker
                latLng={[group.lat, group.lng]}
                onclick={() => handleMarkerClick(group.rides)}
            >
                <Popup>
                    <strong
                        >{group.venue} - {group.rides.length} ride{group.rides
                            .length !== 1
                            ? "s"
                            : ""} starting here</strong
                    >
                </Popup>
            </Marker>
        {/each}
    </Map>
    <Button
        class="absolute top-[85px] bg-black text-white h-10 w-10 z-[1000] right-2.5"
        onclick={handleRecenter}
    >
        <RecenterIcon style="width: 30px; height: 30px;" />
    </Button>
</div>

<LocationCards
    rides={selectedEvents}
    visible={showEventCard}
    on:close={handleCardClose}
/>

<RidesNotShown visible={showNotShown} notShownLength={noAddressRides.length} />

<style>
    :global(.map-container .leaflet-tooltip) {
        background: transparent !important;
        border: none !important;
        box-shadow: none !important;
        font-weight: bold !important;
        color: #000 !important;
        text-shadow:
            1px 1px 2px white,
            -1px -1px 2px white,
            1px -1px 2px white,
            -1px 1px 2px white !important;
    }

    .map-container {
        height: calc(100% - 115px);
        width: 100%;
        margin-top: 60px;
        margin-bottom: 50px;
    }

    .recenter-icon {
        height: 32px;
        width: 32px;
    }

    .recenter-button {
        position: absolute;
        top: 70px;
        right: 10px;
        z-index: 400;
        color: #000;
        background-color: white;
        border: 2px solid #ccc;
        border-radius: 4px;
        padding: 8px 12px;
        font-size: 1.2em;
        cursor: pointer;
        box-shadow: 0 1px 5px rgba(0, 0, 0, 0.4);
        transition: background-color 0.2s;
    }

    .recenter-button:hover {
        background-color: #f4f4f4;
    }
</style>
