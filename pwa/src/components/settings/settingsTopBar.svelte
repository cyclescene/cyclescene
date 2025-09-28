<script>
  import Button from "$lib/components/ui/button/button.svelte";
  import { goBackInHistory } from "$lib/stores";
  import IconChevronLeft from "~icons/bxs/chevron-left";
</script>

<div
  class="flex items-center justify-center gap-5 p-2.5 border-b-[1px] relative z-[500]"
>
  <Button
    disabled={false}
    variant="secondary"
    class="shrink py-2 px-3 text-yellow-400 min-w-10"
    onclick={goBackInHistory}><IconChevronLeft /></Button
  >

  <span class="grow text-center mr-12 text-2xl font-bold">Settings</span>
</div>
