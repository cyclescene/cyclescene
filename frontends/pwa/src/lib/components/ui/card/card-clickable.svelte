<script>
    import { cn } from "$lib/utils"; // Assuming cn utility is at $lib/utils

    // Define props for the clickable card
    let {
        ref = $bindable(null),
        onClick, // Required click handler
        class: customClass = "", // Allows consumer to add custom classes
        children,
        ...restProps // Capture any other props to forward to the Card component
    } = $props();

    // Handle click event
    function handleClick() {
        onClick();
    }

    // Handle keyboard events for accessibility (Enter and Space keys)
    function handleKeyDown(event) {
        if (event.key === "Enter") {
            event.preventDefault(); // Prevent default browser behavior (e.g., form submission)
            onClick();
        }
    }

    function handleKeyUp(event) {
        if (event.key === " ") {
            event.preventDefault(); // Prevent default browser behavior (e.g., page scroll)
            onClick();
        }
    }

    // Define hover/focus styles to make it visually interactive
    const interactiveStyles =
        "cursor-pointer transition-colors " +
        "bg-card text-card-foreground flex flex-col gap-6 rounded-xl border py-6 shadow-sm" +
        "hover:bg-accent hover:text-accent-foreground " +
        "focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2";
</script>

<div
    bind:this={ref}
    class={cn(interactiveStyles, customClass)}
    role="button"
    tabindex="0"
    on:click={handleClick}
    on:keydown={handleKeyDown}
    on:keyup={handleKeyUp}
    {...restProps}
>
    {@render children?.()}
</div>
