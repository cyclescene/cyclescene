<script>
    import * as Card from "$lib/components/ui/card";
    import { navigateTo, setRide, VIEW_RIDE_DETAILS } from "$lib/stores";
    import { cn, formatTime } from "$lib/utils";

    export let ride;

    function onCardClick(ride) {
        navigateTo(VIEW_RIDE_DETAILS);
        setRide(ride);
    }
</script>

{#if ride}
    <Card.Root
        tabindex="0"
        role="button"
        class={cn(
            "mb-2 last:mb-0 bg-black text-white",
            "cursor-pointer transition-colors hover:bg-muted/50 focus-visible:outline-none focus-visible:ring-2, focus-visible:ring-ring focus-visible:ring-offset-2",
        )}
        onclick={() => onCardClick(ride)}
        on:keydown={(e) => onCardClick(ride)}
    >
        <Card.Header>
            <Card.Title>{ride.title}</Card.Title>
        </Card.Header>
        <Card.Content>
            <p>{ride.newsflash.String}</p>
            <p>{ride.venue.String}</p>
            <p>{formatTime(ride.starttime)}</p>
        </Card.Content>
    </Card.Root>
{/if}
