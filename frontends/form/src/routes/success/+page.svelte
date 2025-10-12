<script lang="ts">
  import { page } from "$app/state";
  import { Button } from "$lib/components/ui/button";
  import * as Card from "$lib/components/ui/card";
  import { CircleCheck, Mail, SquarePen } from "@lucide/svelte";

  const editToken = page.url.searchParams.get("token");
  const eventId = page.url.searchParams.get("event_id");
  const editUrl = editToken ? `/edit/${editToken}` : null;
</script>

<div class="container max-w-2xl mx-auto py-16 px-4">
  <div class="text-center mb-8">
    <div
      class="inline-flex items-center justify-center w-16 h-16 rounded-full bg-green-100 mb-4"
    >
      <CircleCheck class="w-8 h-8 text-green-600" />
    </div>
    <h1 class="text-3xl font-bold tracking-tight mb-2">
      Ride Submitted Successfully!
    </h1>
    <p class="text-muted-foreground">Your ride is now pending review</p>
  </div>

  <Card.Root>
    <Card.Content class="pt-6 space-y-6">
      <div class="space-y-4">
        <div class="flex items-start gap-3">
          <Mail class="w-5 h-5 text-muted-foreground mt-0.5" />
          <div>
            <h3 class="font-medium">Check Your Email</h3>
            <p class="text-sm text-muted-foreground mt-1">
              We've sent you a magic link to edit your ride anytime. Keep this
              email safe!
            </p>
          </div>
        </div>

        {#if editUrl}
          <div class="flex items-start gap-3">
            <SquarePen class="w-5 h-5 text-muted-foreground mt-0.5" />
            <div class="flex-1">
              <h3 class="font-medium">Edit Your Ride</h3>
              <p class="text-sm text-muted-foreground mt-1 mb-3">
                You can also use this link to make changes:
              </p>
              <Button size="sm" variant="outline">
                <a href={editUrl}> Go to Edit Page </a>
              </Button>
            </div>
          </div>
        {/if}

        <div class="p-4 bg-muted rounded-lg">
          <h3 class="font-medium mb-2">What happens next?</h3>
          <ol
            class="text-sm text-muted-foreground space-y-2 list-decimal list-inside"
          >
            <li>Our team will review your submission</li>
            <li>You'll receive an email once it's approved</li>
            <li>Your ride will appear on CycleScene</li>
            <li>Riders can find and join your ride!</li>
          </ol>
        </div>
      </div>

      {#if eventId}
        <p class="text-xs text-center text-muted-foreground">
          Event ID: {eventId}
        </p>
      {/if}
    </Card.Content>
  </Card.Root>

  <div class="text-center mt-8">
    <Button variant="ghost">
      <a href="https://cyclescene.cc"> Back to CycleScene </a>
    </Button>
  </div>
</div>
