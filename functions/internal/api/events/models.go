package events

// ImageOptimizationEvent represents an event to trigger image optimization
type ImageOptimizationEvent struct {
	ImageUUID   string `json:"imageUUID"`
	CityCode    string `json:"cityCode"`
	EntityID    string `json:"entityID"`
	EntityType  string `json:"entityType"`   // "ride" or "group"
	MarkerColor string `json:"markerColor"` // Hex color for group markers (e.g., "#FF5733")
}
