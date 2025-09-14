<script>
    import Button from "$lib/components/ui/button/button.svelte";
    import {
        navigateTo,
        VIEW_LIST,
        VIEW_MAP,
        VIEW_SAVED,
        VIEW_SETTINGS,
    } from "../lib/stores";

    const controls = [VIEW_MAP, VIEW_LIST, VIEW_SAVED, VIEW_SETTINGS];

    function onChangeView(view) {
        navigateTo(view);
    }
</script>

<div class="w-full h-full flex justify-evenly">
    {#each controls as view}
        <Button
            class="grow h-full rounded-none"
            onclick={() => onChangeView(view)}
        >
            {view}
        </Button>
    {/each}
</div>
