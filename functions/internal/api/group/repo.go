package group

import (
	"database/sql"
	"strings"
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
	_, err := r.db.Exec(`
		INSERT INTO ride_groups (code, name, description, city, icon_url, web_url, edit_token, is_active)
		VALUES (?, ?, ?, ?, ?, ?, ?, 1)
	`, strings.ToUpper(reg.Code), reg.Name, reg.Description, reg.City, reg.IconURL, reg.WebURL, editToken)

	return err
}

func (r *Repository) GetGroupByEditToken(token string) (*Registration, error) {
	var reg Registration
	err := r.db.QueryRow(`
		SELECT code, name, description, city, icon_url, web_url
		FROM ride_groups WHERE edit_token = ?
	`, token).Scan(&reg.Code, &reg.Name, &reg.Description, &reg.City, &reg.IconURL, &reg.WebURL)

	return &reg, err
}

func (r *Repository) UpdateGroup(token string, reg *Registration) error {
	result, err := r.db.Exec(`
		UPDATE ride_groups SET
			name = ?, description = ?, icon_url = ?, web_url = ?
		WHERE edit_token = ?
	`, reg.Name, reg.Description, reg.IconURL, reg.WebURL, token)

	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}
