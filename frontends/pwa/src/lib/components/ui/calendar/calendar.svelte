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

	// Helper to get days in a month using simple month math
	const getDaysInMonth = (m) => {
		const monthNum = m.month;
		const year = m.year;
		// Days in each month (non-leap year)
		const daysPerMonth = [31, 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31];
		// Check for leap year
		if (monthNum === 2 && ((year % 4 === 0 && year % 100 !== 0) || year % 400 === 0)) {
			return 29;
		}
		return daysPerMonth[monthNum - 1];
	};

	// Generate calendar data for the displayed month
	const getCalendarDays = (month) => {
		try {
			const dayOfWeek = startOfMonth(month).toDate(getLocalTimeZone()).getDay();
			const daysInMonth = getDaysInMonth(month);

			const days = [];

			// Add previous month's days
			const prevMonth = month.subtract({ months: 1 });
			const prevMonthDays = getDaysInMonth(prevMonth);
			const startDay = prevMonthDays - dayOfWeek + 1;

			for (let i = startDay; i <= prevMonthDays; i++) {
				// Create fresh date each iteration
				const d = startOfMonth(prevMonth).set({ day: i });
				days.push({ date: d, isCurrentMonth: false });
			}

			// Add current month's days
			for (let i = 1; i <= daysInMonth; i++) {
				// Create fresh date each iteration
				const d = startOfMonth(month).set({ day: i });
				days.push({ date: d, isCurrentMonth: true });
			}

			// Add next month's days
			const nextMonth = month.add({ months: 1 });
			const nextMonthDays = getDaysInMonth(nextMonth);
			const remainingDays = 42 - days.length;

			for (let i = 1; i <= remainingDays; i++) {
				// Create fresh date each iteration
				const d = startOfMonth(nextMonth).set({ day: i });
				days.push({ date: d, isCurrentMonth: false });
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
			{#each week as dayObj}
				{#if dayObj}
					<button
						onclick={() => handleDayClick(dayObj.date)}
						class={cn(
							"p-2 rounded-md text-sm font-medium transition h-9 w-9 flex items-center justify-center",
							!dayObj.isCurrentMonth ? "text-muted-foreground opacity-50" : "",
							value && dayObj.date.toString() === value.toString()
								? "bg-primary text-primary-foreground"
								: dayObj.date.toString() === todaysDate.toString()
								? "bg-accent text-accent-foreground border-2 border-primary"
								: "hover:bg-accent"
						)}
					>
						{dayObj.date.day}
					</button>
				{:else}
					<div class="h-9 w-9"></div>
				{/if}
			{/each}
		{/each}
	</div>
</div>
