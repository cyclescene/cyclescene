<script lang="ts">
  import { Button } from "$lib/components/ui/button";
</script>

<div class="flex flex-col min-h-screen">
  <div class="container max-w-3xl mx-auto py-12 px-4">
    <div class="space-y-6">
      <div>
        <h1 class="text-3xl font-bold mb-1">Privacy Policy</h1>
        <p class="text-sm text-muted-foreground">Last updated: November 2025</p>
      </div>

      <section>
        <h2 class="text-xl font-semibold mb-2">Analytics & Tracking</h2>
        <p class="text-sm text-muted-foreground mb-2">
          We collect basic analytics data to understand how people use Cycle Scene and to measure the effectiveness of our marketing campaigns (such as QR code stickers).
        </p>
        <p class="text-xs font-semibold mb-1">Data we collect:</p>
        <ul class="list-disc list-inside space-y-0.5 text-sm text-muted-foreground">
          <li>Page visits and timestamps</li>
          <li>Marketing source (from URL parameters)</li>
          <li>Which city you click on</li>
          <li>Browser information (user-agent, accept-language)</li>
        </ul>
      </section>

      <section>
        <h2 class="text-xl font-semibold mb-2">How We Use This Data</h2>
        <ul class="list-disc list-inside space-y-0.5 text-sm text-muted-foreground">
          <li>Track which marketing campaigns drive people to our app</li>
          <li>Understand which cities are most popular</li>
          <li>Improve the user experience</li>
        </ul>
      </section>

      <section>
        <h2 class="text-xl font-semibold mb-2">Data Storage</h2>
        <p class="text-sm text-muted-foreground">
          Analytics data is stored in a separate database and is not shared with third parties. We retain this data indefinitely for trend analysis.
        </p>
      </section>

      <section>
        <h2 class="text-xl font-semibold mb-2">Contact</h2>
        <p class="text-sm text-muted-foreground">
          If you have questions about our privacy practices, please reach out through our main website or GitHub repository.
        </p>
      </section>

      <div class="pt-2">
        <Button onclick={() => (window.location.href = "/")} variant="outline" size="sm">
          Back to Cycle Scene
        </Button>
      </div>
    </div>
  </div>
</div>
