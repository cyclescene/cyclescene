package ride

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log/slog"
	"net/url"
	"strings"
	"time"

	"github.com/spacesedan/cyclescene/functions/internal/api/magiclink"
	"github.com/spacesedan/cyclescene/functions/internal/scraper"
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

// User-submitted rides
func (s *Service) SubmitRide(submission *Submission) (*SubmissionResponse, error) {
	// Generate edit token
	editToken, err := generateSecureToken(32)
	if err != nil {
		return nil, err
	}

	// Geocode the address to get latitude and longitude
	var lat, lng float64
	if submission.Address != "" {
		geocodeQuery := fmt.Sprintf("%s %s", submission.VenueName, submission.Address)
		lat, lng, err = scraper.GeocodeQuery(geocodeQuery, submission.City)
		if err != nil {
			slog.Warn("Failed to geocode address", "geocodequery", geocodeQuery, "city", submission.City, "error", err)
			// Continue with 0,0 coordinates if geocoding fails
			lat, lng = 0.0, 0.0
		} else {
			slog.Info("Successfully geocoded address", "address", submission.Address, "lat", lat, "lng", lng)
		}
	}

	eventID, err := s.repo.CreateRide(submission, editToken, lat, lng)
	if err != nil {
		return nil, err
	}

	// Send magic link email if service is configured and organizer email exists
	if s.magicLinkSvc != nil && submission.OrganizerEmail != "" {
		// Build the full redirect URL with the edit token
		redirectURL := fmt.Sprintf("%s?token=%s", s.editLinkBaseURL, editToken)
		_, err := s.magicLinkSvc.SendMagicLink(context.Background(), magiclink.SendMagicLinkRequest{
			Email:       submission.OrganizerEmail,
			RedirectURL: redirectURL,
		})
		if err != nil {
			// Log but don't fail - ride was created successfully
			slog.Error("Failed to send magic link email", "error", err, "email", submission.OrganizerEmail, "event_id", eventID)
		}
	}

	return &SubmissionResponse{
		Success:   true,
		EventID:   eventID,
		EditToken: editToken,
		Message:   "Ride submitted successfully and is pending review",
	}, nil
}

func (s *Service) GetRideByEditToken(token string) (*EditResponse, error) {
	submission, isPublished, err := s.repo.GetRideByEditToken(token)
	if err != nil {
		return nil, err
	}

	return &EditResponse{
		Event:       *submission,
		IsPublished: isPublished,
	}, nil
}

// UpdateOccurrence updates a single occurrence's details (time, duration, details, newsflash, cancelled status)
func (s *Service) UpdateOccurrence(token string, occurrenceID int64, startTime string, eventDurationMinutes int, eventTimeDetails string, newsflash string, isCancelled bool) error {
	return s.repo.UpdateOccurrence(token, occurrenceID, startTime, eventDurationMinutes, eventTimeDetails, newsflash, isCancelled)
}

func (s *Service) UpdateRide(token string, submission *Submission) (*SubmissionResponse, error) {
	// Geocode the address to get latitude and longitude
	var lat, lng float64
	if submission.Address != "" {
		var err error
		lat, lng, err = scraper.GeocodeQuery(submission.Address, submission.City)
		if err != nil {
			slog.Warn("Failed to geocode address", "address", submission.Address, "city", submission.City, "error", err)
			// Continue with 0,0 coordinates if geocoding fails
			lat, lng = 0.0, 0.0
		} else {
			slog.Info("Successfully geocoded address", "address", submission.Address, "lat", lat, "lng", lng)
		}
	}

	if err := s.repo.UpdateRide(token, submission, lat, lng); err != nil {
		return nil, err
	}

	return &SubmissionResponse{
		Success: true,
		Message: "Ride updated successfully",
	}, nil
}

// Scraped rides from Shift2Bikes
func (s *Service) GetUpcomingRides(city string) ([]ScrapedRide, error) {
	storedRides, err := s.repo.GetUpcomingRides(city)
	if err != nil {
		slog.Error("Failed to query upcoming rides", "error", err)
		return nil, err
	}

	var rides []ScrapedRide
	for i := range storedRides {
		rides = append(rides, storedRides[i].ToScrapedRide())
	}

	if rides == nil {
		rides = []ScrapedRide{}
	}

	return rides, nil
}

func (s *Service) GetPastRides(city string) ([]ScrapedRide, error) {
	storedRides, err := s.repo.GetPastRides(city)
	if err != nil {
		slog.Error("Failed to query past rides", "error", err)
		return nil, err
	}

	var rides []ScrapedRide
	for i := range storedRides {
		rides = append(rides, storedRides[i].ToScrapedRide())
	}

	if rides == nil {
		rides = []ScrapedRide{}
	}

	return rides, nil
}

func (s *Service) GenerateICSFromRide(city, rideID string) (ICSContent, error) {
	storedRide, err := s.repo.GetRide(city, rideID)
	if err != nil {
		slog.Error("Ride not found", "city", city, "rideID", rideID, "error", err)
		return ICSContent{}, err
	}

	if len(storedRide) == 0 {
		slog.Error("Ride not found", "city", city, "rideID", rideID)
		return ICSContent{}, fmt.Errorf("ride not found")
	}

	ride := storedRide[0].ToScrapedRide()

	startTimeStr := fmt.Sprintf("%s %s", ride.Date, ride.StartTime)
	start, err := time.Parse("2006-01-02 15:04:05", startTimeStr)
	if err != nil {
		return ICSContent{}, err
	}

	end := start.Add(2 * time.Hour)
	if ride.EndTime != "" {
		if endTime, err := time.Parse("2006-01-02 15:04:05", fmt.Sprintf("%s %s", ride.Date, ride.EndTime)); err == nil {
			end = endTime
		}
	}

	formatICS := func(t time.Time) string {
		return t.UTC().Format("20060102T150405Z")
	}

	// Clean newlines and commas
	desc := strings.ReplaceAll(ride.Details, "\n", "\\n")
	desc = strings.ReplaceAll(desc, ",", "\\,")

	icsContent := strings.Join([]string{
		"BEGIN:VCALENDAR",
		"VERSION:2.0",
		"PRODID:-//CycleScene//EN",
		"BEGIN:VEVENT",
		fmt.Sprintf("UID:%s@cyclescene.com", ride.ID),
		fmt.Sprintf("DTSTAMP:%s", formatICS(time.Now())),
		fmt.Sprintf("DTSTART:%s", formatICS(start)),
		fmt.Sprintf("DTEND:%s", formatICS(end)),
		fmt.Sprintf("SUMMARY:%s", ride.Title),
		fmt.Sprintf("LOCATION:%s", ride.Venue),
		fmt.Sprintf("DESCRIPTION:%s\\nURL:%s", desc, ride.Shareable),
		"STATUS:CONFIRMED",
		"END:VEVENT",
		"END:VCALENDAR",
		"", // ensure trailing CRLF
	}, "\r\n")

	filename := url.QueryEscape(ride.Title)

	return ICSContent{
		Filename: filename,
		Content:  icsContent,
	}, nil
}

func generateSecureToken(length int) (string, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// GetPendingRides returns all rides that are not yet published
func (s *Service) GetPendingRides() ([]RideForAdmin, error) {
	return s.repo.GetPendingRides()
}

// PublishRide marks a ride as published
func (s *Service) PublishRide(rideID int, moderationNotes string) error {
	return s.repo.PublishRide(rideID, moderationNotes)
}

// ValidateAdminKey checks if an API key is valid
func (s *Service) ValidateAdminKey(apiKey string) (bool, error) {
	return s.repo.ValidateAdminKey(apiKey)
}
