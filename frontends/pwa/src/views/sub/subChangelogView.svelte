<script>
  // Ensure ScrollArea is imported from your Shadcn Svelte library
  import { ScrollArea } from "$lib/components/ui/scroll-area";

  // You would typically fetch this data from an API or a local JSON/MD file
  // For the example, we'll use a simple array.
  const changelogData = [
    {
      version: "1.4.1",
      date: "December 3, 2025",
      changes: {
        Added: [
          "Custom date picker component replacing bits-ui Calendar to fix edge cases with repeated date selection.",
          "Visual indicator for today's date in the calendar (accent background with primary border).",
          "Previous and next month days now displayed in calendar grid with reduced opacity for context.",
          "IndexedDB upsert functionality instead of clearing entire database on ride data refresh, preserving data consistency.",
        ],
        Fixed: [
          "Fixed issue where clicking the same date twice in the date picker would clear the selection.",
          "Fixed map zooming out excessively on background ride data refresh by separating initialization from data updates.",
          "Fixed date picker crashing when selecting the same date consecutively.",
          "Fixed duplicate dates appearing in calendar for months with fewer than 31 days.",
        ],
        Changed: [
          "Map now initializes to starting coordinates/zoom level on load, independent of ride data changes.",
          "Map initialization simplified to use MapLibre component props instead of effects.",
          "Map bounds fitting only occurs when rides change, not on every background sync.",
          "IndexedDB sync now updates/inserts rides instead of clearing and repopulating the entire store.",
          "Calendar date calculation now uses proper calendar math with leap year handling.",
        ],
      },
    },
    {
      version: "1.3.0",
      date: "December 2, 2025",
      changes: {
        Added: [
          "New CardLabel component with contextual icons for ride information.",
        ],
        Changed: [
          "Improved ride card typography and spacing.",
          "Added text truncation for long ride titles and venue names.",
          "Simplified location cards layout for better map visibility.",
          "Enhanced navigation bar styling with optimized icon sizing.",
        ],
      },
    },
    {
      version: "1.2.0",
      date: "November 29, 2025",
      changes: {
        Added: [
          "Custom group markers for organized community rides with distinct colors and visual identity.",
          "Route processing backend infrastructure for tracking ride routes from external sources.",
          "Support for RideWithGPS and Strava route imports with automatic URL parsing.",
          "Route caching system to avoid reprocessing duplicate routes across the database.",
          "Distance calculations for routes (kilometers and miles) using Haversine formula.",
          "Route data propagation to frontend with route_id references in ride API responses.",
          "Elevation profile data extraction from GPX route files.",
          "Interactive route map preview in ride details showing the full cycling route.",
          "Elevation profile chart displaying elevation changes over route distance.",
          "Route statistics including total distance in both kilometers and miles.",
        ],
        Changed: [
          "Database schema extended with routes table and route_id foreign keys.",
          "Ride submission handler now processes and links routes to user-submitted events.",
          "Shift2Bikes scraper enhanced with route extraction and processing capabilities.",
          "Ride API responses now include route_id field for frontend route visualization.",
        ],
      },
    },
    {
      version: "1.1.0",
      date: "November 25, 2025",
      changes: {
        Added: [
          "Seasonal event layers with automatic calendar-based triggering.",
          "Pride Month (June): Falling pride and trans flag emojis with cyclists.",
          "Halloween (October): Flying bats with cyclists across the map.",
          "Christmas (December): Falling snowflakes with cyclists.",
          "Dark mode park styling with distinct color (landcover and landuse layers).",
          "Vercel deployment optimization across all frontends to prevent unnecessary builds.",
        ],
        Changed: [
          "Dark mode parks now use a darker grayish-green color (#4a7c59) for better visual balance.",
          "Improved footer layout to prevent safe area inset conflicts on iOS devices.",
          "Updated all scrollable views to respect safe area insets and prevent footer coverage.",
          "Location card on map now properly positioned above footer with safe area padding.",
        ],
      },
    },
    {
      version: "1.0.0",
      date: "September 29, 2025",
      changes: {
        Added: [
          "Initial launch of the CycleScene PWA.",
          "Universal theme switching (Light/Dark/System).",
          "Map view with dynamic tile layers for light/dark mode.",
          "Initial Portland ride data (via Shift2Bikes).",
          "Settings panel with About/Changelog views.",
          "Support for offline use via PWA Service Worker caching.",
        ],
        Fixed: [
          "Resolved an issue where button text was unreadable in Light Mode.",
        ],
      },
    },
    {
      version: "0.9.0",
      date: "September 15, 2025",
      changes: {
        Added: [
          "Initial implementation of Turso database connection for new ride submissions.",
          "Basic framework for Salt Lake City ride submissions.",
        ],
        Changed: [
          "Updated UI to use Shadcn Select component for theme selection.",
        ],
      },
    },
  ];
</script>

<!-- ============================================== -->
<!--          CHANGELOG START                       -->
<!-- ============================================== -->
<div class="changelog-container">
  <ScrollArea class="scroll-wrapper">
    <div class="p-6 sm:p-8 max-w-4xl mx-auto space-y-12 pb-[calc(var(--footer-height)_+_env(safe-area-inset-bottom)_+_10px)]">
      <section class="space-y-6">
      <h1 class="text-3xl font-bold tracking-tight text-foreground">
        Changelog
      </h1>
      <p class="text-sm text-muted-foreground">
        Track what's new and what's been fixed in CycleScene.
      </p>

      <!-- Iterate over the changelog data -->
      {#each changelogData as entry}
        <div class="relative pl-6 pb-8">
          <!-- The Version/Date Header -->
          <div
            class="absolute -left-[2px] ml-0.5 top-2 size-3 bg-primary rounded-full border-2 border-background"
          ></div>

          <h2 class="text-xl font-bold tracking-tight text-primary">
            Version {entry.version}
          </h2>
          <p class="text-sm text-muted-foreground mb-4">
            Released on {entry.date}
          </p>

          <!-- CHANGES SECTION -->
          <div class="space-y-4">
            <!-- Added Features -->
            {#if entry.changes.Added}
              <h3 class="font-semibold text-green-600 dark:text-green-400">
                Added
              </h3>
              <ul class="list-disc list-inside space-y-1 ml-4 text-sm">
                {#each entry.changes.Added as item}
                  <li>{item}</li>
                {/each}
              </ul>
            {/if}

            <!-- Fixed Issues -->
            {#if entry.changes.Fixed}
              <h3 class="font-semibold text-red-600 dark:text-red-400">
                Fixed
              </h3>
              <ul class="list-disc list-inside space-y-1 ml-4 text-sm">
                {#each entry.changes.Fixed as item}
                  <li>{item}</li>
                {/each}
              </ul>
            {/if}

            <!-- Changed Items (Optional) -->
            {#if entry.changes.Changed}
              <h3 class="font-semibold text-yellow-600 dark:text-yellow-400">
                Changed
              </h3>
              <ul class="list-disc list-inside space-y-1 ml-4 text-sm">
                {#each entry.changes.Changed as item}
                  <li>{item}</li>
                {/each}
              </ul>
            {/if}
          </div>
        </div>
      {/each}

      <!-- Placeholder for end of log -->
      <p class="text-center text-muted-foreground text-sm pt-4">
        — End of Changelog —
      </p>
      </section>
    </div>
  </ScrollArea>
</div>

<style>
  .changelog-container {
    height: 100%;
    width: 100%;
    display: flex;
    flex-direction: column;
    overflow: hidden;
  }

  :global(.scroll-wrapper) {
    height: 100%;
    width: 100%;
  }
</style>
