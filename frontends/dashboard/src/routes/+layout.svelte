<script lang="ts">
  import "../app.css";
  import favicon from "$lib/assets/favicon.svg";
  import DashboardHeader from "$lib/components/DashboardHeader.svelte";
  import { onMount } from "svelte";
  import type { CityCode } from "$lib/config/cities";

  let { children } = $props();

  let selectedCity = $state<CityCode>("all");
  let adminToken = $state("");
  let currentPath = $state("/");

  onMount(() => {
    // Load selected city from localStorage
    const savedCity = localStorage.getItem("selectedCity") as CityCode | null;
    if (savedCity) {
      selectedCity = savedCity;
    }

    // Load admin token to show/hide API key button
    adminToken = localStorage.getItem("adminToken") || "";

    // Track current path for active navigation
    currentPath = window.location.pathname;
  });

  function handleCityChange(city: CityCode) {
    selectedCity = city;
    localStorage.setItem("selectedCity", city);
    // Dispatch custom event so child routes can react to city changes
    window.dispatchEvent(new CustomEvent("citychange", { detail: city }));
  }

  function handleChangeApiKey() {
    // Dispatch event for rides page to handle
    window.dispatchEvent(new CustomEvent("changeapikey"));
  }
</script>

<svelte:head>
  <link rel="icon" href={favicon} />
</svelte:head>

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
