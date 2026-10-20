package mailer

import (
	"context"
	"errors"
	"log/slog"
	"os"

	"github.com/resend/resend-go/v3"
)

var client *resend.Client
var senderEmail string

// Initialize the mailing service. To be called only once at application startup
//
// Requirements from .env file:
//   - RESEND_API_KEY: API key from Resend
//   - SENDER_EMAIL: Name and email of sender (eg. "Sudi from UBCEA <sudi@ubcesports.ca>")
func Init() error {
	resend_api_key := os.Getenv("RESEND_API_KEY")
	if resend_api_key == "" {
		return errors.New("required environment variable 'RESEND_API_KEY' is missing")
	}

	sender := os.Getenv("SENDER_EMAIL")
	if sender == "" {
		return errors.New("required environment variable 'SENDER_EMAIL' is missing")
	}

	client = resend.NewClient(resend_api_key)
	senderEmail = sender
	return nil
}

// Dispatches an HTML email to one or more recipients asynchronously
//
// Params:
//   - to: list of recipients of mail
//   - subject: subject of mail
//   - htmlContent: content of mail in html format
//
// Example usage:
//
//	mailer.SendEmailAsync(
//		[]string{"user@example.com"},
//		"Welcome to our platform!",
//		"<h1>Thanks for signing up!</h1>",
//	)
func SendEmailAsync(to []string, subject string, htmlContent string, requestID string, userID string) {
	if client == nil {
		slog.Error("email delivery skipped: mailer not initialized",
			"request_id", requestID,
			"user_id", userID,
			"recipient_count", len(to),
		)
		return
	}

	go func() {
		params := &resend.SendEmailRequest{
			From:    senderEmail,
			To:      to,
			Html:    htmlContent,
			Subject: subject,
		}

		_, err := client.Emails.SendWithContext(context.Background(), params)
		if err != nil {
			slog.Error("email delivery failed",
				"error", err,
				"request_id", requestID,
				"user_id", userID,
				"recipient_count", len(to),
			)
			return
		}

		slog.Info("email delivered",
			"request_id", requestID,
			"user_id", userID,
			"recipient_count", len(to),
		)
	}()
}
