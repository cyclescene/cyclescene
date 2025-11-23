<script>
  import { Button } from "$lib/components/ui/button";
  import * as Card from "$lib/components/ui/card";
  import ScrollArea from "$lib/components/ui/scroll-area/scroll-area.svelte";
  import { Separator } from "$lib/components/ui/separator";
  import { CITY_CODE } from "$lib/config";
  import {
    navigateTo,
    SUB_VIEW_ABOUT,
    SUB_VIEW_ADULT_ONLY_RIDES,
    SUB_VIEW_APPEARANCE,
    SUB_VIEW_CHANGE_LOG,
    SUB_VIEW_CONTACT,
    SUB_VIEW_COVID_SAFETY_RIDES,
    SUB_VIEW_DATA,
    SUB_VIEW_FAMILY_FRIENDLY_RIDES,
    SUB_VIEW_PRIVACY_POLICY,
    SUB_VIEW_TERMS_OF_USE,
  } from "$lib/stores";
  import IconChevronRight from "~icons/bxs/chevron-right";

  async function handleHostRide() {
    try {
      const url = `${import.meta.env.VITE_API_BASE_URL}/v1/tokens/submission`;

      const response = await fetch(url, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ city: CITY_CODE }),
      });

      if (!response.ok) {
        throw new Error("Failed to get sumbmission token");
      }

      const { token } = await response.json();
      console.log(
        `${import.meta.env.VITE_FORM_BASE_URL}?token=${token}&city=${CITY_CODE}`,
      );

      window.location.href = `${import.meta.env.VITE_FORM_BASE_URL}?token=${token}&city=${CITY_CODE}`;
    } catch (err) {
      console.error("Error getting form access: ", err);
      alert("Unable to access form");
    }
  }
  async function handleRegisterGroup() {
    try {
      const url = `${import.meta.env.VITE_API_BASE_URL}/v1/tokens/submission`;

      const response = await fetch(url, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ city: CITY_CODE }),
      });

      const { token } = await response.json();

      // Redirect to group registration form
      window.location.href = `${import.meta.env.VITE_FORM_BASE_URL}/group?token=${token}&city=${CITY_CODE}`;
    } catch (error) {
      console.error("Error:", error);
    }
  }
</script>

<div
  class="absolute top-0 bottom-[55px] min-h-[calc(100vh_-_135px)] w-full p-5"
