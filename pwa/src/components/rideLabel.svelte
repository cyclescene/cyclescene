<script>
    export let color;
    export let content;
    export let icon;
</script>

<div
    class={`flex flex-row gap-1.5 border-2 border-${color} text-${color} text-sm font-bold rounded-lg px-2 py-1.5`}
>
    <div class={`text-${color}`}>
        <svelte:component this={icon} />
    </div>
    {content}
</div>
