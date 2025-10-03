<script lang="ts">
  import { SymbolLayer } from "svelte-maplibre-gl";
  import { mode } from "mode-watcher";

  const { sourceId, iconName }: { sourceId: string; iconName: string } =
    $props();

  let textColor = $derived(mode.current === "dark" ? "#ffffff" : "#000000");
</script>

<SymbolLayer
  id="ride-icons"
  source={sourceId}
  layout={{
    "icon-image": iconName,
    "icon-size": 0.4,
    "icon-allow-overlap": true,
  }}
  paint={{
    "icon-color": "#0000ff",
  }}
/>
<SymbolLayer
  id="ride-labels"
  source={sourceId}
  layout={{
    "text-field": ["get", "name"],
    "text-font": ["Open Sans Regular"],
    "text-size": 12,
    "text-offset": [0, 2],
    "text-anchor": "top",

    "text-allow-overlap": false,
    "icon-allow-overlap": true,
    "icon-ignore-placement": true,
  }}
  paint={{
    "text-color": textColor,
    "text-halo-color": "#ffffff",
    "icon-halo-width": 1,
  }}
/>
