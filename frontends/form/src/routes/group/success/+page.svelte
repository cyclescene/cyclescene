<script lang="ts">
  import { page } from "$app/state";
  import { Button } from "$lib/components/ui/button";
  import * as Card from "$lib/components/ui/card";
  import { Check, Mail, Edit2, Copy, MapPin, Users } from "lucide-svelte";

  const editToken = page.url.searchParams.get("token");
  const groupCode = page.url.searchParams.get("code");
  const city = page.url.searchParams.get("city");
  const editUrl = editToken && city ? `/group/edit?token=${editToken}&city=${city}` : null;
  const backUrl = city ? `https://${city}.cyclescene.cc` : "https://cyclescene.cc";

  let copied = $state(false);

  function copyCode() {
    if (groupCode) {
      navigator.clipboard.writeText(groupCode);
      copied = true;
      setTimeout(() => {
        copied = false;
      }, 2000);
    }
  }
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
          Group Registered Successfully!
        </h1>
        <p class="text-lg text-slate-600 dark:text-slate-400">
          Your cycling group is now live on CycleScene
        </p>
      </div>

      <!-- Status Badge -->
      <div class="flex justify-center sm:justify-start">
        <div class="inline-flex items-center gap-2 px-4 py-2 bg-emerald-50 dark:bg-emerald-950 border border-emerald-200 dark:border-emerald-800 rounded-full">
          <div class="h-2 w-2 rounded-full bg-emerald-500 animate-pulse"></div>
          <span class="text-sm font-medium text-emerald-700 dark:text-emerald-300">Active Group</span>
        </div>
      </div>
    </header>

    <!-- Group Code Card -->
    {#if groupCode}
      <div class="mb-6 p-6 bg-gradient-to-br from-blue-50 to-indigo-50 dark:from-blue-950/50 dark:to-indigo-950/50 border border-blue-200/60 dark:border-blue-800/60 rounded-2xl shadow-sm animate-slide-in">
        <div class="flex items-start gap-4">
          <div class="flex-shrink-0">
            <div class="h-12 w-12 rounded-full bg-blue-100 dark:bg-blue-900/50 flex items-center justify-center">
              <Users class="h-6 w-6 text-blue-600 dark:text-blue-400" />
            </div>
          </div>
          <div class="flex-1">
            <h3 class="text-base font-semibold text-blue-900 dark:text-blue-100 mb-3">
              Your Group Code
            </h3>
            <div class="flex items-center gap-2">
              <code class="text-2xl font-bold tracking-wider bg-white dark:bg-slate-900 px-5 py-3 rounded-lg border border-blue-200 dark:border-blue-800 flex-1 text-center text-blue-900 dark:text-blue-100">
                {groupCode}
              </code>
              <Button variant="outline" size="icon" onclick={copyCode} class="h-12 w-12 flex-shrink-0">
                {#if copied}
                  <Check class="h-4 w-4" />
                {:else}
                  <Copy class="h-4 w-4" />
                {/if}
              </Button>
            </div>
            <p class="text-xs text-blue-800 dark:text-blue-200/90 mt-3 leading-relaxed">
              Share this code with ride organizers to associate rides with your group
            </p>
          </div>
        </div>
      </div>
    {/if}

    <!-- Main Content Card -->
    <div class="bg-white dark:bg-slate-900 rounded-2xl border border-slate-200 dark:border-slate-800 shadow-sm overflow-hidden animate-fade-in">
      <div class="px-6 py-5 border-b border-slate-200 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-900/50">
        <h2 class="text-xl font-semibold text-slate-900 dark:text-slate-50">Next Steps</h2>
        <p class="text-sm text-slate-500 dark:text-slate-400 mt-0.5">Get started with your group</p>
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
              We've sent you a magic link to edit your group information anytime. Keep this email safe!
            </p>
          </div>
        </div>

        <!-- Edit Your Group -->
        {#if editUrl}
          <div class="flex items-start gap-4 p-4 bg-slate-50 dark:bg-slate-900/50 rounded-xl border border-slate-200 dark:border-slate-800">
            <div class="h-10 w-10 rounded-full bg-slate-100 dark:bg-slate-800 flex items-center justify-center flex-shrink-0">
              <Edit2 class="w-5 h-5 text-slate-600 dark:text-slate-400" />
            </div>
            <div class="flex-1">
              <h3 class="text-base font-semibold text-slate-900 dark:text-slate-50 mb-1">Edit Your Group</h3>
              <p class="text-sm text-slate-600 dark:text-slate-400 leading-relaxed mb-3">
                You can use this link to make changes:
              </p>
              <Button size="sm" variant="outline" class="gap-2 hover:bg-slate-100 dark:hover:bg-slate-800" href={editUrl}>
                <Edit2 class="h-3.5 w-3.5" />
                <span>Go to Edit Page</span>
              </Button>
            </div>
          </div>
        {/if}

        <!-- Next Steps -->
        <div class="p-5 bg-gradient-to-br from-slate-50 to-slate-100/50 dark:from-slate-900/50 dark:to-slate-800/50 border border-slate-200/60 dark:border-slate-800/60 rounded-xl">
          <h3 class="text-base font-semibold text-slate-900 dark:text-slate-50 mb-3">Build Your Community</h3>
          <ul class="text-sm text-slate-700 dark:text-slate-300 space-y-2.5 list-disc list-inside">
            <li>Share your group code with ride organizers</li>
            <li>When hosting rides, enter your code to associate them with your group</li>
            <li>Your group icon will appear on the map for all your rides</li>
            <li>Build your cycling community!</li>
          </ul>
        </div>
      </div>
    </div>

    <!-- Footer Actions -->
    <div class="flex justify-center pt-8">
      <Button variant="outline" class="gap-2" href={backUrl}>
        <span>← Back to CycleScene</span>
      </Button>
    </div>

  </div>
</div>
