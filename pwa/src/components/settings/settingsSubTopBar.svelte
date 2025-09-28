<script>
  import { Button } from "$lib/components/ui/button";
  import { goBackInHistory, navigateTo, VIEW_SETTINGS } from "$lib/stores";
  import IconChevronLeft from "~icons/bxs/chevron-left";
</script>

<div
  class="flex items-center justify-start gap-5 p-2.5 border-b-[1px] relative z-[500]"
>
  <Button
    disabled={false}
    variant="ghost"
    class="py-2 px-3 text-yellow-400 min-w-10 flex items-center justify-center"
    onclick={() => navigateTo(VIEW_SETTINGS)}
    ><IconChevronLeft />
    <p class="text-sm">Settings</p>
  </Button>
</div>
