<script>
    import Button from "$lib/components/ui/button/button.svelte";
    import {
        navigateTo,
        VIEW_LIST,
        VIEW_MAP,
        VIEW_SAVED,
        VIEW_SETTINGS,
    } from "../lib/stores";

    const controls = [
        {
            name: "Map",
            view: VIEW_MAP,
        },
        {
            name: "List",
            view: VIEW_LIST,
        },
        {
            name: "Saved",
            view: VIEW_SAVED,
        },
        {
            name: "Settings",
            view: VIEW_SETTINGS,
        },
    ];

    function onChangeView(view) {
        navigateTo(view);
    }
</script>

<div class="w-full h-full flex justify-evenly">
    {#each controls as view (view.name)}
        <Button
            class="bg-black grow h-full rounded-none"
            onclick={() => onChangeView(view.view)}
        >
            {view.name}
        </Button>
    {/each}
</div>
