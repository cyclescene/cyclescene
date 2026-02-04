<script lang="ts">
  import * as Sheet from "$lib/components/ui/sheet";
  import { Button } from "$lib/components/ui/button";
  import { Separator } from "$lib/components/ui/separator";
  import CitySelector from "./CitySelector.svelte";
  import { Menu, Home, ListOrdered, Key } from "@lucide/svelte";
  import type { CityCode } from "$lib/config/cities";

  interface Props {
    selectedCity: CityCode;
    onCityChange: (city: CityCode) => void;
    onChangeApiKey?: () => void;
    showApiKeyButton?: boolean;
    currentPath?: string;
  }

  let {
    selectedCity = $bindable("all"),
    onCityChange,
    onChangeApiKey,
    showApiKeyButton = false,
    currentPath = "/",
  }: Props = $props();

  let mobileMenuOpen = $state(false);

  function navigate(path: string) {
    window.location.href = path;
    mobileMenuOpen = false;
  }
</script>

<header class="sticky top-0 z-50 w-full border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
  <div class="container max-w-6xl mx-auto">
    <div class="flex h-16 items-center justify-between px-4">
      <!-- Logo / Title -->
      <div class="flex items-center gap-6">
        <a href="/" class="flex items-center gap-2 font-bold text-lg">
          CycleScene
        </a>

        <!-- Desktop Navigation -->
        <nav class="hidden md:flex items-center gap-1">
          <Button
            variant={currentPath === "/" ? "default" : "ghost"}
            size="sm"
            onclick={() => navigate("/")}
          >
            <Home class="h-4 w-4 mr-2" />
            Home
          </Button>
          <Button
            variant={currentPath === "/rides" ? "default" : "ghost"}
            size="sm"
            onclick={() => navigate("/rides")}
          >
            <ListOrdered class="h-4 w-4 mr-2" />
            Rides
          </Button>
        </nav>
      </div>

      <!-- Desktop Actions -->
      <div class="hidden md:flex items-center gap-3">
        <CitySelector
          bind:value={selectedCity}
          onValueChange={onCityChange}
          variant="desktop"
        />
        {#if showApiKeyButton && onChangeApiKey}
          <Button variant="outline" onclick={onChangeApiKey} size="sm">
            <Key class="h-4 w-4 mr-2" />
            Change API Key
          </Button>
        {/if}
      </div>

      <!-- Mobile Menu Button -->
      <div class="flex md:hidden items-center gap-2">
        <CitySelector
          bind:value={selectedCity}
          onValueChange={onCityChange}
          variant="mobile"
        />
        <Sheet.Root bind:open={mobileMenuOpen}>
          <Sheet.Trigger>
            {#snippet child({ props })}
              <Button {...props} variant="outline" size="icon">
                <Menu class="h-5 w-5" />
              </Button>
            {/snippet}
          </Sheet.Trigger>
          <Sheet.Content side="right">
            <Sheet.Header>
              <Sheet.Title>Menu</Sheet.Title>
            </Sheet.Header>
            <div class="mt-6 space-y-1">
              <Button
                variant={currentPath === "/" ? "default" : "ghost"}
                class="w-full justify-start"
                onclick={() => navigate("/")}
              >
                <Home class="h-4 w-4 mr-2" />
                Home
              </Button>
              <Button
                variant={currentPath === "/rides" ? "default" : "ghost"}
                class="w-full justify-start"
                onclick={() => navigate("/rides")}
              >
                <ListOrdered class="h-4 w-4 mr-2" />
                Rides
              </Button>
              {#if showApiKeyButton && onChangeApiKey}
                <Separator class="my-4" />
                <Button
                  variant="outline"
                  class="w-full justify-start"
                  onclick={() => {
                    onChangeApiKey?.();
                    mobileMenuOpen = false;
                  }}
                >
                  <Key class="h-4 w-4 mr-2" />
                  Change API Key
                </Button>
              {/if}
            </div>
          </Sheet.Content>
        </Sheet.Root>
      </div>
    </div>
  </div>
</header>
