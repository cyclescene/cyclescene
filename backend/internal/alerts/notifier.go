package alerts

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"
)

// Notifier handles sending alerts via email (Resend) and push notifications (ntfy.sh)
type Notifier struct {
	resendAPIKey string
	ntfyTopic    string
	adminEmail   string
	httpClient   *http.Client
}

// NewNotifier creates a new notifier from environment variables
func NewNotifier() *Notifier {
	return &Notifier{
		resendAPIKey: os.Getenv("RESEND_API_KEY"),
		ntfyTopic:    os.Getenv("NTFY_TOPIC"),
		adminEmail:   os.Getenv("ADMIN_EMAIL"),
		httpClient:   &http.Client{Timeout: 10 * time.Second},
	}
}

// IsConfigured returns true if at least one notification method is configured
func (n *Notifier) IsConfigured() bool {
	return (n.resendAPIKey != "" && n.adminEmail != "") || n.ntfyTopic != ""
}

// SendCriticalAlert sends both email and push notification for critical failures
func (n *Notifier) SendCriticalAlert(ctx context.Context, title, message string) error {
	if !n.IsConfigured() {
		slog.Debug("alerting_not_configured", "title", title)
		return nil
	}

	var errs []error

	// Send email
	if err := n.sendEmail(ctx, title, message); err != nil {
		slog.Error("failed_to_send_email_alert", "error", err)
		errs = append(errs, fmt.Errorf("email: %w", err))
	}

	// Send push notification
	if err := n.sendPush(ctx, title, message); err != nil {
		slog.Error("failed_to_send_push_alert", "error", err)
		errs = append(errs, fmt.Errorf("push: %w", err))
	}

	if len(errs) > 0 {
		return fmt.Errorf("alert failures: %v", errs)
	}

	slog.Info("critical_alert_sent", "title", title)
	return nil
}

// sendEmail sends an alert via Resend email API
func (n *Notifier) sendEmail(ctx context.Context, title, message string) error {
	if n.resendAPIKey == "" || n.adminEmail == "" {
		slog.Debug("email_alerting_not_configured")
		return nil
	}

	emailBody := struct {
		From    string   `json:"from"`
		To      []string `json:"to"`
		Subject string   `json:"subject"`
		HTML    string   `json:"html"`
	}{
		From:    "CycleScene Alerts <alerts@cyclescene.cc>",
		To:      []string{n.adminEmail},
		Subject: fmt.Sprintf("🚨 %s", title),
		HTML: fmt.Sprintf(`
			<h2>%s</h2>
			<p>%s</p>
			<hr>
			<small>Automated alert from CycleScene</small>
		`, title, message),
	}

	body, err := json.Marshal(emailBody)
	if err != nil {
		return fmt.Errorf("failed to marshal email body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.resend.com/emails", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+n.resendAPIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := n.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("resend API error: status %d", resp.StatusCode)
	}

	slog.Debug("email_alert_sent", "to", n.adminEmail)
	return nil
}

// sendPush sends an alert via ntfy.sh push notification
func (n *Notifier) sendPush(ctx context.Context, title, message string) error {
	if n.ntfyTopic == "" {
		slog.Debug("push_alerting_not_configured")
		return nil
	}

	pushBody := map[string]interface{}{
		"topic":    n.ntfyTopic,
		"title":    title,
		"message":  message,
		"priority": 5, // urgent
		"tags":     []string{"rotating_light", "cyclescene"},
	}

	body, err := json.Marshal(pushBody)
	if err != nil {
		return fmt.Errorf("failed to marshal push body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://ntfy.sh", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := n.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("ntfy error: status %d", resp.StatusCode)
	}

	slog.Debug("push_alert_sent", "topic", n.ntfyTopic)
	return nil
}
