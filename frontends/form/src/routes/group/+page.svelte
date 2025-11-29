<script lang="ts">
  import { superForm } from "sveltekit-superforms";
  import { zod4Client as zodClient } from "sveltekit-superforms/adapters";
  import { groupRegistrationSchema } from "$lib/schemas/ride";
  import { checkGroupCodeAvailability } from "$lib/api/client";
  import ImageUploader from "$lib/components/ride-form/ImageUploader.svelte";

  // shadcn imports
  import { Button } from "$lib/components/ui/button";
  import { Input } from "$lib/components/ui/input";
  import { Label } from "$lib/components/ui/label";
  import { Textarea } from "$lib/components/ui/textarea";
  import * as Card from "$lib/components/ui/card";
  import { CircleCheck, CircleX, Loader, Users } from "@lucide/svelte";

  interface Props {
    data: {
      form: any;
      token: string;
      city: string;
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

  // Code availability checking
  let codeCheckState = $state<
    "idle" | "checking" | "available" | "unavailable"
  >("idle");
  let debounceTimer: ReturnType<typeof setTimeout> | null = $state(null);

  async function checkCodeAvailability(code: string) {
    if (!code || code.length !== 4) {
      codeCheckState = "idle";
      return;
    }

    codeCheckState = "checking";

    try {
      const result = await checkGroupCodeAvailability(code);
      codeCheckState = result.available ? "available" : "unavailable";
    } catch (err) {
      codeCheckState = "unavailable";
      console.error("Code check error:", err);
    }
  }

  function handleCodeInput(e: Event) {
    const target = e.target as HTMLInputElement;
    const newValue = target.value.toUpperCase().slice(0, 4);
    $form.code = newValue;

    // Clear existing timer
    if (debounceTimer) {
      clearTimeout(debounceTimer);
    }

    // Reset check state while typing
    if (newValue.length < 4) {
      codeCheckState = "idle";
      return;
    }

    // Debounce check
    debounceTimer = setTimeout(() => {
      checkCodeAvailability(newValue);
    }, 500);
  }
</script>

<div class="container max-w-3xl mx-auto py-4 sm:py-8 px-4">
  <div class="mb-6 sm:mb-8">
    <div class="flex items-start sm:items-center gap-3 mb-4">
      <div class="p-2 bg-primary/10 rounded-lg flex-shrink-0">
        <Users class="h-5 w-5 sm:h-6 sm:w-6 text-primary" />
      </div>
      <div class="min-w-0">
        <h1 class="text-2xl sm:text-3xl font-bold tracking-tight">
          Register Your Group
        </h1>
        <p class="text-sm sm:text-base text-muted-foreground mt-1">
          Create a group for your cycling community in {data.city.toUpperCase()}
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

  <form method="POST" use:enhance class="space-y-6">
    <!-- Group Code & Name -->
    <Card.Root>
      <Card.Header>
        <Card.Title class="text-lg sm:text-xl">Group Identity</Card.Title>
        <Card.Description class="text-sm">
          Choose a unique 4-character code and name for your group
        </Card.Description>
      </Card.Header>
      <Card.Content class="space-y-4">
        <div class="space-y-2">
          <Label for="code" class="text-sm sm:text-base">
            Group Code *
            <span
              class="text-xs sm:text-sm text-muted-foreground font-normal block sm:inline"
            >
              4 characters (letters and numbers only)
            </span>
          </Label>
          <div class="relative">
            <Input
              id="code"
              type="text"
              value={$form.code}
              oninput={handleCodeInput}
              placeholder="BIKE"
              maxlength={4}
              class={`uppercase pr-10 text-base ${$errors.code ? "border-destructive" : ""}`}
            />

            <div class="absolute right-3 top-1/2 -translate-y-1/2">
              {#if codeCheckState === "checking"}
                <Loader class="h-4 w-4 animate-spin text-muted-foreground" />
              {:else if codeCheckState === "available"}
                <CircleCheck class="h-4 w-4 text-green-600" />
              {:else if codeCheckState === "unavailable"}
                <CircleX class="h-4 w-4 text-destructive" />
              {/if}
            </div>
          </div>

          {#if codeCheckState === "available"}
            <p
              class="text-xs sm:text-sm text-green-600 flex items-center gap-1"
            >
              <CircleX class="h-3 w-3 flex-shrink-0" />
              Code is available!
            </p>
          {:else if codeCheckState === "unavailable"}
            <p
              class="text-xs sm:text-sm text-destructive flex items-center gap-1"
            >
              <CircleX class="h-3 w-3 flex-shrink-0" />
              Code is already taken
            </p>
          {/if}

          {#if $errors.code}
            <p class="text-xs sm:text-sm text-destructive">{$errors.code}</p>
          {/if}

          <p class="text-xs text-muted-foreground">
            This code will be used by ride organizers to associate rides with
            your group
          </p>
        </div>

        <div class="space-y-2">
          <Label for="name" class="text-sm sm:text-base">Group Name *</Label>
          <Input
            id="name"
            type="text"
            bind:value={$form.name}
            placeholder="Portland Bike Club"
            class={`text-base ${$errors.name ? "border-destructive" : ""}`}
          />
          {#if $errors.name}
            <p class="text-xs sm:text-sm text-destructive">{$errors.name}</p>
          {/if}
        </div>
      </Card.Content>
    </Card.Root>

    <!-- Description & Details -->
    <Card.Root>
      <Card.Header>
        <Card.Title class="text-lg sm:text-xl">About Your Group</Card.Title>
        <Card.Description class="text-sm">
          Tell the community what your group is about
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
            class="text-base"
          />
          {#if $errors.web_url}
            <p class="text-xs sm:text-sm text-destructive">{$errors.web_url}</p>
          {/if}
        </div>
      </Card.Content>
    </Card.Root>

    <!-- Group Marker -->
    <Card.Root>
      <Card.Header>
        <Card.Title class="text-lg sm:text-xl">Group Marker</Card.Title>
        <Card.Description class="text-sm">
          Add a custom marker for your group's rides on the map
        </Card.Description>
      </Card.Header>
      <Card.Content class="space-y-4">
        <div class="space-y-2">
          <Label for="marker_color" class="text-sm sm:text-base">Marker Color</Label>
          <div class="flex items-center gap-3">
            <div class="flex-1">
              <Input
                id="marker_color"
                type="color"
                bind:value={$form.marker_color}
                class="h-12 cursor-pointer text-base"
              />
            </div>
            <div class="text-xs sm:text-sm text-muted-foreground font-mono">
              {$form.marker_color}
            </div>
          </div>
          <p class="text-xs sm:text-sm text-muted-foreground">
            Choose a color for your group's marker teardrop on the map
          </p>
          {#if $errors.marker_color}
            <p class="text-xs sm:text-sm text-destructive">{$errors.marker_color}</p>
          {/if}
        </div>

        <ImageUploader
          cityCode={data.city}
          entityType="group"
          label="Upload Group Marker (Optional)"
          description="Recommended: Square image (PNG, JPG, or SVG). Will be resized to 64x64px for the map at high quality."
          acceptedTypes={[
            "image/png",
            "image/svg+xml",
            "image/jpeg",
            "image/webp",
          ]}
          maxSizeMB={5}
          onUploadComplete={(uuid) => {
            $form.image_uuid = uuid;
          }}
          onUploadError={(error) => {
            console.error("Marker upload error:", error);
          }}
        />

        <p class="text-xs sm:text-sm text-muted-foreground">
          Your marker image will be automatically resized to 64x64px and added to your city's marker spritesheet for display on the map.
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
          disabled={$delayed ||
            codeCheckState === "checking" ||
            codeCheckState === "unavailable"}
          size="lg"
          class="w-full sm:w-auto touch-manipulation"
        >
          {#if $delayed}
            Registering...
          {:else}
            Register Group
          {/if}
        </Button>
      </div>
    </div>

    <div class="text-center text-xs sm:text-sm text-muted-foreground pb-4">
      You'll receive a magic link via email to edit your group information.
    </div>
  </form>
</div>
