<script lang="ts">
  import Button from "$lib/components/ui/button/button.svelte";
  import * as Card from "$lib/components/ui/card";
  import ScrollArea from "$lib/components/ui/scroll-area/scroll-area.svelte";
  import MailIcon from "~icons/material-symbols/mail-outline";

  const contactEmail = "info@cyclescene.cc";
  const mailtoHref =
    "mailto:info@cyclescene.cc?subject=Cycle%20Scene%20feedback";
</script>

<div class="contact-view">
  <ScrollArea class="scroll-wrapper">
    <div
      class="flex flex-col gap-4 p-5 pb-[calc(var(--footer-height)_+_env(safe-area-inset-bottom)_+_10px)]"
    >
      <div class="pt-2">
        <h1 class="text-2xl font-bold">Report Bug / Request Feature</h1>
        <p class="text-muted-foreground mt-1">
          Send questions, bug reports, or feature requests by email.
        </p>
      </div>

      <Card.Root class="p-4 gap-3">
        <Card.Header class="p-0">
          <Card.Title class="text-base">Contact Cycle Scene</Card.Title>
          <Card.Description>{contactEmail}</Card.Description>
        </Card.Header>

        <Button class="w-full" href={mailtoHref}>
          <MailIcon class="w-5 h-5" />
          Email Cycle Scene
        </Button>
      </Card.Root>

      <Card.Root class="p-4">
        <Card.Header class="p-0">
          <Card.Title class="text-base">Helpful details</Card.Title>
        </Card.Header>
        <Card.Content class="p-0 mt-2">
          <ul class="text-sm text-muted-foreground space-y-2">
            <li>What you were trying to do</li>
            <li>What went wrong or what you expected</li>
            <li>Your device and browser</li>
          </ul>
        </Card.Content>
      </Card.Root>
    </div>
  </ScrollArea>
</div>

<style>
  .contact-view {
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
