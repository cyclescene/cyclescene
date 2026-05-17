<script lang="ts">
  import { browser } from "$app/environment";
  import { superForm } from "sveltekit-superforms";
  import { zod4Client as zodClient } from "sveltekit-superforms/adapters";
  import {
    rideSubmissionSchema,
    audienceOptions,
    dateTypeOptions,
  } from "$lib/schemas/ride";
  import GroupSelector from "$lib/components/ride-form/GroupSelector.svelte";
  import DateTimePicker from "$lib/components/ride-form/DateTimePicker.svelte";
  import ImageUploader from "$lib/components/ride-form/ImageUploader.svelte";

  // shadcn imports
  import { Button } from "$lib/components/ui/button";
  import { Input } from "$lib/components/ui/input";
  import { Label } from "$lib/components/ui/label";
  import { Textarea } from "$lib/components/ui/textarea";
  import { Checkbox } from "$lib/components/ui/checkbox";
  import * as Select from "$lib/components/ui/select";
  import * as Card from "$lib/components/ui/card";
  import { Separator } from "$lib/components/ui/separator";
  import { CircleX, Link, MapPin, AlertTriangle } from "lucide-svelte";

  interface Props {
    data: {
      form: any;
      token: string;
      city: string;
    };
  }

  let { data }: Props = $props();

  const { form, errors, enhance, delayed, message } = superForm(data.form, {
    id: 'ride-submission-form',
    validators: zodClient(rideSubmissionSchema),
    dataType: "json",
    resetForm: false,
    onError({ result }) {
      $message = result.error.message;
    },
  });

  let linkedEventDate = $state("");
  let linkedEventTime = $state("");

  let isLinkedEvent = $derived(Boolean(($form.web_url || "").trim()));

  $effect(() => {
    if (!isLinkedEvent) {
      return;
    }

    $form.date_type = "S";
    $form.is_loop_ride = false;
    $form.audience = $form.audience || "G";

    if (!$form.web_name) {
      $form.web_name = getLinkLabel($form.web_url || "");
    }

    if (linkedEventDate && linkedEventTime) {
      $form.occurrences = [
        {
          start_date: linkedEventDate,
          start_time: `${linkedEventTime}:00`,
          event_duration_minutes: 0,
          event_time_details: "",
          newsflash: "",
        },
      ];
    } else {
      $form.occurrences = [];
    }
  });

  function getLinkLabel(url: string) {
    try {
      const hostname = new URL(url).hostname.replace(/^www\./, "").toLowerCase();
      if (hostname.includes("strava.com") || hostname.includes("strava.app.link")) return "Strava";
      if (hostname.includes("instagram.com")) return "Instagram";
      if (hostname.includes("facebook.com") || hostname.includes("fb.me")) return "Facebook";
      if (hostname.includes("meetup.com")) return "Meetup";
    } catch {
      return "";
    }
    return "Event link";
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
    <!-- Manual Form Mode -->
    <header class="mb-10">
      <div class="inline-flex items-center gap-2 text-sm font-medium text-slate-500 mb-3">
        <MapPin class="h-4 w-4" />
        <span class="uppercase tracking-wide">{data.city.toUpperCase()}</span>
      </div>

      <div class="flex flex-col sm:flex-row sm:items-start sm:justify-between gap-6">
        <div>
          <h1 class="text-4xl sm:text-5xl font-bold tracking-tight text-card-foreground mb-4 leading-tight">
            Host a Ride
          </h1>
          <p class="text-lg text-muted-foreground">
            Share your ride with the {data.city.toUpperCase()} cycling community
          </p>
        </div>
      </div>
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
    <!-- Hidden input to include image_uuid in form submission -->
    <input type="hidden" name="image_uuid" bind:value={$form.image_uuid} />
    <input type="hidden" name="city" bind:value={$form.city} />

    <!-- Linked Event Entry -->
    <div class="bg-card rounded-xl border border-border shadow-sm overflow-hidden">
      <div class="px-6 py-5 border-b border-border bg-muted/50">
        <div class="flex items-center gap-2">
          <Link class="h-5 w-5 text-slate-500" />
          <h2 class="text-xl font-semibold text-card-foreground">Link an Existing Event</h2>
        </div>
        <p class="text-sm text-muted-foreground mt-0.5">
          Paste a Strava, Instagram, Facebook, Meetup, or event page link to use a simpler submission form.
        </p>
      </div>
      <div class="px-6 py-6 space-y-2">
        <Label for="linked_event_url" class="text-sm sm:text-base">External Event Link</Label>
        <Input
          id="linked_event_url"
          type="url"
          bind:value={$form.web_url}
          placeholder="https://www.strava.com/..."
          aria-invalid={!!$errors.web_url}
          class={`text-base transition-colors ${$errors.web_url ? "border-red-500 bg-red-50/5 dark:bg-red-950/5 focus:border-red-500 focus:ring-1 focus:ring-red-500/20" : ""}`}
        />
        {#if $errors.web_url}
          <p class="text-xs sm:text-sm text-red-600 dark:text-red-400 flex items-center gap-1"><CircleX class="h-3 w-3 flex-shrink-0" />{$errors.web_url}</p>
        {/if}
        {#if isLinkedEvent}
          <p class="text-sm text-muted-foreground">
            Linked event mode is on. Add the basics here; riders will use the link for full details.
          </p>
        {/if}
      </div>
    </div>

    <!-- Basic Information -->
    <div class="bg-card rounded-xl border border-border shadow-sm overflow-hidden">
      <div class="px-6 py-5 border-b border-border bg-muted/50">
        <h2 class="text-xl font-semibold text-card-foreground">Basic Information</h2>
        <p class="text-sm text-muted-foreground mt-0.5">
          Tell riders what your ride is about
        </p>
      </div>
      <div class="px-6 py-6 space-y-4 sm:space-y-4">
        <div class="space-y-2">
          <Label for="title" class="text-sm sm:text-base">Ride Title *</Label>
          <Input
            id="title"
            type="text"
            bind:value={$form.title}
            placeholder="Sunday Morning Coffee Cruise"
            aria-invalid={!!$errors.title}
            class={`text-base transition-colors ${$errors.title ? "border-red-500 bg-red-50/5 dark:bg-red-950/5 focus:border-red-500 focus:ring-1 focus:ring-red-500/20" : ""}`}
          />
          {#if $errors.title}
            <p class="text-xs sm:text-sm text-red-600 dark:text-red-400 flex items-center gap-1"><CircleX class="h-3 w-3 flex-shrink-0" />{$errors.title}</p>
          {/if}
        </div>

        <div class="space-y-2">
          <Label for="tinytitle" class="text-sm sm:text-base">
            Short Title (Optional)
            <span
              class="text-xs sm:text-sm text-muted-foreground font-normal block sm:inline"
            >
              For calendar displays
            </span>
          </Label>
          <Input
            id="tinytitle"
            type="text"
            bind:value={$form.tinytitle}
            placeholder="Coffee Cruise"
            maxlength={50}
            class="text-base"
          />
        </div>

        <div class="space-y-2">
          <Label for="description" class="text-sm sm:text-base"
            >Description {isLinkedEvent ? "(Optional)" : "*"}</Label
          >
          <Textarea
            id="description"
            bind:value={$form.description}
            placeholder={isLinkedEvent ? "Optional extra context. If left blank, Cycle Scene will show See link for full details." : "Join us for a casual morning ride through the city. We'll stop at local coffee shops along the way..."}
            rows={5}
            aria-invalid={!!$errors.description}
            class={`text-base transition-colors ${$errors.description ? "border-red-500 bg-red-50/5 dark:bg-red-950/5 focus:border-red-500 focus:ring-1 focus:ring-red-500/20" : ""}`}
          />
          {#if isLinkedEvent}
            <p class="text-sm text-muted-foreground">
              Optional for linked events. If left blank, Cycle Scene will show “See link for full details.”
            </p>
          {/if}
          {#if $errors.description}
            <p class="text-xs sm:text-sm text-red-600 dark:text-red-400 flex items-center gap-1"><CircleX class="h-3 w-3 flex-shrink-0" />{$errors.description}</p>
          {/if}
        </div>

        {#if !isLinkedEvent}
          <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
          <div class="space-y-2">
            <Label for="audience" class="text-sm sm:text-base">Audience</Label>
            <Select.Root bind:value={$form.audience} type="single">
              <Select.Trigger id="audience" class="text-base">
                {$form.audience ? $form.audience : "Select audience"}
              </Select.Trigger>
              <Select.Content>
                {#each audienceOptions as option}
                  <Select.Item value={option.value} class="text-sm sm:text-base"
                    >{option.label}</Select.Item
                  >
                {/each}
              </Select.Content>
            </Select.Root>
          </div>

          <div class="space-y-2">
            <Label for="ride_length" class="text-sm sm:text-base"
              >Ride Length</Label
            >
            <Input
              id="ride_length"
              type="text"
              bind:value={$form.ride_length}
              placeholder="10 miles, 2 hours"
              class="text-base"
            />
          </div>
          </div>

          <div class="space-y-4">
          <ImageUploader
            cityCode={data.city}
            entityType="ride"
            label="Ride Image (Optional)"
            description="Upload a photo of your ride or cycling community"
            onUploadComplete={(uuid) => {
              $form.image_uuid = uuid;
            }}
            onUploadError={(error) => {
              console.error("Image upload error:", error);
            }}
          />

          <div class="space-y-2">
            <Label for="image_url" class="text-sm sm:text-base"
              >Image URL (Optional)
              <span class="text-xs text-muted-foreground font-normal"
                >Alternative to upload above</span
              ></Label
            >
            <Input
              id="image_url"
              type="url"
              bind:value={$form.image_url}
              placeholder="https://example.com/image.jpg"
              class="text-base"
            />
            {#if $errors.image_url}
              <p class="text-xs sm:text-sm text-destructive">
                {$errors.image_url}
              </p>
            {/if}
          </div>
          </div>
        {/if}
      </div>
    </div>

    <!-- Location Information -->
    <div class="bg-card rounded-xl border border-border shadow-sm overflow-hidden">
      <div class="px-6 py-5 border-b border-border bg-muted/50">
        <h2 class="text-xl font-semibold text-card-foreground">Location</h2>
        <p class="text-sm text-muted-foreground mt-0.5">Where does the ride start and end?</p>
      </div>
      <div class="px-6 py-6 space-y-4">
        <div class="space-y-2">
          <Label for="venue_name">Starting Location Name *</Label>
          <Input
            id="venue_name"
            type="text"
            bind:value={$form.venue_name}
            placeholder="Pioneer Courthouse Square"
            aria-invalid={!!$errors.venue_name}
            class={`text-base transition-colors ${$errors.venue_name ? "border-red-500 bg-red-50/5 dark:bg-red-950/5 focus:border-red-500 focus:ring-1 focus:ring-red-500/20" : ""}`}
          />
          {#if $errors.venue_name}
            <p class="text-xs sm:text-sm text-red-600 dark:text-red-400 flex items-center gap-1"><CircleX class="h-3 w-3 flex-shrink-0" />{$errors.venue_name}</p>
          {/if}
        </div>

        <div class="space-y-2">
          <Label for="address">Address *</Label>
          <Input
            id="address"
            type="text"
            bind:value={$form.address}
            placeholder="701 SW 6th Ave, Portland, OR 97204"
            aria-invalid={!!$errors.address}
            class={`text-base transition-colors ${$errors.address ? "border-red-500 bg-red-50/5 dark:bg-red-950/5 focus:border-red-500 focus:ring-1 focus:ring-red-500/20" : ""}`}
          />
          {#if $errors.address}
            <p class="text-xs sm:text-sm text-red-600 dark:text-red-400 flex items-center gap-1"><CircleX class="h-3 w-3 flex-shrink-0" />{$errors.address}</p>
          {/if}
        </div>

        {#if !isLinkedEvent}
          <div class="space-y-2">
          <Label for="location_details">Location Details (Optional)</Label>
          <Textarea
            id="location_details"
            bind:value={$form.location_details}
            placeholder="Meet on the west side near the fountain"
            rows={2}
          />
          </div>

          <div class="flex items-center space-x-2">
          <Checkbox
            id="is_loop_ride"
            checked={$form.is_loop_ride}
            onCheckedChange={(checked) => {
              $form.is_loop_ride = checked === true;
            }}
          />
          <Label for="is_loop_ride" class="font-normal cursor-pointer">
            This is a loop ride (returns to start)
          </Label>
          </div>

          {#if !$form.is_loop_ride}
          <div class="space-y-2">
            <Label for="ending_location">Ending Location</Label>
            <Input
              id="ending_location"
              type="text"
              bind:value={$form.ending_location}
              placeholder="Waterfront Park"
            />
          </div>
          {/if}

          <div class="space-y-2">
          <Label for="area">Area/Neighborhood</Label>
          <Input
            id="area"
            type="text"
            bind:value={$form.area}
            placeholder="Downtown, Southeast, North Portland"
          />
          </div>
        {/if}
      </div>
    </div>

    <!-- Date & Time Section -->
    <div class="bg-card rounded-xl border border-border shadow-sm overflow-hidden">
      <div class="px-6 py-5 border-b border-border bg-muted/50">
        <h2 class="text-xl font-semibold text-card-foreground">Date & Time</h2>
        <p class="text-sm text-muted-foreground mt-0.5">
          When does your ride happen?
        </p>
      </div>
      <div class="px-6 py-6 space-y-4 sm:space-y-4">
        {#if isLinkedEvent}
          <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div class="space-y-2">
              <Label for="linked_event_date" class="text-sm sm:text-base">Date *</Label>
              <Input
                id="linked_event_date"
                type="date"
                bind:value={linkedEventDate}
                class="text-base"
              />
            </div>
            <div class="space-y-2">
              <Label for="linked_event_time" class="text-sm sm:text-base">Start Time *</Label>
              <Input
                id="linked_event_time"
                type="time"
                bind:value={linkedEventTime}
                class="text-base"
              />
            </div>
          </div>
          {#if $errors.occurrences}
            <p class="text-xs sm:text-sm text-red-600 dark:text-red-400 flex items-center gap-1">
              <CircleX class="h-3 w-3 flex-shrink-0" />
              {Array.isArray($errors.occurrences) ? $errors.occurrences[0] : typeof $errors.occurrences === 'object' && $errors.occurrences?._errors ? $errors.occurrences._errors[0] : $errors.occurrences}
            </p>
          {/if}
        {:else}
          <div class="space-y-2">
          <Label class="text-sm sm:text-base">Date Type *</Label>
          <Select.Root bind:value={$form.date_type} type="single">
            <Select.Trigger id="date_type" class="text-base">
              {$form.date_type ? $form.date_type : "Select date type"}
            </Select.Trigger>
            <Select.Content>
              {#each dateTypeOptions as option}
                <Select.Item value={option.value} class="text-sm sm:text-base"
                  >{option.label}</Select.Item
                >
              {/each}
            </Select.Content>
          </Select.Root>
          {#if $errors.date_type}
            <p class="text-xs sm:text-sm text-destructive">
              {$errors.date_type}
            </p>
          {/if}
        </div>

        {#if browser}
          <DateTimePicker
            bind:occurrences={$form.occurrences}
            dateType={$form.date_type}
            onupdate={(occs) => {
              $form.occurrences = occs;
            }}
            error={Array.isArray($errors.occurrences) ? $errors.occurrences[0] : typeof $errors.occurrences === 'object' && $errors.occurrences?._errors ? $errors.occurrences._errors[0] : typeof $errors.occurrences === 'string' ? $errors.occurrences : undefined}
          />
        {/if}
        {/if}
      </div>
    </div>

    <!-- Contact Information -->
    <div class="bg-card rounded-xl border border-border shadow-sm overflow-hidden">
      <div class="px-6 py-5 border-b border-border bg-muted/50">
        <h2 class="text-xl font-semibold text-card-foreground">Contact Information</h2>
        <p class="text-sm text-muted-foreground mt-0.5">How can riders reach you?</p>
      </div>
      <div class="px-6 py-6 space-y-4">
        <div class="space-y-2">
          <Label for="organizer_name">Your Name *</Label>
          <Input
            id="organizer_name"
            type="text"
            bind:value={$form.organizer_name}
            placeholder="Jane Doe"
            aria-invalid={!!$errors.organizer_name}
            class={`text-base transition-colors ${$errors.organizer_name ? "border-red-500 bg-red-50/5 dark:bg-red-950/5 focus:border-red-500 focus:ring-1 focus:ring-red-500/20" : ""}`}
          />
          {#if $errors.organizer_name}
            <p class="text-xs sm:text-sm text-red-600 dark:text-red-400 flex items-center gap-1"><CircleX class="h-3 w-3 flex-shrink-0" />{$errors.organizer_name}</p>
          {/if}

          <div class="flex items-center space-x-2 mt-2">
            <Checkbox
              id="hide_contact_name"
              checked={$form.hide_contact_name}
              onCheckedChange={(checked) => {
                $form.hide_contact_name = checked === true;
              }}
            />
            <Label
              for="hide_contact_name"
              class="font-normal text-sm cursor-pointer"
            >
              Hide my name from public listing
            </Label>
          </div>
        </div>

        <div class="space-y-2">
          <Label for="organizer_email">Email *</Label>
          <Input
            id="organizer_email"
            type="email"
            bind:value={$form.organizer_email}
            placeholder="jane@example.com"
            aria-invalid={!!$errors.organizer_email}
            class={`text-base transition-colors ${$errors.organizer_email ? "border-red-500 bg-red-50/5 dark:bg-red-950/5 focus:border-red-500 focus:ring-1 focus:ring-red-500/20" : ""}`}
          />
          {#if $errors.organizer_email}
            <p class="text-xs sm:text-sm text-red-600 dark:text-red-400 flex items-center gap-1"><CircleX class="h-3 w-3 flex-shrink-0" />{$errors.organizer_email}</p>
          {/if}

          <div class="flex items-center space-x-2 mt-2">
            <Checkbox
              id="hide_email"
              checked={$form.hide_email}
              onCheckedChange={(checked) => {
                $form.hide_email = checked === true;
              }}
            />
            <Label for="hide_email" class="font-normal text-sm cursor-pointer">
              Hide email from public listing
            </Label>
          </div>
        </div>

        {#if !isLinkedEvent}
          <div class="space-y-2">
          <Label for="organizer_phone">Phone (Optional)</Label>
          <Input
            id="organizer_phone"
            type="tel"
            bind:value={$form.organizer_phone}
            placeholder="(555) 123-4567"
          />

          {#if $form.organizer_phone}
            <div class="flex items-center space-x-2 mt-2">
              <Checkbox
                id="hide_phone"
                checked={$form.hide_phone}
                onCheckedChange={(checked) => {
                  $form.hide_phone = checked === true;
                }}
              />
              <Label
                for="hide_phone"
                class="font-normal text-sm cursor-pointer"
              >
                Hide phone from public listing
              </Label>
            </div>
          {/if}
          </div>

          <Separator />

          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div class="space-y-2">
            <Label for="web_url">Website URL (Optional)</Label>
            <Input
              id="web_url"
              type="url"
              bind:value={$form.web_url}
              placeholder="https://example.com"
              aria-invalid={!!$errors.web_url}
              class={`text-base transition-colors ${$errors.web_url ? "border-red-500 bg-red-50/5 dark:bg-red-950/5 focus:border-red-500 focus:ring-1 focus:ring-red-500/20" : ""}`}
            />
            {#if $errors.web_url}
              <p class="text-xs sm:text-sm text-red-600 dark:text-red-400 flex items-center gap-1"><CircleX class="h-3 w-3 flex-shrink-0" />{$errors.web_url}</p>
            {/if}
          </div>

          <div class="space-y-2">
            <Label for="web_name">Website Name</Label>
            <Input
              id="web_name"
              type="text"
              bind:value={$form.web_name}
              placeholder="Our Cycling Club"
            />
          </div>
          </div>
        {/if}
      </div>
    </div>

    <!-- Group Association -->
    <div class="bg-card rounded-xl border border-border shadow-sm overflow-hidden">
      <div class="px-6 py-5 border-b border-border bg-muted/50">
        <h2 class="text-xl font-semibold text-card-foreground">Group Association</h2>
        <p class="text-sm text-muted-foreground mt-0.5">
          Link this ride to a cycling group
        </p>
      </div>
      <div class="px-6 py-6">
        <GroupSelector
          bind:value={$form.group_code}
          onchange={(val) => {
            $form.group_code = val;
          }}
          city={data.city}
          error={Array.isArray($errors.group_code) ? $errors.group_code[0] : typeof $errors.group_code === 'object' && $errors.group_code?._errors ? $errors.group_code._errors[0] : typeof $errors.group_code === 'string' ? $errors.group_code : undefined}
        />
      </div>
    </div>

    {#if !isLinkedEvent}
      <!-- Additional Details -->
      <div class="bg-card rounded-xl border border-border shadow-sm overflow-hidden">
      <div class="px-6 py-5 border-b border-border bg-muted/50">
        <h2 class="text-xl font-semibold text-card-foreground">Additional Details</h2>
        <p class="text-sm text-muted-foreground mt-0.5">
          Any extra information riders should know
        </p>
      </div>
      <div class="px-6 py-6 space-y-4">
        <div class="space-y-2">
          <Label for="newsflash">
            Newsflash (Optional)
            <span class="text-sm text-muted-foreground font-normal">
              - Important updates or changes
            </span>
          </Label>
          <Textarea
            id="newsflash"
            bind:value={$form.newsflash}
            placeholder="Meeting location changed! Now meeting at the east entrance."
            rows={3}
            maxlength={500}
          />
          <p class="text-xs text-muted-foreground">
            {$form.newsflash?.length || 0}/500 characters
          </p>
        </div>
      </div>
      </div>
    {/if}

    <!-- Submit Button -->
    <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 pt-4">
      <p class="text-sm text-muted-foreground text-center sm:text-left">
        * Required fields
      </p>
      <Button
        type="submit"
        disabled={$delayed}
        size="lg"
        class="w-full sm:w-auto bg-yellow-400 text-black hover:bg-yellow-300 gap-2"
      >
        {#if $delayed}
          <div class="h-4 w-4 border-2 border-white/30 border-t-white rounded-full animate-spin"></div>
          <span>Submitting...</span>
        {:else}
          <span>Submit Ride for Review</span>
        {/if}
      </Button>
    </div>

    <div class="text-center text-sm text-muted-foreground pb-4 pt-6">
      Your ride will be reviewed before appearing on CycleScene.
      <br class="hidden sm:block" />
      <span class="block sm:inline mt-1 sm:mt-0">
        You'll receive a magic link via email to edit your ride.
      </span>
    </div>
  </form>
  </div>
</div>
