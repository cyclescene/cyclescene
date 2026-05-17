<script lang="ts">
  import { superForm } from "sveltekit-superforms";
  import { zod4Client as zodClient } from "sveltekit-superforms/adapters";
  import { groupRegistrationSchema } from "$lib/schemas/ride";
  import { checkGroupCodeAvailability } from "$lib/api/client";
  import ImageUploader from "$lib/components/ride-form/ImageUploader.svelte";
  import CustomMarkerBuilder from "$lib/components/group-form/CustomMarkerBuilder.svelte";

  // shadcn imports
  import { Button } from "$lib/components/ui/button";
  import { Input } from "$lib/components/ui/input";
  import { Label } from "$lib/components/ui/label";
  import { Textarea } from "$lib/components/ui/textarea";
  import * as Card from "$lib/components/ui/card";
  import { CircleCheck, CircleX, Loader, Users, MapPin, AlertTriangle } from "@lucide/svelte";

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
    onSubmit({ formData }) {
      // Check if marker image has been uploaded
      if (!markerImageUUID) {
        // Auto-generate and upload default marker
        return new Promise((resolve, reject) => {
          customMarkerBuilderRef?.autoGenerateAndUploadMarker().then((uuid: string | null) => {
            if (uuid) {
              markerImageUUID = uuid;
              $form.image_uuid = uuid;
              resolve(formData);
            } else {
              $message =
                "Failed to generate marker. Please upload a custom marker or try again.";
              reject(new Error("Marker generation failed"));
            }
          });
        });
      }
    },
  });

  // Code availability checking
  let codeCheckState = $state<
    "idle" | "checking" | "available" | "unavailable"
  >("idle");
  let debounceTimer: ReturnType<typeof setTimeout> | null = $state(null);

  // Marker image tracking
  let markerImageUUID = $state<string | null>(null);
  let customMarkerBuilderRef = $state<{ autoGenerateAndUploadMarker: () => Promise<string | null> }>();

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

