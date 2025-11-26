<script lang="ts">
  import { onMount } from "svelte";

  let canvas: HTMLCanvasElement;
  let animationId: number;

  const emojis = ["🚴", "🗺️", "🏙️", "👥", "🌍", "🚲", "🛣️", "🎯"];
  const emojiSize = 48;
  const speed = 2;
  const animationDuration = 30000; // 30 seconds

  interface EmojiObject {
    emoji: string;
    x: number;
    y: number;
  }

  let emojiObjects: EmojiObject[] = [];

  function initializeEmojis(width: number, height: number) {
    emojiObjects = [];
    // Space emojis evenly across the width
    const spacing = width / emojis.length;

    for (let i = 0; i < emojis.length; i++) {
      emojiObjects.push({
        emoji: emojis[i],
        x: i * spacing - emojiSize,
        y: height / 2 - emojiSize / 2
      });
    }
  }

  function animate() {
    if (!canvas) return;

    const ctx = canvas.getContext("2d");
    if (!ctx) return;

    const width = canvas.width;
    const height = canvas.height;

    // Clear canvas
    ctx.clearRect(0, 0, width, height);

    // Draw emojis
    ctx.font = `${emojiSize}px Arial`;
    ctx.textAlign = "left";
    ctx.textBaseline = "top";

    emojiObjects.forEach((obj) => {
      ctx.fillText(obj.emoji, obj.x, obj.y);

      // Move emoji to the right
      obj.x += speed;

      // Reset position when emoji goes off screen
      if (obj.x > width) {
        obj.x = -emojiSize;
      }
    });

    animationId = requestAnimationFrame(animate);
  }

  function handleResize() {
    if (!canvas) return;
    canvas.width = window.innerWidth;
    canvas.height = emojiSize + 40;
    initializeEmojis(canvas.width, canvas.height);
  }

  onMount(() => {
    if (!canvas) return;

    handleResize();
    window.addEventListener("resize", handleResize);

    // Start animation
    animate();

    return () => {
      window.removeEventListener("resize", handleResize);
      cancelAnimationFrame(animationId);
    };
  });
</script>

<canvas bind:this={canvas} class="w-full bg-gradient-to-r from-transparent via-slate-100 to-transparent dark:via-slate-900" />

<style>
  :global(canvas) {
    display: block;
  }
</style>
