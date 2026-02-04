<script lang="ts">
  import * as Card from "$lib/components/ui/card";
  import { Button } from "$lib/components/ui/button";
  import { ListOrdered, TrendingUp, MapPin } from "@lucide/svelte";
</script>

<div class="container max-w-6xl mx-auto py-8 md:py-16 px-4">
  <!-- Hero Section -->
  <div class="text-center space-y-4 mb-12 md:mb-16">
    <div
      class="inline-block px-4 py-1.5 rounded-full bg-primary/10 text-primary text-sm font-medium mb-4"
    >
      Admin Dashboard
    </div>
    <h1
      class="text-4xl md:text-6xl font-bold tracking-tight bg-gradient-to-br from-foreground to-foreground/70 bg-clip-text text-transparent"
    >
      CycleScene Dashboard
    </h1>
    <p class="text-lg md:text-xl text-muted-foreground max-w-2xl mx-auto">
      Review and publish community-submitted rides across Portland and Salt Lake
      City
    </p>
  </div>

  <!-- Cards Grid -->
  <div class="grid grid-cols-1 md:grid-cols-2 gap-6 max-w-4xl mx-auto">
    <!-- Manage Rides Card -->
    <Card.Root
      class="group hover:shadow-lg transition-all duration-200 cursor-pointer border-2 hover:border-primary/50"
      onclick={() => (window.location.href = "/rides")}
    >
      <Card.Header>
        <div
          class="h-12 w-12 rounded-lg bg-primary/10 flex items-center justify-center mb-4 group-hover:bg-primary/20 transition-colors"
        >
          <ListOrdered class="h-6 w-6 text-primary" />
        </div>
        <Card.Title class="text-2xl">Manage Rides</Card.Title>
        <Card.Description class="text-base">
          Review pending ride submissions and publish them to the community
        </Card.Description>
      </Card.Header>
      <Card.Content>
        <Button
          variant="ghost"
          class="w-full justify-between group-hover:bg-accent"
        >
          Go to Rides
          <span class="group-hover:translate-x-1 transition-transform"
            >&rarr;</span
          >
        </Button>
      </Card.Content>
    </Card.Root>

    <!-- Quick Stats Card -->
    <Card.Root class="border-2">
      <Card.Header>
        <div
          class="h-12 w-12 rounded-lg bg-muted flex items-center justify-center mb-4"
        >
          <TrendingUp class="h-6 w-6 text-muted-foreground" />
        </div>
        <Card.Title class="text-2xl">Quick Stats</Card.Title>
        <Card.Description class="text-base">
          Platform overview and insights
        </Card.Description>
      </Card.Header>
      <Card.Content class="space-y-4">
        <div class="flex items-center gap-3">
          <MapPin class="h-5 w-5 text-muted-foreground" />
          <div>
            <p class="text-sm font-medium">Active Cities</p>
            <p class="text-2xl font-bold">2</p>
          </div>
        </div>
        <div class="pt-2 border-t">
          <p class="text-xs text-muted-foreground">
            Portland, OR • Salt Lake City, UT
          </p>
        </div>
      </Card.Content>
    </Card.Root>
  </div>

  <!-- Footer Info -->
  <div class="mt-12 md:mt-16 text-center">
    <p class="text-sm text-muted-foreground">
      Need help? Contact support or check the documentation
    </p>
  </div>
</div>
