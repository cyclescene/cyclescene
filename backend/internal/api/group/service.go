package group

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/spacesedan/cyclescene/functions/internal/api/magiclink"
)

type Service struct {
	repo            *Repository
	magicLinkSvc    *magiclink.Service
	editLinkBaseURL string
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func NewServiceWithMagicLink(repo *Repository, magicLinkSvc *magiclink.Service, editLinkBaseURL string) *Service {
	return &Service{
		repo:            repo,
		magicLinkSvc:    magicLinkSvc,
		editLinkBaseURL: editLinkBaseURL,
	}
}

func (s *Service) ValidateGroupCode(code string) (*ValidationResponse, error) {
	if len(code) != 4 {
		return &ValidationResponse{Valid: false}, nil
	}

	name, err := s.repo.ValidateGroupCode(code)
	if err != nil {
		return nil, err
	}

	if name == "" {
		return &ValidationResponse{Valid: false}, nil
	}

	return &ValidationResponse{
		Valid: true,
		Name:  name,
	}, nil
}

func (s *Service) CheckCodeAvailability(code string) (*AvailabilityResponse, error) {
	code = strings.ToUpper(strings.TrimSpace(code))

	if len(code) != 4 {
		return &AvailabilityResponse{
			Available: false,
			Message:   "Code must be exactly 4 characters",
		}, nil
	}

	available, err := s.repo.CheckCodeAvailability(code)
	if err != nil {
		return nil, err
	}

	return &AvailabilityResponse{
		Available: available,
		Code:      code,
	}, nil
}

func (s *Service) RegisterGroup(reg *Registration) (*Response, error) {
	// Normalize code
	reg.Code = strings.ToUpper(strings.TrimSpace(reg.Code))

	// Validate code format
	if len(reg.Code) != 4 {
		return nil, errors.New("code must be exactly 4 characters")
	}

	// Check if code is available
	available, err := s.repo.CheckCodeAvailability(reg.Code)
	if err != nil {
		return nil, err
	}
	if !available {
		return nil, errors.New("group code already exists")
	}

	// Generate edit token
	editToken, err := generateSecureToken(32)
	if err != nil {
		return nil, err
	}

	// Create group
	if err := s.repo.CreateGroup(reg, editToken); err != nil {
		return nil, err
	}

	// Send magic link email if service is configured and organizer email exists
	if s.magicLinkSvc != nil && reg.Email != "" {
		// Build the full redirect URL with the edit token and city
		redirectURL := fmt.Sprintf("%s?token=%s&city=%s", s.editLinkBaseURL, editToken, strings.ToLower(reg.City))
		_, err := s.magicLinkSvc.SendMagicLink(context.Background(), magiclink.SendMagicLinkRequest{
			Email:       reg.Email,
			RedirectURL: redirectURL,
		})
		if err != nil {
			// Log but don't fail - group was created successfully
			slog.Error("Failed to send magic link email for group", "error", err, "email", reg.Email, "code", reg.Code)
		}
	}

	return &Response{
		Success:   true,
		Code:      reg.Code,
		EditToken: editToken,
		Message:   "Group registered successfully",
	}, nil
}

func (s *Service) GetGroupByEditToken(token string) (*Registration, error) {
	return s.repo.GetGroupByEditToken(token)
}

func (s *Service) UpdateGroup(token string, reg *Registration) (*Response, error) {
	if err := s.repo.UpdateGroup(token, reg); err != nil {
		return nil, err
	}

	return &Response{
		Success: true,
		Message: "Group updated successfully",
	}, nil
}

func generateSecureToken(length int) (string, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
