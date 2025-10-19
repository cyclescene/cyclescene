package events

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"time"

	publishing "cloud.google.com/go/eventarc/publishing/apiv1"
	publishingpb "cloud.google.com/go/eventarc/publishing/apiv1/publishingpb"
	"github.com/google/uuid"
)

type EventarcClient struct {
	projectID     string
	region        string
	channelName   string
	publishClient *publishing.PublisherClient
}

// CloudEvent represents a CloudEvents v1.0 format event
type CloudEvent struct {
	SpecVersion     string         `json:"specversion"`
	Type            string         `json:"type"`
	Source          string         `json:"source"`
	ID              string         `json:"id"`
	Time            string         `json:"time,omitempty"`
	DataContentType string         `json:"datacontenttype,omitempty"`
	Data            map[string]any `json:"data"`
}

// NewEventarcClient creates a new Eventarc client
func NewEventarcClient() *EventarcClient {
	projectID := os.Getenv("GCP_PROJECT")
	if projectID == "" {
		projectID = "cyclescene"
	}

	region := os.Getenv("GCP_REGION")
	if region == "" {
		region = "us-west1"
	}

	channelName := os.Getenv("EVENTARC_CHANNEL_NAME")
	if channelName == "" {
		channelName = "image-optimization-events"
	}

	// Create Eventarc publishing client
	ctx := context.Background()
	client, err := publishing.NewPublisherClient(ctx)
	if err != nil {
		slog.Warn("failed to create Eventarc client, events will not be published", "error", err)
		return &EventarcClient{
			projectID:     projectID,
			region:        region,
			channelName:   channelName,
			publishClient: nil,
		}
	}

	return &EventarcClient{
		projectID:     projectID,
		region:        region,
		channelName:   channelName,
		publishClient: client,
	}
}

// TriggerOptimization publishes an optimization event to Eventarc
func (ec *EventarcClient) TriggerOptimization(ctx context.Context, event *ImageOptimizationEvent) error {
	if event.ImageUUID == "" || event.CityCode == "" || event.EntityID == "" || event.EntityType == "" {
		return fmt.Errorf("missing required fields in optimization event")
	}

	// If client is not initialized, log and return (dev environment without GCP)
	if ec.publishClient == nil {
		slog.Warn("Eventarc client not available, skipping event publication", "imageUUID", event.ImageUUID)
		return nil
	}

	// Create CloudEvent in v1.0 format
	cloudEvent := CloudEvent{
		SpecVersion:     "1.0",
		Type:            "com.cyclescene.image.optimization",
		Source:          "//cyclescene.com/api",
		ID:              uuid.New().String(),
		Time:            time.Now().UTC().Format(time.RFC3339),
		DataContentType: "application/json",
		Data: map[string]any{
			"imageUUID":  event.ImageUUID,
			"cityCode":   event.CityCode,
			"entityID":   event.EntityID,
			"entityType": event.EntityType,
		},
	}

	// Marshal CloudEvent to JSON
	eventJSON, err := json.Marshal(cloudEvent)
	if err != nil {
		return fmt.Errorf("failed to marshal CloudEvent: %v", err)
	}

	// Build the channel path
	channelPath := fmt.Sprintf("projects/%s/locations/%s/channels/%s",
		ec.projectID, ec.region, ec.channelName)

	// Create the PublishEventsRequest with TextEvents (JSON format)
	req := &publishingpb.PublishEventsRequest{
		Channel:    channelPath,
		TextEvents: []string{string(eventJSON)},
	}

	// Publish the event
	slog.Info("publishing image optimization event to Eventarc",
		"channel", ec.channelName,
		"imageUUID", event.ImageUUID,
		"entityType", event.EntityType,
		"entityID", event.EntityID)

	_, err = ec.publishClient.PublishEvents(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to publish event to Eventarc: %v", err)
	}

	slog.Info("optimization event published successfully",
		"imageUUID", event.ImageUUID)

	return nil
}
