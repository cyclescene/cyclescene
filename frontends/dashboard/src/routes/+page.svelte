<script lang="ts">
  import { Button } from "$lib/components/ui/button";
</script>

<div class="container max-w-4xl mx-auto py-16 px-4">
  <div class="text-center space-y-6">
    <h1 class="text-5xl font-bold tracking-tight">Dashboard</h1>
    <p class="text-lg text-muted-foreground">Manage your cycle scene</p>

    <div class="pt-8">
      <Button onclick={() => window.location.href = '/rides'} size="lg">
        Manage Rides
      </Button>
    </div>
  </div>
</div>
