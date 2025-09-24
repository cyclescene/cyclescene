<script>
    import { cn } from "$lib/utils";

    let {
        ref = $bindable(null),
        class: className,
        children,
        ...restProps
    } = $props();
</script>

<div
    bind:this={ref}
    data-slot="ride-label"
    class={cn(
        "flex flex-row gap-1.5 border-2 text-sm font-bold rounded-lg pl-2 pr-3 py-1.5",
        className,
    )}
    {...restProps}
>
    {@render children?.()}
</div>
