package imageoptimizer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"
)

type Service struct {
	baseURL    string
	httpClient *http.Client
}

func NewService() *Service {
	baseURL := os.Getenv("IMAGE_OPTIMIZER_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080" // Default for local development
	}

	return &Service{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 5 * time.Minute, // 5 minute timeout for image processing
		},
	}
}

// Optimize sends a synchronous optimization request to the image optimizer service
// This blocks until the optimization is complete or times out
func (s *Service) Optimize(ctx context.Context, req *OptimizeRequest) (*OptimizeResponse, error) {
	if err := validateRequest(req); err != nil {
		return nil, err
	}

	resp, err := s.sendOptimizeRequest(ctx, req)
	if err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf("optimizer failed: %s", resp.Error)
	}

	slog.Info("image optimization successful", "entityID", req.EntityID, "imageURL", resp.ImageURL)
	return resp, nil
}

// OptimizeAsync sends an asynchronous optimization request
// Does not wait for the response; useful for non-blocking optimization
func (s *Service) OptimizeAsync(req *OptimizeRequest) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		if _, err := s.Optimize(ctx, req); err != nil {
			slog.Error("async image optimization failed", "error", err, "entityID", req.EntityID, "entityType", req.EntityType)
		}
	}()
}

// sendOptimizeRequest handles the HTTP communication with the optimizer service
func (s *Service) sendOptimizeRequest(ctx context.Context, req *OptimizeRequest) (*OptimizeResponse, error) {
	// Marshal request
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/optimize", s.baseURL), bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	// Send request
	slog.Debug("calling image optimizer", "url", s.baseURL, "entityType", req.EntityType, "entityID", req.EntityID)
	httpResp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to call optimizer: %v", err)
	}
	defer httpResp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("optimizer returned status %d: %s", httpResp.StatusCode, string(respBody))
	}

	// Unmarshal response
	var resp OptimizeResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %v", err)
	}

	return &resp, nil
}

// validateRequest validates the optimization request
func validateRequest(req *OptimizeRequest) error {
	if req == nil {
		return fmt.Errorf("request cannot be nil")
	}

	if req.ImageUUID == "" || req.CityCode == "" || req.EntityID == "" || req.EntityType == "" {
		return fmt.Errorf("missing required fields: imageUUID, cityCode, entityID, entityType")
	}

	if req.EntityType != "ride" && req.EntityType != "group" {
		return fmt.Errorf("entityType must be 'ride' or 'group', got '%s'", req.EntityType)
	}

	return nil
}
