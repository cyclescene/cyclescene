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

<Card.Root class="group relative overflow-hidden border-2 transition-all duration-200 {selected ? 'border-[#FC5200] bg-[#FC5200]/5 shadow-md' : 'hover:border-[#FC5200]/40 hover:shadow-sm'}">
  <!-- Selection indicator bar -->
  {#if selected}
    <div class="absolute left-0 top-0 bottom-0 w-1 bg-[#FC5200]" aria-hidden="true"></div>
  {/if}

  <div class="p-4 pl-5">
    <div class="flex gap-4 items-start">
      <!-- Enhanced checkbox with better visibility and click target -->
      <label class="flex-shrink-0 pt-0.5 cursor-pointer">
        <Checkbox
          checked={selected}
          onCheckedChange={handleCheckChange}
          class="h-6 w-6 border-2 border-muted-foreground/50 hover:border-[#FC5200] hover:bg-[#FC5200]/5 transition-colors data-[state=checked]:bg-[#FC5200] data-[state=checked]:border-[#FC5200] data-[state=checked]:text-white cursor-pointer"
          aria-label="Select {event.title} for import"
        />
      </label>
      <div class="min-w-0 flex-1">
        <h4 class="font-semibold text-base leading-tight mb-2 {selected ? 'text-[#FC5200]' : ''}">{event.title}</h4>

        <!-- Event metadata badges -->
        <div class="flex flex-wrap gap-2 mb-3">
          <span class="inline-flex items-center gap-1.5 px-2.5 py-1.5 rounded-md bg-background border text-xs font-medium min-h-[32px]" title="Location">
            <svg class="h-3.5 w-3.5 flex-shrink-0 text-[#FC5200]" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z" />
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 11a3 3 0 11-6 0 3 3 0 016 0z" />
            </svg>
            <span class="break-words">{event.address || "Location TBD"}</span>
          </span>
          <span class="inline-flex items-center gap-1.5 px-2.5 py-1.5 rounded-md bg-background border text-xs font-medium min-h-[32px] whitespace-nowrap" title="Date">
            <svg class="h-3.5 w-3.5 flex-shrink-0 text-[#FC5200]" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
            </svg>
            {getNextOccurrence()}
          </span>
          {#if event.route?.distance}
            <span class="inline-flex items-center gap-1.5 px-2.5 py-1.5 rounded-md bg-[#FC5200]/10 border border-[#FC5200]/20 text-xs font-semibold text-[#FC5200] min-h-[32px] whitespace-nowrap" title="Distance">
              <svg class="h-3.5 w-3.5 flex-shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
              </svg>
              {formatDistance(event.route.distance)}
            </span>
          {/if}
        </div>

        {#if event.description}
          <p class="text-muted-foreground text-sm line-clamp-2 leading-relaxed">
            {event.description}
          </p>
        {/if}

        <Collapsible.Root bind:open={customizeOpen} class="mt-4">
          <Collapsible.Trigger
            class="inline-flex min-h-[44px] items-center gap-2 rounded-lg px-3 py-2 text-sm font-semibold transition-all duration-200 hover:bg-[#FC5200]/10 hover:text-[#FC5200] {customizeOpen ? 'bg-[#FC5200]/10 text-[#FC5200]' : 'text-muted-foreground'}"
            aria-label="{customizeOpen ? 'Hide' : 'Show'} customization options for {event.title}"
          >
            <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6V4m0 2a2 2 0 100 4m0-4a2 2 0 110 4m-6 8a2 2 0 100-4m0 4a2 2 0 110-4m0 4v2m0-6V4m6 6v10m6-2a2 2 0 100-4m0 4a2 2 0 110-4m0 4v2m0-6V4" />
            </svg>
            {customizeOpen ? "Hide Options" : "Customize Import"}
            <svg
              class="h-4 w-4 transition-transform duration-200 {customizeOpen ? 'rotate-180' : ''}"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
              aria-hidden="true"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M19 9l-7 7-7-7"
              />
            </svg>
          </Collapsible.Trigger>
          <Collapsible.Content class="mt-3 space-y-4 rounded-lg border-2 bg-muted/30 p-4 animate-in fade-in slide-in-from-top-2 duration-300">
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
  </div>
</Card.Root>
