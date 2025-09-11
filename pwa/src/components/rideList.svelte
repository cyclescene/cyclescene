<script>
    import Card from "./card.svelte";

    export let rides = [];
</script>

<div class="card-list-container">
    {#if rides && rides.length > 0}
        {#each rides as ride (ride.id)}
            <Card {ride} />
        {/each}
    {:else}
        No rides happening today :(
    {/if}
</div>

<style>
    .card-list-container {
        position: absolute;
        height: calc(100% - 105px);
        width: 100%;
        margin-top: 75px;
        margin-bottom: 100px;
        box-sizing: border-box;

        padding: 20px;

        overflow-y: scroll;
    }
</style>
