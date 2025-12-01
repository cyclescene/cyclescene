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
	ID       string          `json:"id"`
	GeoJSON  json.RawMessage `json:"geojson"`
}

// GetAllRoutes retrieves all routes for a specific city from the database with their full GeoJSON data
func (r *Repository) GetAllRoutes(ctx context.Context, city string) ([]Route, error) {
	query := `
		SELECT id, geojson
		FROM routes
		WHERE city = ?
		ORDER BY created_at DESC
	`

	slog.Info("[Routes Repo] Executing query", "city", city)
	rows, err := r.db.QueryContext(ctx, query, city)
	if err != nil {
		slog.Error("[Routes Repo] Query error", "error", err, "city", city)
		return nil, err
	}
	defer rows.Close()

	routes := make([]Route, 0) // Initialize as empty slice instead of nil

	for rows.Next() {
		var id string
		var geoJSON string

		if err := rows.Scan(&id, &geoJSON); err != nil {
			slog.Error("[Routes Repo] Scan error", "error", err)
			continue
		}

		route := Route{
			ID:      id,
			GeoJSON: json.RawMessage(geoJSON),
		}

		routes = append(routes, route)
	}

	if err := rows.Err(); err != nil {
		slog.Error("[Routes Repo] Rows error", "error", err)
		return nil, err
	}

	slog.Info("[Routes Repo] Retrieved routes", "count", len(routes), "city", city)
	return routes, nil
}
