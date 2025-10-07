<script lang="ts">
  import Button from "$lib/components/ui/button/button.svelte";
  import * as Card from "$lib/components/ui/card";
  import { ScrollArea } from "$lib/components/ui/scroll-area/index";
  import { downloadCalendarFile } from "$lib/ics";
  import { currentRide } from "$lib/stores";
  import { formatDate, formatTime } from "$lib/utils";
  import RideLabels from "./rideLabels.svelte";
  import RideMap from "./rideMap.svelte";

  const ride = $derived($currentRide);

  const SHIFT2BIKES_URL = "https://www.shift2bikes.org/";

  function handleOpenNativeMapApp() {
    if (ride) {
      const url = `https://www.google.com/maps/search/?api=1&query=${ride.lat},${ride.lng}`;
      window.open(url, "_blank");
    }
  }

  function handleAddtoCalendar() {
    if (ride) {
      downloadCalendarFile(ride);
    }
  }
</script>

{#if ride}
  <div
    class="absolute top-0 bottom-[75px] min-h-[calc(100vh-115px)] left-0 w-full p-5 overflow-hidden z-50"
  >
    <ScrollArea class="h-full w-full" scrollbarYClasses={`hidden`}>
      <div class=" flex flex-col gap-5">
        <!-- <div -->
        <!--   class="h-[400px] w-full bg-blue-500 flex items-center justify-center mx-auto text-5xl" -->
        <!-- > -->
        <RideMap {ride} />
        <!-- </div> -->
        <h2 class="text-3xl">{ride.title}</h2>
        <p>{ride.newsflash}</p>
        <RideLabels {ride} />

        <Card.Root role="button" tabindex="0" onclick={handleAddtoCalendar}>
          <Card.Header>
            <Card.Description>Meetup Time</Card.Description>
            <Card.Title class="text-2xl">
              {formatTime(ride?.starttime)}
              {formatDate(ride.date)}
            </Card.Title>
          </Card.Header>
        </Card.Root>

        <Card.Root role="button" tabindex="0" onclick={handleOpenNativeMapApp}>
          <Card.Header>
            <Card.Description>Meetup Location</Card.Description>
            <Card.Title class="text-2xl">
              {ride.venue}</Card.Title
            >
            <Card.Description>{ride.address}</Card.Description>
            <Card.Description
              >{ride.loopride
                ? "Ride is a loop"
                : "Ride not a loop"}</Card.Description
            >
          </Card.Header>

          {#if ride.locdetails != ""}
            <Card.Footer>{ride.locdetails}</Card.Footer>
          {/if}
        </Card.Root>

        {#if ride.image != ""}
          <img
            src={SHIFT2BIKES_URL + ride.image}
            alt={`Image for ${ride.title} bike ride`}
          />
        {/if}

        <p class="text-lg">{ride.details}</p>
        <Card.Root>
          <Card.Header>
            <Card.Title>{ride.organizer}</Card.Title>

            {#if ride.email}
              <Card.Title>
                {ride.email}
              </Card.Title>
            {/if}

            {#if ride.weburl && ride.webname}
              <a href={ride.weburl} target="_blank" rel="noopener noreferrer">
                <Card.Title class="text-yellow-400 mt-1"
                  >{ride.webname}</Card.Title
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
