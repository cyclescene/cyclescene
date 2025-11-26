<script lang="ts">
import { RawLayer } from "svelte-maplibre-gl";

interface Props {
  isDarkMode: boolean;
}

let { isDarkMode }: Props = $props();

// Dark mode colors for parks
const darkParkColor = "#4a7c59";
const darkParkOpacity = 0.6;
</script>

{#if isDarkMode}
  <!-- Grass/Parks layer - only show in dark mode, positioned before labels -->
  <RawLayer
    id="parks-grass"
    source="carto"
    source-layer="landcover"
    type="fill"
    paint={{
      "fill-color": darkParkColor,
      "fill-opacity": darkParkOpacity,
    }}
    filter={["==", ["get", "class"], "grass"]}
    beforeId="poi_stadium"
  />

  <!-- Cemeteries and Stadiums layer - only show in dark mode, positioned before labels -->
  <RawLayer
    id="parks-landuse"
    source="carto"
    source-layer="landuse"
    type="fill"
    paint={{
      "fill-color": darkParkColor,
      "fill-opacity": darkParkOpacity,
    }}
    filter={[
      "any",
      ["==", ["get", "class"], "cemetery"],
      ["==", ["get", "class"], "stadium"],
    ]}
    beforeId="poi_stadium"
  />
{/if}
