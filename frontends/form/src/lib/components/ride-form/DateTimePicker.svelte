<script lang="ts">
  import {
    CalendarDate,
    DateFormatter,
    type DateValue,
    getLocalTimeZone,
    today,
  } from "@internationalized/date";
  import { Calendar as CalendarIcon, Clock, Plus, X } from "@lucide/svelte";
  import { Button } from "$lib/components/ui/button";
  import { Calendar } from "$lib/components/ui/calendar";
  import { Input } from "$lib/components/ui/input";
  import { Label } from "$lib/components/ui/label";
  import * as Popover from "$lib/components/ui/popover";
  import * as Card from "$lib/components/ui/card";
  import { cn } from "$lib/utils";

  interface Occurrence {
    start_date: string;
    start_time: string;
    event_duration_minutes?: number;
    event_time_details?: string;
  }

  interface Props {
    occurrences: Occurrence[];
    dateType: string;
    onupdate: (occurrences: Occurrence[]) => void;
    error?: string;
  }

  let {
    occurrences = $bindable([]),
    dateType,
    onupdate,
    error,
  }: Props = $props();

  const df = new DateFormatter("en-US", {
    dateStyle: "long",
  });

  const todayDate: CalendarDate = today(getLocalTimeZone());

  let selectedDate = $state<DateValue>(todayDate);
  let selectedTime = $state("18:00");
  let durationMinutes = $state<number>(120);
  let timeDetails = $state("");

  // For recurring rides - day of week selection
  let recurringDay = $state<number | undefined>(undefined);
  let recurringStartDate = $state<CalendarDate>(todayDate);
  let recurringEndDate = $state<CalendarDate>(todayDate.add({ days: 7 }));

  function addOccurrence() {
    if (!selectedDate) return;

    const dateStr = `${selectedDate.year}-${String(selectedDate.month).padStart(2, "0")}-${String(selectedDate.day).padStart(2, "0")}`;
    const timeStr = `${selectedTime}:00`;

    const newOccurrence: Occurrence = {
      start_date: dateStr,
      start_time: timeStr,
      event_duration_minutes: durationMinutes || undefined,
      event_time_details: timeDetails || undefined,
    };

    occurrences = [...occurrences, newOccurrence];
    onupdate(occurrences);

    // Reset form
    selectedDate = todayDate;
    timeDetails = "";
  }

  function generateRecurringOccurrences() {
    if (!recurringStartDate || !recurringEndDate || recurringDay === undefined)
      return;

    const newOccurrences: Occurrence[] = [];
    let currentDate = new Date(
      recurringStartDate.year,
      recurringStartDate.month - 1,
      recurringStartDate.day,
    );
    const endDate = new Date(
      recurringEndDate.year,
      recurringEndDate.month - 1,
      recurringEndDate.day,
    );

    // Find the first occurrence of the selected day of week
    while (currentDate.getDay() !== recurringDay && currentDate <= endDate) {
      currentDate.setDate(currentDate.getDate() + 1);
    }

    // Generate occurrences every week
    while (currentDate <= endDate) {
      const dateStr = `${currentDate.getFullYear()}-${String(currentDate.getMonth() + 1).padStart(2, "0")}-${String(currentDate.getDate()).padStart(2, "0")}`;
      const timeStr = `${selectedTime}:00`;

      newOccurrences.push({
        start_date: dateStr,
        start_time: timeStr,
        event_duration_minutes: durationMinutes || undefined,
      });

      currentDate.setDate(currentDate.getDate() + 7); // Add one week
    }

    occurrences = [...occurrences, ...newOccurrences];
    onupdate(occurrences);

    // Reset
    recurringStartDate = todayDate;
    recurringEndDate = todayDate.add({ days: 7 });
    recurringDay = undefined;
  }

  function removeOccurrence(index: number) {
    occurrences = occurrences.filter((_, i) => i !== index);
    onupdate(occurrences);
  }

  const daysOfWeek = [
    "Sunday",
    "Monday",
    "Tuesday",
    "Wednesday",
    "Thursday",
    "Friday",
    "Saturday",
  ];
</script>

