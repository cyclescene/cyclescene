<script>
  // Ensure ScrollArea is imported from your Shadcn Svelte library
  import { ScrollArea } from "$lib/components/ui/scroll-area";

  // You would typically fetch this data from an API or a local JSON/MD file
  // For the example, we'll use a simple array.
  const changelogData = [
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
<div
  class="p-6 sm:p-8 max-w-4xl mx-auto space-y-12 absolute top-0 bottom-[75px] min-h-[calc(100vh_-_115px)]"
>
  <ScrollArea class="relative" scrollbarYClasses={`hidden`}>
    <section class="space-y-6 h-[calc(100vh_-_160px)] flex flex-col gap-2">
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
  </ScrollArea>
</div>
