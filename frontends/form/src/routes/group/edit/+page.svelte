<script lang="ts">
  import { superForm } from "sveltekit-superforms";
  import { zod4Client as zodClient } from "sveltekit-superforms/adapters";
  import { groupRegistrationSchema } from "$lib/schemas/ride";
  import CustomMarkerBuilder from "$lib/components/group-form/CustomMarkerBuilder.svelte";

  // shadcn imports
  import { Button } from "$lib/components/ui/button";
  import { Input } from "$lib/components/ui/input";
  import { Label } from "$lib/components/ui/label";
  import { Textarea } from "$lib/components/ui/textarea";
  import * as Card from "$lib/components/ui/card";
  import { Edit, Loader, CircleX } from "@lucide/svelte";

  interface Props {
    data: {
      form: any;
      token: string;
      city: string;
      groupCode: string;
      groupName: string;
      groupEmail: string;
    };
  }

  let { data }: Props = $props();

  const { form, errors, enhance, delayed, message } = superForm(data.form, {
    validators: zodClient(groupRegistrationSchema),
    dataType: "json",
    resetForm: false,
    onError({ result }) {
      $message = result.error.message;
    },
  });

  const backUrl = data.city ? `https://${data.city}.cyclescene.cc` : "https://cyclescene.cc";

  let isEditingName = $state(false);
</script>

