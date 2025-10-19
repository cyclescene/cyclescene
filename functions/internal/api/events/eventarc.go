package events

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/compute/metadata"
	"google.golang.org/api/idtoken"
)

type EventarcClient struct {
	optimizerURL string
	httpClient   *http.Client
}

// CloudEvent represents a CloudEvents format event
type CloudEvent struct {
	SpecVersion string            `json:"specversion"`
	Type        string            `json:"type"`
	Source      string            `json:"source"`
	Subject     string            `json:"subject,omitempty"`
	ID          string            `json:"id"`
	Time        string            `json:"time"`
	DataContent string            `json:"datacontenttype"`
	Data        map[string]string `json:"data"`
}

// NewEventarcClient creates a new Eventarc client
func NewEventarcClient() *EventarcClient {
	optimizerURL := os.Getenv("IMAGE_OPTIMIZER_URL")
	if optimizerURL == "" {
		optimizerURL = "http://localhost:8080" // Default for local development
	}

	httpClient := createAuthenticatedHTTPClient(optimizerURL)

	return &EventarcClient{
		optimizerURL: optimizerURL,
		httpClient:   httpClient,
	}
}

// createAuthenticatedHTTPClient creates an HTTP client with Cloud Run authentication
func createAuthenticatedHTTPClient(targetURL string) *http.Client {
	// For local development, use unauthenticated client
	if targetURL == "http://localhost:8080" {
		return &http.Client{
			Timeout: 30 * time.Second,
		}
	}

	// For Cloud Run, use identity token authentication
	return &http.Client{
		Transport: &idTokenRoundTripper{
			targetURL: targetURL,
		},
		Timeout: 30 * time.Second,
	}
}

// idTokenRoundTripper adds identity token to requests for service-to-service authentication
type idTokenRoundTripper struct {
	targetURL string
}

func (rt *idTokenRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	ctx := context.Background()

	// Check if we're running in a Cloud environment by trying to get project ID
	_, err := metadata.ProjectID()
	if err != nil {
		slog.Warn("not running on GCP, making unauthenticated request to image optimizer")
		return http.DefaultTransport.RoundTrip(req)
	}

	// Get an identity token for the target service
	audience := rt.targetURL
	token, err := idtoken.NewClient(ctx, audience)
	if err != nil {
		slog.Error("failed to create identity token client", "error", err)
		return http.DefaultTransport.RoundTrip(req)
	}

	// Add the identity token to the request
	resp, err := token.Transport.RoundTrip(req)
	if err != nil {
		slog.Error("failed to make authenticated request", "error", err)
		return nil, err
	}

	return resp, nil
}

// TriggerOptimization sends an optimization event to the image optimizer
func (ec *EventarcClient) TriggerOptimization(ctx context.Context, event *ImageOptimizationEvent) error {
	if event.ImageUUID == "" || event.CityCode == "" || event.EntityID == "" || event.EntityType == "" {
		return fmt.Errorf("missing required fields in optimization event")
	}

	// Create CloudEvent
	cloudEvent := CloudEvent{
		SpecVersion: "1.0",
		Type:        "com.cyclescene.image.optimization",
		Source:      "cyclescene-api",
		Subject:     fmt.Sprintf("image/%s/%s", event.EntityType, event.EntityID),
		ID:          event.ImageUUID,
		Time:        time.Now().UTC().Format(time.RFC3339),
		DataContent: "application/json",
		Data: map[string]string{
			"imageUUID":  event.ImageUUID,
			"cityCode":   event.CityCode,
			"entityID":   event.EntityID,
			"entityType": event.EntityType,
		},
	}

	// Marshal to JSON
	data, err := json.Marshal(cloudEvent)
	if err != nil {
		return fmt.Errorf("failed to marshal CloudEvent: %v", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/optimize", ec.optimizerURL), bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/cloudevents+json")
	req.Header.Set("ce-specversion", "1.0")
	req.Header.Set("ce-type", "com.cyclescene.image.optimization")
	req.Header.Set("ce-source", "cyclescene-api")
	req.Header.Set("ce-id", event.ImageUUID)
	req.Header.Set("ce-time", time.Now().UTC().Format(time.RFC3339))

	// Send request
	slog.Info("triggering image optimization via Eventarc", "url", ec.optimizerURL, "imageUUID", event.ImageUUID, "entityType", event.EntityType, "entityID", event.EntityID)
	resp, err := ec.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send optimization event: %v", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("optimizer returned status %d", resp.StatusCode)
	}

	slog.Info("optimization event triggered successfully", "imageUUID", event.ImageUUID)
	return nil
}
