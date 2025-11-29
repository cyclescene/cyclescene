/**
 * Marker utilities for loading and extracting group markers from spritesheets
 *
 * This module handles:
 * 1. Fetching spritesheet PNG and metadata JSON from GCS
 * 2. Parsing metadata to understand marker positions
 * 3. Extracting individual 64x64 marker images from the spritesheet using Canvas
 * 4. Creating image objects for use in MapLibre icon layers
 */

interface MarkerInfo {
  x: number
  y: number
  width: number
  height: number
}

interface SpritesheetMetadata {
  markers: Record<string, MarkerInfo>
}

/**
 * Fetches the spritesheet PNG and metadata JSON for a city
 * @param cityCode - The city code (e.g., "pdx", "sea")
 * @returns An object with the spritesheet image and metadata
 */
export async function loadSpritesheet(cityCode: string): Promise<{
  image: HTMLImageElement
  metadata: SpritesheetMetadata
}> {
  const baseUrl = `https://storage.googleapis.com/cyclescene-479119-optimized-images`
  const pngUrl = `${baseUrl}/sprites/${cityCode}/markers.png`
  const jsonUrl = `${baseUrl}/sprites/${cityCode}/markers.json`

  try {
    // Fetch metadata JSON
    const metadataResponse = await fetch(jsonUrl)
    if (!metadataResponse.ok) {
      throw new Error(`Failed to fetch spritesheet metadata: ${metadataResponse.status}`)
    }
    const metadata = (await metadataResponse.json()) as SpritesheetMetadata

    // Load spritesheet image
    const image = new Image()
    image.crossOrigin = "anonymous"

    return new Promise((resolve, reject) => {
      image.onload = () => {
        resolve({ image, metadata })
      }
      image.onerror = () => {
        reject(new Error(`Failed to load spritesheet image from ${pngUrl}`))
      }
      image.src = pngUrl
    })
  } catch (error) {
    console.error(`Error loading spritesheet for city ${cityCode}:`, error)
    throw error
  }
}

/**
 * Extracts a single marker from the spritesheet using Canvas
 * @param spritesheetImage - The spritesheet image
 * @param markerInfo - The marker's position and dimensions in the spritesheet
 * @returns A data URL representing the extracted marker
 */
export function extractMarkerFromSpritesheet(
  spritesheetImage: HTMLImageElement,
  markerInfo: MarkerInfo
): string {
  const canvas = document.createElement("canvas")
  canvas.width = markerInfo.width
  canvas.height = markerInfo.height

  const ctx = canvas.getContext("2d")
  if (!ctx) {
    throw new Error("Failed to get canvas 2D context")
  }

  // Draw the marker region from the spritesheet onto the canvas
  ctx.drawImage(
    spritesheetImage,
    markerInfo.x,
    markerInfo.y,
    markerInfo.width,
    markerInfo.height,
    0,
    0,
    markerInfo.width,
    markerInfo.height
  )

  return canvas.toDataURL("image/png")
}

/**
 * Loads all markers for a city and returns a map of marker key -> data URL
 * @param cityCode - The city code
 * @returns A map of marker keys to extracted marker image data URLs
 */
export async function loadAllMarkersForCity(
  cityCode: string
): Promise<Record<string, string>> {
  try {
    const { image, metadata } = await loadSpritesheet(cityCode)

    const markers: Record<string, string> = {}

    for (const [markerKey, markerInfo] of Object.entries(metadata.markers)) {
      try {
        const markerDataUrl = extractMarkerFromSpritesheet(image, markerInfo)
        markers[markerKey] = markerDataUrl
      } catch (error) {
        console.error(`Failed to extract marker ${markerKey}:`, error)
      }
    }

    return markers
  } catch (error) {
    console.error(`Error loading markers for city ${cityCode}:`, error)
    throw error
  }
}

/**
 * Loads a specific marker by its key from the spritesheet
 * @param cityCode - The city code
 * @param markerKey - The marker key (e.g., "dino-riders-pdx")
 * @returns The extracted marker image data URL
 */
export async function loadMarkerByKey(
  cityCode: string,
  markerKey: string
): Promise<string> {
  try {
    const { image, metadata } = await loadSpritesheet(cityCode)

    const markerInfo = metadata.markers[markerKey]
    if (!markerInfo) {
      throw new Error(`Marker ${markerKey} not found in spritesheet metadata`)
    }

    return extractMarkerFromSpritesheet(image, markerInfo)
  } catch (error) {
    console.error(`Error loading marker ${markerKey} for city ${cityCode}:`, error)
    throw error
  }
}
