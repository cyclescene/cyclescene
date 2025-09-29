<script>
  import { ScrollArea } from "$lib/components/ui/scroll-area";
</script>

<!-- ============================================== -->
<!--          1. PRIVACY POLICY START               -->
<!-- ============================================== -->
<div
  class="p-6 sm:p-8 max-w-4xl mx-auto space-y-12 absolute top-[60px] bottom-[75px] min-h-[calc(100vh_-_115px)]"
>
  <ScrollArea>
    <section class="space-y-6 h-[calc(100vh_-_160px)] flex flex-col gap-2">
      <h1 class="text-3xl font-bold tracking-tight text-foreground">
        CycleScene Privacy Policy
      </h1>
      <p class="text-sm text-muted-foreground">
        Last Updated: September 27, 2025
      </p>

      <h2 class="text-xl font-semibold mt-6">1. Introduction</h2>
      <p>
        This Privacy Policy explains how CycleScene (the "Application")
        collects, uses, and protects information when you use our Progressive
        Web Application (PWA). CycleScene is committed to protecting your
        privacy. We do not collect, store, or transmit any personally
        identifiable information (PII) about our users.
      </p>

      <h2 class="text-xl font-semibold mt-6">
        2. Information We Do NOT Collect
      </h2>
      <p>We want to be clear about what we do not do:</p>
      <ul class="list-disc list-inside space-y-1 ml-4 text-sm">
        <li>
          We <strong>do not</strong> require or collect personal data such as your
          name, email address, or phone number.
        </li>
        <li>
          We <strong>do not</strong> use cookies for tracking or advertising purposes.
        </li>
        <li>
          We <strong>do not</strong> track your usage across other websites or applications.
        </li>
      </ul>

      <h2 class="text-xl font-semibold mt-6">3. Local Data Storage and Use</h2>
      <p>
        To provide a fast, reliable, and offline-capable experience, we store
        strictly non-personal data directly on your device. This data remains
        entirely on your local device and is never transmitted to our servers or
        any third party.
      </p>

      <div class="overflow-x-auto mt-4">
        <table class="w-full text-sm border-collapse">
          <thead>
            <tr class="bg-muted text-muted-foreground">
              <th class="p-3 text-left border-b">Data Type</th>
              <th class="p-3 text-left border-b">What We Store</th>
              <th class="p-3 text-left border-b">Purpose</th>
            </tr>
          </thead>
          <tbody>
            <tr class="hover:bg-accent">
              <td class="p-3 border-b">App Configuration</td>
              <td class="p-3 border-b"
                >Your preferred theme setting (Light/Dark/System).</td
              >
              <td class="p-3 border-b"
                >To remember your visual preferences between sessions.</td
              >
            </tr>
            <tr class="hover:bg-accent">
              <td class="p-3 border-b">Offline Cache</td>
              <td class="p-3 border-b"
                >Map tile images, application code, and static assets.</td
              >
              <td class="p-3 border-b"
                >To allow the app to function when you are offline.</td
              >
            </tr>
            <tr class="hover:bg-accent">
              <td class="p-3 border-b">Application State</td>
              <td class="p-3 border-b"
                >Ride data loaded from remote sources (IndexedDB).</td
              >
              <td class="p-3 border-b"
                >To provide offline access to ride information.</td
              >
            </tr>
          </tbody>
        </table>
      </div>

      <h2 class="text-xl font-semibold mt-6">
        4. Third-Party Services (Map Tiles)
      </h2>
      <p>
        When you load the map, your device's <strong>IP address</strong> is temporarily
        transmitted to the map tile server (e.g., OpenStreetMap) to receive the necessary
        map images. We do not control the data practices of these providers. Please
        refer to their respective privacy policies for details.
      </p>

      <h2 class="text-xl font-semibold mt-6">5. Contact Us</h2>
      <p>
        If you have any questions about this Privacy Policy, please contact us
        at: <strong>[YOUR EMAIL ADDRESS OR SUPPORT LINK]</strong>
      </p>
    </section>
  </ScrollArea>
</div>
