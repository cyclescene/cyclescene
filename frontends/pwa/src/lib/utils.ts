import { clsx, } from "clsx";
import { twMerge } from "tailwind-merge";

import { parseTime, DateFormatter, parseDate, getLocalTimeZone, today } from "@internationalized/date";
import { SvelteMap } from "svelte/reactivity";
import type { RideData, ShiftEvent } from "./types";

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


export function formatDate(dateString: string) {
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

export function getSortedUniqueDatesWithToday(savedRides: RideData[]) {
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



export function filterActiveShiftEvents(scrapedEvents: RideData[], shiftResponse: { events: ShiftEvent[] }): RideData[] {
  const activeEventMap = new Map(shiftResponse.events.map(e => [e.caldaily_id, e]))

  return scrapedEvents.map(event => {
    if (event.ridesource !== 'shift2bikes') {
      return event
    }

    const activeEvent = activeEventMap.get(event.id)
    if (!activeEvent) {
      return null
    }

    // Update cancelled and newsflash if they changed
    const updated = { ...event }
    if (activeEvent.cancelled !== event.cancelled) {
      updated.cancelled = activeEvent.cancelled ? 1 : 0
    }
    if (activeEvent.newsflash && activeEvent.newsflash !== event.newsflash) {
      updated.newsflash = activeEvent.newsflash
    }

    return updated
  }).filter((event): event is RideData => event !== null)
}

