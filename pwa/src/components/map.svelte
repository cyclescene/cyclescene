<script>
    import { LeafletMap, Marker, TileLayer, Tooltip } from "svelte-leafletjs";
    import "leaflet/dist/leaflet.css";

    export let rides = [];

    const mapOptions = {
        center: [45.52, -122.65],
        zoom: 13,
        zoomControl: false,
    };

    const tileURL = "https://tile.openstreetmap.org/{z}/{x}/{y}.png";
    const tileLayerOptions = {
        attribution:
            '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors',
    };
</script>

<div class="map-container">
    <LeafletMap options={mapOptions}>
        <TileLayer url={tileURL} options={tileLayerOptions} />
        {#each rides as ride (ride.id)}
            {#if ride.lat?.Valid && ride.lon?.Valid}
                <Marker latLng={[ride.lat.Float64, ride.lon.Float64]}>
                    <Tooltip
                        options={{
                            permanent: true,
                            direction: "bottom",
                            offset: [-15, 25],
                        }}
                    >
                        {ride.title}
                    </Tooltip>
                    <!-- <Popup> -->
                    <!--     {ride.title} -->
                    <!-- </Popup> -->
                </Marker>
            {/if}
        {/each}
    </LeafletMap>
</div>

<style>
    :global(.leaflet-tile) {
        -webkit-filter: hue-rotate(180deg) invert(100%);
    }

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
</style>
