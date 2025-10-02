<script>
  import { DivIcon, Marker } from "sveaflet";

  const { ride, onMarkerClick, onSetInstance } = $props();

  let marker = $state(null);

  $effect(() => {
    if (marker && onSetInstance) {
      onSetInstance(marker);
    }
  });
</script>

<Marker
  latLng={[ride.lat.Float64, ride.lon.Float64]}
  onclick={() => onMarkerClick(ride)}
  bind:instance={marker}
>
  <DivIcon>
    {ride.title}
  </DivIcon>
</Marker>
