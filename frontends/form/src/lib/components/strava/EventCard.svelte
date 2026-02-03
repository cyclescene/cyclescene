<script lang="ts">
  import * as Card from "$lib/components/ui/card";
  import * as Collapsible from "$lib/components/ui/collapsible";
  import * as Select from "$lib/components/ui/select";
  import { Checkbox } from "$lib/components/ui/checkbox";
  import { Label } from "$lib/components/ui/label";
  import { Button } from "$lib/components/ui/button";
  import ImageUploader from "$lib/components/ride-form/ImageUploader.svelte";
  import {
    AUDIENCE_OPTIONS,
    DURATION_OPTIONS,
    STRAVA_IMPORT_DEFAULTS,
    type StravaGroupEvent,
    type EventImportConfig,
    type EventOverrides,
  } from "$lib/types/strava";

  interface Props {
    event: StravaGroupEvent;
    selected: boolean;
    config: EventImportConfig | null;
    cityCode: string;
    onToggle: (selected: boolean) => void;
    onConfigChange: (config: EventImportConfig) => void;
  }

  let {
    event,
    selected,
    config,
    cityCode,
    onToggle,
    onConfigChange,
  }: Props = $props();

  let customizeOpen = $state(false);

  // Local state for overrides
  let audience = $state(config?.overrides?.audience ?? STRAVA_IMPORT_DEFAULTS.audience);
  let duration = $state<string | undefined>(
    config?.overrides?.event_duration_minutes?.toString()
  );
  let imageUrl = $state<string | null>(config?.overrides?.image_url ?? null);

  // Format date for display
  function formatEventDate(isoString: string): string {
    const date = new Date(isoString);
    return date.toLocaleDateString("en-US", {
      weekday: "short",
      month: "short",
      day: "numeric",
      hour: "numeric",
      minute: "2-digit",
    });
  }

  // Format distance for display
  function formatDistance(meters: number | null | undefined): string {
    if (!meters) return "";
    const miles = meters / 1609.34;
    return `${miles.toFixed(1)} mi`;
  }

  // Get next occurrence date
  function getNextOccurrence(): string {
    if (event.upcoming_occurrences && event.upcoming_occurrences.length > 0) {
      return formatEventDate(event.upcoming_occurrences[0]);
    }
    return "Date TBD";
  }

  // Handle checkbox change
  function handleCheckChange(checked: boolean | "indeterminate") {
    const isChecked = checked === true;
    onToggle(isChecked);

    // If selecting, send initial config
    if (isChecked) {
      updateConfig();
    }
  }

  // Update config when overrides change
  function updateConfig() {
    const overrides: EventOverrides = {};

    if (audience !== STRAVA_IMPORT_DEFAULTS.audience) {
      overrides.audience = audience;
    }
    if (duration) {
      overrides.event_duration_minutes = parseInt(duration, 10);
    }
    if (imageUrl) {
      overrides.image_url = imageUrl;
    }

    const newConfig: EventImportConfig = {
      strava_event_id: event.id, // Already a string from backend
      club_id: event.club_id,
      overrides: Object.keys(overrides).length > 0 ? overrides : undefined,
    };

    onConfigChange(newConfig);
  }

  function handleImageUpload(url: string) {
    imageUrl = url;
    if (selected) updateConfig();
  }

  function handleImageError(err: string) {
    console.error("[Strava] Image upload error:", err);
  }

  // Toggle customize panel
  function toggleCustomize() {
    customizeOpen = !customizeOpen;
  }

  // Get audience label for display
  function getAudienceLabel(): string {
    return AUDIENCE_OPTIONS.find(o => o.value === audience)?.label ?? "Select audience";
  }

  // Get duration label for display
  function getDurationLabel(): string {
    if (!duration) return "Select duration";
    const durationNum = parseInt(duration, 10);
    return DURATION_OPTIONS.find(o => o.value === durationNum)?.label ?? `${duration} min`;
  }

  // Watch for audience changes
  $effect(() => {
    if (selected) updateConfig();
  });

  // Re-sync local state if config changes externally
  $effect(() => {
    if (config?.overrides) {
      if (config.overrides.audience) {
        audience = config.overrides.audience;
      }
      if (config.overrides.event_duration_minutes) {
        duration = config.overrides.event_duration_minutes.toString();
      }
      if (config.overrides.image_url) {
        imageUrl = config.overrides.image_url;
      }
    }
  });
</script>

<Card.Root class="p-4">
  <div class="flex items-start gap-3">
    <Checkbox
      checked={selected}
      onCheckedChange={handleCheckChange}
      class="mt-1"
    />
    <div class="min-w-0 flex-1">
      <h4 class="font-medium leading-tight">{event.title}</h4>
      <p class="text-muted-foreground mt-1 text-sm">
        <span title="Location">📍 {event.address || "Location TBD"}</span>
        <span class="mx-2">•</span>
        <span title="Date">📅 {getNextOccurrence()}</span>
        {#if event.route?.distance}
          <span class="mx-2">•</span>
          <span title="Distance">🚴 {formatDistance(event.route.distance)}</span>
        {/if}
      </p>

      {#if event.description}
        <p class="text-muted-foreground mt-2 line-clamp-2 text-sm">
          {event.description}
        </p>
      {/if}

      <Collapsible.Root bind:open={customizeOpen} class="mt-3">
        <Collapsible.Trigger
          class="inline-flex h-8 items-center gap-1 rounded-md px-2 text-sm font-medium hover:bg-accent hover:text-accent-foreground"
        >
          {customizeOpen ? "Hide Options" : "Customize"}
          <svg
            class="h-4 w-4 transition-transform {customizeOpen ? 'rotate-180' : ''}"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M19 9l-7 7-7-7"
            />
          </svg>
        </Collapsible.Trigger>
        <Collapsible.Content class="mt-3 space-y-4 rounded-md border p-4">
          <!-- Audience Selection -->
          <div class="space-y-2">
            <Label>Audience</Label>
            <Select.Root type="single" bind:value={audience}>
              <Select.Trigger class="w-full">
                {getAudienceLabel()}
              </Select.Trigger>
              <Select.Content>
                {#each AUDIENCE_OPTIONS as option}
                  <Select.Item value={option.value} label={option.label}>
                    {option.label}
                  </Select.Item>
                {/each}
              </Select.Content>
            </Select.Root>
          </div>

          <!-- Duration Selection -->
          <div class="space-y-2">
            <Label>Duration (optional)</Label>
            <Select.Root type="single" bind:value={duration}>
              <Select.Trigger class="w-full">
                {getDurationLabel()}
              </Select.Trigger>
              <Select.Content>
                {#each DURATION_OPTIONS as option}
                  <Select.Item value={option.value.toString()} label={option.label}>
                    {option.label}
                  </Select.Item>
                {/each}
              </Select.Content>
            </Select.Root>
          </div>

          <!-- Image Upload -->
          <div class="space-y-2">
            <Label>Image (optional)</Label>
            <ImageUploader
              label="Upload Event Image"
              entityType="ride"
              cityCode={cityCode}
              onUploadComplete={handleImageUpload}
              onUploadError={handleImageError}
            />
            {#if imageUrl}
              <p class="text-muted-foreground text-xs">Image uploaded</p>
            {/if}
          </div>
        </Collapsible.Content>
      </Collapsible.Root>
    </div>
  </div>
</Card.Root>
