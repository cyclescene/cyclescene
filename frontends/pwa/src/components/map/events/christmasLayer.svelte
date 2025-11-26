<script lang="ts">
  import { onMount } from "svelte";

  interface Props {
    isDarkMode: boolean;
  }

  let { isDarkMode }: Props = $props();
  let containerDiv: HTMLDivElement | undefined = $state();
  let isAnimating = $state(false);

  // Christmas emojis - snowflakes and cyclists
  const christmasEmojis = ["❄️", "❄️", "❄️", "❄️", "❄️", "❄️", "❄️", "❄️", "❄️", "🚴"]; // 90% snowflakes, 10% cyclists

  onMount(() => {
    if (!containerDiv) return;

    // Create and animate snow every 8 seconds
    const animationInterval = setInterval(() => {
      if (isAnimating) return;

      isAnimating = true;
      let activeSnowflakes = 250;

      // Create snowflakes and cyclists spread across entire map
      for (let i = 0; i < 250; i++) {
        const snowflake = document.createElement("div");
        const randomEmoji = christmasEmojis[Math.floor(Math.random() * christmasEmojis.length)];
        snowflake.textContent = randomEmoji;
        snowflake.style.position = "fixed";
        snowflake.style.left = Math.random() * window.innerWidth + "px";
        snowflake.style.top = (Math.random() * window.innerHeight * 2 - window.innerHeight) + "px";
        snowflake.style.fontSize = (8 + Math.random() * 12) + "px";
        const initialOpacity = 0.5 + Math.random() * 0.5;
        snowflake.style.opacity = initialOpacity.toString();
        snowflake.style.color = isDarkMode ? "#ffffff" : "#e0f2ff";
        snowflake.style.pointerEvents = "none";
        containerDiv!.appendChild(snowflake);

        // Animate snowflake falling
        const startYPos = parseFloat(snowflake.style.top);
        const distanceToFall = window.innerHeight - startYPos;
        const duration = 2000 + (distanceToFall / window.innerHeight) * 4000;
        const startTime = Date.now();
        const windOffset = Math.random() * 60 - 30;

        const animate = () => {
          const elapsed = Date.now() - startTime;
          const progress = elapsed / duration;

          if (progress < 1) {
            const y = progress * window.innerHeight;
            const x = parseFloat(snowflake.style.left) + Math.sin(progress * Math.PI * 4) * windOffset;

            // Fade out as it reaches the bottom
            const opacity = initialOpacity * (1 - progress);

            snowflake.style.top = y + "px";
            snowflake.style.left = x + "px";
            snowflake.style.opacity = opacity.toString();
            requestAnimationFrame(animate);
          } else {
            snowflake.remove();
            activeSnowflakes--;

            // End animation once all snowflakes have reached the bottom
            if (activeSnowflakes === 0) {
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
  style="position: absolute; top: 0; left: 0; width: 100%; height: 100%; pointer-events: none; z-index: 1; overflow: hidden;"
/>
