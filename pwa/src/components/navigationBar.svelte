<script>
    import { currentView } from "../lib/stores";

    const controls = ["map", "list", "saved", "settings"];

    function onViewChange(view) {
        $currentView = view;
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
