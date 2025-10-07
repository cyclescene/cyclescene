<script lang="ts">
  import Button from "$lib/components/ui/button/button.svelte";
  import * as Card from "$lib/components/ui/card";
  import { ScrollArea } from "$lib/components/ui/scroll-area/index";
  import { currentRide } from "$lib/stores";
  import { formatDate, formatTime, formatToICS } from "$lib/utils";
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

  function getRideDateTime(date: string, time: string): Date {
    return new Date(`${date} ${time}`);
  }

  function handleAddtoCalendar() {
    // Safety check that the ride object is available
    if (!ride) {
      console.error("Cannot add to calendar: ride object is null.");
      return;
    }

    // --- 1. Determine Start and End Times ---
    const rideDate = ride.date;
    const startTime = ride.starttime || "00:00:00";
    let endTime = ride.endtime || "";

    if (!endTime) {
      // Estimate 2 hours long if end time is missing
      const startDateTime = getRideDateTime(rideDate, startTime);
      const estimatedEndDateTime = new Date(
        startDateTime.getTime() + 2 * 60 * 60 * 1000,
      );

      // Manually build the 24-hour time string
      const pad = (num: number) => String(num).padStart(2, "0");
      endTime = `${pad(estimatedEndDateTime.getHours())}:${pad(estimatedEndDateTime.getMinutes())}:${pad(estimatedEndDateTime.getSeconds())}`;
    }

    // --- 2. Format to ICS (Using your assumed helper functions) ---
    // Generate current timestamp for file creation
    const DTSTAMP = formatToICS(
      new Date().toISOString().slice(0, 10),
      new Date().toLocaleTimeString("en-US", { hour12: false }),
    );

    const DTSTART = formatToICS(rideDate, startTime);
    const DTEND = formatToICS(rideDate, endTime);

    // --- 3. Construct ICS Content (The text file) ---
    // Clean up description and add shareable link to details
    const DETAILS = `DESCRIPTION:${ride.details.replace(/[\r\n]/g, "\\n")}\\nURL:${ride.shareable}`;
    const LOCATION = `${ride.venue || ride.address}`;

    const ICS_CONTENT = `BEGIN:VCALENDAR
      VERSION:2.0
      PRODID:-//CycleScene//NONSGML V1.0//EN
      BEGIN:VEVENT
      UID:${ride.id}@cyclescene.com
      DTSTAMP:${DTSTAMP}
      DTSTART:${DTSTART}
      DTEND:${DTEND}
      SUMMARY:${ride.title}
      LOCATION:${LOCATION}
      ${DETAILS}
      END:VEVENT
      END:VCALENDAR`;

    // --- 4. Trigger Download (The iOS/Safari Fix) ---
    const blob = new Blob([ICS_CONTENT], {
      type: "text/calendar;charset=utf-8",
    });
    const url = URL.createObjectURL(blob);

    // Create a temporary link element
    const link = document.createElement("a");
    link.href = url;
    link.download = `${ride.title.replace(/[\s\W]+/g, "_")}_${ride.date}.ics`;

    // Append to body (required for some browsers)
    document.body.appendChild(link);

    link.click();

    // 5. Clean up temporary objects
    // Use a slight delay to ensure the browser registers the download before cleanup
    setTimeout(() => {
      document.body.removeChild(link);
      URL.revokeObjectURL(url);
    }, 100);
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
