package auth

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"time"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

type TokenResponse struct {
	Token     string `json:"token"`
	ExpiresAt string `json:"expires_at"`
}

type ValidateTokenResponse struct {
	Valid bool   `json:"valid"`
	City  string `json:"city,omitempty"`
}

func (s *Service) GenerateSubmissionToken(city string) (*TokenResponse, error) {
	token, err := generateSecureToken(32)
	if err != nil {
		return nil, err
	}

	expiresAt := time.Now().Add(30 * time.Minute)

	if err := s.repo.CreateSubmissionToken(token, city, expiresAt); err != nil {
		return nil, err
	}

	return &TokenResponse{
		Token:     token,
		ExpiresAt: expiresAt.Format(time.RFC3339),
	}, nil
}

func (s *Service) ValidateSubmissionToken(token, city string) (*ValidateTokenResponse, error) {
	st, err := s.repo.GetSubmissionToken(token)
	if err == sql.ErrNoRows {
		return &ValidateTokenResponse{Valid: false}, nil
	}
	if err != nil {
		return nil, err
	}

	// Check if already used
	if st.Used {
		return &ValidateTokenResponse{Valid: false}, nil
	}

	// Check if expired
	if time.Now().After(st.ExpiresAt) {
		return &ValidateTokenResponse{Valid: false}, nil
	}

	// Check city matches if provided
	if city != "" && st.City != city {
		return &ValidateTokenResponse{Valid: false}, nil
	}

	return &ValidateTokenResponse{
		Valid: true,
		City:  st.City,
	}, nil
}

func (s *Service) MarkTokenAsUsed(token string) error {
	return s.repo.MarkTokenAsUsed(token)
}

func (s *Service) VerifyAndConsumeToken(token, city string) error {
	validation, err := s.ValidateSubmissionToken(token, city)
	if err != nil {
		return err
	}

	if !validation.Valid {
		return errors.New("invalid or expired token")
	}

	if validation.City != city {
		return errors.New("token city mismatch")
	}

	return s.MarkTokenAsUsed(token)
}

func generateSecureToken(length int) (string, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
