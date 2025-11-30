<script lang="ts">
  import { Label } from "$lib/components/ui/label";
  import Slider from "$lib/components/ui/slider/slider.svelte";
  import { Button } from "$lib/components/ui/button";
  import {
    Upload,
    Loader,
    CircleAlert,
    CircleCheck,
    RefreshCcw,
  } from "@lucide/svelte";
  import {
    generateSignedUploadURL,
    uploadImageToGCS,
    type SignedURLResponse,
  } from "$lib/api/client";

  interface Props {
    cityCode: string;
    onUploadComplete?: (imageUUID: string) => void;
    onUploadError?: (error: string) => void;
  }

  let { cityCode, onUploadComplete, onUploadError }: Props = $props();

  const DEFAULT_MARKER_COLOR = "#3B82F6";

  let imageFile = $state<File | null>(null);
  let imagePreview = $state<string | null>(null);
  let imageScale = $state(1.0);
  let maskScale = $state(0.8);
  let markerColor = $state(DEFAULT_MARKER_COLOR);
  let pendingMarkerColor = $state(DEFAULT_MARKER_COLOR);
  let colorChangeTimeout: ReturnType<typeof setTimeout> | null = null;
  let isUploading = $state(false);
  let error = $state<string | null>(null);
  let canvasRef = $state<HTMLCanvasElement>();
  let fileInput = $state<HTMLInputElement>();
  let generatedPreviewURL = $state<string | null>(null);
  let imageUUID = $state<string | null>(null);

  // Constants
  const RENDER_SIZE = 256; // Internal rendering resolution for quality
  const SVG_PATH =
    "M12 2C8.13 2 5 5.13 5 9c0 5.25 7 13 7 13s7-7.75 7-13c0-3.87-3.13-7-7-7zm0 9.5c-1.38 0-2.5-1.12-2.5-2.5s1.12-2.5 2.5-2.5 2.5 1.12 2.5 2.5-1.12 2.5-2.5 2.5z";

  // Re-render when image preview loads
  $effect(() => {
    if (imagePreview && canvasRef) {
      renderCanvas();
    }
  });

  // Re-render only when sliders change, NOT on color changes
  // Color is baked into the preview but doesn't require full re-render
  $effect(() => {
    // Dependency on scales to trigger re-render
    if (imageScale || maskScale) {
      if (imagePreview && canvasRef) {
        renderCanvas();
      }
    }
  });

  function handleColorChange(color: string) {
    pendingMarkerColor = color;

    // Clear existing timeout
    if (colorChangeTimeout) {
      clearTimeout(colorChangeTimeout);
    }

    // Debounce color updates by 150ms to avoid excessive re-renders
    colorChangeTimeout = setTimeout(() => {
      markerColor = color;
    }, 150);
  }

  function handleFileSelect(e: Event) {
    const input = e.target as HTMLInputElement;
    const file = input.files?.[0];

    if (!file) return;

    if (!file.type.startsWith("image/")) {
      error = "Please upload a valid image file";
      return;
    }

    if (file.size > 5 * 1024 * 1024) {
      error = "File size must be less than 5MB";
      return;
    }

    error = null;
    imageFile = file;

    const reader = new FileReader();
    reader.onload = (e) => {
      imagePreview = e.target?.result as string;
    };
    reader.readAsDataURL(file);
  }

  async function renderCanvas() {
    if (!canvasRef || !imagePreview) return;

    const ctx = canvasRef.getContext("2d");
    if (!ctx) return;

    // Clear canvas
    ctx.clearRect(0, 0, RENDER_SIZE, RENDER_SIZE);

    // Scale context to match RENDER_SIZE (24x24 coordinate system -> RENDER_SIZE)
    const scale = RENDER_SIZE / 24;

    // 1. Draw SVG Path (Marker Body)
    ctx.save();
    ctx.scale(scale, scale);
    const p = new Path2D(SVG_PATH);
    ctx.fillStyle = markerColor;
    ctx.fill(p);
    ctx.restore();

    // 2. Draw White Circle Background
    const centerX = 12 * scale;
    const centerY = 9 * scale;
    const baseRadius = 2.5 * scale;
    const currentRadius = baseRadius * maskScale;

    ctx.beginPath();
    ctx.arc(centerX, centerY, currentRadius, 0, Math.PI * 2);
    ctx.fillStyle = "white";
    ctx.fill();
    ctx.closePath();

    // 3. Draw Image (with proper async handling)
    // Load image first, then draw it clipped to the circle
    try {
      const img = await new Promise<HTMLImageElement>((resolve, reject) => {
        const image = new Image();
        image.crossOrigin = "anonymous";
        image.src = imagePreview;

        console.log("CustomMarkerBuilder: Loading image...", {
          srcLength: imagePreview.length,
          srcStart: imagePreview.substring(0, 50),
        });

        if (image.complete) {
          console.log("CustomMarkerBuilder: Image already complete", {
            width: image.width,
            height: image.height,
          });
          if (image.width > 0 && image.height > 0) {
            resolve(image);
          } else {
            reject(new Error("Image has invalid dimensions"));
          }
        } else {
          image.onload = () => {
            console.log("CustomMarkerBuilder: Image loaded", {
              width: image.width,
              height: image.height,
            });
            resolve(image);
          };
          image.onerror = (e) => {
            console.error("CustomMarkerBuilder: Image load error", e);
            reject(new Error("Failed to load image"));
          };
        }
      });

      if (img.width === 0 || img.height === 0) {
        console.error("CustomMarkerBuilder: Image has 0 dimensions");
        return;
      }

      // Now that image is loaded, clip and draw it
      ctx.save();
      ctx.beginPath();
      ctx.arc(centerX, centerY, currentRadius, 0, Math.PI * 2);
      ctx.clip();

      // Calculate aspect ratio to fit
      const aspect = img.width / img.height;
      let drawWidth, drawHeight;

      // Base size: cover the circle diameter
      const diameter = currentRadius * 2;

      if (aspect > 1) {
        drawHeight = diameter;
        drawWidth = diameter * aspect;
      } else {
        drawWidth = diameter;
        drawHeight = diameter / aspect;
      }

      // Apply image scale (zoom)
      drawWidth *= imageScale;
      drawHeight *= imageScale;

      console.log("CustomMarkerBuilder: Drawing image", {
        centerX,
        centerY,
        drawWidth,
        drawHeight,
        destX: centerX - drawWidth / 2,
        destY: centerY - drawHeight / 2,
      });

      ctx.drawImage(
        img,
        centerX - drawWidth / 2,
        centerY - drawHeight / 2,
        drawWidth,
        drawHeight,
      );

      console.log("CustomMarkerBuilder: Image drawn to canvas successfully");
      ctx.restore();
    } catch (err) {
      console.error("CustomMarkerBuilder: Error rendering canvas", err);
      console.error("CustomMarkerBuilder: Error details", {
        errorMessage: err instanceof Error ? err.message : String(err),
        errorType: err instanceof Error ? err.name : typeof err,
        stack: err instanceof Error ? err.stack : undefined,
      });
      // Canvas still shows the marker teardrop and white circle even if image fails
    }
  }

  async function handleUpload() {
    if (!canvasRef) return;

    try {
      isUploading = true;

      // Ensure canvas is fully rendered before converting to blob
      console.log("CustomMarkerBuilder: Ensuring canvas is fully rendered...");
      await renderCanvas();

      // Add a small delay to ensure canvas is fully rendered
      await new Promise(resolve => setTimeout(resolve, 100));

      console.log("CustomMarkerBuilder: Converting canvas to blob...", {
        canvasWidth: canvasRef.width,
        canvasHeight: canvasRef.height,
      });

      // Verify canvas content before saving
      const ctx = canvasRef.getContext("2d");
      if (ctx) {
        const imageData = ctx.getImageData(0, 0, canvasRef.width, canvasRef.height);
        const data = imageData.data;

        // Check if canvas has any non-transparent pixels
        let hasContent = false;
        for (let i = 3; i < data.length; i += 4) {
          if (data[i] > 0) { // Check alpha channel
            hasContent = true;
            break;
          }
        }

        console.log("CustomMarkerBuilder: Canvas content validation", {
          hasContent,
          pixelDataLength: data.length,
          markerColor,
          imagePreviewExists: !!imagePreview,
        });

        if (!hasContent) {
          throw new Error("Canvas appears to be empty. Please ensure your image and marker color are set correctly.");
        }

        // Verify marker color is applied (check for pixels with that color)
        const markerColorInt = parseInt(markerColor.slice(1), 16);
        const expectedR = (markerColorInt >> 16) & 255;
        const expectedG = (markerColorInt >> 8) & 255;
        const expectedB = markerColorInt & 255;

        console.log("CustomMarkerBuilder: Expected marker color", {
          hex: markerColor,
          r: expectedR,
          g: expectedG,
          b: expectedB,
        });
      }

      // Convert canvas to blob
      const blob = await new Promise<Blob | null>((resolve) =>
        canvasRef!.toBlob(resolve, "image/png"),
      );

      if (!blob) throw new Error("Failed to generate image");

      console.log("CustomMarkerBuilder: Blob generated", {
        size: blob.size,
        type: blob.type,
      });

      // Verify blob is not too small (indicates empty image)
      if (blob.size < 500) {
        console.warn("CustomMarkerBuilder: Blob size is suspiciously small", {
          size: blob.size,
        });
        throw new Error("Generated image appears to be empty or too small. Please check your marker settings.");
      }

      const file = new File([blob], "custom-marker.png", {
        type: "image/png",
      });

      // Generate signed URL
      const signedURLResponse: SignedURLResponse =
        await generateSignedUploadURL(
          file.name,
          file.type,
          "group", // entityType
          cityCode,
        );

      if (!signedURLResponse.success) {
        throw new Error(
          signedURLResponse.error || "Failed to generate upload URL",
        );
      }

      // Upload to GCS
      await uploadImageToGCS(signedURLResponse.signed_url, file);

      imageUUID = signedURLResponse.image_uuid;
      generatedPreviewURL = URL.createObjectURL(blob);

      onUploadComplete?.(imageUUID);
    } catch (err) {
      console.error(err);
      error = err instanceof Error ? err.message : "Upload failed";
      onUploadError?.(error);
    } finally {
      isUploading = false;
    }
  }

  function reset() {
    imageFile = null;
    imagePreview = null;
    imageUUID = null;
    generatedPreviewURL = null;
    imageScale = 1.0;
    maskScale = 0.8;
    error = null;
  }


  // Export method to auto-generate and upload marker with default settings
  export async function autoGenerateAndUploadMarker(): Promise<string | null> {
    if (!canvasRef) {
      error = "Canvas not ready";
      return null;
    }

    try {
      isUploading = true;

      // If user has uploaded an image, use the fully rendered canvas
      if (imagePreview) {
        console.log("AutoGenerateMarker: Using uploaded image");
        // First ensure the canvas is fully rendered
        await renderCanvas();

        // Add a small delay to ensure canvas is fully rendered
        await new Promise(resolve => setTimeout(resolve, 100));
      } else {
        // Create a simple marker with just the teardrop shape (no uploaded image)
        console.log("AutoGenerateMarker: Creating default marker without image");
        const ctx = canvasRef.getContext("2d");
        if (!ctx) {
          error = "Failed to create canvas context";
          return null;
        }

        // Clear canvas
        ctx.clearRect(0, 0, RENDER_SIZE, RENDER_SIZE);

        // Draw just the teardrop marker
        ctx.save();
        const scale = RENDER_SIZE / 24;
        ctx.scale(scale, scale);
        const p = new Path2D(SVG_PATH);
        ctx.fillStyle = markerColor;
        ctx.fill(p);
        ctx.restore();

        // Draw white circle in the center
        const centerX = (12 * RENDER_SIZE) / 24;
        const centerY = (9 * RENDER_SIZE) / 24;
        const baseRadius = (2.5 * RENDER_SIZE) / 24;
        ctx.beginPath();
        ctx.arc(centerX, centerY, baseRadius, 0, Math.PI * 2);
        ctx.fillStyle = "white";
        ctx.fill();
        ctx.closePath();
      }

      // Verify canvas content before saving
      const ctx = canvasRef.getContext("2d");
      if (ctx) {
        const imageData = ctx.getImageData(0, 0, canvasRef.width, canvasRef.height);
        const data = imageData.data;

        // Check if canvas has any non-transparent pixels
        let hasContent = false;
        for (let i = 3; i < data.length; i += 4) {
          if (data[i] > 0) { // Check alpha channel
            hasContent = true;
            break;
          }
        }

        console.log("AutoGenerateMarker: Canvas content validation", {
          hasContent,
          hasImage: !!imagePreview,
        });

        if (!hasContent) {
          throw new Error("Canvas appears to be empty. Please check your marker settings.");
        }
      }

      // Convert canvas to blob
      const blob = await new Promise<Blob | null>((resolve) =>
        canvasRef.toBlob(resolve, "image/png"),
      );

      if (!blob) {
        error = "Failed to generate image";
        return null;
      }

      console.log("AutoGenerateMarker: Blob generated", {
        size: blob.size,
        hasImage: !!imagePreview,
      });

      const file = new File([blob], "marker.png", {
        type: "image/png",
      });

      // Generate signed URL
      const signedURLResponse: SignedURLResponse =
        await generateSignedUploadURL(file.name, file.type, "group", cityCode);

      if (!signedURLResponse.success) {
        error = signedURLResponse.error || "Failed to generate upload URL";
        return null;
      }

      // Upload to GCS
      await uploadImageToGCS(signedURLResponse.signed_url, file);

      imageUUID = signedURLResponse.image_uuid;
      generatedPreviewURL = URL.createObjectURL(blob);
      onUploadComplete?.(imageUUID);

      return imageUUID;
    } catch (err) {
      console.error(err);
      error = err instanceof Error ? err.message : "Upload failed";
      return null;
    } finally {
      isUploading = false;
    }
  }
