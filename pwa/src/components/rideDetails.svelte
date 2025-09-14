<script>
    import * as Card from "$lib/components/ui/card";
    import {
        ScrollArea,
        Scrollbar,
    } from "$lib/components/ui/scroll-area/index";
    import { currentRide } from "$lib/stores";
    import { formatDate, formatTime } from "$lib/utils";

    const SHIFT2BIKES_IMG_URL = "https://www.shift2bikes.org/";

    $: {
        console.log($currentRide);
    }
</script>

{#if $currentRide}
    <div
        class="absolute top-[60px] bottom-[75px] min-h-[calc(100vh-115px)] left-0 w-full p-5 bg-black text-white overflow-hidden z-50"
    >
        <ScrollArea class="h-full w-full">
            <div class=" flex flex-col gap-5">
                <div
                    class="h-[400px] w-full bg-blue-500 flex items-center justify-center mx-auto text-5xl"
                >
                    MAP
                </div>
                <h2 class="text-3xl">{$currentRide.title}</h2>
                <p>{$currentRide.newsflash?.String}</p>
                <div class="flex flex-row">
                    {$currentRide.cancelled}
                    {$currentRide.audience}
                    {$currentRide.safetyplan}
                </div>

                <Card.Root class="bg-black text-white">
                    <Card.Header>
                        <Card.Description>Meetup Time</Card.Description>
                        <Card.Title class="text-2xl">
                            {formatTime($currentRide?.starttime)}
                            {formatDate($currentRide.date)}
                        </Card.Title>
                    </Card.Header>
                </Card.Root>

                <Card.Root class="bg-black text-white">
                    <Card.Header>
                        <Card.Description>Meetup Location</Card.Description>
                        <Card.Title class="text-2xl">
                            {$currentRide.venue?.String}</Card.Title
                        >
                        <Card.Description
                            >{$currentRide.address}</Card.Description
                        >
                        <Card.Description
                            >{$currentRide.loopride}</Card.Description
                        >
                    </Card.Header>
                    <Card.Content>
                        <div
                            class="h-96 w-full bg-blue-500 flex items-center justify-center mx-auto text-5xl"
                        >
                            MAP
                        </div>
                    </Card.Content>
                    <Card.Footer>{$currentRide.locdetails?.String}</Card.Footer>
                </Card.Root>
                {#if $currentRide.image.String != ""}
                    <img
                        src={SHIFT2BIKES_IMG_URL + $currentRide.image.String}
                        alt={`Image for ${$currentRide.title} bike ride`}
                    />
                {/if}

                <p class="text-lg">{$currentRide.details?.String}</p>
                <Card.Root class="bg-black text-white">
                    <Card.Header>
                        <Card.Title>{$currentRide.organizer?.String}</Card.Title
                        >
                    </Card.Header>
                </Card.Root>
            </div>
            <Scrollbar orientation="veritcal" />
        </ScrollArea>
    </div>
{/if}
