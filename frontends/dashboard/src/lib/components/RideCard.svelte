<script lang="ts">
  import * as Card from "$lib/components/ui/card";
  import { Button } from "$lib/components/ui/button";
  import { Badge } from "$lib/components/ui/badge";
  import { CITIES } from "$lib/config/cities";

  interface Ride {
    id: number;
    title: string;
    description: string;
    city: string;
    venue_name: string;
    organizer_name: string;
    organizer_email: string;
    image_url: string;
    image_uuid: string;
    is_published: boolean;
    is_loop_ride: boolean;
    created_at: string;
  }

  interface Props {
    ride: Ride;
    onPublish: (id: number) => void;
    isPublishing: boolean;
    showCityBadge?: boolean;
  }

  let {
    ride,
    onPublish,
    isPublishing = false,
    showCityBadge = false,
  }: Props = $props();

  let expanded = $state(false);

  function toggleExpanded() {
    expanded = !expanded;
  }

  const cityInfo = $derived(
    CITIES.find((c) => c.code === ride.city) || { name: ride.city, state: "" }
  );
</script>

<div class="space-y-2">
  <button
    onclick={toggleExpanded}
    class="w-full text-left p-4 border rounded-lg hover:bg-accent/50 transition-colors {expanded
      ? 'bg-accent/30'
      : ''}"
  >
    <div class="flex items-start sm:items-center justify-between gap-4 flex-col sm:flex-row">
      <div class="flex-1 min-w-0">
        <div class="flex items-center gap-2 flex-wrap">
          <p class="font-medium truncate">{ride.title}</p>
          {#if showCityBadge}
            <Badge variant="secondary" class="text-xs">
              {cityInfo.name}
            </Badge>
          {/if}
        </div>
        <p class="text-sm text-muted-foreground mt-1">
          {ride.venue_name} • {ride.organizer_name}
        </p>
      </div>
      <Button
        onclick={(e) => {
          e.stopPropagation();
          onPublish(ride.id);
        }}
        disabled={isPublishing}
        size="sm"
        class="shrink-0 w-full sm:w-auto"
      >
        {isPublishing ? "Publishing..." : "Publish"}
      </Button>
    </div>
  </button>

  {#if expanded}
    <Card.Root class="ml-0 sm:ml-4">
      <Card.Content class="pt-6 space-y-4">
        {#if ride.image_url}
          <div>
            <img
              src={ride.image_url}
              alt={ride.title}
              class="w-full h-48 object-cover rounded-lg border"
            />
          </div>
        {/if}

        <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
          <div>
            <p class="text-xs font-medium text-muted-foreground">City</p>
            <p class="text-sm mt-1">
              {cityInfo.name}{#if cityInfo.state}, {cityInfo.state}{/if}
            </p>
          </div>
          <div>
            <p class="text-xs font-medium text-muted-foreground">Venue</p>
            <p class="text-sm mt-1">{ride.venue_name}</p>
          </div>
          <div>
            <p class="text-xs font-medium text-muted-foreground">Type</p>
            <p class="text-sm mt-1">
              {ride.is_loop_ride ? "Loop Ride" : "Point-to-Point"}
            </p>
          </div>
          <div>
            <p class="text-xs font-medium text-muted-foreground">Organizer</p>
            <p class="text-sm mt-1">{ride.organizer_name}</p>
          </div>
          <div class="sm:col-span-2">
            <p class="text-xs font-medium text-muted-foreground">Email</p>
            <p class="text-sm mt-1 break-all">{ride.organizer_email}</p>
          </div>
        </div>

        <div>
          <p class="text-xs font-medium text-muted-foreground">Description</p>
          <p class="text-sm mt-2 whitespace-pre-wrap">
            {ride.description}
          </p>
        </div>

        {#if ride.image_uuid}
          <div class="pt-2 border-t">
            <p class="text-xs font-medium text-muted-foreground">Image UUID</p>
            <p class="text-xs mt-1 break-all text-muted-foreground">
              {ride.image_uuid}
            </p>
          </div>
        {/if}
      </Card.Content>
    </Card.Root>
  {/if}
</div>
