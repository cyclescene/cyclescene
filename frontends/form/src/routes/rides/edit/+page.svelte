<script lang="ts">
  import { page } from "$app/state";
  import { Button } from "$lib/components/ui/button";
  import { Input } from "$lib/components/ui/input";
  import { Checkbox } from "$lib/components/ui/checkbox";
  import { Textarea } from "$lib/components/ui/textarea";
  import * as Select from "$lib/components/ui/select";
  import * as Card from "$lib/components/ui/card";
  import * as Alert from "$lib/components/ui/alert";
  import { CircleX, Edit2, Check, X, AlertTriangle, Calendar, Clock, MapPin, CheckCircle2, AlertCircle } from "@lucide/svelte";
  import { audienceOptions } from "$lib/schemas/ride";

  interface RideData {
    event: {
      id: number;
      title: string;
      description: string;
      venue_name: string;
      address: string;
      location_details: string;
      ending_location: string;
      is_loop_ride: boolean;
      audience: string;
      ride_length: string;
      area: string;
      organizer_name: string;
      organizer_email: string;
      organizer_phone: string;
      web_url: string;
      web_name: string;
      city: string;
      image_url: string;
      newsflash: string;
      source?: string;
      occurrences: Occurrence[];
    };
    is_published: boolean;
  }

  interface Occurrence {
    id: number;
    start_date: string;
    start_time: string;
    event_duration_minutes: number;
    event_time_details: string;
    newsflash?: string;
    is_cancelled: boolean;
  }

  interface EditingOccurrence extends Occurrence {
    isEditing?: boolean;
    isSaving?: boolean;
  }

  let { data }: any = $props();
  let rideData: RideData | null = $state(data?.rideData || null);
  let occurrences: EditingOccurrence[] = $state([]);
  let successMessage = $state("");
  let errorMessage = $state("");
  let token = $state(page.url.searchParams.get("token") || "");

  // Event editing state
  let isEditingEvent = $state(false);
  let isSavingEvent = $state(false);
  let editedDescription = $state("");
  let editedAudience = $state("");
  let editedRideLength = $state("");

  $effect(() => {
    if (rideData?.event.occurrences) {
      occurrences = rideData.event.occurrences.map(o => ({
        ...o,
        is_cancelled: o.is_cancelled || false
      }));
    }
  });

  const today = new Date().toISOString().split("T")[0];

  let pastOccurrences = $derived(
    occurrences.filter((o) => o.start_date < today),
  );
  let upcomingOccurrences = $derived(
    occurrences.filter((o) => o.start_date >= today),
  );

  const audienceTriggerContent = $derived(
    audienceOptions.find((opt) => opt.value === editedAudience)?.label ?? "Select audience"
  );

  const getCycleSceneDomain = (cityCode?: string): string => {
    if (!cityCode) return "https://cyclescene.cc";
    const cityDomains: Record<string, string> = {
      pdx: "https://pdx.cyclescene.cc",
      slc: "https://slc.cyclescene.cc",
    };
    return cityDomains[cityCode.toLowerCase()] || "https://cyclescene.cc";
  };

  const toggleEditEvent = () => {
    if (!isEditingEvent && rideData) {
      editedDescription = rideData.event.description;
      editedAudience = rideData.event.audience;
      editedRideLength = rideData.event.ride_length;
    }
    isEditingEvent = !isEditingEvent;
  };

  const saveEventDetails = async () => {
    if (!rideData) return;

    if (editedDescription.length < 10) {
      errorMessage = "Description must be at least 10 characters";
      return;
    }

    if (rideData.event.source === 'strava') {
      const confirmed = confirm(
        "⚠️ This event is from Strava.\n\n" +
        "Editing will convert it to a CycleScene event and stop syncing with Strava.\n\n" +
        "Continue?"
      );
      if (!confirmed) return;
    }

    isSavingEvent = true;
    errorMessage = "";

    try {
      const formData = new FormData();
      formData.append('description', editedDescription);
      formData.append('audience', editedAudience || '');
      formData.append('ride_length', editedRideLength || '');

      const response = await fetch(`?/updateEventDetails&token=${encodeURIComponent(token)}`, {
        method: 'POST',
        body: formData
      });

      if (!response.ok) {
        const text = await response.text();
        console.error('Update failed:', text);
        throw new Error("Failed to update event details");
      }

      const result = await response.json();

      if (result.type === 'failure') {
        throw new Error(result.data?.error || "Failed to update event details");
      }

      // Update local state
      if (rideData) {
        rideData.event.description = editedDescription;
        rideData.event.audience = editedAudience;
        rideData.event.ride_length = editedRideLength;
        if (rideData.event.source === 'strava') {
          rideData.event.source = undefined;
        }
      }

      isEditingEvent = false;
      successMessage = "Event details updated successfully!";
      setTimeout(() => {
        successMessage = "";
      }, 3000);
    } catch (err) {
      errorMessage = err instanceof Error ? err.message : "Failed to save changes";
    } finally {
      isSavingEvent = false;
    }
  };

  const toggleEdit = (occurrence: EditingOccurrence) => {
    const occurrenceToUpdate = occurrences.find((o) => o.id === occurrence.id);
    if (occurrenceToUpdate) {
      occurrenceToUpdate.isEditing = !occurrenceToUpdate.isEditing;
    }
  };

  const saveOccurrence = async (occurrence: EditingOccurrence) => {
    const occurrenceToUpdate = occurrences.find((o) => o.id === occurrence.id);
    if (!occurrenceToUpdate) return;

    occurrenceToUpdate.isSaving = true;
    errorMessage = "";

    try {
      const formData = new FormData();
      formData.append('occurrence_id', occurrence.id.toString());
      formData.append('start_time', occurrence.start_time);
      formData.append('event_duration_minutes', occurrence.event_duration_minutes.toString());
      formData.append('event_time_details', occurrence.event_time_details || '');
      formData.append('newsflash', occurrence.newsflash || '');
      formData.append('is_cancelled', occurrence.is_cancelled.toString());

      const response = await fetch(`?/updateOccurrence&token=${encodeURIComponent(token)}`, {
        method: 'POST',
        body: formData
      });

      if (!response.ok) {
        const text = await response.text();
        console.error('Server action failed:', text);
        throw new Error("Failed to save occurrence");
      }

      const result = await response.json();
      const actionResult = result?.data?.[0] || result;

      if (actionResult?.type === 'failure') {
        throw new Error(actionResult?.data?.error || "Failed to save occurrence");
      }

      occurrenceToUpdate.isEditing = false;
      successMessage = "Occurrence updated successfully!";
      setTimeout(() => {
        successMessage = "";
      }, 3000);
    } catch (err) {
      errorMessage =
        err instanceof Error ? err.message : "Failed to save changes";
    } finally {
      occurrenceToUpdate.isSaving = false;
    }
  };

  function formatDate(dateStr: string): string {
    const date = new Date(dateStr);
    return date.toLocaleDateString('en-US', { weekday: 'short', month: 'short', day: 'numeric' });
  }

  function formatTime(timeStr: string): string {
    const [hours, minutes] = timeStr.split(':');
    const hour = parseInt(hours);
    const ampm = hour >= 12 ? 'PM' : 'AM';
    const displayHour = hour % 12 || 12;
    return `${displayHour}:${minutes} ${ampm}`;
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

  .edit-mode-enter {
    animation: slideIn 0.4s cubic-bezier(0.16, 1, 0.3, 1);
  }
</style>

<div class="min-h-screen bg-gradient-to-b from-slate-50 to-white dark:from-slate-950 dark:to-slate-900">
  <div class="max-w-5xl mx-auto px-4 sm:px-6 lg:px-8 py-8 sm:py-12">

    <!-- Header -->
    <header class="mb-10">
      <div class="inline-flex items-center gap-2 text-sm font-medium text-slate-500 mb-3">
        <MapPin class="h-4 w-4" />
        <span class="uppercase tracking-wide">{rideData?.event.city || 'CycleScene'}</span>
      </div>

      <h1 class="text-4xl sm:text-5xl font-bold tracking-tight text-slate-900 dark:text-slate-50 mb-4 leading-tight">
        {rideData?.event.title}
      </h1>

      <p class="text-lg text-slate-600 dark:text-slate-400 mb-6">
        Manage your ride details and upcoming occurrences
      </p>

      <!-- Status Badge -->
      <div class="flex flex-wrap gap-3">
        {#if rideData?.is_published}
          <div class="inline-flex items-center gap-2 px-4 py-2 bg-emerald-50 dark:bg-emerald-950 border border-emerald-200 dark:border-emerald-800 rounded-full">
            <div class="h-2 w-2 rounded-full bg-emerald-500 animate-pulse"></div>
            <span class="text-sm font-medium text-emerald-700 dark:text-emerald-300">Live & Visible</span>
          </div>
        {:else}
          <div class="inline-flex items-center gap-2 px-4 py-2 bg-blue-50 dark:bg-blue-950 border border-blue-200 dark:border-blue-800 rounded-full">
            <Clock class="h-3.5 w-3.5 text-blue-600 dark:text-blue-400" />
            <span class="text-sm font-medium text-blue-700 dark:text-blue-300">Pending Review</span>
          </div>
        {/if}

        {#if rideData?.event.source === 'strava'}
          <div class="inline-flex items-center gap-2 px-4 py-2 bg-orange-50 dark:bg-orange-950 border border-orange-200 dark:border-orange-800 rounded-full">
            <svg class="h-3.5 w-3.5 text-orange-600 dark:text-orange-400" fill="currentColor" viewBox="0 0 24 24">
              <path d="M15.387 17.944l-2.089-4.116h-3.065L15.387 24l5.15-10.172h-3.066m-7.008-5.599l2.836 5.598h4.172L10.463 0l-7 13.828h4.169"/>
            </svg>
            <span class="text-sm font-medium text-orange-700 dark:text-orange-300">Strava Synced</span>
          </div>
        {/if}
      </div>

      <!-- Strava Warning (Expanded) -->
      {#if rideData?.event.source === 'strava'}
        <Alert.Root class="mt-6 border-amber-200 dark:border-amber-800 bg-gradient-to-br from-amber-50 to-orange-50 dark:from-amber-950/50 dark:to-orange-950/50 animate-slide-in">
          <AlertTriangle class="h-4 w-4 text-amber-600 dark:text-amber-400" />
          <Alert.Title class="text-amber-900 dark:text-amber-100">Imported from Strava</Alert.Title>
          <Alert.Description class="text-amber-800 dark:text-amber-200/90">
            <p class="leading-relaxed">
              Any edits will convert this to a CycleScene event and stop syncing with Strava.
            </p>
          </Alert.Description>
        </Alert.Root>
      {/if}
    </header>

    <!-- Messages -->
    {#if successMessage}
      <Alert.Root class="mb-6 animate-slide-in">
        <CheckCircle2 class="h-4 w-4" />
        <Alert.Title>Success</Alert.Title>
        <Alert.Description>{successMessage}</Alert.Description>
      </Alert.Root>
    {/if}

    {#if errorMessage}
      <Alert.Root variant="destructive" class="mb-6 animate-slide-in">
        <AlertCircle class="h-4 w-4" />
        <Alert.Title>Error</Alert.Title>
        <Alert.Description>{errorMessage}</Alert.Description>
      </Alert.Root>
    {/if}

    <!-- Main Content -->
    <div class="space-y-6">

      <!-- Ride Information Card -->
      <div class="bg-white dark:bg-slate-900 rounded-2xl border border-slate-200 dark:border-slate-800 shadow-sm overflow-hidden">
        <div class="px-6 py-5 border-b border-slate-200 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-900/50">
          <div class="flex items-center justify-between">
            <div>
              <h2 class="text-xl font-semibold text-slate-900 dark:text-slate-50">Event Details</h2>
              <p class="text-sm text-slate-500 dark:text-slate-400 mt-0.5">Core information for all occurrences</p>
            </div>
            {#if !isEditingEvent}
              <Button
                variant="outline"
                size="sm"
                onclick={toggleEditEvent}
                class="gap-2 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors"
              >
                <Edit2 class="h-3.5 w-3.5" />
                <span>Edit</span>
              </Button>
            {/if}
          </div>
        </div>

        <div class="px-6 py-6">
          {#if !isEditingEvent}
            <!-- Read Mode -->
            <div class="space-y-8 animate-fade-in">
              <!-- Description -->
              <div>
                <h3 class="text-xs font-semibold uppercase tracking-wider text-slate-500 dark:text-slate-400 mb-2">Description</h3>
                <p class="text-base text-slate-700 dark:text-slate-300 leading-relaxed whitespace-pre-wrap">
                  {rideData?.event.description}
                </p>
              </div>

              <!-- Details Grid -->
              <div class="grid grid-cols-1 sm:grid-cols-2 gap-x-8 gap-y-6">
                <div>
                  <h3 class="text-xs font-semibold uppercase tracking-wider text-slate-500 dark:text-slate-400 mb-2">Venue</h3>
                  <p class="text-base text-slate-900 dark:text-slate-100 font-medium">{rideData?.event.venue_name}</p>
                </div>
                <div>
                  <h3 class="text-xs font-semibold uppercase tracking-wider text-slate-500 dark:text-slate-400 mb-2">Address</h3>
                  <p class="text-base text-slate-900 dark:text-slate-100">{rideData?.event.address}</p>
                </div>
                <div>
                  <h3 class="text-xs font-semibold uppercase tracking-wider text-slate-500 dark:text-slate-400 mb-2">Audience</h3>
                  <p class="text-base text-slate-900 dark:text-slate-100">
                    {audienceOptions.find(opt => opt.value === rideData?.event.audience)?.label || "Not specified"}
                  </p>
                </div>
                <div>
                  <h3 class="text-xs font-semibold uppercase tracking-wider text-slate-500 dark:text-slate-400 mb-2">Ride Length</h3>
                  <p class="text-base text-slate-900 dark:text-slate-100">
                    {rideData?.event.ride_length || "Not specified"}
                  </p>
                </div>
              </div>

              {#if rideData?.event.location_details}
                <div>
                  <h3 class="text-xs font-semibold uppercase tracking-wider text-slate-500 dark:text-slate-400 mb-2">Location Details</h3>
                  <p class="text-base text-slate-700 dark:text-slate-300 leading-relaxed">{rideData.event.location_details}</p>
                </div>
              {/if}

              <!-- Organizer -->
              <div class="pt-6 border-t border-slate-200 dark:border-slate-800">
                <h3 class="text-xs font-semibold uppercase tracking-wider text-slate-500 dark:text-slate-400 mb-4">Organizer</h3>
                <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
                  <div class="flex items-center gap-3">
                    <div class="h-10 w-10 rounded-full bg-slate-100 dark:bg-slate-800 flex items-center justify-center">
                      <span class="text-sm font-semibold text-slate-600 dark:text-slate-300">
                        {rideData?.event.organizer_name?.charAt(0).toUpperCase()}
                      </span>
                    </div>
                    <div>
                      <p class="text-sm font-medium text-slate-900 dark:text-slate-100">{rideData?.event.organizer_name}</p>
                      <p class="text-sm text-slate-500 dark:text-slate-400">{rideData?.event.organizer_email}</p>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          {:else}
            <!-- Edit Mode -->
            <div class="space-y-6 edit-mode-enter">
              <div>
                <label for="description" class="block text-sm font-semibold text-slate-700 dark:text-slate-300 mb-2">
                  Description
                </label>
                <Textarea
                  id="description"
                  bind:value={editedDescription}
                  rows={6}
                  class="resize-none font-sans text-base"
                  placeholder="Describe your ride..."
                />
                <p class="text-xs text-slate-500 dark:text-slate-400 mt-1.5">Minimum 10 characters</p>
              </div>

              <div class="grid grid-cols-1 sm:grid-cols-2 gap-6">
                <div>
                  <label for="audience" class="block text-sm font-semibold text-slate-700 dark:text-slate-300 mb-2">
                    Audience
                  </label>
                  <Select.Root type="single" bind:value={editedAudience}>
                    <Select.Trigger class="w-full">
                      {audienceTriggerContent}
                    </Select.Trigger>
                    <Select.Content>
                      {#each audienceOptions as option}
                        <Select.Item value={option.value} label={option.label}>
                          {option.label}
                        </Select.Item>
                      {/each}
                    </Select.Content>
                  </Select.Root>
                </div>

                <div>
                  <label for="ride_length" class="block text-sm font-semibold text-slate-700 dark:text-slate-300 mb-2">
                    Ride Length
                  </label>
                  <Input
                    id="ride_length"
                    bind:value={editedRideLength}
                    placeholder="e.g., 15 miles, 90 minutes"
                    maxlength={50}
                  />
                </div>
              </div>

              <div class="flex gap-3 pt-4">
                <Button
                  disabled={isSavingEvent}
                  onclick={saveEventDetails}
                  class="gap-2 bg-slate-900 hover:bg-slate-800 dark:bg-slate-50 dark:hover:bg-slate-200"
                >
                  {#if isSavingEvent}
                    <div class="h-4 w-4 border-2 border-white/30 border-t-white rounded-full animate-spin"></div>
                    <span>Saving...</span>
                  {:else}
                    <Check class="h-4 w-4" />
                    <span>Save Changes</span>
                  {/if}
                </Button>
                <Button
                  variant="outline"
                  disabled={isSavingEvent}
                  onclick={toggleEditEvent}
                  class="gap-2"
                >
                  <X class="h-4 w-4" />
                  <span>Cancel</span>
                </Button>
              </div>
            </div>
          {/if}
        </div>
      </div>

      <!-- Upcoming Occurrences -->
      {#if upcomingOccurrences.length > 0}
        <div class="bg-white dark:bg-slate-900 rounded-2xl border border-slate-200 dark:border-slate-800 shadow-sm overflow-hidden">
          <div class="px-6 py-5 border-b border-slate-200 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-900/50">
            <h2 class="text-xl font-semibold text-slate-900 dark:text-slate-50">Upcoming Rides</h2>
            <p class="text-sm text-slate-500 dark:text-slate-400 mt-0.5">Edit timing or cancel individual occurrences</p>
          </div>

          <div class="divide-y divide-slate-200 dark:divide-slate-800">
            {#each upcomingOccurrences as occurrence (occurrence.id)}
              <div class="px-6 py-5 {occurrence.is_cancelled ? 'bg-slate-50/50 dark:bg-slate-900/50' : ''}">
                {#if !occurrence.isEditing}
                  <!-- View Mode -->
                  <div class="flex items-start justify-between gap-4">
                    <div class="flex-1 space-y-3">
                      <div class="flex items-baseline gap-3 flex-wrap">
                        <div class="flex items-center gap-2">
                          <Calendar class="h-4 w-4 text-slate-400" />
                          <span class="text-lg font-semibold text-slate-900 dark:text-slate-50">
                            {formatDate(occurrence.start_date)}
                          </span>
                        </div>

                        {#if occurrence.is_cancelled}
                          <span class="line-through text-slate-400 dark:text-slate-500">
                            {formatTime(occurrence.start_time)}
                          </span>
                          <span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-red-100 dark:bg-red-900/30 text-red-700 dark:text-red-400">
                            Cancelled
                          </span>
                        {:else}
                          <div class="flex items-center gap-1.5 text-slate-600 dark:text-slate-400">
                            <Clock class="h-3.5 w-3.5" />
                            <span class="font-medium">{formatTime(occurrence.start_time)}</span>
                          </div>
                        {/if}
                      </div>

                      {#if occurrence.event_duration_minutes}
                        <p class="text-sm text-slate-600 dark:text-slate-400">
                          Duration: <span class="font-medium">{occurrence.event_duration_minutes} minutes</span>
                        </p>
                      {/if}

                      {#if occurrence.event_time_details}
                        <p class="text-sm text-slate-700 dark:text-slate-300 leading-relaxed">
                          {occurrence.event_time_details}
                        </p>
                      {/if}

                      {#if occurrence.newsflash}
                        <div class="flex items-start gap-2 p-3 bg-amber-50 dark:bg-amber-950/30 border border-amber-200 dark:border-amber-800/50 rounded-lg">
                          <AlertTriangle class="h-4 w-4 text-amber-600 dark:text-amber-400 flex-shrink-0 mt-0.5" />
                          <p class="text-sm text-amber-900 dark:text-amber-100 font-medium">
                            {occurrence.newsflash}
                          </p>
                        </div>
                      {/if}
                    </div>

                    <Button
                      variant="outline"
                      size="sm"
                      onclick={() => toggleEdit(occurrence)}
                      class="gap-2 flex-shrink-0"
                    >
                      <Edit2 class="h-3.5 w-3.5" />
                      <span>Edit</span>
                    </Button>
                  </div>
                {:else}
                  <!-- Edit Mode -->
                  <div class="space-y-5 edit-mode-enter">
                    <div class="grid grid-cols-1 sm:grid-cols-2 gap-5">
                      <div>
                        <label class="block text-sm font-semibold text-slate-700 dark:text-slate-300 mb-2">
                          Start Time
                        </label>
                        <Input
                          type="time"
                          bind:value={occurrence.start_time}
                        />
                      </div>

                      <div>
                        <label class="block text-sm font-semibold text-slate-700 dark:text-slate-300 mb-2">
                          Duration (minutes)
                        </label>
                        <Input
                          type="number"
                          bind:value={occurrence.event_duration_minutes}
                          min="0"
                        />
                      </div>
                    </div>

                    <div>
                      <label class="block text-sm font-semibold text-slate-700 dark:text-slate-300 mb-2">
                        Time Details <span class="text-slate-400 font-normal">(Optional)</span>
                      </label>
                      <Input
                        type="text"
                        bind:value={occurrence.event_time_details}
                        placeholder="e.g., Meet at the fountain"
                      />
                    </div>

                    <div>
                      <label class="block text-sm font-semibold text-slate-700 dark:text-slate-300 mb-2">
                        Alert/Update <span class="text-slate-400 font-normal">(Optional)</span>
                      </label>
                      <Input
                        type="text"
                        bind:value={occurrence.newsflash}
                        placeholder="e.g., Route change due to construction"
                        maxlength={500}
                      />
                      <p class="text-xs text-slate-500 dark:text-slate-400 mt-1.5">Special message for this date (max 500 characters)</p>
                    </div>

                    <div class="flex items-center gap-3 p-4 bg-slate-50 dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-lg">
                      <Checkbox
                        id={`cancel-${occurrence.id}`}
                        bind:checked={occurrence.is_cancelled}
                      />
                      <label
                        for={`cancel-${occurrence.id}`}
                        class="text-sm font-medium text-slate-700 dark:text-slate-300 cursor-pointer"
                      >
                        Cancel this occurrence
                      </label>
                    </div>

                    <div class="flex gap-3 pt-2">
                      <Button
                        disabled={occurrence.isSaving}
                        onclick={() => saveOccurrence(occurrence)}
                        class="gap-2 bg-slate-900 hover:bg-slate-800 dark:bg-slate-50 dark:hover:bg-slate-200"
                      >
                        {#if occurrence.isSaving}
                          <div class="h-4 w-4 border-2 border-white/30 border-t-white rounded-full animate-spin"></div>
                          <span>Saving...</span>
                        {:else}
                          <Check class="h-4 w-4" />
                          <span>Save</span>
                        {/if}
                      </Button>
                      <Button
                        variant="outline"
                        disabled={occurrence.isSaving}
                        onclick={() => toggleEdit(occurrence)}
                        class="gap-2"
                      >
                        <X class="h-4 w-4" />
                        <span>Cancel</span>
                      </Button>
                    </div>
                  </div>
                {/if}
              </div>
            {/each}
          </div>
        </div>
      {/if}

      <!-- Past Occurrences -->
      {#if pastOccurrences.length > 0}
        <div class="bg-white dark:bg-slate-900 rounded-2xl border border-slate-200 dark:border-slate-800 shadow-sm overflow-hidden opacity-75">
          <div class="px-6 py-5 border-b border-slate-200 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-900/50">
            <h2 class="text-xl font-semibold text-slate-900 dark:text-slate-50">Past Rides</h2>
            <p class="text-sm text-slate-500 dark:text-slate-400 mt-0.5">These occurrences have already happened</p>
          </div>

          <div class="px-6 py-5">
            <div class="space-y-3">
              {#each pastOccurrences as occurrence (occurrence.id)}
                <div class="flex items-center gap-3 p-3 bg-slate-50 dark:bg-slate-900/50 rounded-lg border border-slate-200 dark:border-slate-800">
                  <Calendar class="h-4 w-4 text-slate-400" />
                  <span class="text-sm font-medium text-slate-600 dark:text-slate-400">
                    {formatDate(occurrence.start_date)}
                  </span>
                  <span class="text-sm text-slate-500 dark:text-slate-500">
                    {formatTime(occurrence.start_time)}
                  </span>
                  {#if occurrence.is_cancelled}
                    <span class="text-xs px-2 py-0.5 bg-red-100 dark:bg-red-900/30 text-red-700 dark:text-red-400 rounded-full">
                      Cancelled
                    </span>
                  {/if}
                </div>
              {/each}
            </div>
          </div>
        </div>
      {/if}

      <!-- Footer Actions -->
      {#if successMessage}
        <div class="flex justify-center pt-4">
          <Button variant="outline" class="gap-2">
            <a href={getCycleSceneDomain(rideData?.event.city)} class="flex items-center gap-2">
              <span>← Back to {rideData?.event.city?.toUpperCase() || "CycleScene"}</span>
            </a>
          </Button>
        </div>
      {/if}

    </div>
  </div>
</div>
