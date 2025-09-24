import { clsx, } from "clsx";
import { twMerge } from "tailwind-merge";

import { format, parse, parseISO, isToday, isTomorrow, isYesterday } from 'date-fns';
import { parseTime } from "@internationalized/date";
import { DateFormatter } from "@internationalized/date";
import { parseDate } from "@internationalized/date";
import { getLocalTimeZone } from "@internationalized/date";
import { today } from "@internationalized/date";
import { SvelteMap } from "svelte/reactivity";
export function cn(...inputs) {
    return twMerge(clsx(inputs));
}

export function formatTime(timeString) {
    const parsedTime = parseTime(timeString)

    let now = new Date()
    let dateForFormatting = new Date(
        now.getFullYear(),
        now.getMonth(),
        now.getDate(),
        parsedTime.hour,
        parsedTime.minute,
        parsedTime.second
    )

    const timeFormatter = new DateFormatter("en-US", {
        hour: "numeric",
        minute: "2-digit",
        hour12: true
    })

    return timeFormatter.format(dateForFormatting)
}

export function formatDate(dateString) {
    if (!dateString) {
        return "";
    }

    // Parse the ISO date string into a Date object
    const date = parseDate(dateString);

    // Check if the parsed date is valid before proceeding
    if (date.toString() === "Invalid Date") {
        console.warn(`Invalid date string provided to formatDate: ${dateString}`);
        return "Invalid Date"; // Or some other fallback like "Invalid Date"
    }

    const todaysDate = today(getLocalTimeZone())
    const tomorrowsDate = todaysDate.add({ days: 1 })
    const yesterdaysDate = todaysDate.subtract({ days: -1 })

    const dateFormatter = new DateFormatter("en-US", {
        weekday: "short",
        month: 'short',
        day: "numeric"
    })

    if (date.compare(todaysDate) === 0) {
        return "Today"
    } else if (date.compare(tomorrowsDate) === 0) {
        return "Tomorrow"
    } else if (date.compare(yesterdaysDate) === 0) {
        return "Yesterday"
    } else {
        return dateFormatter.format(date.toDate(getLocalTimeZone()))
    }

}

export function getSortedUniqueDatesWithToday(savedRides) {
    const existingCalendarDates = savedRides.map(ride => ride.date);

    const todaysCalendarDate = today(getLocalTimeZone())

    const allCalendarDates = [...existingCalendarDates, todaysCalendarDate]

    const uniqueDatesMap = new SvelteMap()
    for (const calendarDate of allCalendarDates) {
        uniqueDatesMap.set(calendarDate.toString(), calendarDate)
    }

    const uniqueSortedDates = Array.from(uniqueDatesMap.values()).sort((a, b) => a.compare(b))

    return uniqueSortedDates
}
