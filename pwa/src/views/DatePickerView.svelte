<script>
    import Calendar from "$lib/components/ui/calendar/calendar.svelte";
    import { dateStore, goBackInHistory } from "$lib/stores";
    import { getLocalTimeZone, today } from "@internationalized/date";

    let value = $state(today(getLocalTimeZone()));

    $effect(() => {
        dateStore.setSpecificDate(value);
        goBackInHistory();
    });
</script>

<div class="h-[100vh] bg-black mx-auto">
    <Calendar
        bind:value
        class=" absolute bg-black text-white p-5 border-2 border-white rounded-xl text-2xl"
    />
</div>
