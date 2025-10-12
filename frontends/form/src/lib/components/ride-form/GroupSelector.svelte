<script lang="ts">
  import { validateGroupCode } from "$lib/api/client";
  import { Input } from "$lib/components/ui/input";
  import { Label } from "$lib/components/ui/label";
  import { CircleCheck, CircleX, Loader } from "@lucide/svelte";

  interface Props {
    value: string;
    onchange: (value: string) => void;
    error?: string;
  }

  let { value = $bindable(), onchange, error }: Props = $props();

  let validationState = $state<"idle" | "validating" | "valid" | "invalid">(
    "idle",
  );
  let groupName = $state<string>("");
  let debounceTimer: ReturnType<typeof setTimeout> | null = $state(null);

  async function handleRegisterGroup() {
    try {
      const url = `http://localhost:8080/v1/tokens/submission`;

      const response = await fetch(url, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ city: "pdx" }),
      });

      const { token } = await response.json();

      // Redirect to group registration form
      window.location.href = `/group?token=${token}&city=pdx`;
    } catch (error) {
      console.error("Error:", error);
    }
  }

  async function handleValidation(code: string) {
    if (!code || code.length !== 4) {
      validationState = "idle";
      groupName = "";
      return;
    }

    validationState = "validating";

    try {
      const result = await validateGroupCode(code);

      if (result.valid && result.name) {
        validationState = "valid";
        groupName = result.name;
      } else {
        validationState = "invalid";
        groupName = "";
      }
    } catch (err) {
      validationState = "invalid";
      groupName = "";
      console.error("Group validation error:", err);
    }
  }

  function handleInput(e: Event) {
    const target = e.target as HTMLInputElement;
    const newValue = target.value.toUpperCase().slice(0, 4);
    value = newValue;
    onchange(newValue);

    // Clear existing timer
    if (debounceTimer) {
      clearTimeout(debounceTimer);
    }

    // Reset validation state while typing
    if (newValue.length < 4) {
      validationState = "idle";
      groupName = "";
      return;
    }

    // Debounce validation
    debounceTimer = setTimeout(() => {
      handleValidation(newValue);
    }, 500);
  }
</script>

<div class="space-y-2">
  <Label for="group-code">
    Group Code (Optional)
    <span class="text-sm text-muted-foreground font-normal">
      - 4 characters
    </span>
  </Label>

  <div class="relative">
    <Input
      id="group-code"
      type="text"
      {value}
      oninput={handleInput}
      placeholder="BIKE"
      maxlength={4}
      class="uppercase pr-10"
    />

    <div class="absolute right-3 top-1/2 -translate-y-1/2">
      {#if validationState === "validating"}
        <Loader class="h-4 w-4 animate-spin text-muted-foreground" />
      {:else if validationState === "valid"}
        <CircleCheck class="h-4 w-4 text-green-600" />
      {:else if validationState === "invalid"}
        <CircleX class="h-4 w-4 text-destructive" />
      {/if}
    </div>
  </div>

  {#if validationState === "valid" && groupName}
    <p class="text-sm text-green-600 flex items-center gap-1">
      <CircleCheck class="h-3 w-3" />
      Associated with {groupName}
    </p>
  {:else if validationState === "invalid"}
    <p class="text-sm text-destructive flex items-center gap-1">
      <CircleX class="h-3 w-3" />
      Group code not found
    </p>
  {/if}

  {#if error}
    <p class="text-sm text-destructive">{error}</p>
  {/if}

  <p class="text-xs text-muted-foreground">
    Don't have a group? <button
      onclick={handleRegisterGroup}
      class="underline hover:text-foreground">Register one here</button
    >
  </p>
</div>
