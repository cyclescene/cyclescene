<script>
    export let rides = [];
</script>

{#if rides && rides.length > 0}
    <pre>{JSON.stringify(rides, null, 2)}</pre>
{:else}
    No rides happening today :(
{/if}