>
  <ScrollArea class={`relative`} scrollbarYClasses={`hidden`}>
    <div class="h-[calc(100vh_-_140px)] flex flex-col gap-4">
      <Card.Root class="p-2 gap-2">
        <Card.Header class=" flex p-0">
          <Button
            disabled={false}
            variant="ghost"
            onclick={() => navigateTo(SUB_VIEW_ABOUT)}
            class="w-full justify-center"
          >
            <Card.Title class="grow text-left">About</Card.Title>
            <IconChevronRight class="shrink" />
          </Button>
        </Card.Header>
        <Separator class="" />
        <Card.Header class=" flex p-0">
          <Button
            disabled={false}
            variant="ghost"
            onclick={() => navigateTo(SUB_VIEW_APPEARANCE)}
            class="w-full justify-center"
          >
            <Card.Title class="grow text-left">Appearance</Card.Title>
            <IconChevronRight class="shrink" />
          </Button>
        </Card.Header>
      </Card.Root>
      <Card.Root class="p-2 gap-2">
        {#if CITY_CODE === "pdx"}
          <Card.Header class=" flex p-0">
            <Button
              disabled={false}
              variant="ghost"
              href="https://www.shift2bikes.org/addevent/"
              ref="noopener noreferrer"
              target="_blank"
              class="w-full justify-center"
            >
              <Card.Title class="grow text-left"
                >Host a Ride on Shift2Bikes</Card.Title
              >
              <IconChevronRight class="shrink" />
            </Button>
          </Card.Header>
          <Separator />
        {/if}
        <Card.Header class=" flex p-0">
          <Button
            disabled={false}
            onclick={handleHostRide}
            variant="ghost"
            class="w-full justify-center"
          >
            <Card.Title class="grow text-left">Host a Ride</Card.Title>
            <IconChevronRight class="shrink" />
          </Button>
        </Card.Header>
        <Separator />
        <Card.Header class=" flex p-0">
          <Button
            disabled={false}
            variant="ghost"
            onclick={handleRegisterGroup}
            class="w-full justify-center"
          >
            <Card.Title class="grow text-left">Register Group</Card.Title>
            <IconChevronRight class="shrink" />
          </Button>
        </Card.Header>
        <Separator class="" />
        <Card.Header class=" flex p-0">
          <Button
            disabled={false}
            variant="ghost"
            onclick={() => navigateTo(SUB_VIEW_ADULT_ONLY_RIDES)}
            class="w-full justify-center"
          >
            <Card.Title class="grow text-left">Adults Only Rides</Card.Title>
            <IconChevronRight class="shrink" />
          </Button>
        </Card.Header>
        <Separator class="" />
        <Card.Header class=" flex p-0">
          <Button
            disabled={false}
            variant="ghost"
            onclick={() => navigateTo(SUB_VIEW_FAMILY_FRIENDLY_RIDES)}
            class="w-full justify-center"
          >
            <Card.Title class="grow text-left">Family Friendly Rides</Card.Title
            >
            <IconChevronRight class="shrink" />
          </Button>
        </Card.Header>
        <Separator class="" />
        <Card.Header class=" flex p-0">
          <Button
            disabled={false}
            variant="ghost"
            onclick={() => navigateTo(SUB_VIEW_COVID_SAFETY_RIDES)}
            class="w-full justify-center"
          >
            <Card.Title class="grow text-left">Covid Safety Rides</Card.Title>
            <IconChevronRight class="shrink" />
          </Button>
        </Card.Header>
      </Card.Root>
      <Card.Root class="p-2 gap-2">
        <Card.Header class=" flex p-0">
          <Button
            disabled={false}
            variant="ghost"
            href="https://www.shift2bikes.org/pages/donate/"
            ref="noopener noreferrer"
            target="_blank"
            class="w-full justify-center"
          >
            <Card.Title class="grow text-left">Donate to Shift2Bikes</Card.Title
            >
            <IconChevronRight class="shrink" />
          </Button>
        </Card.Header>
        <Separator class="" />
        <Card.Header class=" flex p-0">
          <Button
            disabled={false}
            variant="ghost"
            href="https://www.paypal.com/donate/?business=D47RHCG7M2XPA&no_recurring=1&item_name=Support+Cycle+Scene&currency_code=USD"
            ref="noopener noreferrer"
            target="_blank"
            class="w-full justify-center"
          >
            <Card.Title class="grow text-left">Support CycleScene</Card.Title>
            <IconChevronRight class="shrink" />
          </Button>
        </Card.Header>
      </Card.Root>
      <Card.Root class="p-2 gap-2">
        <Card.Header class=" flex p-0">
          <Button
            disabled={false}
            variant="ghost"
            onclick={() => navigateTo(SUB_VIEW_DATA)}
            class="w-full justify-center"
          >
            <Card.Title class="grow text-left">Refresh Data</Card.Title>
            <IconChevronRight class="shrink" />
          </Button>
        </Card.Header>
        <Separator />
        <Card.Header class=" flex p-0">
          <Button
            disabled={false}
            variant="ghost"
            onclick={() => navigateTo(SUB_VIEW_PRIVACY_POLICY)}
            class="w-full justify-center"
          >
            <Card.Title class="grow text-left">Privacy Policy</Card.Title>
            <IconChevronRight class="shrink" />
          </Button>
        </Card.Header>
        <Separator />
        <Card.Header class="flex p-0">
          <Button
            disabled={false}
            variant="ghost"
            onclick={() => navigateTo(SUB_VIEW_TERMS_OF_USE)}
            class="w-full justify-center"
          >
            <Card.Title class="grow text-left">Terms of Use</Card.Title>
            <IconChevronRight class="shrink" />
          </Button>
        </Card.Header>
        <Separator />
        <Card.Header class="flex p-0">
          <Button
            disabled={false}
            variant="ghost"
            onclick={() => navigateTo(SUB_VIEW_CHANGE_LOG)}
            class="w-full justify-center"
          >
            <Card.Title class="grow text-left">Changelog</Card.Title>
            <IconChevronRight class="shrink" />
          </Button>
        </Card.Header>
        <Separator />
        <Card.Header class="flex p-0">
          <Button
            disabled={false}
            variant="ghost"
            onclick={() => navigateTo(SUB_VIEW_CONTACT)}
            class="w-full justify-center"
          >
            <Card.Title class="grow text-left"
              >Report Bug / Request Feature</Card.Title
            >
            <IconChevronRight class="shrink" />
          </Button>
        </Card.Header>
      </Card.Root>
    </div>
  </ScrollArea>
</div>
