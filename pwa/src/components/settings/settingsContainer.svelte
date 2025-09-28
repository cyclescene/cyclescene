<script>
  import { Button } from "$lib/components/ui/button";
  import * as Card from "$lib/components/ui/card";
  import { Separator } from "$lib/components/ui/separator";
  import IconChevronRight from "~icons/bxs/chevron-right";
</script>

<div
  class="absolute top-[60px] bottom-[75px] min-h-[calc(100vh_-_115px)] w-full p-10 flex flex-col gap-5"
>
  <Card.Root class="p-2 gap-2">
    <Card.Header class=" flex p-0">
      <Button
        disabled={false}
        variant="ghost"
        onclick={() => console.log("clicked")}
        class="grow h-full w-full flex flex-row justify-center items-center"
      >
        <Card.Title class="grow text-left">Appearance</Card.Title>
        <IconChevronRight class="shrink" />
      </Button>
    </Card.Header>
    <Separator class="" />
    <Card.Header class=" flex p-0">
      <Button
        disabled={false}
        variant="ghost"
        onclick={() => console.log("clicked")}
        class="grow h-full w-full flex flex-row justify-center items-center"
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
        onclick={() => console.log("clicked")}
        class="grow h-full w-full flex flex-row justify-center items-center"
      >
        <Card.Title class="grow text-left"
          >Report Bug / Request Feature</Card.Title
        >
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
        class="grow h-full w-full flex flex-row justify-center items-center"
      >
        <Card.Title class="grow text-left">Donate to Shift2Bikes</Card.Title>
        <IconChevronRight class="shrink" />
      </Button>
    </Card.Header>
    <Separator class="" />
    <Card.Header class=" flex p-0">
      <Button
        disabled={false}
        variant="ghost"
        href="https://www.shift2bikes.org/pages/donate/"
        ref="noopener noreferrer"
        target="_blank"
        class="grow h-full w-full flex flex-row justify-center items-center"
      >
        <Card.Title class="grow text-left">Support CycleScene</Card.Title>
        <IconChevronRight class="shrink" />
      </Button>
    </Card.Header>
  </Card.Root>
</div>
