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

  @keyframes fadeIn {
    from {
      opacity: 0;
    }
    to {
      opacity: 1;
    }
  }

  @keyframes scaleIn {
    from {
      opacity: 0;
      transform: scale(0.98);
    }
    to {
      opacity: 1;
      transform: scale(1);
    }
  }

  .edit-mode-enter {
    animation: scaleIn 0.35s cubic-bezier(0.16, 1, 0.3, 1);
  }

  .read-mode-enter {
    animation: fadeIn 0.3s cubic-bezier(0.16, 1, 0.3, 1);
  }

  .field-group {
    transition: all 0.2s cubic-bezier(0.16, 1, 0.3, 1);
  }

  .field-group:focus-within {
    transform: translateY(-1px);
  }
</style>

<div class="min-h-screen bg-gradient-to-b from-slate-50 to-white dark:from-slate-950 dark:to-slate-900">
  <div class="max-w-5xl mx-auto px-4 sm:px-6 lg:px-8 py-8 sm:py-12">

    <!-- Header -->
    <header class="mb-12">
      <div class="inline-flex items-center gap-2.5 text-sm font-bold text-slate-500 dark:text-slate-400 mb-4">
        <MapPin class="h-4 w-4" />
        <span class="uppercase tracking-widest">{rideData?.event.city || 'CycleScene'}</span>
      </div>

      <h1 class="text-4xl sm:text-5xl lg:text-6xl font-black tracking-tight text-slate-900 dark:text-slate-50 mb-5 leading-[1.1]">
        {rideData?.event.title}
      </h1>

      <p class="text-lg text-slate-600 dark:text-slate-400 mb-8 max-w-2xl">
        Manage your ride details and upcoming occurrences
      </p>

      <!-- Status Badge -->
      <div class="flex flex-wrap gap-3 mb-6">
        {#if rideData?.is_published}
          <div class="inline-flex items-center gap-2.5 px-4 py-2.5 bg-emerald-50 dark:bg-emerald-950 border border-emerald-200 dark:border-emerald-800 rounded-full shadow-sm">
            <div class="h-2 w-2 rounded-full bg-emerald-500 animate-pulse"></div>
            <span class="text-sm font-bold text-emerald-700 dark:text-emerald-300">Live & Visible</span>
          </div>
        {:else}
          <div class="inline-flex items-center gap-2.5 px-4 py-2.5 bg-blue-50 dark:bg-blue-950 border border-blue-200 dark:border-blue-800 rounded-full shadow-sm">
            <Clock class="h-4 w-4 text-blue-600 dark:text-blue-400" />
            <span class="text-sm font-bold text-blue-700 dark:text-blue-300">Pending Review</span>
          </div>
        {/if}

      </div>
    </header>

    <!-- Messages -->
    {#if successMessage}
      <Alert.Root class="mb-6 animate-slide-in border-2 border-emerald-300 dark:border-emerald-700 bg-gradient-to-br from-emerald-50 to-emerald-100/50 dark:from-emerald-950/50 dark:to-emerald-900/30 shadow-md">
        <CheckCircle2 class="h-5 w-5 text-emerald-600 dark:text-emerald-400" />
        <Alert.Title class="text-emerald-900 dark:text-emerald-100 font-bold">Success</Alert.Title>
        <Alert.Description class="text-emerald-800 dark:text-emerald-200 font-medium">{successMessage}</Alert.Description>
      </Alert.Root>
    {/if}

    {#if errorMessage}
      <Alert.Root variant="destructive" class="mb-6 animate-slide-in border-2 shadow-md">
        <AlertCircle class="h-5 w-5" />
        <Alert.Title class="font-bold">Error</Alert.Title>
        <Alert.Description class="font-medium">{errorMessage}</Alert.Description>
      </Alert.Root>
    {/if}

    <!-- Main Content -->
    <div class="space-y-6">

      <!-- Ride Information Card -->
      <div class="bg-white dark:bg-slate-900 rounded-2xl border-2 border-slate-200 dark:border-slate-800 shadow-sm overflow-hidden transition-all hover:shadow-md">
        <div class="px-6 py-6 border-b-2 border-slate-200 dark:border-slate-800 bg-gradient-to-br from-slate-50 to-slate-100/30 dark:from-slate-900/50 dark:to-slate-800/30">
          <div class="flex items-center justify-between">
            <div>
              <h2 class="text-2xl font-black text-slate-900 dark:text-slate-50 tracking-tight">Event Details</h2>
              <p class="text-sm text-slate-600 dark:text-slate-400 mt-1.5 font-medium">Core information for all occurrences</p>
            </div>
            {#if !isEditingEvent}
              <Button
                variant="outline"
                size="sm"
                onclick={toggleEditEvent}
                class="gap-2 hover:bg-white dark:hover:bg-slate-800 transition-all hover:shadow-sm font-semibold"
              >
                <Edit2 class="h-3.5 w-3.5" />
                <span>Edit</span>
              </Button>
            {/if}
          </div>
        </div>

        <div class="px-6 py-8">
          {#if !isEditingEvent}
            <!-- Read Mode -->
            <div class="space-y-10 read-mode-enter">
              <!-- Description -->
              <div class="group">
                <h3 class="text-xs font-bold uppercase tracking-widest text-slate-400 dark:text-slate-500 mb-3 flex items-center gap-2">
                  <span class="h-px w-8 bg-slate-300 dark:bg-slate-700"></span>
                  Description
                </h3>
                <p class="text-base text-slate-700 dark:text-slate-300 leading-[1.75] whitespace-pre-wrap pl-10">
                  {rideData?.event.description}
                </p>
              </div>

              <!-- Details Grid -->
              <div class="grid grid-cols-1 sm:grid-cols-2 gap-x-12 gap-y-8 pl-10">
                <div class="group">
                  <h3 class="text-xs font-bold uppercase tracking-widest text-slate-400 dark:text-slate-500 mb-2.5">Venue</h3>
                  <p class="text-base text-slate-900 dark:text-slate-50 font-semibold">{rideData?.event.venue_name}</p>
                </div>
                <div class="group">
                  <h3 class="text-xs font-bold uppercase tracking-widest text-slate-400 dark:text-slate-500 mb-2.5">Address</h3>
                  <p class="text-base text-slate-700 dark:text-slate-300">{rideData?.event.address}</p>
                </div>
                <div class="group">
                  <h3 class="text-xs font-bold uppercase tracking-widest text-slate-400 dark:text-slate-500 mb-2.5">Audience</h3>
                  <p class="text-base text-slate-900 dark:text-slate-50 font-medium">
                    {audienceOptions.find(opt => opt.value === rideData?.event.audience)?.label || "Not specified"}
                  </p>
                </div>
                <div class="group">
                  <h3 class="text-xs font-bold uppercase tracking-widest text-slate-400 dark:text-slate-500 mb-2.5">Ride Length</h3>
                  <p class="text-base text-slate-900 dark:text-slate-50 font-medium">
                    {rideData?.event.ride_length || "Not specified"}
                  </p>
                </div>
              </div>

              {#if rideData?.event.location_details}
                <div class="group pl-10">
                  <h3 class="text-xs font-bold uppercase tracking-widest text-slate-400 dark:text-slate-500 mb-3">Location Details</h3>
                  <p class="text-base text-slate-700 dark:text-slate-300 leading-[1.75]">{rideData.event.location_details}</p>
                </div>
              {/if}

              <!-- Organizer -->
              <div class="pt-8 border-t border-slate-200 dark:border-slate-800">
                <h3 class="text-xs font-bold uppercase tracking-widest text-slate-400 dark:text-slate-500 mb-5 flex items-center gap-2">
                  <span class="h-px w-8 bg-slate-300 dark:bg-slate-700"></span>
                  Organizer
                </h3>
                <div class="pl-10">
                  <div class="flex items-center gap-4">
                    <div class="h-12 w-12 rounded-full bg-gradient-to-br from-slate-100 to-slate-200 dark:from-slate-800 dark:to-slate-700 flex items-center justify-center shadow-sm">
                      <span class="text-base font-bold text-slate-700 dark:text-slate-200">
                        {rideData?.event.organizer_name?.charAt(0).toUpperCase()}
                      </span>
                    </div>
                    <div>
                      <p class="text-base font-semibold text-slate-900 dark:text-slate-50">{rideData?.event.organizer_name}</p>
                      <p class="text-sm text-slate-500 dark:text-slate-400 mt-0.5">{rideData?.event.organizer_email}</p>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          {:else}
            <!-- Edit Mode -->
            <div class="space-y-8 edit-mode-enter">
              <div class="field-group">
                <label for="description" class="block text-xs font-bold uppercase tracking-widest text-slate-500 dark:text-slate-400 mb-3">
                  Description <span class="text-rose-500">*</span>
                </label>
                <Textarea
                  id="description"
                  bind:value={editedDescription}
                  rows={6}
                  class="resize-none font-sans text-base leading-relaxed focus:ring-2 focus:ring-slate-900 dark:focus:ring-slate-100"
                  placeholder="Describe your ride..."
                />
                <p class="text-xs text-slate-500 dark:text-slate-400 mt-2 pl-0.5">Minimum 10 characters</p>
              </div>

              <div class="grid grid-cols-1 sm:grid-cols-2 gap-8">
                <div class="field-group">
                  <label for="audience" class="block text-xs font-bold uppercase tracking-widest text-slate-500 dark:text-slate-400 mb-3">
                    Audience
                  </label>
                  <Select.Root type="single" bind:value={editedAudience}>
                    <Select.Trigger class="w-full focus:ring-2 focus:ring-slate-900 dark:focus:ring-slate-100">
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

                <div class="field-group">
                  <label for="ride_length" class="block text-xs font-bold uppercase tracking-widest text-slate-500 dark:text-slate-400 mb-3">
                    Ride Length
                  </label>
                  <Input
                    id="ride_length"
                    bind:value={editedRideLength}
                    placeholder="e.g., 15 miles, 90 minutes"
                    maxlength={50}
                    class="focus:ring-2 focus:ring-slate-900 dark:focus:ring-slate-100"
                  />
                </div>
              </div>

              <div class="flex gap-3 pt-6 border-t border-slate-200 dark:border-slate-800">
                <Button
                  disabled={isSavingEvent}
                  onclick={saveEventDetails}
                  size="lg"
                  class="gap-2.5 bg-slate-900 hover:bg-slate-800 dark:bg-slate-50 dark:hover:bg-slate-200 font-semibold shadow-sm hover:shadow transition-all"
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
                  size="lg"
                  disabled={isSavingEvent}
                  onclick={toggleEditEvent}
                  class="gap-2 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors"
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
        <div class="bg-white dark:bg-slate-900 rounded-2xl border-2 border-slate-200 dark:border-slate-800 shadow-sm overflow-hidden transition-all hover:shadow-md">
          <div class="px-6 py-6 border-b-2 border-slate-200 dark:border-slate-800 bg-gradient-to-br from-slate-50 to-slate-100/30 dark:from-slate-900/50 dark:to-slate-800/30">
            <h2 class="text-2xl font-black text-slate-900 dark:text-slate-50 tracking-tight">Upcoming Rides</h2>
            <p class="text-sm text-slate-600 dark:text-slate-400 mt-1.5 font-medium">Edit timing or cancel individual occurrences</p>
          </div>

          <div class="divide-y divide-slate-200 dark:divide-slate-800">
            {#each upcomingOccurrences as occurrence (occurrence.id)}
              <div class="px-6 py-6 {occurrence.is_cancelled ? 'bg-slate-50/50 dark:bg-slate-900/50' : ''} hover:bg-slate-50/50 dark:hover:bg-slate-900/30 transition-colors">
                {#if !occurrence.isEditing}
                  <!-- View Mode -->
                  <div class="flex items-start justify-between gap-6 read-mode-enter">
                    <div class="flex-1 space-y-4">
                      <div class="flex items-baseline gap-4 flex-wrap">
                        <div class="flex items-center gap-2.5">
                          <Calendar class="h-4.5 w-4.5 text-slate-400 dark:text-slate-500" />
                          <span class="text-lg font-bold text-slate-900 dark:text-slate-50 tracking-tight">
                            {formatDate(occurrence.start_date)}
                          </span>
                        </div>

                        {#if occurrence.is_cancelled}
                          <span class="line-through text-slate-400 dark:text-slate-500 text-base">
                            {formatTime(occurrence.start_time)}
                          </span>
                          <span class="inline-flex items-center px-3 py-1 rounded-full text-xs font-bold uppercase tracking-wide bg-red-100 dark:bg-red-900/40 text-red-700 dark:text-red-400 shadow-sm">
                            Cancelled
                          </span>
                        {:else}
                          <div class="flex items-center gap-2 text-slate-600 dark:text-slate-400">
                            <Clock class="h-4 w-4" />
                            <span class="font-semibold text-base">{formatTime(occurrence.start_time)}</span>
                          </div>
                        {/if}
                      </div>

                      {#if occurrence.event_duration_minutes}
                        <div class="flex items-center gap-2 text-sm text-slate-500 dark:text-slate-400">
                          <span class="h-1 w-1 rounded-full bg-slate-300 dark:bg-slate-600"></span>
                          <span>Duration: <span class="font-semibold text-slate-700 dark:text-slate-300">{occurrence.event_duration_minutes} minutes</span></span>
                        </div>
                      {/if}

                      {#if occurrence.event_time_details}
                        <p class="text-sm text-slate-600 dark:text-slate-400 leading-relaxed pl-6 border-l-2 border-slate-200 dark:border-slate-700">
                          {occurrence.event_time_details}
                        </p>
                      {/if}

                      {#if occurrence.newsflash}
                        <div class="flex items-start gap-3 p-4 bg-gradient-to-br from-amber-50 to-orange-50 dark:from-amber-950/30 dark:to-orange-950/30 border border-amber-200 dark:border-amber-800/50 rounded-xl shadow-sm">
                          <AlertTriangle class="h-4.5 w-4.5 text-amber-600 dark:text-amber-400 flex-shrink-0 mt-0.5" />
                          <p class="text-sm text-amber-900 dark:text-amber-100 font-medium leading-relaxed">
                            {occurrence.newsflash}
                          </p>
                        </div>
                      {/if}
                    </div>

                    <Button
                      variant="outline"
                      size="sm"
                      onclick={() => toggleEdit(occurrence)}
                      class="gap-2 flex-shrink-0 hover:bg-slate-100 dark:hover:bg-slate-800 transition-all hover:shadow-sm"
                    >
                      <Edit2 class="h-3.5 w-3.5" />
                      <span class="font-medium">Edit</span>
                    </Button>
                  </div>
                {:else}
                  <!-- Edit Mode -->
                  <div class="space-y-6 edit-mode-enter">
                    <div class="grid grid-cols-1 sm:grid-cols-2 gap-6">
                      <div class="field-group">
                        <label for={`start-time-${occurrence.id}`} class="block text-xs font-bold uppercase tracking-widest text-slate-500 dark:text-slate-400 mb-3">
                          Start Time
                        </label>
                        <Input
                          id={`start-time-${occurrence.id}`}
                          type="time"
                          bind:value={occurrence.start_time}
                          class="focus:ring-2 focus:ring-slate-900 dark:focus:ring-slate-100 text-base"
                        />
                      </div>

                      <div class="field-group">
                        <label for={`duration-${occurrence.id}`} class="block text-xs font-bold uppercase tracking-widest text-slate-500 dark:text-slate-400 mb-3">
                          Duration (minutes)
                        </label>
                        <Input
                          id={`duration-${occurrence.id}`}
                          type="number"
                          bind:value={occurrence.event_duration_minutes}
                          min="0"
                          class="focus:ring-2 focus:ring-slate-900 dark:focus:ring-slate-100 text-base"
                        />
                      </div>
                    </div>

                    <div class="field-group">
                      <label for={`time-details-${occurrence.id}`} class="block text-xs font-bold uppercase tracking-widest text-slate-500 dark:text-slate-400 mb-3">
                        Time Details <span class="text-slate-400 dark:text-slate-500 font-normal normal-case tracking-normal text-xs">(Optional)</span>
                      </label>
                      <Input
                        id={`time-details-${occurrence.id}`}
                        type="text"
                        bind:value={occurrence.event_time_details}
                        placeholder="e.g., Meet at the fountain"
                        class="focus:ring-2 focus:ring-slate-900 dark:focus:ring-slate-100 text-base"
                      />
                    </div>

                    <div class="field-group">
                      <label for={`newsflash-${occurrence.id}`} class="block text-xs font-bold uppercase tracking-widest text-slate-500 dark:text-slate-400 mb-3">
                        Alert/Update <span class="text-slate-400 dark:text-slate-500 font-normal normal-case tracking-normal text-xs">(Optional)</span>
                      </label>
                      <Input
                        id={`newsflash-${occurrence.id}`}
                        type="text"
                        bind:value={occurrence.newsflash}
                        placeholder="e.g., Route change due to construction"
                        maxlength={500}
                        class="focus:ring-2 focus:ring-slate-900 dark:focus:ring-slate-100 text-base"
                      />
                      <p class="text-xs text-slate-500 dark:text-slate-400 mt-2 pl-0.5">Special message for this date (max 500 characters)</p>
                    </div>

                    <div class="flex items-center gap-3 p-4 bg-gradient-to-br from-slate-50 to-slate-100/50 dark:from-slate-900/50 dark:to-slate-800/30 border border-slate-200 dark:border-slate-700 rounded-xl">
                      <Checkbox
                        id={`cancel-${occurrence.id}`}
                        bind:checked={occurrence.is_cancelled}
                      />
                      <label
                        for={`cancel-${occurrence.id}`}
                        class="text-sm font-semibold text-slate-700 dark:text-slate-300 cursor-pointer"
                      >
                        Cancel this occurrence
                      </label>
                    </div>

                    <div class="flex gap-3 pt-4 border-t border-slate-200 dark:border-slate-800">
                      <Button
                        disabled={occurrence.isSaving}
                        onclick={() => saveOccurrence(occurrence)}
                        size="lg"
                        class="gap-2.5 bg-slate-900 hover:bg-slate-800 dark:bg-slate-50 dark:hover:bg-slate-200 font-semibold shadow-sm hover:shadow transition-all"
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
                        size="lg"
                        disabled={occurrence.isSaving}
                        onclick={() => toggleEdit(occurrence)}
                        class="gap-2 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors"
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
        <div class="bg-white dark:bg-slate-900 rounded-2xl border-2 border-slate-200 dark:border-slate-800 shadow-sm overflow-hidden opacity-80 hover:opacity-100 transition-all">
          <div class="px-6 py-6 border-b-2 border-slate-200 dark:border-slate-800 bg-gradient-to-br from-slate-50 to-slate-100/30 dark:from-slate-900/50 dark:to-slate-800/30">
            <h2 class="text-2xl font-black text-slate-900 dark:text-slate-50 tracking-tight">Past Rides</h2>
            <p class="text-sm text-slate-600 dark:text-slate-400 mt-1.5 font-medium">These occurrences have already happened</p>
          </div>

          <div class="px-6 py-6">
            <div class="space-y-2.5">
              {#each pastOccurrences as occurrence (occurrence.id)}
                <div class="flex items-center gap-3.5 p-4 bg-gradient-to-r from-slate-50 to-slate-100/50 dark:from-slate-900/50 dark:to-slate-800/30 rounded-xl border border-slate-200 dark:border-slate-800 shadow-sm">
                  <Calendar class="h-4.5 w-4.5 text-slate-400 dark:text-slate-500" />
                  <span class="text-sm font-bold text-slate-600 dark:text-slate-400">
                    {formatDate(occurrence.start_date)}
                  </span>
                  <span class="text-sm text-slate-500 dark:text-slate-500 font-medium">
                    {formatTime(occurrence.start_time)}
                  </span>
                  {#if occurrence.is_cancelled}
                    <span class="text-xs px-2.5 py-1 bg-red-100 dark:bg-red-900/30 text-red-700 dark:text-red-400 rounded-full font-bold uppercase tracking-wide">
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
        <div class="flex justify-center pt-8">
          <Button
            variant="outline"
            size="lg"
            class="gap-2.5 hover:bg-slate-100 dark:hover:bg-slate-800 transition-all font-semibold shadow-sm hover:shadow"
            href={getCycleSceneDomain(rideData?.event.city)}
          >
            <span>← Back to {rideData?.event.city?.toUpperCase() || "CycleScene"}</span>
          </Button>
        </div>
      {/if}

    </div>
  </div>
</div>
