<script>
    export let rides = [];
</script>

{#if rides && rides.length > 1}
    <pre>{JSON.stringify(rides, null, 2)}</pre>
{:else}
    <p>Nothing to see here</p>
{/if}
