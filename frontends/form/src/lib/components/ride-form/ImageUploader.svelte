<script lang="ts">
  import { Label } from "$lib/components/ui/label";
  import { Upload, X, Loader, CircleAlert, CircleCheck } from "@lucide/svelte";
  import {
    generateSignedUploadURL,
    uploadImageToGCS,
    type SignedURLResponse,
  } from "$lib/api/client";

  interface Props {
    label?: string;
    entityType: string;
    cityCode: string;
    description?: string;
    onUploadComplete?: (imageUUID: string) => void;
    onUploadError?: (error: string) => void;
    maxSizeMB?: number;
    acceptedTypes?: string[];
  }

  let {
    label = "Upload Image",
    entityType,
    cityCode,
    description,
    onUploadComplete,
    onUploadError,
    maxSizeMB = 10,
    acceptedTypes = ["image/jpeg", "image/png", "image/webp", "image/gif"],
  }: Props = $props();

  let imageUUID = $state<string | null>(null);
  let previewURL = $state<string | null>(null);
  let isUploading = $state(false);
  let error = $state<string | null>(null);
  let uploadProgress = $state(0);
  let fileInput = $state<HTMLInputElement>();

  const maxSizeBytes = maxSizeMB * 1024 * 1024;

  function getMimeType(file: File): string {
    return file.type || "application/octet-stream";
  }

  function validateFile(file: File): string | null {
    if (!acceptedTypes.includes(file.type)) {
      return `Invalid file type. Accepted types: ${acceptedTypes.join(", ")}`;
    }

    if (file.size > maxSizeBytes) {
      return `File size exceeds ${maxSizeMB}MB limit`;
    }

    return null;
  }

  async function handleFileSelect(e: Event) {
    const input = e.target as HTMLInputElement;
    const file = input.files?.[0];

    if (!file) return;

    const validationError = validateFile(file);
    if (validationError) {
      error = validationError;
      onUploadError?.(validationError);
      return;
    }

    error = null;
    imageUUID = null;
    uploadProgress = 0;

    try {
      isUploading = true;

      // Create preview
      const reader = new FileReader();
      reader.onload = (e) => {
        previewURL = e.target?.result as string;
      };
      reader.readAsDataURL(file);

      // Generate signed URL
      const signedURLResponse: SignedURLResponse =
        await generateSignedUploadURL(
          file.name,
          getMimeType(file),
          entityType,
          cityCode,
        );

      if (!signedURLResponse.success) {
        throw new Error(
          signedURLResponse.error || "Failed to generate upload URL",
        );
      }

      // Upload to GCS
      await uploadImageToGCS(signedURLResponse.signed_url, file);

      // Store UUID and call callback
      imageUUID = signedURLResponse.image_uuid;
      uploadProgress = 100;
      onUploadComplete?.(imageUUID);
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : "Upload failed";
      error = errorMessage;
      onUploadError?.(errorMessage);
      previewURL = null;
      imageUUID = null;
    } finally {
      isUploading = false;
      if (fileInput) {
        fileInput.value = "";
      }
    }
  }

  function clearUpload() {
    imageUUID = null;
    previewURL = null;
    error = null;
    uploadProgress = 0;
    if (fileInput) {
      fileInput.value = "";
    }
  }

  function triggerFileInput() {
    fileInput?.click();
  }
</script>

<div class="space-y-3">
  <div>
    <Label class="text-sm sm:text-base">{label}</Label>
    {#if description}
      <p class="text-xs sm:text-sm text-muted-foreground mt-1">{description}</p>
    {/if}
  </div>

  {#if !imageUUID}
    <div
      class="border-2 border-dashed rounded-lg p-6 sm:p-8 text-center cursor-pointer transition-colors hover:bg-muted/50"
      role="button"
      tabindex="0"
      onkeydown={(e) => e.key === "Enter" && triggerFileInput()}
      onclick={triggerFileInput}
      ondrop={(e) => {
        e.preventDefault();
        const file = e.dataTransfer?.files?.[0];
        if (file) {
          const input = fileInput as HTMLInputElement;
          const dataTransfer = new DataTransfer();
          dataTransfer.items.add(file);
          input.files = dataTransfer.files;
          const event = new Event("change", { bubbles: true });
          input.dispatchEvent(event);
        }
      }}
      ondragover={(e) => {
        e.preventDefault();
        e.currentTarget.classList.add("bg-muted");
      }}
      ondragleave={(e) => {
        e.currentTarget.classList.remove("bg-muted");
      }}
    >
      <div class="flex flex-col items-center gap-2">
        {#if isUploading}
          <Loader class="h-8 w-8 animate-spin text-muted-foreground" />
          <p class="text-sm font-medium">
            Uploading... {uploadProgress}%
          </p>
        {:else}
          <Upload class="h-8 w-8 text-muted-foreground" />
          <div>
            <p class="text-sm font-medium">Click to upload or drag and drop</p>
            <p class="text-xs text-muted-foreground mt-1">
              {acceptedTypes
                .map((t) => t.split("/")[1])
                .join(", ")
                .toUpperCase()} up to {maxSizeMB}MB
            </p>
          </div>
        {/if}
      </div>
    </div>
  {:else}
    <div class="border rounded-lg p-4 bg-green-50 dark:bg-green-950">
      <div class="flex items-start gap-3">
        <CircleCheck
          class="h-5 w-5 text-green-600 dark:text-green-400 flex-shrink-0 mt-0.5"
        />
        <div class="flex-1 min-w-0">
          <p class="text-sm font-medium text-green-900 dark:text-green-100">
            Image uploaded successfully
          </p>
          <p class="text-xs text-green-700 dark:text-green-300 mt-1 break-all">
            UUID: {imageUUID}
          </p>
        </div>
        <button
          onclick={clearUpload}
          class="flex-shrink-0 text-green-600 dark:text-green-400 hover:text-green-700 dark:hover:text-green-300"
          aria-label="Remove image"
        >
          <X class="h-4 w-4" />
        </button>
      </div>

      {#if previewURL}
        <div class="mt-3">
          <img
            src={previewURL}
            alt="Uploaded image preview"
            class="max-h-40 rounded object-cover"
          />
        </div>
      {/if}
    </div>
  {/if}

  {#if error}
    <div
      class="border border-destructive bg-destructive/10 rounded-lg p-3 flex items-start gap-2"
    >
      <CircleAlert class="h-4 w-4 text-destructive flex-shrink-0 mt-0.5" />
      <p class="text-xs sm:text-sm text-destructive">{error}</p>
    </div>
  {/if}

  <input
    bind:this={fileInput}
    type="file"
    accept={acceptedTypes.join(",")}
    onchange={handleFileSelect}
    class="hidden"
    aria-label="File input for image upload"
  />
</div>
