<script lang="ts">
  import Button from "$lib/components/ui/button/button.svelte";
  import * as Card from "$lib/components/ui/card/index";
  import { ScrollArea } from "$lib/components/ui/scroll-area/index";
  import { currentRide, currentRoute } from "$lib/stores";
  import { formatDate, formatTime } from "$lib/utils";
  import { STARTING_LAT, STARTING_LNG } from "$lib/config";
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

  type DescriptionPart =
    | { type: "text"; content: string }
    | { type: "link"; content: string; href: string };

  const ride = $derived($currentRide);
  const route = $derived($currentRoute);
  const hasMapLocation = $derived.by(() => {
    if (!ride) return false;

    const lat = Number(ride.lat);
    const lng = Number(ride.lng);

    return (
      Number.isFinite(lat) &&
      Number.isFinite(lng) &&
      lat !== 0 &&
      lng !== 0 &&
      !(lat === STARTING_LAT && lng === STARTING_LNG)
    );
  });

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
  const descriptionParts = $derived.by(() =>
    toDescriptionParts(ride?.details ?? ""),
  );

  function isSafeExternalURL(url: string): boolean {
    try {
      const parsed = new URL(url);
      return parsed.protocol === "https:" || parsed.protocol === "http:";
    } catch {
      return false;
    }
  }

  function addTextParts(parts: DescriptionPart[], text: string) {
    const urlPattern = /https?:\/\/[^\s<>"']+/g;
    let lastIndex = 0;

    for (const match of text.matchAll(urlPattern)) {
      const index = match.index ?? 0;
      if (index > lastIndex) {
        parts.push({ type: "text", content: text.slice(lastIndex, index) });
      }

      const url = match[0].replace(/[.,!?;:)]*$/, "");
      if (isSafeExternalURL(url)) {
        parts.push({ type: "link", content: url, href: url });
      } else {
        parts.push({ type: "text", content: match[0] });
      }
      lastIndex = index + match[0].length;
    }

    if (lastIndex < text.length) {
      parts.push({ type: "text", content: text.slice(lastIndex) });
    }
  }

  function toDescriptionParts(description: string): DescriptionPart[] {
    const parts: DescriptionPart[] = [];

    if (typeof DOMParser === "undefined") {
      addTextParts(parts, description);
      return parts;
    }

    const document = new DOMParser().parseFromString(description, "text/html");
    const visit = (node: Node) => {
      if (node.nodeType === Node.TEXT_NODE) {
        addTextParts(parts, node.textContent ?? "");
        return;
      }

      if (node.nodeType !== Node.ELEMENT_NODE) return;

      const element = node as HTMLElement;
      if (element.tagName === "BR") {
        parts.push({ type: "text", content: "\n" });
        return;
      }

      if (element.tagName === "A") {
        const href = element.getAttribute("href") ?? "";
        if (isSafeExternalURL(href)) {
          parts.push({
            type: "link",
            content: element.textContent?.trim() || href,
            href,
          });
          return;
        }
      }

      for (const child of element.childNodes) visit(child);
    };

    for (const child of document.body.childNodes) visit(child);
    return parts;
  }

  $inspect(ride);

  function handleOpenNativeMapApp() {
    if (ride && hasMapLocation) {
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
        {#if hasMapLocation}
          <RideMap {ride} />
        {/if}

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

        <Card.Root
          role={hasMapLocation ? "button" : undefined}
          tabindex={hasMapLocation ? 0 : undefined}
          onclick={handleOpenNativeMapApp}
        >
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

        {#if descriptionParts.length > 0}
          <p class="text-lg whitespace-pre-wrap">
            {#each descriptionParts as part}
              {#if part.type === "link"}
                <a
                  class="text-yellow-400 underline"
                  href={part.href}
                  target="_blank"
                  rel="noopener noreferrer"
                >{part.content}</a
                >
              {:else}
                {part.content}
              {/if}
            {/each}
          </p>
        {/if}
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
