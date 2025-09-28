<script>
  import { Button } from "$lib/components/ui/button";
  import * as Card from "$lib/components/ui/card";
  import ScrollArea from "$lib/components/ui/scroll-area/scroll-area.svelte";
  import { Separator } from "$lib/components/ui/separator";
  import { navigateTo, SUB_VIEW_ABOUT, SUB_VIEW_APPEARANCE } from "$lib/stores";
  import IconChevronRight from "~icons/bxs/chevron-right";

  // https://www.shift2bikes.org/addevent/
</script>

<div
  class="absolute top-[60px] bottom-[55px] min-h-[calc(100vh_-_135px)] w-full p-5"
>
  <ScrollArea class={`relative`}>
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
        <Card.Header class=" flex p-0">
          <Button
            disabled={false}
            variant="ghost"
            href="https://www.shift2bikes.org/addevent/"
            ref="noopener noreferrer"
            target="_blank"
            class="w-full justify-center"
          >
            <Card.Title class="grow text-left">Host a Ride</Card.Title>
            <IconChevronRight class="shrink" />
          </Button>
        </Card.Header>
        <Separator class="" />
        <Card.Header class=" flex p-0">
          <Button
            disabled={false}
            variant="ghost"
            onclick={() => console.log("clicked")}
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
            onclick={() => console.log("clicked")}
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
            onclick={() => console.log("clicked")}
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
            onclick={() => console.log("clicked")}
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
            onclick={() => console.log("clicked")}
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
            onclick={() => console.log("clicked")}
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
            onclick={() => console.log("clicked")}
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
            onclick={() => console.log("clicked")}
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
