package imageoptimizer

// OptimizeRequest represents a request to optimize an image
type OptimizeRequest struct {
	ImageUUID  string `json:"imageUUID"`
	CityCode   string `json:"cityCode"`
	EntityID   string `json:"entityID"`
	EntityType string `json:"entityType"` // "ride" or "group"
}

// OptimizeResponse represents the response from the optimizer service
type OptimizeResponse struct {
	Success  bool   `json:"success"`
	ImageURL string `json:"imageURL"`
	Error    string `json:"error,omitempty"`
}