<style>
  @keyframes slideIn {
    from {
      opacity: 0;
      transform: translateY(-8px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }

  .animate-slide-in {
    animation: slideIn 0.3s cubic-bezier(0.16, 1, 0.3, 1);
  }
</style>

<div class="min-h-screen bg-background">
  <div class="max-w-5xl mx-auto px-4 sm:px-6 lg:px-8 py-8 sm:py-12">

    <header class="mb-10">
      <div class="inline-flex items-center gap-2 text-sm font-medium text-slate-500 mb-3">
        <MapPin class="h-4 w-4" />
        <span class="uppercase tracking-wide">{data.city.toUpperCase()}</span>
      </div>

      <h1 class="text-4xl sm:text-5xl font-bold tracking-tight text-card-foreground mb-4 leading-tight">
        Register Your Group
      </h1>
      <p class="text-lg text-muted-foreground">
        Create a group for your cycling community in {data.city.toUpperCase()}
      </p>
    </header>

  {#if $message}
    <div class="mb-6 p-4 bg-red-50 dark:bg-red-950 border-l-4 border-red-500 rounded-lg shadow-sm animate-slide-in">
      <div class="flex items-center gap-3">
        <CircleX class="h-5 w-5 text-red-600 dark:text-red-400 flex-shrink-0" />
        <p class="text-sm font-medium text-red-900 dark:text-red-100">{$message}</p>
      </div>
    </div>
  {/if}

  {#if Object.keys($errors).length > 0}
    <div class="mb-6 p-5 bg-gradient-to-br from-amber-50 to-orange-50 dark:from-amber-950/50 dark:to-orange-950/50 border border-amber-200/60 dark:border-amber-800/60 rounded-2xl shadow-sm animate-slide-in">
      <div class="flex gap-4">
        <div class="flex-shrink-0">
          <div class="h-10 w-10 rounded-full bg-amber-100 dark:bg-amber-900/50 flex items-center justify-center">
            <AlertTriangle class="h-5 w-5 text-amber-600 dark:text-amber-400" />
          </div>
        </div>
        <div class="flex-1">
          <h3 class="text-base font-semibold text-amber-900 dark:text-amber-100 mb-2">
            Please fix the following errors:
          </h3>
          <ul class="list-disc list-inside space-y-1">
            {#each Object.entries($errors) as [field, message]}
              {#if message}
                <li class="text-sm text-amber-800 dark:text-amber-200/90">
                  {typeof message === 'string' ? message : message[0]}
                </li>
              {/if}
            {/each}
          </ul>
        </div>
      </div>
    </div>
  {/if}

  <form method="POST" use:enhance class="space-y-6">
    <!-- Group Code & Name -->
    <div class="bg-card rounded-xl border border-border shadow-sm overflow-hidden">
      <div class="px-6 py-5 border-b border-border bg-muted/50">
        <h2 class="text-xl font-semibold text-card-foreground">Group Identity</h2>
        <p class="text-sm text-muted-foreground mt-0.5">
          Choose a unique 4-character code and name for your group
        </p>
      </div>
      <div class="px-6 py-6 space-y-4">
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
              aria-invalid={!!$errors.code}
              class={`uppercase pr-10 text-base transition-colors ${
                $errors.code
                  ? "border-red-500 bg-red-50/5 dark:bg-red-950/5 focus:border-red-500 focus:ring-1 focus:ring-red-500/20"
                  : ""
              }`}
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
              class="text-xs sm:text-sm text-emerald-600 dark:text-emerald-400 flex items-center gap-1"
            >
              <CircleCheck class="h-3 w-3 flex-shrink-0" />
              Code is available!
            </p>
          {:else if codeCheckState === "unavailable"}
            <p
              class="text-xs sm:text-sm text-red-600 dark:text-red-400 flex items-center gap-1"
            >
              <CircleX class="h-3 w-3 flex-shrink-0" />
              Code is already taken
            </p>
          {/if}

          {#if $errors.code}
            <p class="text-xs sm:text-sm text-red-600 dark:text-red-400">{$errors.code}</p>
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
            aria-invalid={!!$errors.name}
            class={`text-base transition-colors ${
              $errors.name
                ? "border-red-500 bg-red-50/5 dark:bg-red-950/5 focus:border-red-500 focus:ring-1 focus:ring-red-500/20"
                : ""
            }`}
          />
          {#if $errors.name}
            <p class="text-xs sm:text-sm text-red-600 dark:text-red-400 flex items-center gap-1">
              <CircleX class="h-3 w-3 flex-shrink-0" />
              {$errors.name}
            </p>
          {/if}
        </div>
      </div>
    </div>

    <!-- Description & Details -->
    <div class="bg-card rounded-xl border border-border shadow-sm overflow-hidden">
      <div class="px-6 py-5 border-b border-border bg-muted/50">
        <h2 class="text-xl font-semibold text-card-foreground">About Your Group</h2>
        <p class="text-sm text-muted-foreground mt-0.5">
          Tell the community what your group is about
        </p>
      </div>
      <div class="px-6 py-6 space-y-4">
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
                ? "border-red-500 bg-red-50/5 dark:bg-red-950/5 focus:border-red-500 focus:ring-1 focus:ring-red-500/20"
                : ""
            }`}
          />
          {#if $errors.web_url}
            <p class="text-xs sm:text-sm text-red-600 dark:text-red-400 flex items-center gap-1">
              <CircleX class="h-3 w-3 flex-shrink-0" />
              {$errors.web_url}
            </p>
          {/if}
        </div>
      </div>
    </div>

    <!-- Contact Information -->
    <div class="bg-card rounded-xl border border-border shadow-sm overflow-hidden">
      <div class="px-6 py-5 border-b border-border bg-muted/50">
        <h2 class="text-xl font-semibold text-card-foreground">Contact Information</h2>
        <p class="text-sm text-muted-foreground mt-0.5">
          We'll use this to send you a magic link to edit your group
        </p>
      </div>
      <div class="px-6 py-6 space-y-4">
        <div class="space-y-2">
          <Label for="email" class="text-sm sm:text-base">
            Contact Email *
          </Label>
          <Input
            id="email"
            type="email"
            bind:value={$form.email}
            placeholder="organizer@portlandbikeclub.com"
            aria-invalid={!!$errors.email}
            class={`text-base transition-colors ${
              $errors.email
                ? "border-red-500 bg-red-50/5 dark:bg-red-950/5 focus:border-red-500 focus:ring-1 focus:ring-red-500/20"
                : ""
            }`}
          />
          {#if $errors.email}
            <p class="text-xs sm:text-sm text-red-600 dark:text-red-400 flex items-center gap-1">
              <CircleX class="h-3 w-3 flex-shrink-0" />
              {$errors.email}
            </p>
          {/if}
          <p class="text-xs text-muted-foreground">
            You'll receive a magic link via this email to edit your group
            information anytime
          </p>
        </div>
      </div>
    </div>

    <!-- Group Marker -->
    <div class="bg-card rounded-xl border border-border shadow-sm overflow-hidden">
      <div class="px-6 py-5 border-b border-border bg-muted/50">
        <h2 class="text-xl font-semibold text-card-foreground">Group Marker</h2>
        <p class="text-sm text-muted-foreground mt-0.5">
          Add a custom marker for your group's rides on the map
        </p>
      </div>
      <div class="px-6 py-6 space-y-4">
        <CustomMarkerBuilder
          bind:this={customMarkerBuilderRef}
          cityCode={data.city}
          onUploadComplete={(uuid) => {
            markerImageUUID = uuid;
            $form.image_uuid = uuid;
          }}
          onUploadError={(error) => {
            console.error("Marker upload error:", error);
          }}
        />

        <p class="text-xs sm:text-sm text-muted-foreground">
          Your marker image will be automatically resized to 64x64px and added
          to your city's marker spritesheet for display on the map.
        </p>
      </div>
    </div>

    <!-- Submit Button -->
    <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 pt-4">
      <p class="text-sm text-muted-foreground text-center sm:text-left">
        * Required fields
      </p>
      <Button
        type="submit"
        disabled={$delayed ||
          codeCheckState === "checking" ||
          codeCheckState === "unavailable"}
        size="lg"
        class="w-full sm:w-auto bg-yellow-400 text-black hover:bg-yellow-300 gap-2"
      >
        {#if $delayed}
          <div class="h-4 w-4 border-2 border-white/30 border-t-white rounded-full animate-spin"></div>
          {#if !markerImageUUID}
            <span>Generating marker...</span>
          {:else}
            <span>Registering...</span>
          {/if}
        {:else}
          <span>Register Group</span>
        {/if}
      </Button>
    </div>

    <div class="text-center text-sm text-muted-foreground pb-4 pt-6">
      You'll receive a magic link via email to edit your group information.
    </div>
  </form>
  </div>
</div>
