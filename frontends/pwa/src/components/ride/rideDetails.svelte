<script lang="ts">
  import Button from "$lib/components/ui/button/button.svelte";
  import * as Card from "$lib/components/ui/card";
  import { ScrollArea } from "$lib/components/ui/scroll-area/index";
  import { currentRide } from "$lib/stores";
  import { formatDate, formatTime } from "$lib/utils";
  import RideLabels from "./rideLabels.svelte";
  import RideMap from "./rideMap.svelte";

  const SHIFT2BIKES_URL = "https://www.shift2bikes.org/";
</script>

{#if $currentRide}
  <div
    class="absolute top-[60px] bottom-[75px] min-h-[calc(100vh-115px)] left-0 w-full p-5 overflow-hidden z-50"
  >
    <ScrollArea class="h-full w-full" scrollbarYClasses={`hidden`}>
      <div class=" flex flex-col gap-5">
        <!-- <div -->
        <!--   class="h-[400px] w-full bg-blue-500 flex items-center justify-center mx-auto text-5xl" -->
        <!-- > -->
        <RideMap ride={$currentRide} />
        <!-- </div> -->
        <h2 class="text-3xl">{$currentRide.title}</h2>
        <p>{$currentRide.newsflash?.String}</p>
        <RideLabels ride={$currentRide} />

        <Card.Root>
          <Card.Header>
            <Card.Description>Meetup Time</Card.Description>
            <Card.Title class="text-2xl">
              {formatTime($currentRide?.starttime)}
              {formatDate($currentRide.date)}
            </Card.Title>
          </Card.Header>
        </Card.Root>

        <Card.Root>
          <Card.Header>
            <Card.Description>Meetup Location</Card.Description>
            <Card.Title class="text-2xl">
              {$currentRide.venue?.String}</Card.Title
            >
            <Card.Description>{$currentRide.address}</Card.Description>
            <Card.Description
              >{$currentRide.loopride
                ? "Ride is a loop"
                : "Ride not a loop"}</Card.Description
            >
          </Card.Header>

          {#if $currentRide.locdetails?.String != ""}
            <Card.Footer>{$currentRide.locdetails?.String}</Card.Footer>
          {/if}
        </Card.Root>

        {#if $currentRide.image.String != ""}
          <img
            src={SHIFT2BIKES_URL + $currentRide.image.String}
            alt={`Image for ${$currentRide.title} bike ride`}
          />
        {/if}

        <p class="text-lg">{$currentRide.details?.String}</p>
        <Card.Root>
          <Card.Header>
            <Card.Title>{$currentRide.organizer?.String}</Card.Title>

            {#if $currentRide.email?.Valid}
              <Card.Title>
                {$currentRide.email?.String}
              </Card.Title>
            {/if}

            {#if $currentRide.weburl?.Valid && $currentRide.webname?.Valid}
              <a
                href={$currentRide.weburl?.String}
                target="_blank"
                rel="noopener noreferrer"
              >
                <Card.Title class="text-yellow-400 mt-1"
                  >{$currentRide.webname?.String}</Card.Title
                >
              </a>
            {/if}
          </Card.Header>
        </Card.Root>

        <Button
          disabled={false}
          variant="ghost"
          href="https://www.shift2bikes.org/pages/donate/"
          ref="noopener noreferrer"
          target="_blank"
          class="grow h-full w-full flex flex-row justify-center items-center"
        >
          Donate to Shift2Bikes
        </Button>
      </div>
    </ScrollArea>
  </div>
{/if}
