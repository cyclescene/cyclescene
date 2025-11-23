<script lang="ts">
  import { page } from "$app/state";
  import { Button } from "$lib/components/ui/button";
  import { Input } from "$lib/components/ui/input";
  import { Checkbox } from "$lib/components/ui/checkbox";
  import * as Card from "$lib/components/ui/card";
  import { API_URL } from "$env/static/private";

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
    is_cancelled: boolean;
  }

  interface EditingOccurrence extends Occurrence {
    isEditing?: boolean;
    isSaving?: boolean;
  }

  let { data }: any = $props();
  let rideData: RideData | null = $state(data?.rideData || null);
  let occurrences: EditingOccurrence[] = $state(rideData?.event.occurrences || []);
  let successMessage = $state('');
  let errorMessage = $state('');
  let token = $state(page.url.searchParams.get('token') || '');

  const today = new Date().toISOString().split('T')[0];

  // Separate occurrences into past and upcoming
  let pastOccurrences = $derived(occurrences.filter(o => o.start_date < today));
  let upcomingOccurrences = $derived(occurrences.filter(o => o.start_date >= today));

  const getCycleSceneDomain = (cityCode?: string): string => {
    if (!cityCode) return 'https://cyclescene.cc';
    const cityDomains: Record<string, string> = {
      pdx: 'https://pdx.cyclescene.cc',
      slc: 'https://slc.cyclescene.cc',
    };
    return cityDomains[cityCode.toLowerCase()] || 'https://cyclescene.cc';
  };

  const toggleEdit = (index: number) => {
    upcomingOccurrences[index].isEditing = !upcomingOccurrences[index].isEditing;
  };

  const saveOccurrence = async (occurrence: EditingOccurrence, index: number) => {
    upcomingOccurrences[index].isSaving = true;
    errorMessage = '';

    try {
      const response = await fetch(`${API_URL}/v1/rides/edit/${token}/occurrences/${occurrence.id}`, {
        method: 'PATCH',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          start_time: occurrence.start_time,
          event_duration_minutes: occurrence.event_duration_minutes,
          event_time_details: occurrence.event_time_details,
          is_cancelled: occurrence.is_cancelled,
        }),
      });

      if (!response.ok) {
        throw new Error('Failed to save occurrence');
      }

      upcomingOccurrences[index].isEditing = false;
      successMessage = 'Occurrence updated successfully!';
      setTimeout(() => { successMessage = ''; }, 3000);
    } catch (err) {
      errorMessage = err instanceof Error ? err.message : 'Failed to save changes';
    } finally {
      upcomingOccurrences[index].isSaving = false;
    }
  };
</script>

