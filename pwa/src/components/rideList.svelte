<script>
    import { ScrollArea } from "$lib/components/ui/scroll-area/index";
    import Card from "./card.svelte";

    export let rides = [];
</script>

<div
    class="absolute top-[60px] bottom-[75px] min-h-[calc(100vh_-_115px)] w-full p-5 bg-black"
>
    <ScrollArea>
        <div class="h-[calc(100vh_-_160px)] flex flex-col gap-2">
            {#if rides && rides.length > 0}
                {#each rides as ride (ride.id)}
                    <Card {ride} />
                {/each}
            {:else}
                No rides happening today :(
            {/if}
        </div>
    </ScrollArea>
</div>
