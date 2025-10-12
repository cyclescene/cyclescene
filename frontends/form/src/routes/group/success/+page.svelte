<script lang="ts">
  import { page } from "$app/state";
  import { Button } from "$lib/components/ui/button";
  import * as Card from "$lib/components/ui/card";
  import { CircleCheck, Mail, SquarePen, Copy } from "@lucide/svelte";

  const editToken = page.url.searchParams.get("token");
  const groupCode = page.url.searchParams.get("code");
  const editUrl = editToken ? `/group/edit/${editToken}` : null;

  let copied = $state(false);

  function copyCode() {
    if (groupCode) {
      navigator.clipboard.writeText(groupCode);
      copied = true;
      setTimeout(() => {
        copied = false;
      }, 2000);
    }
  }
</script>

<div class="container max-w-2xl mx-auto py-16 px-4">
  <div class="text-center mb-8">
    <div
      class="inline-flex items-center justify-center w-16 h-16 rounded-full bg-green-100 mb-4"
    >
      <CircleCheck class="w-8 h-8 text-green-600" />
    </div>
    <h1 class="text-3xl font-bold tracking-tight mb-2">
      Group Registered Successfully!
    </h1>
    <p class="text-muted-foreground">
      Your cycling group is now live on CycleScene
    </p>
  </div>

  <Card.Root>
    <Card.Content class="pt-6 space-y-6">
      {#if groupCode}
        <div class="p-4 bg-primary/5 border-2 border-primary/20 rounded-lg">
          <p class="text-sm font-medium mb-2">Your Group Code:</p>
          <div class="flex items-center gap-2">
            <code
              class="text-2xl font-bold tracking-wider bg-background px-4 py-2 rounded border flex-1 text-center"
            >
              {groupCode}
            </code>
            <Button variant="outline" size="icon" onclick={copyCode}>
              {#if copied}
                <CircleCheck class="h-4 w-4" />
              {:else}
                <Copy class="h-4 w-4" />
              {/if}
            </Button>
          </div>
          <p class="text-xs text-muted-foreground mt-2">
            Share this code with ride organizers to associate rides with your
            group
          </p>
        </div>
      {/if}

      <div class="space-y-4">
        <div class="flex items-start gap-3">
          <Mail class="w-5 h-5 text-muted-foreground mt-0.5" />
          <div>
            <h3 class="font-medium">Check Your Email</h3>
            <p class="text-sm text-muted-foreground mt-1">
              We've sent you a magic link to edit your group information
              anytime. Keep this email safe!
            </p>
          </div>
        </div>

        {#if editUrl}
          <div class="flex items-start gap-3">
            <SquarePen class="w-5 h-5 text-muted-foreground mt-0.5" />
            <div class="flex-1">
              <h3 class="font-medium">Edit Your Group</h3>
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
          <h3 class="font-medium mb-2">Next Steps:</h3>
          <ul
            class="text-sm text-muted-foreground space-y-2 list-disc list-inside"
          >
            <li>Share your group code with ride organizers</li>
            <li>
              When hosting rides, enter your code to associate them with your
              group
            </li>
            <li>Your group icon will appear on the map for all your rides</li>
            <li>Build your cycling community!</li>
          </ul>
        </div>
      </div>
    </Card.Content>
  </Card.Root>

  <div class="text-center mt-8">
    <Button variant="ghost">
      <a href="https://cyclescene.cc"> Back to CycleScene </a>
    </Button>
  </div>
</div>
