package group

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) ValidateGroupCode(code string) (string, error) {
	var name string
	err := r.db.QueryRow(`
		SELECT name FROM ride_groups WHERE code = ? AND is_active = 1
	`, strings.ToUpper(code)).Scan(&name)

	if err == sql.ErrNoRows {
		return "", nil
	}
	return name, err
}

func (r *Repository) CheckCodeAvailability(code string) (bool, error) {
	var count int
	err := r.db.QueryRow(`
		SELECT COUNT(*) FROM ride_groups WHERE code = ?
	`, strings.ToUpper(code)).Scan(&count)

	if err != nil {
		return false, err
	}
	return count == 0, nil
}

func (r *Repository) CreateGroup(reg *Registration, editToken string) error {
	groupID := uuid.New().String()

	_, err := r.db.Exec(`
		INSERT INTO ride_groups (id, code, name, description, city, web_url, edit_token, is_active, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, 1, CURRENT_TIMESTAMP)
	`, groupID, strings.ToUpper(reg.Code), reg.Name, reg.Description, reg.City, reg.WebURL, editToken)

	return err
}

func (r *Repository) GetGroupByEditToken(token string) (*Registration, error) {
	var reg Registration
	err := r.db.QueryRow(`
		SELECT code, name, description, city, web_url
		FROM ride_groups WHERE edit_token = ?
	`, token).Scan(&reg.Code, &reg.Name, &reg.Description, &reg.City, &reg.WebURL)

	return &reg, err
}

func (r *Repository) UpdateGroup(token string, reg *Registration) error {
	result, err := r.db.Exec(`
		UPDATE ride_groups SET
			name = ?, description = ?, web_url = ?
		WHERE edit_token = ?
	`, reg.Name, reg.Description, reg.WebURL, token)

	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// SetGroupMarker updates the marker field for a group after marker processing
func (r *Repository) SetGroupMarker(publicID, markerKey string) error {
	result, err := r.db.Exec(`
		UPDATE ride_groups SET marker = ? WHERE public_id = ?
	`, markerKey, publicID)

	if err != nil {
		return fmt.Errorf("failed to update group marker: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("group with public_id %s not found", publicID)
	}

	return nil
}

// GetGroupByPublicID retrieves a group by its public_id
func (r *Repository) GetGroupByPublicID(publicID string) (string, error) {
	var id string
	err := r.db.QueryRow(`
		SELECT id FROM ride_groups WHERE public_id = ?
	`, publicID).Scan(&id)

	if err == sql.ErrNoRows {
		return "", fmt.Errorf("group with public_id %s not found", publicID)
	}
	return id, err
}
