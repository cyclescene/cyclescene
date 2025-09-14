import { clsx, } from "clsx";
import { format, isToday, isTomorrow, isYesterday, parse, parseISO } from "date-fns";
import { twMerge } from "tailwind-merge";

export function cn(...inputs) {
    return twMerge(clsx(inputs));
}

export function formatTime(timeString) {
    if (!timeString) return "N/A";
    const parsedTime = parse(timeString, "HH:mm:ss", new Date());
    return format(parsedTime, "h:mm a");
}

export function formatDate(dateString) {
    if (!dateString) {
        return "";
    }

    const date = parseISO(dateString);

    if (isNaN(date.getTime())) {
        console.warn(`Invalid date string provided to formatDate: ${dateString}`);
        return ""; // Or some other fallback like "Invalid Date"
    }


    if (isToday(date)) {
        return "Today";
    } else if (isTomorrow(date)) {
        return "Tomorrow";
    } else if (isYesterday(date)) {
        return "Yesterday";
    } else {
        // Format as "Day, Mon D" e.g., "Sat, Sep 14"
        return format(date, "eee, MMM d");
    }
}
