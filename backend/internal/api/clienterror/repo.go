package clienterror

import (
	"database/sql"
	"log/slog"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// LogError inserts a new client error entry into the database
func (r *Repository) LogError(err *ClientError) error {
	query := `
		INSERT INTO client_errors (client_id, error_type, error_msg, stack_trace, component, action, url, user_agent, os, city_code, timestamp)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, dbErr := r.db.Exec(
		query,
		err.ClientID,
		err.ErrorType,
		err.ErrorMsg,
		err.StackTrace,
		err.Component,
		err.Action,
		err.URL,
		err.UserAgent,
		err.OS,
		err.CityCode,
		err.Timestamp,
	)

	if dbErr != nil {
		slog.Error("Failed to log client error", "error", dbErr, "client_id", err.ClientID)
		return dbErr
	}

	id, dbErr := result.LastInsertId()
	if dbErr != nil {
		slog.Error("Failed to get last insert ID", "error", dbErr)
		return dbErr
	}

	err.ID = id
	return nil
}
