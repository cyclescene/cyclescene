<script>
  import { Button } from "$lib/components/ui/button";
  import { jumpToView, VIEW_SETTINGS } from "$lib/stores";
  import IconChevronLeft from "~icons/bxs/chevron-left";

  function closeSubView() {
    jumpToView(VIEW_SETTINGS);
  }
</script>

<div
  class="flex items-center justify-start gap-5 p-2.5 border-b-[1px] relative z-[500]"
>
  <Button
    disabled={false}
    variant="ghost"
    class="py-2 px-3 text-yellow-400 min-w-10 flex items-center justify-center"
    onclick={closeSubView}
    ><IconChevronLeft />
    <p class="text-sm">Settings</p>
  </Button>
</div>
