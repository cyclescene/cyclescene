<script>
  import Button from "$lib/components/ui/button/button.svelte";
  import {
    navigateTo,
    VIEW_LIST,
    VIEW_MAP,
    VIEW_SAVED,
    VIEW_SETTINGS,
  } from "../lib/stores";

  import MapIcon from "~icons/solar/map-linear";
  import StackIcon from "~icons/picon/stacked";
  import SavedIcon from "~icons/material-symbols/bookmark-sharp";
  import SettingsIcon from "~icons/material-symbols/settings";

  const controls = [
    {
      name: "Map",
      icon: MapIcon,
      view: VIEW_MAP,
    },
    {
      name: "List",
      icon: StackIcon,
      view: VIEW_LIST,
    },
    {
      name: "Saved",
      icon: SavedIcon,
      view: VIEW_SAVED,
    },
    {
      name: "Settings",
      icon: SettingsIcon,
      view: VIEW_SETTINGS,
    },
  ];

  function onChangeView(view) {
    navigateTo(view);
  }
</script>

<div class="w-full h-full bg-background flex items-center justify-evenly z-[5]">
  {#each controls as view (view.name)}
    <Button
      disabled={false}
      variant="ghost"
      class="grow h-full rounded-none flex flex-col"
      onclick={() => onChangeView(view.view)}
    >
      <div class="h-5 w-5">
        <svelte:component this={view.icon} />
      </div>
      <p class="text-xs">
        {view.name}
      </p>
    </Button>
  {/each}
</div>
