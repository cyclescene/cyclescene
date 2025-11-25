package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/spacesedan/cyclescene/functions/internal/imageprocessing"
)

type OptimizeRequest struct {
	ImageUUID  string `json:"imageUUID"`
	CityCode   string `json:"cityCode"`
	EntityID   string `json:"entityID"`
	EntityType string `json:"entityType"` // "ride" or "group"
}

type OptimizeResponse struct {
	Success  bool   `json:"success"`
	ImageURL string `json:"imageURL"`
	Error    string `json:"error,omitempty"`
}

type HealthResponse struct {
	Status string `json:"status"`
}

func setupRoutes(router *chi.Mux, dbConnector *imageprocessing.DBConnector) {
	router.Get("/health", handleHealth)
	router.Post("/optimize", handleOptimize(dbConnector))
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(HealthResponse{Status: "ok"}); err != nil {
		slog.Error("failed to encode health response", "error", err)
	}
}

func handleOptimize(dbConnector *imageprocessing.DBConnector) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var req OptimizeRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			slog.Error("failed to decode request", "error", err)
			w.WriteHeader(http.StatusBadRequest)
			if err := json.NewEncoder(w).Encode(OptimizeResponse{
				Success: false,
				Error:   "invalid request body",
			}); err != nil {
				slog.Error("failed to encode error response", "error", err)
			}
			return
		}

		// Validate request
		if req.ImageUUID == "" || req.CityCode == "" || req.EntityID == "" || req.EntityType == "" {
			slog.Error("missing required fields", "request", req)
			w.WriteHeader(http.StatusBadRequest)
			if err := json.NewEncoder(w).Encode(OptimizeResponse{
				Success: false,
				Error:   "missing required fields: imageUUID, cityCode, entityID, entityType",
			}); err != nil {
				slog.Error("failed to encode error response", "error", err)
			}
			return
		}

		if req.EntityType != "ride" && req.EntityType != "group" {
			slog.Error("invalid entityType", "entityType", req.EntityType)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(OptimizeResponse{
				Success: false,
				Error:   "entityType must be 'ride' or 'group'",
			})
			return
		}

		slog.Info("processing image", "imageUUID", req.ImageUUID, "cityCode", req.CityCode, "entityID", req.EntityID, "entityType", req.EntityType)

		// Process image
		imageURL, err := processImageOptimization(r.Context(), req)
		if err != nil {
			slog.Error("failed to process image", "error", err, "imageUUID", req.ImageUUID)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(OptimizeResponse{
				Success: false,
				Error:   "failed to process image: " + err.Error(),
			})
			return
		}

		// Update database
		if err := dbConnector.UpdateImageURL(req.EntityType, req.EntityID, imageURL); err != nil {
			slog.Error("failed to update database", "error", err, "entityID", req.EntityID)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(OptimizeResponse{
				Success: false,
				Error:   "image processed but failed to update database: " + err.Error(),
			})
			return
		}

		slog.Info("successfully optimized image", "imageUUID", req.ImageUUID, "imageURL", imageURL)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(OptimizeResponse{
			Success:  true,
			ImageURL: imageURL,
		})
	}
}

func processImageOptimization(ctx context.Context, req OptimizeRequest) (string, error) {
	stagingBucket := os.Getenv("STAGING_BUCKET")
	optimizedBucket := os.Getenv("OPTIMIZED_BUCKET")

	if stagingBucket == "" || optimizedBucket == "" {
		return "", fmt.Errorf("missing required environment variables: STAGING_BUCKET or OPTIMIZED_BUCKET not set")
	}

	processor, err := imageprocessing.NewImageProcessor(ctx, stagingBucket, optimizedBucket)
	if err != nil {
		return "", err
	}
	defer processor.Close()

	// Get the object path from image processing
	objectPath, err := processor.ProcessImage(ctx, req.ImageUUID, req.CityCode, req.EntityID, req.EntityType)
	if err != nil {
		return "", err
	}

	// Generate a signed URL for the optimized image
	signedURL, err := processor.GenerateSignedURL(ctx, objectPath)
	if err != nil {
		return "", fmt.Errorf("failed to generate signed URL for optimized image: %v", err)
	}

	return signedURL, nil
}
