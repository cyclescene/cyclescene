package imageoptimizer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new image optimizer client
func NewClient() *Client {
	baseURL := os.Getenv("IMAGE_OPTIMIZER_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080" // Default for local development
	}

	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// TriggerOptimization triggers image optimization asynchronously
// It does not wait for the response and logs any errors
func (c *Client) TriggerOptimization(ctx context.Context, req *OptimizeRequest) {
	go func() {
		// Create a new context with timeout for the async operation
		optimizeCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := c.optimize(optimizeCtx, req); err != nil {
			slog.Error("failed to trigger image optimization", "error", err, "entityID", req.EntityID, "entityType", req.EntityType)
		}
	}()
}

// optimize makes the actual HTTP call to the image optimizer service
func (c *Client) optimize(ctx context.Context, req *OptimizeRequest) error {
	if req.ImageUUID == "" || req.CityCode == "" || req.EntityID == "" || req.EntityType == "" {
		return fmt.Errorf("missing required fields")
	}

	// Marshal request
	reqBody, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/optimize", c.baseURL), bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	// Send request
	slog.Info("calling image optimizer", "url", c.baseURL, "entityType", req.EntityType, "entityID", req.EntityID)
	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to call optimizer: %v", err)
	}
	defer httpResp.Body.Close()

	// Check response status
	if httpResp.StatusCode != http.StatusOK {
		return fmt.Errorf("optimizer returned status %d", httpResp.StatusCode)
	}

	// Decode response
	var resp OptimizeResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}

	if !resp.Success {
		return fmt.Errorf("optimizer failed: %s", resp.Error)
	}

	slog.Info("image optimization triggered successfully", "entityID", req.EntityID, "imageURL", resp.ImageURL)
	return nil
}
