package routes

import (
	"context"
	"database/sql"
	"encoding/json"
	"log/slog"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// Route represents a route with its GeoJSON data
type Route struct {
	ID         string          `json:"id"`
	Type       string          `json:"type"`
	Geometry   json.RawMessage `json:"geometry"`
	Properties json.RawMessage `json:"properties"`
}

// GetAllRoutes retrieves all routes from the database and formats them as GeoJSON features
func (r *Repository) GetAllRoutes(ctx context.Context) ([]Route, error) {
	query := `
		SELECT id, geojson
		FROM routes
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		slog.Error("[Routes Repo] Query error", "error", err)
		return nil, err
	}
	defer rows.Close()

	var routes []Route

	for rows.Next() {
		var id string
		var geoJSON string

		if err := rows.Scan(&id, &geoJSON); err != nil {
			slog.Error("[Routes Repo] Scan error", "error", err)
			continue
		}

		// Parse the GeoJSON string and reconstruct with id
		var feature map[string]interface{}
		if err := json.Unmarshal([]byte(geoJSON), &feature); err != nil {
			slog.Error("[Routes Repo] Failed to unmarshal GeoJSON", "error", err, "route_id", id)
			continue
		}

		// Extract geometry and properties from the stored GeoJSON
		var geometryBytes, propertiesBytes json.RawMessage

		if geom, ok := feature["geometry"]; ok {
			geometryBytes, _ = json.Marshal(geom)
		}
		if props, ok := feature["properties"]; ok {
			propertiesBytes, _ = json.Marshal(props)
		}

		route := Route{
			ID:         id,
			Type:       "Feature",
			Geometry:   geometryBytes,
			Properties: propertiesBytes,
		}

		routes = append(routes, route)
	}

	if err := rows.Err(); err != nil {
		slog.Error("[Routes Repo] Rows error", "error", err)
		return nil, err
	}

	slog.Info("[Routes Repo] Retrieved routes", "count", len(routes))
	return routes, nil
}
