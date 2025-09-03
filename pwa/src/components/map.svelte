<script>
    import { Map, TileLayer, Tooltip, Marker, Popup } from "sveaflet";
    import "leaflet/dist/leaflet.css";

    import EventDetailsCard from "./eventDetailsCard.svelte";
    import { SvelteMap } from "svelte/reactivity";

    const ORIGINAL_MAP_CENTER = [45.52, -122.65];
    const ORIGINAL_MAP_ZOOM = 12;

    export let rides = [];
    let mapCenter = ORIGINAL_MAP_CENTER;
    let mapZoom = ORIGINAL_MAP_ZOOM;

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
        if (ridesAtLocation.length > 0) {
            mapCenter = [
                ridesAtLocation[0].lat.Float64,
                ridesAtLocation[0].lon.Float64,
            ];
        }
    }

    function recenterMap() {
        mapCenter = [...ORIGINAL_MAP_CENTER];
        mapZoom = ORIGINAL_MAP_ZOOM;
    }

    let selectedEvents = null;
    let showEventCard = false;

    function handleCardClose() {
        selectedEvents = null;
        showEventCard = false;
    }
</script>

<div class="map-container">
    <Map
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
    <button class="recenter-button" onclick={recenterMap}> ⟲ </button>
</div>

<EventDetailsCard
    events={selectedEvents}
    visible={showEventCard}
    on:close={handleCardClose}
/>

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
        height: 100%;
        width: 100%;
    }

    .recenter-button {
        position: absolute;
        top: 60px;
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
