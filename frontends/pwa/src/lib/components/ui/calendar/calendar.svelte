<script>
	import { getLocalTimeZone, today, startOfMonth } from "@internationalized/date";
	import * as Calendar from "./index.js";
	import { cn } from "$lib/utils.js";
	import { isEqualMonth } from "@internationalized/date";

	let {
		value = $bindable(),
		class: className,
		weekdayFormat = "short",
		buttonVariant = "ghost",
		captionLayout = "label",
		locale = "en-US",
		months: monthsProp,
		years,
		monthFormat: monthFormatProp = "long",
		yearFormat = "numeric",
		day,
		disableDaysOutsideMonth = false,
		...restProps
	} = $props();

	// Initialize value if not set
	if (!value) {
		value = today(getLocalTimeZone());
	}

	let displayMonth = $state(startOfMonth(value));

	// Get today's date for highlighting
	const todaysDate = today(getLocalTimeZone());

	// Generate calendar data for the displayed month
	const getCalendarDays = (month) => {
		try {
			const start = startOfMonth(month);
			const dayOfWeek = start.toDate(getLocalTimeZone()).getDay();

			// Calculate days in month by getting the last day of the month
			let daysInMonth = 31;
			for (let i = 31; i >= 28; i--) {
				try {
					start.set({ day: i });
					daysInMonth = i;
					break;
				} catch {
					// Day doesn't exist in this month, try previous day
				}
			}

			const days = [];

			// Add empty cells for days before month starts
			for (let i = 0; i < dayOfWeek; i++) {
				days.push(null);
			}

			// Add days of the month
			for (let i = 1; i <= daysInMonth; i++) {
				days.push(start.set({ day: i }));
			}

			// Group into weeks
			const weeks = [];
			for (let i = 0; i < days.length; i += 7) {
				weeks.push(days.slice(i, i + 7));
			}

			return weeks;
		} catch (e) {
			console.error('Calendar error:', e);
			return [[]];
		}
	};

	const weeks = $derived(getCalendarDays(displayMonth));

	// Get weekday names
	const getWeekdays = () => {
		try {
			const date = new Date(2024, 0, 1); // Monday, Jan 1, 2024
			const weekdays = [];
			for (let i = 0; i < 7; i++) {
				const formatter = new Intl.DateTimeFormat(locale, { weekday: weekdayFormat });
				weekdays.push(formatter.format(new Date(date.getTime() + i * 24 * 60 * 60 * 1000)));
			}
			return weekdays;
		} catch (e) {
			return ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'];
		}
	};

	const weekdays = getWeekdays();

	const handlePrevMonth = () => {
		displayMonth = displayMonth.subtract({ months: 1 });
	};

	const handleNextMonth = () => {
		displayMonth = displayMonth.add({ months: 1 });
	};

	const handleDayClick = (date) => {
		value = date;
		displayMonth = startOfMonth(date);
	};

	const monthNames = ['January', 'February', 'March', 'April', 'May', 'June', 'July', 'August', 'September', 'October', 'November', 'December'];
	const getMonthYear = () => {
		try {
			const date = displayMonth.toDate(getLocalTimeZone());
			const month = monthNames[date.getMonth()];
			const year = date.getFullYear();
			return `${month} ${year}`;
		} catch (e) {
			return 'Calendar';
		}
	};
</script>

<div class={cn(
	"bg-background group/calendar p-3 [--cell-size:--spacing(8)]",
	className
)} {...restProps}>
	<!-- Header with month/year and navigation -->
	<div class="flex items-center justify-between gap-2 mb-4">
		<button
			onclick={handlePrevMonth}
			class="p-2 hover:bg-accent rounded-md transition"
			aria-label="Previous month"
		>
			←
		</button>
		<div class="font-semibold text-center min-w-32">
			{getMonthYear()}
		</div>
		<button
			onclick={handleNextMonth}
			class="p-2 hover:bg-accent rounded-md transition"
			aria-label="Next month"
		>
			→
		</button>
	</div>

	<!-- Weekday headers -->
	<div class="grid grid-cols-7 gap-2 mb-2">
		{#each weekdays as weekday}
			<div class="text-center text-sm font-semibold text-muted-foreground py-2">
				{weekday.slice(0, 2)}
			</div>
		{/each}
	</div>

	<!-- Calendar grid -->
	<div class="grid grid-cols-7 gap-2">
		{#each weeks as week}
			{#each week as date}
				{#if date}
					<button
						onclick={() => handleDayClick(date)}
						class={cn(
							"p-2 rounded-md text-sm font-medium transition h-9 w-9 flex items-center justify-center",
							value && date.toString() === value.toString()
								? "bg-primary text-primary-foreground"
								: date.toString() === todaysDate.toString()
								? "bg-accent text-accent-foreground border-2 border-primary"
								: "hover:bg-accent"
						)}
					>
						{date.day}
					</button>
				{:else}
					<div class="h-9 w-9"></div>
				{/if}
			{/each}
		{/each}
	</div>
</div>
