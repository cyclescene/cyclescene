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
  import { Edit2, Loader, CircleX, MapPin, AlertTriangle, Check } from "@lucide/svelte";

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

<div class="min-h-screen bg-gradient-to-b from-slate-50 to-white dark:from-slate-950 dark:to-slate-900">
  <div class="max-w-5xl mx-auto px-4 sm:px-6 lg:px-8 py-8 sm:py-12">

    <header class="mb-10">
      <div class="inline-flex items-center gap-2 text-sm font-medium text-slate-500 mb-3">
        <MapPin class="h-4 w-4" />
        <span class="uppercase tracking-wide">{data.city.toUpperCase()}</span>
      </div>

      <h1 class="text-4xl sm:text-5xl font-bold tracking-tight text-slate-900 dark:text-slate-50 mb-4 leading-tight">
        Edit Group Settings
      </h1>
      <p class="text-lg text-slate-600 dark:text-slate-400">
        Update your group information for {data.city.toUpperCase()}
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

  <!-- Group Summary Card -->
  <div class="mb-6 bg-white dark:bg-slate-900 rounded-2xl border border-slate-200 dark:border-slate-800 shadow-sm overflow-hidden">
    <div class="px-6 py-5 border-b border-slate-200 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-900/50">
      <h2 class="text-xl font-semibold text-slate-900 dark:text-slate-50">Group Summary</h2>
    </div>
    <div class="px-6 py-6 space-y-6">
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
                ? "border-red-500 bg-red-50/5 dark:bg-red-950/5 focus:border-red-500 focus:ring-1 focus:ring-red-500/20"
                : ""
            }`}
            autofocus
          />
          {#if $errors.name}
            <p class="text-xs sm:text-sm text-red-600 dark:text-red-400 flex items-center gap-1 mt-2">
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
    </div>
  </div>

  <form method="POST" use:enhance class="space-y-6">

    <!-- Description & Details -->
    <div class="bg-white dark:bg-slate-900 rounded-2xl border border-slate-200 dark:border-slate-800 shadow-sm overflow-hidden">
      <div class="px-6 py-5 border-b border-slate-200 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-900/50">
        <h2 class="text-xl font-semibold text-slate-900 dark:text-slate-50">About Your Group</h2>
        <p class="text-sm text-slate-500 dark:text-slate-400 mt-0.5">
          Update your group description and website
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

    <!-- Group Marker -->
    <div class="bg-white dark:bg-slate-900 rounded-2xl border border-slate-200 dark:border-slate-800 shadow-sm overflow-hidden">
      <div class="px-6 py-5 border-b border-slate-200 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-900/50">
        <h2 class="text-xl font-semibold text-slate-900 dark:text-slate-50">Group Marker (Optional)</h2>
        <p class="text-sm text-slate-500 dark:text-slate-400 mt-0.5">
          Update your group's marker with a new image
        </p>
      </div>
      <div class="px-6 py-6 space-y-4">
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
      </div>
    </div>

    <!-- Submit Button -->
    <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 pt-4">
      <p class="text-sm text-slate-500 dark:text-slate-400 text-center sm:text-left">
        * Required fields
      </p>
      <Button
        type="submit"
        disabled={$delayed}
        size="lg"
        class="w-full sm:w-auto bg-slate-900 hover:bg-slate-800 dark:bg-slate-50 dark:hover:bg-slate-200 gap-2"
      >
        {#if $delayed}
          <div class="h-4 w-4 border-2 border-white/30 border-t-white rounded-full animate-spin"></div>
          <span>Saving Changes...</span>
        {:else}
          <Check class="h-4 w-4" />
          <span>Save Changes</span>
        {/if}
      </Button>
    </div>
  </form>

  <div class="text-center text-sm text-slate-500 dark:text-slate-400 pb-4 mt-6">
    <p>
      Your group code and city cannot be changed. Upload a new marker image above
      if you'd like to update your group's appearance on the map.
    </p>
  </div>

  <div class="flex justify-center mt-8">
    <Button variant="outline" class="gap-2" href={backUrl}>
      <span>← Back to CycleScene</span>
    </Button>
  </div>
  </div>
</div>
