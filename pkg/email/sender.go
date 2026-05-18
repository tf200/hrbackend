package email

import (
	"context"
	"errors"
	"fmt"
	"log"

	brevo "github.com/getbrevo/brevo-go/lib"
)

type BrevoConf struct {
	SenderName  string
	Senderemail string
	ApiKey      string
	client      *brevo.APIClient
}

func NewBrevoConf(senderName, senderEmail, apiKey string) *BrevoConf {
	cfg := brevo.NewConfiguration()
	cfg.AddDefaultHeader("api-key", apiKey)
	return &BrevoConf{
		SenderName:  senderName,
		Senderemail: senderEmail,
		ApiKey:      apiKey,
		client:      brevo.NewAPIClient(cfg),
	}
}

func (b *BrevoConf) SendHTML(ctx context.Context, to []string, subject, htmlContent string) error {
	if len(to) == 0 {
		return errors.New("no recipient addresses provided")
	}
	if b.SenderName == "" || b.Senderemail == "" {
		return errors.New("invalid sender configuration")
	}
	if b.ApiKey == "" {
		return errors.New("invalid API key")
	}

	sender := brevo.SendSmtpEmailSender{
		Name:  b.SenderName,
		Email: b.Senderemail,
	}

	recipients := make([]brevo.SendSmtpEmailTo, 0, len(to))
	for _, recipient := range to {
		recipients = append(recipients, brevo.SendSmtpEmailTo{
			Email: recipient,
			Name:  recipient,
		})
	}

	emailContent := brevo.SendSmtpEmail{
		Sender:      &sender,
		To:          recipients,
		Subject:     subject,
		HtmlContent: htmlContent,
	}

	result, response, err := b.client.TransactionalEmailsApi.SendTransacEmail(ctx, emailContent)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	if response.StatusCode != 201 {
		return fmt.Errorf("failed to send email, status code: %d", response.StatusCode)
	}

	log.Printf("Email sent to %s", to)
	log.Printf("Response: %s", result)
	log.Printf("Response Status Code: %d", response.StatusCode)
	log.Printf("Response Headers: %v", response.Header)
	log.Printf("Response Body: %s", response.Body)

	return nil
}
