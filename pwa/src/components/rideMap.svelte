<script>
    import { Map, TileLayer, Marker, Popup, Tooltip } from "sveaflet";

    export let ride;

    const RIDE_COORDS = [ride.lat?.Float64, ride.lon?.Float64];

    const rideMapOptions = {
        center: RIDE_COORDS,
        zoom: 18,
        zoomControl: false,
        dragging: false,
        boxZoom: false,
        scrollWheelZoom: false,
        doubleClickZoom: false,
        touchZoom: false,
    };

    const markerOptions = {
        permanent: true,
    };

    const tileURL =
        "https://{s}.basemaps.cartocdn.com/dark_all/{z}/{x}/{y}{r}.png";
    const tileLayerOptions = {
        attribution:
            "Map tiles by Carto, under CC BY 3.0. Data by OpenStreetMap, under ODbL.",
    };
</script>

<Map options={rideMapOptions}>
    <TileLayer url={tileURL} options={tileLayerOptions} />
    <Marker options={markerOptions} latLng={RIDE_COORDS} />
</Map>
