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

<Card.Root class="mx-auto max-w-md border-2 shadow-lg">
  <Card.Header class="space-y-3 pb-6">
    <div class="flex items-center gap-3">
      <div class="flex h-12 w-12 items-center justify-center rounded-xl bg-[#FC5200]/10 ring-2 ring-[#FC5200]/20">
        <svg class="h-6 w-6 text-[#FC5200]" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
        </svg>
      </div>
      <div>
        <Card.Title class="text-xl">Import Settings</Card.Title>
        <Card.Description class="text-sm">
          Configure before selecting events
        </Card.Description>
      </div>
    </div>
  </Card.Header>
  <Card.Content>
    <form onsubmit={handleSubmit}>
      <div class="space-y-6">
        <!-- Email Input -->
        <div class="space-y-2">
          <Label for="organizer-email" class="text-sm font-semibold">Email Address</Label>
          <div class="relative">
            <Input
              id="organizer-email"
              type="email"
              placeholder="you@example.com"
              bind:value={email}
              oninput={handleEmailInput}
              aria-invalid={!!error}
              aria-describedby={error ? "email-error" : "email-help"}
              class="pl-10 h-11 text-base {error ? 'border-red-500 focus:ring-red-500' : 'focus:ring-[#FC5200] focus:border-[#FC5200]'}"
              autofocus
            />
            <svg class="absolute left-3 top-1/2 -translate-y-1/2 h-5 w-5 text-muted-foreground pointer-events-none" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 12a4 4 0 10-8 0 4 4 0 008 0zm0 0v1.5a2.5 2.5 0 005 0V12a9 9 0 10-9 9m4.5-1.206a8.959 8.959 0 01-4.5 1.207" />
            </svg>
          </div>
          {#if error}
            <p id="email-error" class="text-sm text-red-600 font-medium flex items-center gap-1.5 animate-in fade-in slide-in-from-top-1 duration-300" role="alert">
              <svg class="h-4 w-4 flex-shrink-0" fill="currentColor" viewBox="0 0 20 20" aria-hidden="true">
                <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clip-rule="evenodd" />
              </svg>
              {error}
            </p>
          {/if}
          <p id="email-help" class="text-muted-foreground text-xs flex items-start gap-1.5">
            <svg class="h-4 w-4 flex-shrink-0 mt-0.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            <span>We'll send edit links for your imported events to this email</span>
          </p>
        </div>

        <!-- Group Selector -->
        <div class="rounded-lg border bg-muted/30 p-4">
          <GroupSelector
            value={groupCode}
            onchange={handleGroupChange}
          />
        </div>

        <div class="flex flex-col sm:flex-row gap-3 pt-4">
          <Button type="button" variant="outline" size="lg" onclick={onBack} class="w-full sm:flex-1">
            <svg class="h-4 w-4 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 19l-7-7m0 0l7-7m-7 7h18" />
            </svg>
            Back
          </Button>
          <Button type="submit" size="lg" class="w-full sm:flex-1 bg-[#FC5200] hover:bg-[#E04A00] text-white shadow-sm hover:shadow-md transition-all duration-200">
            Continue
            <svg class="h-4 w-4 ml-2" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14 5l7 7m0 0l-7 7m7-7H3" />
            </svg>
          </Button>
        </div>
      </div>
    </form>
  </Card.Content>
</Card.Root>