<div class="container max-w-3xl mx-auto py-4 sm:py-8 px-4">
  <div class="mb-6 sm:mb-8">
    <div class="flex items-start sm:items-center gap-3 mb-4">
      <div class="p-2 bg-primary/10 rounded-lg flex-shrink-0">
        <Edit class="h-5 w-5 sm:h-6 sm:w-6 text-primary" />
      </div>
      <div class="min-w-0">
        <h1 class="text-2xl sm:text-3xl font-bold tracking-tight">
          Edit Group Settings
        </h1>
        <p class="text-sm sm:text-base text-muted-foreground mt-1">
          Update your group information for {data.city.toUpperCase()}
        </p>
      </div>
    </div>
  </div>

  {#if $message}
    <div
      class="mb-4 sm:mb-6 p-3 sm:p-4 border border-destructive bg-destructive/10 rounded-lg"
    >
      <p class="text-xs sm:text-sm text-destructive">{$message}</p>
    </div>
  {/if}

  {#if Object.keys($errors).length > 0}
    <div
      class="mb-4 sm:mb-6 p-3 sm:p-4 border border-destructive bg-destructive/10 rounded-lg"
    >
      <p class="text-xs sm:text-sm font-semibold text-destructive mb-2">
        Please fix the following errors:
      </p>
      <ul class="list-disc list-inside space-y-1">
        {#each Object.entries($errors) as [field, message]}
          {#if message}
            <li class="text-xs sm:text-sm text-destructive">
              {typeof message === 'string' ? message : message[0]}
            </li>
          {/if}
        {/each}
      </ul>
    </div>
  {/if}

  <!-- Group Summary Card -->
  <Card.Root class="mb-6">
    <Card.Header>
      <Card.Title class="text-lg sm:text-xl">Group Summary</Card.Title>
    </Card.Header>
    <Card.Content class="space-y-6">
      <!-- Row 1: Code and Email (read-only) -->
      <div class="grid gap-4 sm:grid-cols-2">
        <div>
          <p class="text-xs sm:text-sm text-muted-foreground mb-2">Group Code</p>
          <code class="block text-2xl font-bold tracking-wider bg-background px-4 py-3 rounded border text-center">
            {data.groupCode}
          </code>
        </div>
        <div>
          <p class="text-xs sm:text-sm text-muted-foreground mb-2">Contact Email</p>
          <p class="text-sm sm:text-base break-all">{data.groupEmail}</p>
        </div>
      </div>

      <!-- Row 2: Group Name (editable) -->
      <div>
        <p class="text-xs sm:text-sm text-muted-foreground mb-2">Group Name</p>
        {#if isEditingName}
          <Input
            type="text"
            bind:value={$form.name}
            aria-invalid={!!$errors.name}
            class={`text-sm sm:text-base transition-colors ${
              $errors.name
                ? "border-destructive bg-destructive/5 focus:border-destructive focus:ring-1 focus:ring-destructive/20"
                : ""
            }`}
            autofocus
          />
          {#if $errors.name}
            <p class="text-xs sm:text-sm text-destructive flex items-center gap-1 mt-2">
              <CircleX class="h-3 w-3 flex-shrink-0" />
              {$errors.name}
            </p>
          {/if}
          <div class="flex gap-2 mt-2">
            <button
              type="button"
              onclick={() => {
                isEditingName = false;
                $form.name = data.groupName;
              }}
              class="text-xs text-muted-foreground hover:underline"
            >
              Cancel
            </button>
          </div>
        {:else}
          <p class="text-sm sm:text-base font-medium">{data.groupName}</p>
          <button
            type="button"
            onclick={() => (isEditingName = true)}
            class="text-xs text-primary hover:underline mt-1"
          >
            Edit name
          </button>
        {/if}
      </div>
    </Card.Content>
  </Card.Root>

  <form method="POST" use:enhance class="space-y-6">

    <!-- Description & Details -->
    <Card.Root>
      <Card.Header>
        <Card.Title class="text-lg sm:text-xl">About Your Group</Card.Title>
        <Card.Description class="text-sm">
          Update your group description and website
        </Card.Description>
      </Card.Header>
      <Card.Content class="space-y-4">
        <div class="space-y-2">
          <Label for="description" class="text-sm sm:text-base"
            >Description (Optional)</Label
          >
          <Textarea
            id="description"
            bind:value={$form.description}
            placeholder="We're a community of casual riders who love exploring the city on two wheels. All skill levels welcome!"
            rows={4}
            maxlength={500}
            class="text-base"
          />
          <p class="text-xs text-muted-foreground text-right">
            {$form.description?.length || 0}/500 characters
          </p>
        </div>

        <div class="space-y-2">
          <Label for="web_url" class="text-sm sm:text-base"
            >Website URL (Optional)</Label
          >
          <Input
            id="web_url"
            type="url"
            bind:value={$form.web_url}
            placeholder="https://portlandbikeclub.com"
            aria-invalid={!!$errors.web_url}
            class={`text-base transition-colors ${
              $errors.web_url
                ? "border-destructive bg-destructive/5 focus:border-destructive focus:ring-1 focus:ring-destructive/20"
                : ""
            }`}
          />
          {#if $errors.web_url}
            <p class="text-xs sm:text-sm text-destructive flex items-center gap-1">
              <CircleX class="h-3 w-3 flex-shrink-0" />
              {$errors.web_url}
            </p>
          {/if}
        </div>
      </Card.Content>
    </Card.Root>

    <!-- Group Marker -->
    <Card.Root>
      <Card.Header>
        <Card.Title class="text-lg sm:text-xl">Group Marker (Optional)</Card.Title>
        <Card.Description class="text-sm">
          Update your group's marker with a new image
        </Card.Description>
      </Card.Header>
      <Card.Content class="space-y-4">
        <CustomMarkerBuilder
          cityCode={data.city}
          onUploadComplete={(uuid) => {
            $form.image_uuid = uuid;
          }}
          onUploadError={(error) => {
            console.error("Marker upload error:", error);
          }}
        />

        <p class="text-xs sm:text-sm text-muted-foreground">
          If you upload a new marker image, it will be automatically resized to 64x64px and added
          to your city's marker spritesheet. This is optional - you can leave this blank to keep
          your current marker.
        </p>
      </Card.Content>
    </Card.Root>

    <!-- Submit Button -->
    <div
      class="sticky bottom-0 bg-background border-t pt-4 pb-4 sm:pb-6 -mx-4 px-4 sm:mx-0 sm:px-0 sm:static"
    >
      <div
        class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3"
      >
        <p
          class="text-xs sm:text-sm text-muted-foreground text-center sm:text-left"
        >
          * Required fields
        </p>
        <Button
          type="submit"
          disabled={$delayed}
          size="lg"
          class="w-full sm:w-auto touch-manipulation"
        >
          {#if $delayed}
            <Loader class="mr-2 h-4 w-4 animate-spin" />
            Saving Changes...
          {:else}
            Save Changes
          {/if}
        </Button>
      </div>
    </div>
  </form>

  <div class="text-center text-xs sm:text-sm text-muted-foreground pb-4 mt-6">
    <p>
      Your group code and city cannot be changed. Upload a new marker image above
      if you'd like to update your group's appearance on the map.
    </p>
  </div>

  <div class="text-center mt-8">
    <Button variant="ghost">
      <a href={backUrl}> Back to CycleScene </a>
    </Button>
  </div>
</div>
