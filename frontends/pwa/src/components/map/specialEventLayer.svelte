<script lang="ts">
  import { onMount } from "svelte";

  interface Props {
    isDarkMode: boolean;
  }

  let { isDarkMode }: Props = $props();
  let canvasContainer: HTMLDivElement | undefined = $state();
  let opacityValue = $state(0.25);

  // Rainbow colors for dark mode
  const darkRainbowColors = [
    "#ff0000", // Red
    "#ff7f00", // Orange
    "#ffff00", // Yellow
    "#00ff00", // Green
    "#0000ff", // Blue
    "#9400d3", // Purple
  ];

  // Brighter rainbow colors for light mode
  const lightRainbowColors = [
    "#ff1744", // Bright Red
    "#ff6e40", // Bright Orange
    "#ffd600", // Bright Yellow
    "#00e676", // Bright Green
    "#2979f3", // Bright Blue
    "#d500f9", // Bright Purple
  ];

  const rainbowColors = isDarkMode ? darkRainbowColors : lightRainbowColors;

  onMount(() => {
    if (!canvasContainer) return;

    const canvas = document.createElement("canvas");
    const ctx = canvas.getContext("2d");
    if (!ctx) return;

    // Set canvas size to window size
    const updateCanvasSize = () => {
      canvas.width = window.innerWidth;
      canvas.height = window.innerHeight;
      drawGradient();
    };

    const drawGradient = () => {
      // Create vertical linear gradient
      const gradient = ctx!.createLinearGradient(0, 0, 0, canvas.height);

      // Add rainbow color stops
      rainbowColors.forEach((color, index) => {
        const position = index / (rainbowColors.length - 1);
        gradient.addColorStop(position, color);
      });

      // Fill canvas with gradient
      ctx!.fillStyle = gradient;
      ctx!.fillRect(0, 0, canvas.width, canvas.height);
    };

    // Set initial size and draw
    updateCanvasSize();

    // Redraw on window resize
    window.addEventListener("resize", updateCanvasSize);

    // Append canvas to container
    canvasContainer!.appendChild(canvas);

    // Flickering animation
    let animationFrameId: number;
    const flicker = () => {
      // Random opacity between 0.15 and 0.35 for flickering effect
      opacityValue = 0.15 + Math.random() * 0.2;

      // Vary the interval (50-200ms) for irregular flickering
      const flickerInterval = 50 + Math.random() * 150;
      setTimeout(() => {
        animationFrameId = requestAnimationFrame(flicker);
      }, flickerInterval);
    };

    // Start the flickering animation
    animationFrameId = requestAnimationFrame(flicker);

    return () => {
      window.removeEventListener("resize", updateCanvasSize);
      cancelAnimationFrame(animationFrameId);
      canvas.remove();
    };
  });
</script>

<!-- Canvas container for rainbow gradient overlay -->
<div
  bind:this={canvasContainer}
  style="position: absolute; top: 0; left: 0; width: 100%; height: 100%; pointer-events: none; opacity: {opacityValue}; z-index: 1; transition: opacity 0.1s ease-in-out;"
/>
