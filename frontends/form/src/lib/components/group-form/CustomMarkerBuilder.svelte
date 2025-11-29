<script lang="ts">
    import { Label } from "$lib/components/ui/label";
    import Slider from "$lib/components/ui/slider/slider.svelte";
    import { Button } from "$lib/components/ui/button";
    import {
        Upload,
        X,
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
    import { onMount } from "svelte";

    interface Props {
        cityCode: string;
        onUploadComplete?: (imageUUID: string) => void;
        onUploadError?: (error: string) => void;
    }

    let { cityCode, onUploadComplete, onUploadError }: Props =
        $props();

    const DEFAULT_MARKER_COLOR = "#3B82F6";

    let imageFile = $state<File | null>(null);
    let imagePreview = $state<string | null>(null);
    let imageScale = $state(1.0);
    let maskScale = $state(0.8);
    let isUploading = $state(false);
    let error = $state<string | null>(null);
    let canvasRef = $state<HTMLCanvasElement>();
    let fileInput = $state<HTMLInputElement>();
    let generatedPreviewURL = $state<string | null>(null);
    let imageUUID = $state<string | null>(null);

    // Constants
    const CANVAS_SIZE = 64; // Final output size
    const RENDER_SIZE = 256; // Internal rendering resolution for quality
    const SVG_PATH =
        "M12 2C8.13 2 5 5.13 5 9c0 5.25 7 13 7 13s7-7.75 7-13c0-3.87-3.13-7-7-7zm0 9.5c-1.38 0-2.5-1.12-2.5-2.5s1.12-2.5 2.5-2.5 2.5 1.12 2.5 2.5-1.12 2.5-2.5 2.5z";

    // Derived state for canvas updates
    $effect(() => {
        if (imagePreview && canvasRef) {
            renderCanvas();
        }
    });

    // Re-render when sliders change
    $effect(() => {
        // Dependency on scales to trigger re-render
        if (imageScale || maskScale) {
            if (imagePreview && canvasRef) {
                renderCanvas();
            }
        }
    });

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
        ctx.fillStyle = DEFAULT_MARKER_COLOR;
        ctx.fill(p);
        ctx.restore();

        // 2. Draw White Circle Background
        // The circle in the SVG is cx="12" cy="9" r="2.5"
        // We want to allow this "window" to be larger based on maskScale
        // Default r=2.5 is quite small for an image, let's say base is slightly larger for visibility
        // But we must respect the user's request: "size of the circle mask to change fo a increased mark marker border size"
        // Wait, if the mask size changes, does the white circle size change?
        // The user said: "size of the image to change and the size of the circle mask to change fo a increased mark marker border size"
        // This implies the white circle is the "border" around the image.
        // So we draw a white circle, and then draw the image masked by a slightly smaller circle?
        // Or just draw the image masked by a circle.
        // Let's interpret:
        // - The "hole" in the marker is at 12, 9.
        // - We draw a white circle at 12, 9.
        // - We draw the image on top, masked by a circle at 12, 9.

        const centerX = 12 * scale;
        const centerY = 9 * scale;
        // Base radius from SVG is 2.5. Let's make the "window" adjustable.
        // Actually, the SVG path has a hole or is it solid? The path provided is solid with a hole cut out?
        // "M12 2... ...z" - it's a single path.
        // The SVG provided in prompt has a separate <circle> element:
        // <circle cx="12" cy="9" r="2.5" fill="white" fillOpacity="0.9" />
        // So we should draw this white circle.

        const baseRadius = 2.5 * scale;
        // We'll use maskScale to adjust the size of this white circle area (and thus the image area)
        // If maskScale is large, the white circle is large.
        // But the marker body is fixed. If we make it too big it will overlap the marker borders.
        // Let's allow maskScale to go from 0.5 to 1.5 roughly.

        const currentRadius = baseRadius * maskScale;

        ctx.beginPath();
        ctx.arc(centerX, centerY, currentRadius, 0, Math.PI * 2);
        ctx.fillStyle = "white";
        ctx.fill();
        ctx.closePath();

        // 3. Draw Image
        // We want to mask the image to be inside the circle.
        // But wait, if we want a "border", the image should be slightly smaller than the white circle?
        // The prompt says: "size of the circle mask to change fo a increased mark marker border size"
        // This implies the white circle is the background, and the image is on top.
        // If the image is smaller than the white circle, we see a white border.
        // So we need TWO sizes? Or just one mask size and one image size?
        // "one that will allow foir the size of the image to change and the size of the circle mask to change"
        // Okay.
        // Circle Mask Size -> controls the clipping area of the image? Or the white background?
        // Let's assume:
        // - White Circle is drawn at `currentRadius`.
        // - Image is drawn masked by `currentRadius`.
        // - `imageScale` controls the zoom level of the image INSIDE that mask.

        // Let's try this:
        // Clip to the circle
        ctx.save();
        ctx.beginPath();
        ctx.arc(centerX, centerY, currentRadius, 0, Math.PI * 2);
        ctx.clip();

        // Draw image centered at centerX, centerY
        const img = new Image();
        img.src = imagePreview;

        console.log("CustomMarkerBuilder: Loading image...", {
            srcLength: imagePreview.length,
        });

        await new Promise((resolve) => {
            if (img.complete) {
                console.log("CustomMarkerBuilder: Image already complete");
                resolve(null);
            } else {
                img.onload = () => {
                    console.log("CustomMarkerBuilder: Image loaded", {
                        width: img.width,
                        height: img.height,
                    });
                    resolve(null);
                };
                img.onerror = (e) => {
                    console.error("CustomMarkerBuilder: Image load error", e);
                    resolve(null); // Resolve anyway to avoid hanging
                };
            }
        });

        if (img.width === 0 || img.height === 0) {
            console.error("CustomMarkerBuilder: Image has 0 dimensions");
            ctx.restore();
            return;
        }

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

        ctx.restore();
    }

    async function handleUpload() {
        if (!canvasRef) return;

        try {
            isUploading = true;

            // Convert canvas to blob
            const blob = await new Promise<Blob | null>((resolve) =>
                canvasRef!.toBlob(resolve, "image/png"),
            );

            if (!blob) throw new Error("Failed to generate image");

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
        // Create a default marker with default image (placeholder blue teardrop)
        if (!canvasRef) {
            error = "Canvas not ready";
            return null;
        }

        try {
            isUploading = true;

            // Create a simple SVG canvas with just the teardrop shape
            const svgCanvas = document.createElement("canvas");
            svgCanvas.width = RENDER_SIZE;
            svgCanvas.height = RENDER_SIZE;
            const ctx = svgCanvas.getContext("2d");
            if (!ctx) {
                error = "Failed to create canvas context";
                return null;
            }

            // Draw just the teardrop marker in default blue
            ctx.save();
            const scale = RENDER_SIZE / 24;
            ctx.scale(scale, scale);
            const p = new Path2D(SVG_PATH);
            ctx.fillStyle = DEFAULT_MARKER_COLOR;
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

            // Convert canvas to blob
            const blob = await new Promise<Blob | null>((resolve) =>
                svgCanvas.toBlob(resolve, "image/png"),
            );

            if (!blob) {
                error = "Failed to generate image";
                return null;
            }

            const file = new File([blob], "default-marker.png", {
                type: "image/png",
            });

            // Generate signed URL
            const signedURLResponse: SignedURLResponse =
                await generateSignedUploadURL(
                    file.name,
                    file.type,
                    "group",
                    cityCode,
                );

            if (!signedURLResponse.success) {
                error =
                    signedURLResponse.error || "Failed to generate upload URL";
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
                            fill={DEFAULT_MARKER_COLOR}
                        />
                        <circle cx="12" cy="9" r="2.5" fill="white" />
                    </svg>
                {/if}
            </div>
            <p class="text-xs text-muted-foreground">
                Actual size on map: 64x64px
            </p>
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
                        onkeydown={(e) =>
                            e.key === "Enter" && fileInput?.click()}
                    >
                        <div class="flex flex-col items-center gap-2">
                            <Upload class="h-6 w-6 text-muted-foreground" />
                            <p class="text-sm font-medium">
                                Click to upload image
                            </p>
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
                        <Slider
                            bind:value={imageScale}
                            min={0.5}
                            max={3.0}
                            step={0.1}
                        />
                    </div>

                    <div class="space-y-2">
                        <div class="flex justify-between">
                            <Label>Mask Size</Label>
                            <span class="text-xs text-muted-foreground"
                                >{Math.round(maskScale * 100)}%</span
                            >
                        </div>
                        <Slider
                            bind:value={maskScale}
                            min={0.5}
                            max={2.25}
                            step={0.05}
                        />
                    </div>

                    <div class="flex gap-2 pt-2">
                        <Button
                            variant="outline"
                            class="flex-1"
                            onclick={reset}
                        >
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
                <div
                    class="border rounded-lg p-4 bg-green-50 dark:bg-green-950"
                >
                    <div class="flex items-start gap-3">
                        <CircleCheck
                            class="h-5 w-5 text-green-600 dark:text-green-400 flex-shrink-0 mt-0.5"
                        />
                        <div class="flex-1">
                            <p
                                class="text-sm font-medium text-green-900 dark:text-green-100"
                            >
                                Custom marker saved!
                            </p>
                            <p
                                class="text-xs text-green-700 dark:text-green-300 mt-1"
                            >
                                Your group will appear on the map with this
                                custom icon.
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