<div class="space-y-4">
  {#if dateType === "R"}
    <!-- Recurring Date Pattern -->
    <Card.Root>
      <Card.Header>
        <Card.Title class="text-base sm:text-lg">Recurring Pattern</Card.Title>
      </Card.Header>
      <Card.Content class="space-y-3 sm:space-y-4">
        <div class="space-y-2">
          <Label class="text-sm sm:text-base">Which day of the week?</Label>
          <div class="grid grid-cols-2 sm:grid-cols-4 gap-2">
            {#each daysOfWeek as day, index}
              <Button
                type="button"
                variant={recurringDay === index ? "default" : "outline"}
                size="sm"
                class="touch-manipulation text-sm"
                onclick={() => {
                  recurringDay = index;
                }}
              >
                <span class="hidden sm:inline">{day}</span>
                <span class="sm:hidden">{day.slice(0, 3)}</span>
              </Button>
            {/each}
          </div>
        </div>

        <div class="grid grid-cols-1 gap-3 sm:gap-4">
          <div class="space-y-2">
            <Label class="text-sm sm:text-base">Start Date</Label>
            <Popover.Root>
              <Popover.Trigger>
                <Button
                  variant="outline"
                  class={cn(
                    "w-full justify-start text-left font-normal text-sm sm:text-base touch-manipulation",
                    !recurringStartDate && "text-muted-foreground",
                  )}
                >
                  <CalendarIcon class="mr-2 h-4 w-4 flex-shrink-0" />
                  <span class="truncate">
                    {recurringStartDate
                      ? df.format(recurringStartDate.toDate(getLocalTimeZone()))
                      : "Pick start date"}
                  </span>
                </Button>
              </Popover.Trigger>
              <Popover.Content class="w-auto p-0">
                <Calendar bind:value={recurringStartDate} />
              </Popover.Content>
            </Popover.Root>
          </div>

          <div class="space-y-2">
            <Label class="text-sm sm:text-base">End Date</Label>
            <Popover.Root>
              <Popover.Trigger>
                <Button
                  variant="outline"
                  class={cn(
                    "w-full justify-start text-left font-normal text-sm sm:text-base touch-manipulation",
                    !recurringEndDate && "text-muted-foreground",
                  )}
                >
                  <CalendarIcon class="mr-2 h-4 w-4 flex-shrink-0" />
                  <span class="truncate">
                    {recurringEndDate
                      ? df.format(recurringEndDate.toDate(getLocalTimeZone()))
                      : "Pick end date"}
                  </span>
                </Button>
              </Popover.Trigger>
              <Popover.Content class="w-auto p-0">
                <Calendar bind:value={recurringEndDate} />
              </Popover.Content>
            </Popover.Root>
          </div>
        </div>

        <div class="grid grid-cols-1 sm:grid-cols-2 gap-3 sm:gap-4">
          <div class="space-y-2">
            <Label for="recurring-time" class="text-sm sm:text-base"
              >Start Time</Label
            >
            <div class="flex items-center gap-2">
              <Clock class="h-4 w-4 text-muted-foreground flex-shrink-0" />
              <Input
                id="recurring-time"
                type="time"
                bind:value={selectedTime}
                class="text-base"
              />
            </div>
          </div>

          <div class="space-y-2">
            <Label for="recurring-duration" class="text-sm sm:text-base"
              >Duration (minutes)</Label
            >
            <Input
              id="recurring-duration"
              type="number"
              bind:value={durationMinutes}
              placeholder="120"
              min="15"
              step="15"
              class="text-base"
            />
          </div>
        </div>

        <Button
          type="button"
          onclick={generateRecurringOccurrences}
          disabled={!recurringDay || !recurringStartDate || !recurringEndDate}
          class="w-full touch-manipulation"
        >
          <Plus class="mr-2 h-4 w-4" />
          Generate Recurring Dates
        </Button>
      </Card.Content>
    </Card.Root>
  {/if}

  <!-- Single Date Selection (for all types) -->
  <Card.Root>
    <Card.Header>
      <Card.Title class="text-base sm:text-lg">
        {dateType === "R" ? "Add Individual Date" : "Add Date"}
      </Card.Title>
    </Card.Header>
    <Card.Content class="space-y-3 sm:space-y-4">
      <div class="space-y-2">
        <Label class="text-sm sm:text-base">Select Date</Label>
        <Popover.Root>
          <Popover.Trigger>
            <Button
              variant="outline"
              class={cn(
                "w-full justify-start text-left font-normal text-sm sm:text-base touch-manipulation",
                !selectedDate && "text-muted-foreground",
              )}
            >
              <CalendarIcon class="mr-2 h-4 w-4 flex-shrink-0" />
              <span class="truncate">
                {selectedDate
                  ? df.format(selectedDate.toDate(getLocalTimeZone()))
                  : "Pick a date"}
              </span>
            </Button>
          </Popover.Trigger>
          <Popover.Content class="w-auto p-0">
            <Calendar bind:value={selectedDate} />
          </Popover.Content>
        </Popover.Root>
      </div>

      <div class="grid grid-cols-1 sm:grid-cols-2 gap-3 sm:gap-4">
        <div class="space-y-2">
          <Label for="time" class="text-sm sm:text-base">Start Time</Label>
          <div class="flex items-center gap-2">
            <Clock class="h-4 w-4 text-muted-foreground flex-shrink-0" />
            <Input
              id="time"
              type="time"
              bind:value={selectedTime}
              class="text-base"
            />
          </div>
        </div>

        <div class="space-y-2">
          <Label for="duration" class="text-sm sm:text-base"
            >Duration (minutes)</Label
          >
          <Input
            id="duration"
            type="number"
            bind:value={durationMinutes}
            placeholder="120"
            min="15"
            step="15"
            class="text-base"
          />
        </div>
      </div>

      <div class="space-y-2">
        <Label for="time-details" class="text-sm sm:text-base">
          Time Details (Optional)
          <span
            class="text-xs sm:text-sm text-muted-foreground font-normal block sm:inline"
          >
            When should riders arrive vs. when you depart
          </span>
        </Label>
        <Input
          id="time-details"
          type="text"
          bind:value={timeDetails}
          placeholder="Meet at 6:00 PM, ride out at 6:30 PM"
          class="text-base"
        />
        <p class="text-xs text-muted-foreground">
          Let riders know if there's a different meet time and departure time
        </p>
      </div>

      <Button
        type="button"
        onclick={addOccurrence}
        disabled={!selectedDate}
        class="w-full touch-manipulation"
      >
        <Plus class="mr-2 h-4 w-4" />
        Add Date
      </Button>
    </Card.Content>
  </Card.Root>

  <!-- List of Added Occurrences -->
  {#if occurrences.length > 0}
    <Card.Root>
      <Card.Header>
        <Card.Title class="text-base sm:text-lg">
          Scheduled Dates ({occurrences.length})
        </Card.Title>
      </Card.Header>
      <Card.Content>
        <div class="space-y-2">
          {#each occurrences as occurrence, index}
            <div
              class="flex items-start sm:items-center justify-between p-3 border rounded-lg gap-2"
            >
              <div class="flex-1 min-w-0">
                <div class="font-medium text-sm sm:text-base truncate">
                  {new Date(occurrence.start_date).toLocaleDateString("en-US", {
                    weekday: "short",
                    year: "numeric",
                    month: "short",
                    day: "numeric",
                  })}
                </div>
                <div class="text-xs sm:text-sm text-muted-foreground">
                  {occurrence.start_time.slice(0, 5)}
                  {#if occurrence.event_duration_minutes}
                    • {occurrence.event_duration_minutes} min
                  {/if}
                  {#if occurrence.event_time_details}
                    <span class="block sm:inline"
                      >• {occurrence.event_time_details}</span
                    >
                  {/if}
                </div>
              </div>
              <Button
                type="button"
                variant="ghost"
                size="icon"
                class="flex-shrink-0 touch-manipulation"
                onclick={() => removeOccurrence(index)}
              >
                <X class="h-4 w-4" />
              </Button>
            </div>
          {/each}
        </div>
      </Card.Content>
    </Card.Root>
  {/if}

  {#if error}
    <p class="text-xs sm:text-sm text-destructive">{error}</p>
  {/if}
</div>
