<script lang="ts">
  import "../app.css";
  import favicon from "$lib/assets/favicon.svg";

  let { children } = $props();
</script>

<svelte:head>
  <link rel="icon" href={favicon} />
</svelte:head>

<div class="flex flex-col min-h-screen">
  <div class="flex-1">
    {@render children()}
  </div>

  <footer class="border-t bg-muted/30 py-4">
    <div class="container max-w-4xl mx-auto px-4">
      <div class="flex items-center justify-between text-xs text-muted-foreground">
        <p>We use analytics to improve your experience and measure campaign effectiveness.</p>
        <a href="/privacy" class="underline hover:text-foreground transition-colors">
          Privacy Policy
        </a>
      </div>
    </div>
  </footer>
</div>