<div class="container max-w-4xl mx-auto py-4 sm:py-8 px-4">
  <!-- Header -->
  <div class="mb-8">
    <h1 class="text-3xl sm:text-4xl font-bold tracking-tight">{rideData?.event.title}</h1>
    <p class="text-muted-foreground mt-2">Edit your ride details and manage occurrences</p>
    {#if rideData?.is_published}
      <div class="mt-3 p-3 bg-green-50 border border-green-200 rounded-md text-sm text-green-700">
        ✓ Published - Visible to the community
      </div>
    {:else}
      <div class="mt-3 p-3 bg-blue-50 border border-blue-200 rounded-md text-sm text-blue-700">
        Pending review - Will be visible once approved
      </div>
    {/if}
  </div>

  <!-- Messages -->
  {#if successMessage}
    <div class="mb-4 p-3 border border-green-200 bg-green-50 rounded-lg">
      <p class="text-sm text-green-700">✓ {successMessage}</p>
    </div>
  {/if}

  {#if errorMessage}
    <div class="mb-4 p-3 border border-destructive bg-destructive/10 rounded-lg">
      <p class="text-sm text-destructive">{errorMessage}</p>
    </div>
  {/if}

  <!-- Ride Information (Read-Only) -->
  <Card.Root class="mb-8">
    <Card.Header>
      <Card.Title>Ride Information</Card.Title>
      <Card.Description>These details apply to all occurrences</Card.Description>
    </Card.Header>
    <Card.Content class="space-y-6">
      <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
        <div>
          <p class="text-sm font-medium text-muted-foreground">Venue</p>
          <p class="text-base mt-1">{rideData?.event.venue_name}</p>
        </div>
        <div>
          <p class="text-sm font-medium text-muted-foreground">Address</p>
          <p class="text-base mt-1">{rideData?.event.address}</p>
        </div>
        <div>
          <p class="text-sm font-medium text-muted-foreground">Audience</p>
          <p class="text-base mt-1">{rideData?.event.audience || 'Not specified'}</p>
        </div>
        <div>
          <p class="text-sm font-medium text-muted-foreground">Ride Length</p>
          <p class="text-base mt-1">{rideData?.event.ride_length || 'Not specified'}</p>
        </div>
      </div>

      <div>
        <p class="text-sm font-medium text-muted-foreground">Description</p>
        <p class="text-base mt-2 whitespace-pre-wrap">{rideData?.event.description}</p>
      </div>

      {#if rideData?.event.location_details}
        <div>
          <p class="text-sm font-medium text-muted-foreground">Location Details</p>
          <p class="text-base mt-2">{rideData.event.location_details}</p>
        </div>
      {/if}

      <div class="grid grid-cols-1 md:grid-cols-2 gap-6 pt-4 border-t">
        <div>
          <p class="text-sm font-medium text-muted-foreground">Organizer</p>
          <p class="text-base mt-1">{rideData?.event.organizer_name}</p>
        </div>
        <div>
          <p class="text-sm font-medium text-muted-foreground">Email</p>
          <p class="text-base mt-1">{rideData?.event.organizer_email}</p>
        </div>
      </div>
    </Card.Content>
  </Card.Root>

  <!-- Upcoming Occurrences (Editable) -->
  {#if upcomingOccurrences.length > 0}
    <Card.Root class="mb-8">
      <Card.Header>
        <Card.Title>Upcoming Occurrences</Card.Title>
        <Card.Description>Click Edit to modify time or cancel this occurrence</Card.Description>
      </Card.Header>
      <Card.Content>
        <div class="space-y-4">
          {#each upcomingOccurrences as occurrence, index (occurrence.id)}
            <div class={`border rounded-lg p-4 ${occurrence.is_cancelled ? 'bg-gray-50' : ''}`}>
              {#if !occurrence.isEditing}
                <!-- View Mode -->
                <div class="flex items-start justify-between">
                  <div class="flex-1">
                    <div class="flex items-center gap-3">
                      <div>
                        <p class="font-medium">
                          {occurrence.start_date}
                          {#if occurrence.is_cancelled}
                            <span class="text-sm text-muted-foreground ml-2 line-through">
                              {occurrence.start_time}
                            </span>
                            <span class="inline-block ml-2 px-2 py-1 bg-red-100 text-red-700 text-xs rounded">
                              Cancelled
                            </span>
                          {:else}
                            <span class="text-sm text-muted-foreground">{occurrence.start_time}</span>
                          {/if}
                        </p>
                        {#if occurrence.event_duration_minutes}
                          <p class="text-sm text-muted-foreground mt-1">
                            Duration: {occurrence.event_duration_minutes} minutes
                          </p>
                        {/if}
                        {#if occurrence.event_time_details}
                          <p class="text-sm mt-2">{occurrence.event_time_details}</p>
                        {/if}
                      </div>
                    </div>
                  </div>
                  <Button
                    variant="outline"
                    size="sm"
                    on:click={() => toggleEdit(index)}
                    class="ml-4"
                  >
                    Edit
                  </Button>
                </div>
              {:else}
                <!-- Edit Mode -->
                <div class="space-y-4">
                  <div>
                    <label class="text-sm font-medium">Start Time</label>
                    <Input
                      type="time"
                      bind:value={occurrence.start_time}
                      class="mt-1"
                    />
                  </div>

                  <div>
                    <label class="text-sm font-medium">Duration (minutes)</label>
                    <Input
                      type="number"
                      bind:value={occurrence.event_duration_minutes}
                      class="mt-1"
                    />
                  </div>

                  <div>
                    <label class="text-sm font-medium">Time Details (Optional)</label>
                    <Input
                      type="text"
                      bind:value={occurrence.event_time_details}
                      placeholder="e.g., Meet at the fountain"
                      class="mt-1"
                    />
                  </div>

                  <div class="flex items-center gap-2 p-3 bg-gray-50 rounded">
                    <Checkbox
                      id={`cancel-${occurrence.id}`}
                      bind:checked={occurrence.is_cancelled}
                    />
                    <label for={`cancel-${occurrence.id}`} class="text-sm font-medium cursor-pointer">
                      Cancel this occurrence
                    </label>
                  </div>

                  <div class="flex gap-2">
                    <Button
                      size="sm"
                      disabled={occurrence.isSaving}
                      on:click={() => saveOccurrence(occurrence, index)}
                    >
                      {occurrence.isSaving ? 'Saving...' : 'Save'}
                    </Button>
                    <Button
                      variant="outline"
                      size="sm"
                      disabled={occurrence.isSaving}
                      on:click={() => toggleEdit(index)}
                    >
                      Cancel
                    </Button>
                  </div>
                </div>
              {/if}
            </div>
          {/each}
        </div>
      </Card.Content>
    </Card.Root>
  {/if}

  <!-- Past Occurrences (Read-Only) -->
  {#if pastOccurrences.length > 0}
    <Card.Root class="mb-8">
      <Card.Header>
        <Card.Title>Past Occurrences</Card.Title>
        <Card.Description>These occurrences have already happened</Card.Description>
      </Card.Header>
      <Card.Content>
        <div class="space-y-3">
          {#each pastOccurrences as occurrence (occurrence.id)}
            <div class="border rounded-lg p-3 bg-gray-50">
              <p class="text-sm">
                <span class="font-medium">{occurrence.start_date}</span>
                <span class="text-muted-foreground ml-2">{occurrence.start_time}</span>
                {#if occurrence.is_cancelled}
                  <span class="text-xs bg-red-100 text-red-700 px-2 py-1 rounded ml-2">Cancelled</span>
                {/if}
              </p>
            </div>
          {/each}
        </div>
      </Card.Content>
    </Card.Root>
  {/if}

  <!-- Back Button -->
  {#if successMessage}
    <div class="flex gap-3 justify-center sm:justify-start">
      <Button variant="outline">
        <a href={getCycleSceneDomain(rideData?.event.city)}>
          Back to {rideData?.event.city?.toUpperCase() || 'CycleScene'}
        </a>
      </Button>
    </div>
  {/if}
</div>
