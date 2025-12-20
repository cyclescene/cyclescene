<script lang="ts">
  import { currentRide, mapStore, ridesWithLocations } from "$lib/stores";
  import type { RideData } from "$lib/types";
  import Card from "./card.svelte";

  let { onNavigateToRide }: { onNavigateToRide: (ride: RideData) => void } = $props();

  let ride = $state<RideData | null>();

  $effect(() => {
    ride = $currentRide;
  });

  function navigateToNextRide() {
    const rides = $ridesWithLocations;
    const current = $currentRide;
    if (!current || rides.length === 0) return;

    const currentIndex = rides.findIndex((r) => r.id === current.id);
    if (currentIndex === -1) return;

    // Circular: wrap to first when reaching end
    const nextIndex = (currentIndex + 1) % rides.length;
    onNavigateToRide(rides[nextIndex]);
  }

  function navigateToPreviousRide() {
    const rides = $ridesWithLocations;
    const current = $currentRide;
    if (!current || rides.length === 0) return;

    const currentIndex = rides.findIndex((r) => r.id === current.id);
    if (currentIndex === -1) return;

    // Circular: wrap to last when going before first
    const previousIndex = currentIndex - 1 < 0 ? rides.length - 1 : currentIndex - 1;
    onNavigateToRide(rides[previousIndex]);
  }

  // Swipe detection state
  let startX = $state(0);
  let startY = $state(0);
  let currentX = $state(0);
  let isDragging = $state(false);
  let wasSwipe = $state(false);
  let cardElement = $state<HTMLElement>();

  function handlePointerDown(e: PointerEvent) {
    startX = e.clientX;
    startY = e.clientY;
    currentX = e.clientX;
    isDragging = true;
    // Don't capture yet - wait to see if user is actually dragging
  }

  function handlePointerMove(e: PointerEvent) {
    if (!isDragging) return;

    currentX = e.clientX;
    const deltaX = currentX - startX;
    const deltaY = Math.abs(e.clientY - startY);

    // Only horizontal swipes - capture pointer once we detect movement
    if (Math.abs(deltaX) > deltaY) {
      if (!cardElement?.hasPointerCapture(e.pointerId)) {
        cardElement?.setPointerCapture(e.pointerId);
      }
      e.preventDefault();
      // Apply transform with dampening (divide by 3 for resistance)
      if (cardElement) {
        cardElement.style.transform = `translateX(${deltaX / 3}px)`;
      }
    }
  }

  function handlePointerUp(e: PointerEvent) {
    if (!isDragging) return;

    const deltaX = currentX - startX;
    const swipeThreshold = 50; // minimum swipe distance in pixels

    wasSwipe = false; // Default to not a swipe

    if (Math.abs(deltaX) > swipeThreshold) {
      // Swipe detected
      wasSwipe = true;
      if (deltaX < 0) {
        // Swiped left → next ride
        navigateToNextRide();
      } else {
        // Swiped right → previous ride
        navigateToPreviousRide();
      }
      // Prevent click event from firing after swipe
      e.preventDefault();
      e.stopPropagation();
    }
    // If deltaX < swipeThreshold, wasSwipe remains false - allow click to propagate

    // Reset transform
    if (cardElement) {
      cardElement.style.transform = "";
    }

    isDragging = false;
    cardElement?.releasePointerCapture(e.pointerId);
  }
</script>

{#if $mapStore.showCurrentRide && $currentRide}
  <div
    class="location-card-container"
    bind:this={cardElement}
    onpointerdown={handlePointerDown}
    onpointermove={handlePointerMove}
    onpointerup={handlePointerUp}
    onpointercancel={handlePointerUp}
    style="touch-action: pan-y; transition: transform 0.2s ease-out;"
  >
    <Card {ride} />
  </div>
{/if}

<style>
  .location-card-container {
    position: absolute;
    bottom: 15px;
    left: 0;
    width: 100%;
    background-color: transparent;
    padding: 0 1.25rem;
    z-index: 1000;
    max-height: 20vh;
    display: flex;
    flex-direction: column;
    justify-content: flex-end;
  }
</style>
