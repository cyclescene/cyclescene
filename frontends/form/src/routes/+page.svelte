<script lang="ts">
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

  interface Props {
    data: {
      form: any;
      token: string;
      city: string;
    };
  }

  let { data }: Props = $props();

  const { form, errors, enhance, delayed, message } = superForm(data.form, {
    validators: zodClient(rideSubmissionSchema),
    dataType: "json",
    resetForm: false,
    onError({ result }) {
      $message = result.error.message;
    },
  });
</script>

<div class="container max-w-4xl mx-auto py-4 sm:py-8 px-4">
  <div class="mb-6 sm:mb-8">
    <h1 class="text-2xl sm:text-3xl font-bold tracking-tight">Host a Ride</h1>
    <p class="text-sm sm:text-base text-muted-foreground mt-1 sm:mt-2">
      Share your ride with the {data.city.toUpperCase()} cycling community
    </p>
  </div>

  {#if $message}
    <div
      class="mb-4 sm:mb-6 p-3 sm:p-4 border border-destructive bg-destructive/10 rounded-lg"
    >
      <p class="text-xs sm:text-sm text-destructive">{$message}</p>
    </div>
  {/if}

  <form method="POST" use:enhance class="space-y-6 sm:space-y-8">
    <!-- Basic Information -->
    <Card.Root>
      <Card.Header>
        <Card.Title class="text-lg sm:text-xl">Basic Information</Card.Title>
        <Card.Description class="text-sm">
          Tell riders what your ride is about
        </Card.Description>
      </Card.Header>
      <Card.Content class="space-y-4 sm:space-y-4">
        <div class="space-y-2">
          <Label for="title" class="text-sm sm:text-base">Ride Title *</Label>
          <Input
            id="title"
            type="text"
            bind:value={$form.title}
            placeholder="Sunday Morning Coffee Cruise"
            class={`text-base ${$errors.title ? "border-destructive" : ""}`}
          />
          {#if $errors.title}
            <p class="text-xs sm:text-sm text-destructive">{$errors.title}</p>
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
            >Description *</Label
          >
          <Textarea
            id="description"
            bind:value={$form.description}
            placeholder="Join us for a casual morning ride through the city. We'll stop at local coffee shops along the way..."
            rows={5}
            class={`text-base ${$errors.description ? "border-destructive" : ""}`}
          />
          {#if $errors.description}
            <p class="text-xs sm:text-sm text-destructive">
              {$errors.description}
            </p>
          {/if}
        </div>

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
      </Card.Content>
    </Card.Root>

    <!-- Location Information -->
    <Card.Root>
      <Card.Header>
        <Card.Title>Location</Card.Title>
        <Card.Description>Where does the ride start and end?</Card.Description>
      </Card.Header>
      <Card.Content class="space-y-4">
        <div class="space-y-2">
          <Label for="venue_name">Starting Location Name *</Label>
          <Input
            id="venue_name"
            type="text"
            bind:value={$form.venue_name}
            placeholder="Pioneer Courthouse Square"
            class={$errors.venue_name ? "border-destructive" : ""}
          />
          {#if $errors.venue_name}
            <p class="text-sm text-destructive">{$errors.venue_name}</p>
          {/if}
        </div>

        <div class="space-y-2">
          <Label for="address">Address *</Label>
          <Input
            id="address"
            type="text"
            bind:value={$form.address}
            placeholder="701 SW 6th Ave, Portland, OR 97204"
            class={$errors.address ? "border-destructive" : ""}
          />
          {#if $errors.address}
            <p class="text-sm text-destructive">{$errors.address}</p>
          {/if}
        </div>

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
      </Card.Content>
    </Card.Root>

    <!-- Date & Time Section -->
    <Card.Root>
      <Card.Header>
        <Card.Title class="text-lg sm:text-xl">Date & Time</Card.Title>
        <Card.Description class="text-sm">
          When does your ride happen?
        </Card.Description>
      </Card.Header>
      <Card.Content class="space-y-4 sm:space-y-4">
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

        <DateTimePicker
          bind:occurrences={$form.occurrences}
          dateType={$form.date_type}
          onupdate={(occs) => {
            $form.occurrences = occs;
          }}
          error={$errors.occurrences}
        />
      </Card.Content>
    </Card.Root>

    <!-- Contact Information -->
    <Card.Root>
      <Card.Header>
        <Card.Title>Contact Information</Card.Title>
        <Card.Description>How can riders reach you?</Card.Description>
      </Card.Header>
      <Card.Content class="space-y-4">
        <div class="space-y-2">
          <Label for="organizer_name">Your Name *</Label>
          <Input
            id="organizer_name"
            type="text"
            bind:value={$form.organizer_name}
            placeholder="Jane Doe"
            class={$errors.organizer_name ? "border-destructive" : ""}
          />
          {#if $errors.organizer_name}
            <p class="text-sm text-destructive">{$errors.organizer_name}</p>
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
            class={$errors.organizer_email ? "border-destructive" : ""}
          />
          {#if $errors.organizer_email}
            <p class="text-sm text-destructive">{$errors.organizer_email}</p>
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
            />
            {#if $errors.web_url}
              <p class="text-sm text-destructive">{$errors.web_url}</p>
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
      </Card.Content>
    </Card.Root>

    <!-- Group Association -->
    <Card.Root>
      <Card.Header>
        <Card.Title>Group Association</Card.Title>
        <Card.Description>Link this ride to a cycling group</Card.Description>
      </Card.Header>
      <Card.Content>
        <GroupSelector
          bind:value={$form.group_code}
          onchange={(val) => {
            $form.group_code = val;
          }}
          error={$errors.group_code}
        />
      </Card.Content>
    </Card.Root>

    <!-- Additional Details -->
    <Card.Root>
      <Card.Header>
        <Card.Title>Additional Details</Card.Title>
        <Card.Description>
          Any extra information riders should know
        </Card.Description>
      </Card.Header>
      <Card.Content class="space-y-4">
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
      </Card.Content>
    </Card.Root>

    <!-- Submit Button -->
    <div
      class="sticky bottom-0 bg-background border-t pt-4 pb-4 sm:pb-6 -mx-4 px-4 sm:mx-0 sm:px-0 sm:static"
    >
      <div
        class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3"
      >
        <p
          class="text-xs sm:text-sm text-muted-foreground text-center sm:text-left"
        >
          * Required fields
        </p>
        <Button
          type="submit"
          disabled={$delayed}
          size="lg"
          class="w-full sm:w-auto touch-manipulation"
        >
          {#if $delayed}
            Submitting...
          {:else}
            Submit Ride for Review
          {/if}
        </Button>
      </div>
    </div>

    <div class="text-center text-xs sm:text-sm text-muted-foreground pb-4">
      Your ride will be reviewed before appearing on CycleScene.
      <br class="hidden sm:block" />
      <span class="block sm:inline mt-1 sm:mt-0"
        >You'll receive a magic link via email to edit your ride.</span
      >
    </div>
  </form>
</div>
