<script lang="ts">
  import * as Card from "$lib/components/ui/card";
  import { Input } from "$lib/components/ui/input";
  import { Label } from "$lib/components/ui/label";
  import { Button } from "$lib/components/ui/button";
  import GroupSelector from "$lib/components/ride-form/GroupSelector.svelte";

  interface Props {
    onSubmit: (email: string, groupCode: string) => void;
    onBack: () => void;
  }

  let { onSubmit, onBack }: Props = $props();

  let email = $state("");
  let groupCode = $state("");
  let error = $state<string | null>(null);

  // Simple email validation
  function isValidEmail(value: string): boolean {
    return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(value);
  }

  function handleSubmit(e: Event) {
    e.preventDefault();

    const trimmedEmail = email.trim();

    if (!trimmedEmail) {
      error = "Email is required";
      return;
    }

    if (!isValidEmail(trimmedEmail)) {
      error = "Please enter a valid email address";
      return;
    }

    error = null;
    onSubmit(trimmedEmail, groupCode);
  }

  function handleEmailInput() {
    // Clear error when user starts typing
    if (error) {
      error = null;
    }
  }

  function handleGroupChange(value: string) {
    groupCode = value;
  }
</script>

<Card.Root class="mx-auto max-w-md">
  <Card.Header>
    <Card.Title>Import Settings</Card.Title>
    <Card.Description>
      Configure your import before selecting events.
    </Card.Description>
  </Card.Header>
  <Card.Content>
    <form onsubmit={handleSubmit}>
      <div class="space-y-6">
        <!-- Email Input -->
        <div class="space-y-2">
          <Label for="organizer-email">Email Address</Label>
          <Input
            id="organizer-email"
            type="email"
            placeholder="you@example.com"
            bind:value={email}
            oninput={handleEmailInput}
            aria-invalid={!!error}
            class={error ? "border-red-500" : ""}
          />
          {#if error}
            <p class="text-sm text-red-500">{error}</p>
          {/if}
          <p class="text-muted-foreground text-xs">
            We'll send you edit links for your imported events.
          </p>
        </div>

        <!-- Group Selector -->
        <GroupSelector
          value={groupCode}
          onchange={handleGroupChange}
        />

        <div class="flex gap-2 pt-2">
          <Button type="button" variant="outline" onclick={onBack}>
            Back
          </Button>
          <Button type="submit" class="flex-1">
            Continue
          </Button>
        </div>
      </div>
    </form>
  </Card.Content>
</Card.Root>
