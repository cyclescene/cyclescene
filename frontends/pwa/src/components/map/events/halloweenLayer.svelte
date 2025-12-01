<script lang="ts">
  import { onMount } from "svelte";

  interface Props {
    isDarkMode: boolean;
  }

  let { isDarkMode }: Props = $props();
  let containerDiv: HTMLDivElement | undefined = $state();
  let isAnimating = $state(false);

  // Halloween emojis - bats and cyclists
  const halloweenEmojis = [
    "🦇",
    "🦇",
    "🦇",
    "🦇",
    "🦇",
    "🦇",
    "🦇",
    "🦇",
    "🚴",
  ]; // 89% bats, 11% cyclists

  onMount(() => {
    if (!containerDiv) return;

    // Create and animate bats every 5 seconds
    const animationInterval = setInterval(() => {
      if (isAnimating) return;

      isAnimating = true;

      // Create 100 bats flying diagonally corner to corner
      for (let i = 0; i < 100; i++) {
        const bat = document.createElement("div");
        const randomEmoji =
          halloweenEmojis[Math.floor(Math.random() * halloweenEmojis.length)];
        bat.textContent = randomEmoji;
        bat.style.fontSize = "40px";

        // All start from top right, spread horizontally
        const offset = (i - 49) * 40; // i goes 0-99, so offset goes -1960 to 1960
        const startX = window.innerWidth + offset;
        const startY = 0;
        const endX = -100;
        const endY = window.innerHeight;

        console.log(
          `Bat ${i} - Start: (${startX}, ${startY}), End: (${endX}, ${endY})`,
        );

        bat.style.position = "fixed";
        bat.style.left = startX + "px";
        bat.style.top = startY + "px";
        containerDiv!.appendChild(bat);

        // Animate bat flying from top-right to bottom-left
        const duration = 4000 + Math.random() * 2000;
        const startTime = Date.now();

        const animate = () => {
          const elapsed = Date.now() - startTime;
          const progress = elapsed / duration;

          if (progress < 1) {
            const x = startX + progress * (endX - startX);
            const y = startY + progress * (endY - startY);
            console.log(
              `Bat progress: ${(progress * 100).toFixed(1)}% - X: ${x.toFixed(0)}, Y: ${y.toFixed(0)}`,
            );
            bat.style.left = x + "px";
            bat.style.top = y + "px";
            requestAnimationFrame(animate);
          } else {
            console.log("Bat animation complete");
            bat.remove();
          }
        };

        requestAnimationFrame(animate);
      }

      setTimeout(() => {
        isAnimating = false;
      }, 6000);
    }, 30000); // Run every 30 seconds

    return () => clearInterval(animationInterval);
  });
</script>

<div
  bind:this={containerDiv}
  style="position: fixed; top: 0; left: 0; width: 100%; height: 100%; pointer-events: none; z-index: 1; overflow: hidden;"
></div>
