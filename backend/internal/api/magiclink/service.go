package magiclink

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/resendlabs/resend-go"
)

type Service struct {
	resendClient *resend.Client
}

// SendMagicLinkRequest contains the information needed to send a magic link
type SendMagicLinkRequest struct {
	Email       string // ride organizer email
	RedirectURL string // Full URL with token (e.g., http://localhost:5174/rides/edit?token=xyz)
	IPAddress   string // client IP for security (not used with Resend but kept for compatibility)
	EventTitle  string // optional: title of the ride/event (for personalized emails)
}

// SendMagicLinkResponse contains the result of sending a magic link
type SendMagicLinkResponse struct {
	MessageID string `json:"message_id"`
	Email     string `json:"email"`
}

// ImportedEvent represents an event that was imported (for summary emails)
type ImportedEvent struct {
	Title     string // Event title
	EventID   int64  // CycleScene event ID
	EditToken string // Token for editing the event
	EditURL   string // Full URL for editing the event
}

func NewService(apiKey string) *Service {
	return &Service{
		resendClient: resend.NewClient(apiKey),
	}
}

// SendMagicLink sends a magic link email via Resend
func (s *Service) SendMagicLink(_ context.Context, req SendMagicLinkRequest) (*SendMagicLinkResponse, error) {
	if req.Email == "" {
		return nil, fmt.Errorf("email is required")
	}
	if req.RedirectURL == "" {
		return nil, fmt.Errorf("redirect URL is required")
	}

	// Create email body
	// Customize subject and body based on whether event title is provided
	subject := "Edit Your CycleScene Ride"
	headerText := "Edit Your Ride"
	bodyText := "Your ride has been submitted! Click the button below to edit it anytime."

	if req.EventTitle != "" {
		subject = fmt.Sprintf("Edit Your Ride: %s", req.EventTitle)
		headerText = fmt.Sprintf("Edit: %s", req.EventTitle)
		bodyText = "Your ride has been submitted! Click the button below to edit it anytime."
	}

	htmlBody := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <style>
    body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
    .container { max-width: 600px; margin: 0 auto; padding: 20px; }
    .header { margin-bottom: 30px; }
    .button { display: inline-block; background-color: #000; color: #fff; padding: 12px 24px; text-decoration: none; border-radius: 4px; margin: 20px 0; }
    .footer { margin-top: 40px; padding-top: 20px; border-top: 1px solid #eee; font-size: 12px; color: #666; }
    .event-title { color: #666; font-size: 14px; margin-bottom: 20px; }
  </style>
</head>
<body>
  <div class="container">
    <div class="header">
      <h2>%s</h2>
    </div>

    <p>%s</p>

    <a href="%s" class="button">Edit Your Ride</a>

    <p style="color: #666; font-size: 14px;">
      Or copy and paste this link:<br>
      <code>%s</code>
    </p>

    <div class="footer">
      <p>This link will remain active so you can edit your ride whenever you need to.</p>
      <p>CycleScene</p>
    </div>
  </div>
</body>
</html>
`, headerText, bodyText, req.RedirectURL, req.RedirectURL)

	// Send email via Resend
	slog.Info("Sending magic link email via Resend", "email", req.Email, "redirect_url", req.RedirectURL)

	params := &resend.SendEmailRequest{
		From:    "CycleScene <noreply@email.cyclescene.cc>",
		To:      []string{req.Email},
		Subject: subject,
		Html:    htmlBody,
	}

	sent, err := s.resendClient.Emails.Send(params)
	if err != nil {
		slog.Error("Failed to send magic link via Resend", "error", err, "email", req.Email, "redirect_url", req.RedirectURL)
		return nil, fmt.Errorf("failed to send magic link: %w", err)
	}

	slog.Info("Sent magic link email via Resend", "email", req.Email, "message_id", sent.Id)

	return &SendMagicLinkResponse{
		MessageID: sent.Id,
		Email:     req.Email,
	}, nil
}

// SendImportSummaryEmail sends a summary email after batch importing events from Strava
func (s *Service) SendImportSummaryEmail(_ context.Context, email string, events []ImportedEvent) (*SendMagicLinkResponse, error) {
	if email == "" {
		return nil, fmt.Errorf("email is required")
	}
	if len(events) == 0 {
		return nil, fmt.Errorf("at least one event is required")
	}

	// Build the event list HTML
	eventListHTML := ""
	for _, event := range events {
		eventListHTML += fmt.Sprintf(`
      <tr>
        <td style="padding: 12px; border-bottom: 1px solid #eee;">
          <strong>%s</strong>
        </td>
        <td style="padding: 12px; border-bottom: 1px solid #eee; text-align: right;">
          <a href="%s" style="color: #000; text-decoration: underline;">Edit</a>
        </td>
      </tr>`, event.Title, event.EditURL)
	}

	// Determine subject based on number of events
	subject := "Your Strava Rides Have Been Imported"
	headerText := "Strava Import Complete"
	var countText string
	if len(events) == 1 {
		countText = "1 ride has"
	} else {
		countText = fmt.Sprintf("%d rides have", len(events))
	}

	htmlBody := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <style>
    body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
    .container { max-width: 600px; margin: 0 auto; padding: 20px; }
    .header { margin-bottom: 30px; }
    .event-table { width: 100%%; border-collapse: collapse; margin: 20px 0; }
    .footer { margin-top: 40px; padding-top: 20px; border-top: 1px solid #eee; font-size: 12px; color: #666; }
  </style>
</head>
<body>
  <div class="container">
    <div class="header">
      <h2>%s</h2>
    </div>

    <p>%s been imported from Strava to CycleScene! Your rides are pending review and will appear on the calendar once approved.</p>

    <p>Use the links below to edit any of your imported rides:</p>

    <table class="event-table">
      <tbody>
        %s
      </tbody>
    </table>

    <p style="color: #666; font-size: 14px;">
      These links will remain active so you can edit your rides whenever you need to.
    </p>

    <div class="footer">
      <p>Thank you for sharing your rides with the cycling community!</p>
      <p>CycleScene</p>
    </div>
  </div>
</body>
</html>
`, headerText, countText, eventListHTML)

	// Send email via Resend
	slog.Info("Sending import summary email via Resend",
		"email", email,
		"event_count", len(events),
	)

	params := &resend.SendEmailRequest{
		From:    "CycleScene <noreply@email.cyclescene.cc>",
		To:      []string{email},
		Subject: subject,
		Html:    htmlBody,
	}

	sent, err := s.resendClient.Emails.Send(params)
	if err != nil {
		slog.Error("Failed to send import summary email via Resend",
			"error", err,
			"email", email,
			"event_count", len(events),
		)
		return nil, fmt.Errorf("failed to send import summary email: %w", err)
	}

	slog.Info("Sent import summary email via Resend",
		"email", email,
		"message_id", sent.Id,
		"event_count", len(events),
	)

	return &SendMagicLinkResponse{
		MessageID: sent.Id,
		Email:     email,
	}, nil
}
