<script>
    import {
        ScrollArea,
        Scrollbar,
    } from "$lib/components/ui/scroll-area/index";
    import Card from "./card.svelte";

    export let rides = null;
    export let visible = false;
</script>

{#if visible && rides && rides.length > 0}
    <div
        class="absolute bottom-[50px] mb-14 left-0 w-full bg-transparent py-0 px-5 z-[1000]"
    >
        <ScrollArea>
            <div class="max-h-[35vh] w-full flex flex-col gap-1">
                {#each rides as ride (ride.id)}
                    <Card {ride} />
                {/each}
            </div>

            <Scrollbar orientation="veritcal" />
        </ScrollArea>
    </div>
{/if}
