<script lang="ts">
  import { page } from "$app/state";
  import { Button } from "$lib/components/ui/button";
  import * as Card from "$lib/components/ui/card";
  import { Check, Mail, Edit2, MapPin } from "lucide-svelte";

  const editToken = page.url.searchParams.get("token");
  const eventId = page.url.searchParams.get("event_id");
  const city = page.url.searchParams.get("city");
  const editUrl = editToken ? `/rides/edit?token=${editToken}` : null;

  // Map city code to PWA domain
  const getCycleSceneDomain = (cityCode: string | null): string => {
    if (!cityCode) return 'https://cyclescene.cc';
    const cityDomains: Record<string, string> = {
      pdx: 'https://pdx.cyclescene.cc',
      slc: 'https://slc.cyclescene.cc',
    };
    return cityDomains[cityCode.toLowerCase()] || 'https://cyclescene.cc';
  };
</script>

<style>
  @keyframes slideIn {
    from {
      opacity: 0;
      transform: translateY(-8px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }

  @keyframes fadeIn {
    from { opacity: 0; }
    to { opacity: 1; }
  }

  .animate-slide-in {
    animation: slideIn 0.3s cubic-bezier(0.16, 1, 0.3, 1);
  }

  .animate-fade-in {
    animation: fadeIn 0.2s ease-out;
  }
</style>

<div class="min-h-screen bg-gradient-to-b from-slate-50 to-white dark:from-slate-950 dark:to-slate-900">
  <div class="max-w-5xl mx-auto px-4 sm:px-6 lg:px-8 py-8 sm:py-12">

    <!-- Header -->
    <header class="mb-10 animate-slide-in">
      <div class="inline-flex items-center gap-2 text-sm font-medium text-slate-500 mb-3">
        <MapPin class="h-4 w-4" />
        <span class="uppercase tracking-wide">{city?.toUpperCase() || 'CycleScene'}</span>
      </div>

      <div class="text-center sm:text-left mb-8">
        <div class="inline-flex items-center justify-center w-16 h-16 rounded-full bg-emerald-50 dark:bg-emerald-950 border border-emerald-200 dark:border-emerald-800 mb-6">
          <Check class="w-8 h-8 text-emerald-600 dark:text-emerald-400" />
        </div>
        <h1 class="text-4xl sm:text-5xl font-bold tracking-tight text-slate-900 dark:text-slate-50 mb-4 leading-tight">
          Ride Submitted Successfully!
        </h1>
        <p class="text-lg text-slate-600 dark:text-slate-400">
          Your ride is now pending review
        </p>
      </div>

      <!-- Status Badge -->
      <div class="flex justify-center sm:justify-start">
        <div class="inline-flex items-center gap-2 px-4 py-2 bg-blue-50 dark:bg-blue-950 border border-blue-200 dark:border-blue-800 rounded-full">
          <div class="h-2 w-2 rounded-full bg-blue-500 animate-pulse"></div>
          <span class="text-sm font-medium text-blue-700 dark:text-blue-300">Under Review</span>
        </div>
      </div>
    </header>

    <!-- Main Content Card -->
    <div class="bg-white dark:bg-slate-900 rounded-2xl border border-slate-200 dark:border-slate-800 shadow-sm overflow-hidden animate-fade-in">
      <div class="px-6 py-5 border-b border-slate-200 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-900/50">
        <h2 class="text-xl font-semibold text-slate-900 dark:text-slate-50">Next Steps</h2>
        <p class="text-sm text-slate-500 dark:text-slate-400 mt-0.5">What to expect after submission</p>
      </div>

      <div class="px-6 py-6 space-y-6">
        <!-- Check Email -->
        <div class="flex items-start gap-4 p-4 bg-slate-50 dark:bg-slate-900/50 rounded-xl border border-slate-200 dark:border-slate-800">
          <div class="h-10 w-10 rounded-full bg-slate-100 dark:bg-slate-800 flex items-center justify-center flex-shrink-0">
            <Mail class="w-5 h-5 text-slate-600 dark:text-slate-400" />
          </div>
          <div class="flex-1">
            <h3 class="text-base font-semibold text-slate-900 dark:text-slate-50 mb-1">Check Your Email</h3>
            <p class="text-sm text-slate-600 dark:text-slate-400 leading-relaxed">
              We've sent you a magic link to edit your ride anytime. Keep this email safe!
            </p>
          </div>
        </div>

        <!-- Edit Your Ride -->
        {#if editUrl}
          <div class="flex items-start gap-4 p-4 bg-slate-50 dark:bg-slate-900/50 rounded-xl border border-slate-200 dark:border-slate-800">
            <div class="h-10 w-10 rounded-full bg-slate-100 dark:bg-slate-800 flex items-center justify-center flex-shrink-0">
              <Edit2 class="w-5 h-5 text-slate-600 dark:text-slate-400" />
            </div>
            <div class="flex-1">
              <h3 class="text-base font-semibold text-slate-900 dark:text-slate-50 mb-1">Edit Your Ride</h3>
              <p class="text-sm text-slate-600 dark:text-slate-400 leading-relaxed mb-3">
                You can use this link to make changes:
              </p>
              <Button size="sm" variant="outline" class="gap-2 hover:bg-slate-100 dark:hover:bg-slate-800">
                <a href={editUrl} class="flex items-center gap-2">
                  <Edit2 class="h-3.5 w-3.5" />
                  <span>Go to Edit Page</span>
                </a>
              </Button>
            </div>
          </div>
        {/if}

        <!-- What Happens Next -->
        <div class="p-5 bg-gradient-to-br from-slate-50 to-slate-100/50 dark:from-slate-900/50 dark:to-slate-800/50 border border-slate-200/60 dark:border-slate-800/60 rounded-xl">
          <h3 class="text-base font-semibold text-slate-900 dark:text-slate-50 mb-3">What happens next?</h3>
          <ol class="text-sm text-slate-700 dark:text-slate-300 space-y-2.5 list-decimal list-inside">
            <li>Our team will review your submission</li>
            <li>You'll receive an email once it's approved</li>
            <li>Your ride will appear on CycleScene</li>
            <li>Riders can find and join your ride!</li>
          </ol>
        </div>

        {#if eventId}
          <p class="text-xs text-center text-slate-400 dark:text-slate-500 pt-2">
            Event ID: {eventId}
          </p>
        {/if}
      </div>
    </div>

    <!-- Footer Actions -->
    <div class="flex justify-center pt-8">
      <Button variant="outline" class="gap-2">
        <a href={getCycleSceneDomain(city)} class="flex items-center gap-2">
          <span>← Back to {city ? city.toUpperCase() : 'CycleScene'}</span>
        </a>
      </Button>
    </div>

  </div>
</div>