</script>

<div class="space-y-4">
  <div class="flex flex-col sm:flex-row gap-6">
    <!-- Preview Area -->
    <div class="flex-shrink-0 flex flex-col items-center gap-2">
      <Label>Preview</Label>
      <div
        class="relative w-32 h-32 bg-muted/30 rounded-lg border flex items-center justify-center p-4"
      >
        {#if imagePreview}
          <canvas
            bind:this={canvasRef}
            width={RENDER_SIZE}
            height={RENDER_SIZE}
            class="w-full h-full object-contain"
          ></canvas>
        {:else}
          <!-- Placeholder SVG -->
          <svg
            width="64"
            height="64"
            viewBox="0 0 24 24"
            fill="none"
            xmlns="http://www.w3.org/2000/svg"
          >
            <path
              d="M12 2C8.13 2 5 5.13 5 9c0 5.25 7 13 7 13s7-7.75 7-13c0-3.87-3.13-7-7-7zm0 9.5c-1.38 0-2.5-1.12-2.5-2.5s1.12-2.5 2.5-2.5 2.5 1.12 2.5 2.5-1.12 2.5-2.5 2.5z"
              fill={markerColor}
            />
            <circle cx="12" cy="9" r="2.5" fill="white" />
          </svg>
        {/if}
      </div>
      <p class="text-xs text-muted-foreground">Actual size on map: 64x64px</p>
    </div>

    <!-- Controls Area -->
    <div class="flex-1 space-y-4">
      {#if !imageFile}
        <div class="space-y-2">
          <Label>Upload Logo/Image</Label>
          <div
            class="border-2 border-dashed rounded-lg p-6 text-center cursor-pointer hover:bg-muted/50 transition-colors"
            role="button"
            tabindex="0"
            onclick={() => fileInput?.click()}
            onkeydown={(e) => e.key === "Enter" && fileInput?.click()}
          >
            <div class="flex flex-col items-center gap-2">
              <Upload class="h-6 w-6 text-muted-foreground" />
              <p class="text-sm font-medium">Click to upload image</p>
              <p class="text-xs text-muted-foreground">
                PNG, JPG, SVG up to 5MB
              </p>
            </div>
          </div>
          <input
            bind:this={fileInput}
            type="file"
            accept="image/*"
            class="hidden"
            onchange={handleFileSelect}
          />
        </div>
      {:else if !imageUUID}
        <!-- Editor Controls -->
        <div class="space-y-4">
          <div class="space-y-2">
            <div class="flex justify-between">
              <Label>Image Size (Zoom)</Label>
              <span class="text-xs text-muted-foreground"
                >{Math.round(imageScale * 100)}%</span
              >
            </div>
            <Slider bind:value={imageScale} min={0.5} max={3.0} step={0.1} />
          </div>

          <div class="space-y-2">
            <div class="flex justify-between">
              <Label>Mask Size</Label>
              <span class="text-xs text-muted-foreground"
                >{Math.round(maskScale * 100)}%</span
              >
            </div>
            <Slider bind:value={maskScale} min={0.5} max={2.25} step={0.05} />
          </div>

          <div class="space-y-2">
            <Label for="markerColor" class="text-sm">Marker Color</Label>
            <div class="flex items-center gap-3">
              <div class="flex-1">
                <input
                  id="markerColor"
                  type="color"
                  value={pendingMarkerColor}
                  onchange={(e) => handleColorChange((e.target as HTMLInputElement).value)}
                  oninput={(e) => pendingMarkerColor = (e.target as HTMLInputElement).value}
                  class="h-12 cursor-pointer w-full rounded border border-input"
                />
              </div>
              <div class="text-xs text-muted-foreground font-mono">
                {pendingMarkerColor.toUpperCase()}
              </div>
            </div>
          </div>

          <div class="flex gap-2 pt-2">
            <Button variant="outline" class="flex-1" onclick={reset}>
              Change Image
            </Button>
            <Button
              class="flex-1"
              onclick={handleUpload}
              disabled={isUploading}
            >
              {#if isUploading}
                <Loader class="mr-2 h-4 w-4 animate-spin" />
                Saving...
              {:else}
                Use This Marker
              {/if}
            </Button>
          </div>
        </div>
      {:else}
        <!-- Success State -->
        <div class="border rounded-lg p-4 bg-green-50 dark:bg-green-950">
          <div class="flex items-start gap-3">
            <CircleCheck
              class="h-5 w-5 text-green-600 dark:text-green-400 flex-shrink-0 mt-0.5"
            />
            <div class="flex-1">
              <p class="text-sm font-medium text-green-900 dark:text-green-100">
                Custom marker saved!
              </p>
              <p class="text-xs text-green-700 dark:text-green-300 mt-1">
                Your group will appear on the map with this custom icon.
              </p>
            </div>
            <Button
              variant="ghost"
              size="icon"
              class="text-green-600 hover:text-green-700 hover:bg-green-100"
              onclick={reset}
            >
              <RefreshCcw class="h-4 w-4" />
            </Button>
          </div>
        </div>
      {/if}

      {#if error}
        <div
          class="p-3 rounded-lg bg-destructive/10 text-destructive text-sm flex items-center gap-2"
        >
          <CircleAlert class="h-4 w-4" />
          {error}
        </div>
      {/if}
    </div>
  </div>
</div>
