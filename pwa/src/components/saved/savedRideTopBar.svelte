<script>
    import Button from "$lib/components/ui/button/button.svelte";
    import { savedRidesGroupedByDate } from "$lib/stores";
    import IconChevronRight from "~icons/bxs/chevron-right";
    import IconChevronLeft from "~icons/bxs/chevron-left";
    import { getLocalTimeZone, today } from "@internationalized/date";
    import { formatDate } from "$lib/utils";

    const savedDates = $savedRidesGroupedByDate.map((rides) => {
        return rides.date;
    });

    const todaysDate = formatDate(today(getLocalTimeZone()).toString());
    $: {
        console.log(savedDates);
    }
</script>

<div
    class="flex items-center justify-center p-2.5 bg-black text-white border-b-[1px] relative z-[500]"
>
    <Button class="bg-black py-2 px-3 text-yellow-400 min-w-10"
        ><IconChevronLeft /></Button
    >

    <button
        class="text-xl text-yellow-400 grow font-bold text-center py-2 px-3"
    >
        {todaysDate}
    </button>
    <Button class="bg-black py-2 px-3 text-yellow-400 min-w-10"
        ><IconChevronRight /></Button
    >
</div>
