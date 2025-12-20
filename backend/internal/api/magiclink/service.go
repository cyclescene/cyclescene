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
}

// SendMagicLinkResponse contains the result of sending a magic link
type SendMagicLinkResponse struct {
	MessageID string `json:"message_id"`
	Email     string `json:"email"`
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
	subject := "Edit Your CycleScene Ride"
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
  </style>
</head>
<body>
  <div class="container">
    <div class="header">
      <h2>Edit Your Ride</h2>
    </div>

    <p>Your ride has been submitted! Click the button below to edit it anytime.</p>

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
`, req.RedirectURL, req.RedirectURL)

	// Send email via Resend
	slog.Info("Sending magic link email via Resend", "email", req.Email, "redirect_url", req.RedirectURL)

	params := &resend.SendEmailRequest{
		From:    "magic@cyclescene.cc",
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
