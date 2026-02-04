<script lang="ts">
  import { Button } from "$lib/components/ui/button";
</script>

<div class="flex flex-col min-h-screen">
  <div class="container max-w-3xl mx-auto py-12 px-4">
    <div class="space-y-6">
      <div>
        <h1 class="text-3xl font-bold mb-1">Privacy Policy</h1>
        <p class="text-sm text-muted-foreground">Last updated: February 2026</p>
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
        <h2 class="text-xl font-semibold mb-2">Strava Integration</h2>
        <p class="text-sm text-muted-foreground mb-2">
          When a ride organizer connects their Strava account to CycleScene, we access the following data through the Strava API:
        </p>
        <ul class="list-disc list-inside space-y-0.5 text-sm text-muted-foreground">
          <li><strong>Club information:</strong> Names and details of Strava clubs where the organizer holds an admin or owner role</li>
          <li><strong>Club events:</strong> Upcoming event details (name, date, time, description) from those clubs</li>
        </ul>
        <p class="text-sm text-muted-foreground mt-2">
          We do not access individual activity data, GPS data, performance metrics, or personal health information from Strava.
        </p>
      </section>

      <section>
        <h2 class="text-xl font-semibold mb-2">How We Use Strava Data</h2>
        <p class="text-sm text-muted-foreground">
          Strava club event data is used solely to display upcoming group rides on CycleScene so that local cyclists can discover them. Strava data is not sold, shared with third parties, used for advertising, or used for AI/machine learning purposes.
        </p>
      </section>

      <section>
        <h2 class="text-xl font-semibold mb-2">Data Retention and Refresh</h2>
        <p class="text-sm text-muted-foreground">
          Strava event data is refreshed at least every 7 days. If an event is deleted or modified on Strava, those changes are reflected on CycleScene within 48 hours. Strava data is not cached beyond 7 days without a refresh from the Strava API.
        </p>
      </section>

      <section>
        <h2 class="text-xl font-semibold mb-2">Strava Usage Data</h2>
        <p class="text-sm text-muted-foreground">
          Strava may monitor and collect usage data related to our use of the Strava API, as described in the Strava API Agreement. For more information, see <a href="https://www.strava.com/legal/privacy" target="_blank" rel="noopener noreferrer" class="text-orange-500 hover:underline">Strava's Privacy Policy</a>.
        </p>
      </section>

      <section>
        <h2 class="text-xl font-semibold mb-2">Revoking Strava Access</h2>
        <p class="text-sm text-muted-foreground mb-2">
          You may disconnect your Strava account from CycleScene at any time through your account settings. Upon disconnection, we will delete all Strava data associated with your account.
        </p>
        <p class="text-sm text-muted-foreground">
          You may also revoke access directly from your Strava settings at <a href="https://www.strava.com/settings/apps" target="_blank" rel="noopener noreferrer" class="text-orange-500 hover:underline">strava.com/settings/apps</a>.
        </p>
      </section>

      <section>
        <h2 class="text-xl font-semibold mb-2">Requesting Data Deletion</h2>
        <p class="text-sm text-muted-foreground">
          You may request deletion of your Strava data by contacting us at <a href="mailto:info@cyclescene.cc" class="text-orange-500 hover:underline">info@cyclescene.cc</a>. We will process deletion requests promptly and confirm when your data has been removed.
        </p>
      </section>

      <section>
        <h2 class="text-xl font-semibold mb-2">Contact</h2>
        <p class="text-sm text-muted-foreground">
          If you have questions about our privacy practices, please contact us at <a href="mailto:info@cyclescene.cc" class="text-orange-500 hover:underline">info@cyclescene.cc</a>.
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
