<script lang="ts">
  import { page } from "$app/state";
  import { Button } from "$lib/components/ui/button";
  import * as Card from "$lib/components/ui/card";
  import { CircleAlert } from "@lucide/svelte";

  const message = page.url.searchParams.get("message") || "An error occurred";

  const href = "http://localhost:5174";
</script>

<div class="container max-w-2xl mx-auto py-16 px-4">
  <div class="text-center mb-8">
    <div
      class="inline-flex items-center justify-center w-16 h-16 rounded-full bg-destructive/10 mb-4"
    >
      <CircleAlert class="w-8 h-8 text-destructive" />
    </div>
    <h1 class="text-3xl font-bold tracking-tight mb-2">
      Oops! Something Went Wrong
    </h1>
    <p class="text-muted-foreground">We couldn't process your request</p>
  </div>

  <Card.Root>
    <Card.Content class="pt-6">
      <div
        class="p-4 bg-destructive/10 border border-destructive/20 rounded-lg mb-6"
      >
        <p class="text-sm text-destructive">
          {message}
        </p>
      </div>

      <div class="space-y-4">
        <h3 class="font-medium">What you can do:</h3>
        <ul
          class="text-sm text-muted-foreground space-y-2 list-disc list-inside"
        >
          <li>
            Make sure you clicked the "Host a Ride" button from the CycleScene
            app
          </li>
          <li>
            Check that your link hasn't expired (tokens are valid for 30
            minutes)
          </li>
          <li>Try requesting a new submission link from the app</li>
          <li>Contact support if the problem persists</li>
        </ul>
      </div>
    </Card.Content>
  </Card.Root>

  <div class="text-center mt-8">
    <Button>
      <a {href}> Back to CycleScene </a>
    </Button>
  </div>
</div>
