<script>
  // Ensure ScrollArea is imported from your Shadcn Svelte library
  import { ScrollArea } from "$lib/components/ui/scroll-area";
</script>

<!-- ============================================== -->
<!--          ABOUT SECTION START                   -->
<!-- ============================================== -->
<div
  class="p-6 sm:p-8 max-w-4xl mx-auto space-y-12 absolute top-[60px] bottom-[75px] min-h-[calc(100vh_-_115px)]"
>
  <ScrollArea>
    <section class="space-y-6 h-[calc(100vh_-_160px)] flex flex-col gap-2">
      <h1 class="text-3xl font-bold tracking-tight text-foreground">
        About CycleScene
      </h1>
      <p class="text-sm text-muted-foreground">The story behind the app.</p>

      <h2 class="text-xl font-semibold mt-6">
        1. The Inspiration: A Game-Changer Called BikeFun
      </h2>
      <p>
        CycleScene was born out of my deep appreciation for a single app:
        <strong>BikeFun</strong>. The BikeFun native iOS app was, and continues
        to be, a game-changer for navigating the daily schedule of rides during
        events like Portland's Pedalpalooza. It's an indispensable tool that
        truly embodies the spirit of local cycling.
      </p>

      <h2 class="text-xl font-semibold mt-6">
        2. The Problem: Making the Fun Inclusive
      </h2>
      <p>
        However, a simple moment of sharing revealed a significant barrier: when
        I shared my favorite rides with friends, I quickly learned that many
        cyclists were on Android or other devices and were completely locked out
        of using BikeFun. A fantastic, community-driven resource was
        unintentionally unavailable to a large portion of the cycling community.
      </p>

      <h2 class="text-xl font-semibold mt-6">
        3. Our Solution: A Universal PWA for Everyone
      </h2>
      <p>
        This realization sparked the mission for CycleScene: to provide the same
        essential functionality—the ability to easily find and join the city's
        best rides—to <strong>everyone</strong>, regardless of their platform.
      </p>
      <p>
        That's why <strong
          >CycleScene is built as a Progressive Web App (PWA)</strong
        >. By being platform-agnostic, CycleScene ensures no cyclist is ever
        left out of the scene, working seamlessly on iOS, Android, and desktop.
      </p>

      <!-- COMMUNITY NOTE / BIKEFUN SUPPORT BLOCK -->
      <div class="p-4 border border-border bg-muted rounded-lg space-y-3">
        <h3 class="text-lg font-bold">
          A Note to iOS Users: Support the Original
        </h3>
        <p class="text-sm">
          If you are currently on an iOS device and love the polished, native
          experience, I wholeheartedly encourage you to download the original
          app. We strongly recommend using BikeFun for the definitive native
          experience.
        </p>
        <a
          href="https://apps.apple.com/us/app/bike-fun-group-ride-calendar/id6451367875"
          target="_blank"
          rel="noopener noreferrer"
          class="inline-block text-primary hover:underline font-medium text-sm"
        >
          Download BikeFun on the App Store ↗
        </a>
      </div>

      <h2 class="text-xl font-semibold mt-6">4. What We Are Building Today</h2>
      <p>
        CycleScene is actively expanding to help local communities organize
        their own events. While we start with core data from established
        sources, we are building the tools necessary to support self-organized
        communities, starting with areas like Salt Lake City, by enabling hosts
        to easily submit and update local rides directly.
      </p>
      <p class="text-lg font-bold mt-4">Now go ride!</p>
    </section>
  </ScrollArea>
</div>
