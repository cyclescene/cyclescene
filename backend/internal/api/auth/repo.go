package auth

import (
	"database/sql"
	"time"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

type SubmissionToken struct {
	Token     string
	City      string
	ExpiresAt time.Time
	Used      bool
}

func (r *Repository) CreateSubmissionToken(token, city string, expiresAt time.Time) error {
	_, err := r.db.Exec(`
		INSERT INTO submission_tokens (token, city, expires_at)
		VALUES (?, ?, ?)
	`, token, city, expiresAt.Format(time.RFC3339))
	return err
}

func (r *Repository) GetSubmissionToken(token string) (*SubmissionToken, error) {
	var st SubmissionToken
	var expiresAt string

	err := r.db.QueryRow(`
		SELECT token, city, expires_at, used 
		FROM submission_tokens 
		WHERE token = ?
	`, token).Scan(&st.Token, &st.City, &expiresAt, &st.Used)

	if err != nil {
		return nil, err
	}

	st.ExpiresAt, err = time.Parse(time.RFC3339, expiresAt)
	if err != nil {
		return nil, err
	}

	return &st, nil
}

func (r *Repository) MarkTokenAsUsed(token string) error {
	_, err := r.db.Exec(`UPDATE submission_tokens SET used = 1 WHERE token = ?`, token)
	return err
}
