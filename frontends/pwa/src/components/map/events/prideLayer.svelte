<script lang="ts">
  import { onMount } from "svelte";

  interface Props {
    isDarkMode: boolean;
  }

  let { isDarkMode }: Props = $props();
  let containerDiv: HTMLDivElement | undefined = $state();
  let isAnimating = $state(false);

  // Pride emojis - mix of flags and cyclists
  const prideEmojis = [
    "🏳️‍🌈", "🏳️‍🌈", "🏳️‍🌈", "🏳️‍🌈", "🏳️‍🌈", // 50% pride flags
    "🏳️‍⚧️", "🏳️‍⚧️", "🏳️‍⚧️", // 30% trans flags
    "🚴", "🚴" // 20% cyclists
  ];

  onMount(() => {
    if (!containerDiv) return;

    // Create and animate flags every 8 seconds
    const animationInterval = setInterval(() => {
      if (isAnimating) return;

      isAnimating = true;
      let activeFlagCount = 250;

      // Create flags spread across entire map
      for (let i = 0; i < 250; i++) {
        const flagDiv = document.createElement("div");
        const randomEmoji = prideEmojis[Math.floor(Math.random() * prideEmojis.length)];
        flagDiv.textContent = randomEmoji;
        flagDiv.style.position = "fixed";
        flagDiv.style.left = Math.random() * window.innerWidth + "px";
        flagDiv.style.top = (Math.random() * window.innerHeight * 3 - window.innerHeight * 1.5) + "px";
        flagDiv.style.fontSize = (20 + Math.random() * 20) + "px";
        const initialOpacity = 0.6 + Math.random() * 0.4;
        flagDiv.style.opacity = initialOpacity.toString();
        flagDiv.style.pointerEvents = "none";
        containerDiv!.appendChild(flagDiv);

        // Animate flag falling
        const startYPos = parseFloat(flagDiv.style.top);
        const distanceToFall = window.innerHeight - startYPos;
        const duration = 5000 + (distanceToFall / window.innerHeight) * 8000;
        const startTime = Date.now();
        const windOffset = Math.random() * 60 - 30;

        const animate = () => {
          const elapsed = Date.now() - startTime;
          const progress = elapsed / duration;

          if (progress < 1) {
            const y = progress * window.innerHeight;
            const x = parseFloat(flagDiv.style.left) + Math.sin(progress * Math.PI * 4) * windOffset;

            // Fade out as it reaches the bottom
            const opacity = initialOpacity * (1 - progress);

            flagDiv.style.top = y + "px";
            flagDiv.style.left = x + "px";
            flagDiv.style.opacity = opacity.toString();
            requestAnimationFrame(animate);
          } else {
            flagDiv.remove();
            activeFlagCount--;

            // End animation once all flags have reached the bottom
            if (activeFlagCount === 0) {
              isAnimating = false;
            }
          }
        };

        requestAnimationFrame(animate);
      }
    }, 30000); // Run every 30 seconds

    return () => clearInterval(animationInterval);
  });
</script>

<div
  bind:this={containerDiv}
  style="position: fixed; top: 0; left: 0; width: 100%; height: 100%; pointer-events: none; z-index: 1; overflow: hidden;"
/>
