<script lang="ts">
  import {
    mapStore,
    navigateTo,
    ridesWithoutLocations,
    VIEW_OTHER_RIDES,
  } from "$lib/stores";

  function onShowNotShown() {
    navigateTo(VIEW_OTHER_RIDES);
  }
</script>

{#if $mapStore.showNoLocationRideCard && $ridesWithoutLocations.length > 0}
  <div class="event-not-shown-card">
    <button class="ride-not-shown-button" onclick={() => onShowNotShown()}>
      <p>
        {$ridesWithoutLocations.length} ride{$ridesWithoutLocations.length == 1
          ? ""
          : "s"} not shown
      </p>
    </button>
  </div>
{/if}

<style>
  .event-not-shown-card {
    position: absolute;
    bottom: 75px;
    right: 0;
    width: 55%;
    max-height: 35vh;
    background: transparent;
    color: lightgray;
    border: 5px lightgray;
    box-shadow: 0 -2px 10px rgba(0, 0, 0, 0.1);
    padding: 0px 20px 20px 20px;
    box-sizing: border-box;
    z-index: 1000;
    transform: translateY(0%);
    transition: transform 0.3s ease-out;
    overflow-y: auto;
    overflow-x: hidden;
  }

  .ride-not-shown-button {
    all: unset;
    cursor: pointer;
    border: none;
    display: block;
    box-sizing: border-box;

    width: 100%;
    border-radius: 25px;
    background-color: #242424;
    padding: 15px;
  }
  .ride-not-shown-button p {
    margin-top: 0;
    margin-bottom: 5px;
    font-size: 1.3em;
  }
</style>
