package email

import (
	"fmt"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

// Sender handles sending emails via SendGrid
type Sender struct {
	apiKey string
}

// NewSender creates a new email sender
func NewSender(apiKey string) *Sender {
	return &Sender{apiKey: apiKey}
}

// Send sends an HTML email via SendGrid
func (s *Sender) Send(fromEmail, toEmail, subject, htmlContent string) error {
	from := mail.NewEmail("The Paper", fromEmail)
	to := mail.NewEmail("", toEmail)

	message := mail.NewSingleEmail(from, subject, to, "", htmlContent)

	client := sendgrid.NewSendClient(s.apiKey)
	response, err := client.Send(message)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	if response.StatusCode >= 400 {
		return fmt.Errorf("sendgrid error: status %d, body: %s", response.StatusCode, response.Body)
	}

	return nil
}
