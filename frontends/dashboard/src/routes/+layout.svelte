<script lang="ts">
  import "../app.css";
  import DashboardHeader from "$lib/components/DashboardHeader.svelte";
  import * as Card from "$lib/components/ui/card";
  import { Button } from "$lib/components/ui/button";
  import { Input } from "$lib/components/ui/input";
  import { onMount } from "svelte";
  import type { CityCode } from "$lib/config/cities";

  const API_URL = import.meta.env.PUBLIC_API_URL || "https://api.cyclescene.cc";

  let { children } = $props();

  let selectedCity = $state<CityCode>("all");
  let adminToken = $state("");
  let currentPath = $state("/");
  let isAuthenticated = $state(false);
  let isValidating = $state(true);
  let showApiKeyForm = $state(false);
  let apiKeyInput = $state("");
  let authError = $state("");

  onMount(() => {
    const removeThemeListener = syncThemeWithSystem();

    // Track current path for active navigation
    currentPath = window.location.pathname;

    // Listen for API key change requests
    window.addEventListener("changeapikey", clearApiKey);

    // Load admin token from localStorage
    adminToken = localStorage.getItem("adminToken") || "";

    if (!adminToken) {
      showApiKeyForm = true;
      isValidating = false;
    } else {
      // Validate the token by making a test API call
      validateToken();
    }

    // Load selected city from localStorage
    const savedCity = localStorage.getItem("selectedCity") as CityCode | null;
    if (savedCity) selectedCity = savedCity;

    return () => {
      window.removeEventListener("changeapikey", clearApiKey);
      removeThemeListener();
    };
  });

  function syncThemeWithSystem() {
    const query = window.matchMedia("(prefers-color-scheme: dark)");
    const themeColor = document.querySelector<HTMLMetaElement>('meta[name="theme-color"]');

    const applyTheme = () => {
      document.documentElement.classList.toggle("dark", query.matches);
      document.documentElement.style.colorScheme = query.matches ? "dark" : "light";
      if (themeColor) {
        themeColor.content = query.matches ? "black" : "white";
      }
    };

    applyTheme();
    query.addEventListener("change", applyTheme);

    return () => {
      query.removeEventListener("change", applyTheme);
    };
  }

  async function validateToken() {
    try {
      isValidating = true;
      authError = "";

      // Validate token by calling an admin endpoint
      const response = await fetch(`${API_URL}/v1/rides/admin/pending`, {
        headers: {
          "X-Admin-Token": adminToken,
        },
      });

      if (response.ok) {
        isAuthenticated = true;
      } else if (response.status === 401 || response.status === 403) {
        // Invalid token
        clearApiKey();
        authError = "Invalid API key. Please try again.";
      } else {
        // Other error, but token might be valid
        isAuthenticated = true;
      }
    } catch (err) {
      // Network error, assume token might be valid
      isAuthenticated = true;
    } finally {
      isValidating = false;
    }
  }

  function setApiKey() {
    if (!apiKeyInput.trim()) {
      authError = "Please enter an API key";
      return;
    }
    adminToken = apiKeyInput.trim();
    localStorage.setItem("adminToken", adminToken);
    apiKeyInput = "";
    authError = "";
    validateToken();
  }

  function clearApiKey() {
    localStorage.removeItem("adminToken");
    adminToken = "";
    isAuthenticated = false;
    showApiKeyForm = true;
  }

  function handleCityChange(city: CityCode) {
    selectedCity = city;
    localStorage.setItem("selectedCity", city);
    // Dispatch custom event so child routes can react to city changes
    window.dispatchEvent(new CustomEvent("citychange", { detail: city }));
  }

  function handleChangeApiKey() {
    clearApiKey();
  }
</script>

{#if isValidating}
  <!-- Loading state while validating token -->
  <div class="min-h-screen flex items-center justify-center bg-background">
    <div class="text-center space-y-4">
      <div class="animate-spin h-8 w-8 border-4 border-primary border-t-transparent rounded-full mx-auto"></div>
      <p class="text-muted-foreground">Validating access...</p>
    </div>
  </div>
{:else if !isAuthenticated || showApiKeyForm}
  <!-- Authentication required -->
  <div class="min-h-screen flex items-center justify-center bg-background px-4">
    <div class="w-full max-w-md">
      <div class="text-center mb-8">
        <h1 class="text-3xl font-bold tracking-tight">CycleScene Dashboard</h1>
        <p class="text-muted-foreground mt-2">Admin access required</p>
      </div>

      <Card.Root>
        <Card.Header>
          <Card.Title>Enter API Key</Card.Title>
          <Card.Description>
            Your API key provides access to the admin dashboard
          </Card.Description>
        </Card.Header>
        <Card.Content class="space-y-4">
          {#if authError}
            <div class="p-3 border border-destructive bg-destructive/10 rounded-lg">
              <p class="text-sm text-destructive">{authError}</p>
            </div>
          {/if}
          <Input
            type="password"
            placeholder="Paste your API key..."
            bind:value={apiKeyInput}
            onkeydown={(e) => {
              if (e.key === "Enter") setApiKey();
            }}
          />
          <div class="flex gap-2">
            <Button onclick={setApiKey} class="flex-1">Continue</Button>
            <Button variant="outline" onclick={() => (apiKeyInput = "")}>
              Clear
            </Button>
          </div>
        </Card.Content>
      </Card.Root>
    </div>
  </div>
{:else}
  <!-- Authenticated - show dashboard -->
  <div class="min-h-screen flex flex-col">
    <DashboardHeader
      bind:selectedCity
      onCityChange={handleCityChange}
      onChangeApiKey={handleChangeApiKey}
      showApiKeyButton={!!adminToken}
      {currentPath}
    />
    <main class="flex-1">
      {@render children()}
    </main>
  </div>
{/if}
