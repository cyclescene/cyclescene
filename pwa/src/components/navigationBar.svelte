<script>
    import {
        navigateTo,
        VIEW_LIST,
        VIEW_MAP,
        VIEW_SAVED,
        VIEW_SETTINGS,
    } from "../lib/stores";

    const controls = [VIEW_MAP, VIEW_LIST, VIEW_SAVED, VIEW_SETTINGS];

    function onViewChange(view) {
        navigateTo(view);
    }
</script>

<div class="navigation-controls-container">
    {#each controls as control}
        <button onclick={() => onViewChange(control)}
            >{control.charAt(0).toLocaleUpperCase() + control.slice(1)}</button
        >
    {/each}
</div>

<style>
    .navigation-controls-container {
        width: 100%;
        height: 100%;
        display: flex;
        justify-content: space-evenly;
    }

    button {
        flex: 1;
    }
</style>
