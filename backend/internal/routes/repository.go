package routes

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"
)

// Route represents a route in the database
type Route struct {
	ID          string `json:"id"`
	Source      string `json:"source"`
	SourceID    string `json:"source_id"`
	SourceURL   string `json:"source_url"`
	GeoJSON     string `json:"geojson"`
	DistanceKm  *float64 `json:"distance_km"`
	DistanceMi  *float64 `json:"distance_mi"`
	CreatedAt   string `json:"created_at"`
}

// Repository handles database operations for routes
type Repository struct {
	db *sql.DB
}

// NewRepository creates a new Repository instance
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// CreateRoute stores a new route in the database
// Returns the route ID and handles UNIQUE constraint violations by returning existing route ID
func (r *Repository) CreateRoute(ctx context.Context, source, sourceID, sourceURL, city string, feature GeoJSONFeature, distanceKm, distanceMi float64) (string, error) {
	// Check if route already exists
	existing, err := r.GetRouteBySourceID(ctx, source, sourceID)
	if err == nil && existing != nil {
		slog.Info("route already exists", "source", source, "sourceID", sourceID, "routeID", existing.ID)
		return existing.ID, nil
	}

	// Encode full GeoJSON Feature to string (including geometry and properties)
	geoJSONBytes, err := json.Marshal(feature)
	if err != nil {
		return "", fmt.Errorf("failed to marshal GeoJSON: %w", err)
	}
	geoJSONStr := string(geoJSONBytes)

	// Generate UUID for route ID
	routeID := generateID()

	query := `
		INSERT INTO routes (id, source, source_id, source_url, geojson, distance_km, distance_mi, city)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = r.db.ExecContext(ctx, query, routeID, source, sourceID, sourceURL, geoJSONStr, distanceKm, distanceMi, city)
	if err != nil {
		// If unique constraint violation, fetch and return existing
		if isUniqueConstraintError(err) {
			existing, err := r.GetRouteBySourceID(ctx, source, sourceID)
			if err != nil {
				return "", fmt.Errorf("failed to fetch existing route: %w", err)
			}
			if existing != nil {
				slog.Info("route created by another process, using existing", "source", source, "sourceID", sourceID, "routeID", existing.ID)
				return existing.ID, nil
			}
		}
		return "", fmt.Errorf("failed to create route: %w", err)
	}

	slog.Info("route created", "source", source, "sourceID", sourceID, "routeID", routeID, "distanceKm", distanceKm)
	return routeID, nil
}

// GetRouteBySourceID retrieves an existing route by source and source ID
func (r *Repository) GetRouteBySourceID(ctx context.Context, source, sourceID string) (*Route, error) {
	query := `
		SELECT id, source, source_id, source_url, geojson, distance_km, distance_mi, created_at
		FROM routes
		WHERE source = ? AND source_id = ?
	`

	var route Route
	err := r.db.QueryRowContext(ctx, query, source, sourceID).Scan(
		&route.ID,
		&route.Source,
		&route.SourceID,
		&route.SourceURL,
		&route.GeoJSON,
		&route.DistanceKm,
		&route.DistanceMi,
		&route.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to query route: %w", err)
	}

	return &route, nil
}

// GetRouteByID retrieves a route by its ID
func (r *Repository) GetRouteByID(ctx context.Context, id string) (*Route, error) {
	query := `
		SELECT id, source, source_id, source_url, geojson, distance_km, distance_mi, created_at
		FROM routes
		WHERE id = ?
	`

	var route Route
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&route.ID,
		&route.Source,
		&route.SourceID,
		&route.SourceURL,
		&route.GeoJSON,
		&route.DistanceKm,
		&route.DistanceMi,
		&route.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to query route: %w", err)
	}

	return &route, nil
}

// GetAllRoutes returns all routes for a specific city for cache initialization
func (r *Repository) GetAllRoutes(ctx context.Context, city string) ([]Route, error) {
	query := `
		SELECT id, source, source_id, source_url, geojson, distance_km, distance_mi, created_at
		FROM routes
		WHERE city = ?
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, city)
	if err != nil {
		return nil, fmt.Errorf("failed to query routes: %w", err)
	}
	defer rows.Close()

	routes := make([]Route, 0) // Initialize as empty slice instead of nil
	for rows.Next() {
		var route Route
		err := rows.Scan(
			&route.ID,
			&route.Source,
			&route.SourceID,
			&route.SourceURL,
			&route.GeoJSON,
			&route.DistanceKm,
			&route.DistanceMi,
			&route.CreatedAt,
		)
		if err != nil {
			slog.Warn("failed to scan route", "error", err)
			continue
		}
		routes = append(routes, route)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating routes: %w", err)
	}

	return routes, nil
}

// generateID generates a unique ID using UUID v4
func generateID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		// Fallback to timestamp-based ID if random fails
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	// Simple UUID v4 format
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

// RouteToUpsert represents a route ready to be inserted/updated
type RouteToUpsert struct {
	ID         string
	Source     string
	SourceID   string
	SourceURL  string
	GeoJSON    string
	DistanceKm float64
	DistanceMi float64
	City       string
}

// BulkUpsertRoutes inserts or replaces multiple routes in a single transaction
func (r *Repository) BulkUpsertRoutes(ctx context.Context, routesToUpsert []RouteToUpsert) error {
	if len(routesToUpsert) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT OR REPLACE INTO routes (id, source, source_id, source_url, geojson, distance_km, distance_mi, city, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, route := range routesToUpsert {
		_, err := stmt.ExecContext(ctx, route.ID, route.Source, route.SourceID, route.SourceURL, route.GeoJSON, route.DistanceKm, route.DistanceMi, route.City)
		if err != nil {
			return fmt.Errorf("failed to execute upsert for route %s: %w", route.SourceID, err)
		}
		slog.Info("upserted route", "source", route.Source, "sourceID", route.SourceID, "routeID", route.ID)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	slog.Info("bulk upserted routes", "count", len(routesToUpsert))
	return nil
}

// isUniqueConstraintError checks if error is a unique constraint violation
func isUniqueConstraintError(err error) bool {
	return err != nil && (err.Error() == "UNIQUE constraint failed: routes.source, routes.source_id" ||
		err.Error() == "unique constraint violation" ||
		err.Error() == "UNIQUE constraint failed")
}
