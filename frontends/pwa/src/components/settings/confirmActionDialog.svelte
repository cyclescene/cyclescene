<script lang="ts">
  import Button from "$lib/components/ui/button/button.svelte";

  let {
    open = $bindable(false),
    title,
    description,
    confirmLabel = "Confirm",
    destructive = false,
    onConfirm,
  }: {
    open: boolean;
    title: string;
    description: string;
    confirmLabel?: string;
    destructive?: boolean;
    onConfirm: () => void | Promise<void>;
  } = $props();

  async function handleConfirm() {
    open = false;
    await onConfirm();
  }

  function handleKeydown(event: KeyboardEvent) {
    if (open && event.key === "Escape") {
      open = false;
    }
  }
</script>

<svelte:window onkeydown={handleKeydown} />

{#if open}
  <div
    class="fixed inset-0 z-[1200] bg-black/45 flex items-end sm:items-center justify-center p-4"
    role="presentation"
    onclick={() => (open = false)}
  >
    <div
      aria-modal="true"
      aria-labelledby="confirm-action-title"
      tabindex="-1"
      class="w-full max-w-sm rounded-lg border bg-background text-foreground shadow-xl p-4"
      role="dialog"
      onkeydown={(event) => event.stopPropagation()}
      onclick={(event) => event.stopPropagation()}
    >
      <h2 id="confirm-action-title" class="text-lg font-semibold leading-tight">
        {title}
      </h2>
      <p class="mt-2 text-sm text-muted-foreground leading-relaxed">
        {description}
      </p>

      <div class="mt-5 grid grid-cols-2 gap-2">
        <Button variant="outline" onclick={() => (open = false)}>
          Cancel
        </Button>
        <Button
          class={destructive ? "bg-destructive text-white hover:bg-destructive/90" : ""}
          onclick={handleConfirm}
        >
          {confirmLabel}
        </Button>
      </div>
    </div>
  </div>
{/if}
