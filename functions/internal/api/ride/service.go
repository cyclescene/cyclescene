package ride

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log/slog"
	"net/url"
	"strings"
	"time"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// User-submitted rides
func (s *Service) SubmitRide(submission *Submission) (*SubmissionResponse, error) {
	// Generate edit token
	editToken, err := generateSecureToken(32)
	if err != nil {
		return nil, err
	}

	eventID, err := s.repo.CreateRide(submission, editToken)
	if err != nil {
		return nil, err
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

func (s *Service) UpdateRide(token string, submission *Submission) (*SubmissionResponse, error) {
	if err := s.repo.UpdateRide(token, submission); err != nil {
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
