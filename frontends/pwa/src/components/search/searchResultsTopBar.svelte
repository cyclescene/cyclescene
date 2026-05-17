<script lang="ts">
  import Button from "$lib/components/ui/button/button.svelte";
  import { goBackInHistory } from "$lib/stores";
  import IconChevronLeft from "~icons/bxs/chevron-left";
</script>

<div class="relative z-[500] flex items-center gap-3 border-b-[1px] p-2.5">
  <Button
    variant="ghost"
    class="flex min-w-10 items-center justify-center px-3 py-2 text-yellow-400"
    onclick={goBackInHistory}
    aria-label="Back"
  >
    <IconChevronLeft />
  </Button>

  <h1 class="absolute left-1/2 -translate-x-1/2 text-xl font-bold">
    Results
  </h1>
</div>
