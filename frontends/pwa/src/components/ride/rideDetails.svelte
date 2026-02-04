<script lang="ts">
  import Button from "$lib/components/ui/button/button.svelte";
  import * as Card from "$lib/components/ui/card/index";
  import { ScrollArea } from "$lib/components/ui/scroll-area/index";
  import { currentRide, currentRoute } from "$lib/stores";
  import { formatDate, formatTime } from "$lib/utils";
  import RideLabels from "./rideLabels.svelte";
  import RideMap from "./rideMap.svelte";
  import RideRouteMap from "./rideRouteMap.svelte";
  import RideRouteDetails from "./rideRouteDetails.svelte";
  import MapPinIcon from "~icons/mingcute/map-pin-line";
  import ClockIcon from "~icons/hugeicons/clock-01";
  import EmailIcon from "~icons/clarity/email-line";
  import LinkIcon from "~icons/bx/link";
  import InfoIcon from "~icons/charm/info";
  import Flash from "~icons/lets-icons/flash";

  const API_BASE = import.meta.env.VITE_API_BASE_URL;
  const CITY_CODE = import.meta.env.VITE_CITY_CODE;
  const SHIFT2BIKES_URL = "https://www.shift2bikes.org/";

  const ride = $derived($currentRide);
  const route = $derived($currentRoute);

  const imageUrl = $derived.by(() =>
    ride && ride?.ridesource === "Shift2Bikes"
      ? SHIFT2BIKES_URL + ride?.image
      : ride?.image,
  );

  function normalizeUrl(url: string): string {
    if (!url) return "";
    // If URL already has a protocol, return as-is
    if (/^https?:\/\//.test(url)) return url;
    // Otherwise prepend https://
    return `https://${url}`;
  }

  const webUrl = $derived.by(() =>
    ride?.weburl ? normalizeUrl(ride.weburl) : "",
  );

  $inspect(ride);

  function handleOpenNativeMapApp() {
    if (ride) {
      const url = `https://www.google.com/maps/search/?api=1&query=${ride.lat},${ride.lng}`;
      if (
        window.matchMedia("(display-mode: standalone)").matches ||
        /iPhone|iPad|iPod|Android/i.test(navigator.userAgent)
      ) {
        // 2. Mobile/Installed PWA Mode: Use window.open(url, '_self')
        // This is the most reliable command to force the OS to trigger the Calendar intent.
        // It causes a brief, full-screen navigation event that is necessary.
        window.open(url, "_self");
      } else {
        // 3. Desktop Browser Mode: Use the classic window.open(url, '_blank')
        // This opens a new tab and prevents the user from losing their SPA window.
        window.open(url, "_blank");
      }
    }
  }

  function handleAddtoCalendar() {
    if (!ride) return;

    const url = `${API_BASE}/v1/rides/ics?id=${ride.id}&city=${CITY_CODE}`;

    // 1. Check if the code is being run in a mobile PWA/browser context
    if (
      window.matchMedia("(display-mode: standalone)").matches ||
      /iPhone|iPad|iPod|Android/i.test(navigator.userAgent)
    ) {
      // 2. Mobile/Installed PWA Mode: Use window.open(url, '_self')
      // This is the most reliable command to force the OS to trigger the Calendar intent.
      // It causes a brief, full-screen navigation event that is necessary.
      window.open(url, "_self");
    } else {
      // 3. Desktop Browser Mode: Use the classic window.open(url, '_blank')
      // This opens a new tab and prevents the user from losing their SPA window.
      window.open(url, "_blank");
    }
  }
</script>

{#if ride}
  <div class="ride-details-container">
    <ScrollArea class="scroll-wrapper">
      <div
        class="flex flex-col gap-5 p-5 pb-[calc(var(--footer-height)_+_env(safe-area-inset-bottom)_+_10px)]"
      >
        <RideMap {ride} />

        <h2 class="text-3xl">{ride.title}</h2>
        {#if ride.newsflash}
          <div class="flex flex-row gap-1 items-center">
            <Flash
              height="22"
              width="22"
              style="color: orange; min-width: 22px;"
            />
            <p class="whitespace-pre-wrap text-lg">{ride.newsflash}</p>
          </div>
        {/if}
        <RideLabels {ride} />

        <Card.Root role="button" tabindex="0" onclick={handleAddtoCalendar}>
          <Card.Header>
            <Card.Description class="flex flex-row gap-1 items-center p-0">
              <ClockIcon height="15" width="15" style="color: orange;" />
              <span> Meetup Time </span>
            </Card.Description>
            <Card.Title class="text-2xl">
              {formatTime(ride?.starttime)}
              {formatDate(ride.date)}
            </Card.Title>
          </Card.Header>
          {#if ride.timedetails != ""}
            <Card.Footer class="flex flex-row gap-1 items-start">
              <InfoIcon
                height="20"
                width="20"
                style="color: orange; min-width: 20px;"
              />
              <span class="text-lg p-0 -mt-1">
                {ride?.timedetails}
              </span>
            </Card.Footer>
          {/if}
        </Card.Root>

        {#if ride.endtime}
          <Card.Root>
            <Card.Header>
              <Card.Description class="flex flex-row gap-1 items-center p-0">
                <ClockIcon height="15" width="15" style="color: orange;" />
                <span> End Time </span>
              </Card.Description>
              <Card.Title class="text-2xl">
                {formatTime(ride?.endtime)}</Card.Title
              >
            </Card.Header>
          </Card.Root>
        {/if}

        <Card.Root role="button" tabindex="0" onclick={handleOpenNativeMapApp}>
          <Card.Header>
            <Card.Description class="flex flex-row gap-1 items-center">
              <MapPinIcon height="15" width="15" style="color: orange;" />
              <span> Meetup Location </span>
            </Card.Description>
            <Card.Title class="text-2xl">
              {ride.venue}</Card.Title
            >
            <Card.Description class="text-lg">{ride.address}</Card.Description>
          </Card.Header>

          {#if ride.locdetails != ""}
            <Card.Footer class="flex flex-row gap-1 items-start">
              <InfoIcon
                height="20"
                width="20"
                style="color: orange; min-width: 20px;"
              />
              <span class="text-lg p-0 -mt-1">
                {ride?.locdetails}
              </span>
            </Card.Footer>
          {/if}
        </Card.Root>

        <!-- Route Visualization -->
        {#if $currentRoute}
          <Card.Root>
            <Card.Header>
              <Card.Title>Ride Route</Card.Title>
            </Card.Header>
            <Card.Content class="space-y-6">
              <RideRouteMap {route} />
              <RideRouteDetails {ride} />
            </Card.Content>
          </Card.Root>
        {/if}

        {#if ride.image != ""}
          <img src={imageUrl} alt={`Image for ${ride.title} bike ride`} />
        {/if}

        <p class="text-lg whitespace-pre-wrap">{ride.details}</p>
        <Card.Root>
          <Card.Header>
            <Card.Title>{ride.organizer}</Card.Title>

            {#if ride.email}
              <Card.Title class="flex flex-row gap-1 items-center">
                <EmailIcon height="22" width="22" style="color: orange;" />
                <span> {ride.email} </span>
              </Card.Title>
            {/if}

            {#if ride.weburl && ride.webname}
              <a href={webUrl} target="_blank" rel="noopener noreferrer">
                <Card.Title
                  class="text-yellow-400 mt-1 flex flex-row gap-1 items-center"
                >
                  <LinkIcon height="22" width="22" style="color: orange;" />
                  <span>
                    {ride.webname}
                  </span>
                </Card.Title>
              </a>
            {/if}
          </Card.Header>
        </Card.Root>

        {#if ride && ride.ridesource === "Shift2Bikes"}
          <p class="text-sm text-gray-500 my-4">
            Event data provided by Shift2Bikes
          </p>
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
        {/if}

        {#if ride && ride.ridesource === "strava"}
          <div class="flex items-center justify-center my-4 py-4 bg-[#2D2D32] rounded-lg">
            <img
              src="/api_logo_pwrdBy_strava_horiz_white.png"
              alt="Powered by Strava"
              class="h-5"
            />
          </div>
        {/if}
      </div>
    </ScrollArea>
  </div>
{/if}

<style>
  .ride-details-container {
    height: 100%;
    width: 100%;
    display: flex;
    flex-direction: column;
    overflow: hidden;
  }

  :global(.scroll-wrapper) {
    height: 100%;
    width: 100%;
  }
</style>
