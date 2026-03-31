package email

import (
	"bytes"
	"context"
	"embed"
	"encoding/base64"
	"errors"
	"fmt"
	"html/template"
	"log"
	"strings"
	"time"

	brevo "github.com/getbrevo/brevo-go/lib"
	"github.com/wneessen/go-mail"
)

type SmtpConf struct {
	Name          string
	Address       string
	Athentication string
	SmtpHost      string
	SmtpPort      int
}

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

type Credentials struct {
	Name     string
	Email    string
	Password string
}

type Incident struct {
	IncidentID   string
	ReportedBy   string
	ClientName   string
	IncidentType string
	Severity     string
	Location     string
	DocumentLink string
}

type AcceptedRegitrationForm struct {
	ReferrerName        string
	ChildName           string
	ChildBSN            string
	AppointmentDate     string
	AppointmentLocation string
}

type ClientContractReminder struct {
	ClientID           string
	ClientFirstName    string
	ClientLastName     string
	ContractID         string
	CareType           string
	ContractStartDate  time.Time
	ContractEndDate    time.Time
	ContractStatus     string
	ReminderType       string
	LastReminderSentAt *time.Time
	CurrentDate        time.Time
	CurrentYear        int
}

func NewSmtpConf(name, address, authentication, smtpHost string, smtpPort int) *SmtpConf {
	return &SmtpConf{
		Name:          name,
		Address:       address,
		Athentication: authentication,
		SmtpHost:      smtpHost,
		SmtpPort:      smtpPort,
	}
}

func (e *SmtpConf) Send(subject, body string, to []string) error {
	message := mail.NewMsg()
	if err := message.From("dev@maicare.online"); err != nil {
		log.Fatalf("failed to set From address: %s", err)
	}
	if err := message.To(to[0]); err != nil {
		log.Fatalf("failed to set To address: %s", err)
	}
	message.Subject(subject)
	message.SetBodyString(mail.TypeTextPlain, body)
	_, err := mail.NewClient(e.SmtpHost, mail.WithSMTPAuth(mail.SMTPAuthPlain),
		mail.WithUsername(e.Address), mail.WithPassword(e.Athentication), mail.WithPort(e.SmtpPort))
	if err != nil {
		log.Fatalf("failed to create mail client: %s", err)
	}

	return nil
}

//go:embed templates/credentials.html
var credentialsTemplateFS embed.FS

type ProcessRegistrationForm struct {
	RecipientName string
	ClientName    string
	Location      string
	Link          string
}

//go:embed templates/process_registration_form.html
var processRegistrationFormTemplateFS embed.FS

func (b *BrevoConf) SendProcessRegistrationForm(ctx context.Context, to []string, data ProcessRegistrationForm) error {
	if len(to) == 0 {
		return errors.New("no recipient addresses provided")
	}
	if b.SenderName == "" || b.Senderemail == "" {
		return errors.New("invalid sender configuration")
	}
	if b.ApiKey == "" {
		return errors.New("invalid API key")
	}

	tmpl, err := template.ParseFS(processRegistrationFormTemplateFS, "templates/process_registration_form.html")
	if err != nil {
		return fmt.Errorf("failed to parse HTML template: %w", err)
	}

	var body bytes.Buffer
	if err := tmpl.Execute(&body, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	htmlContent := body.String()
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
		Subject:     "Intake Planning for " + data.ClientName,
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

func (b *BrevoConf) SendCredentials(ctx context.Context, to []string, data Credentials) error {
	if len(to) == 0 {
		return errors.New("no recipient addresses provided")
	}

	if b.SenderName == "" || b.Senderemail == "" {
		return errors.New("invalid sender configuration")
	}
	if b.ApiKey == "" {
		return errors.New("invalid API key")
	}

	tmpl, err := template.ParseFS(credentialsTemplateFS, "templates/credentials.html")
	if err != nil {
		return fmt.Errorf("failed to parse HTML template: %w", err)
	}

	var body bytes.Buffer
	if err := tmpl.Execute(&body, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}
	htmlContent := body.String()

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
		Subject:     "Maicare Credentials",
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

//go:embed templates/incident.html
var incidentTemplateFS embed.FS

func (b *BrevoConf) SendIncident(ctx context.Context, to []string, data Incident) error {
	if len(to) == 0 {
		return errors.New("no recipient addresses provided")
	}
	if b.SenderName == "" || b.Senderemail == "" {
		return errors.New("invalid sender configuration")
	}
	if b.ApiKey == "" {
		return errors.New("invalid API key")
	}

	tmpl, err := template.New("incident.html").Funcs(template.FuncMap{
		"toLower": strings.ToLower,
	}).ParseFS(incidentTemplateFS, "templates/incident.html")
	if err != nil {
		return fmt.Errorf("failed to parse HTML template: %w", err)
	}

	var body bytes.Buffer
	if err := tmpl.Execute(&body, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	htmlContent := body.String()
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
		Subject:     "Incident Report",
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

func (b *BrevoConf) SendIncidentWithAttachment(ctx context.Context, to []string, data Incident, attachmentName string, attachmentBytes []byte) error {
	if len(to) == 0 {
		return errors.New("no recipient addresses provided")
	}
	if b.SenderName == "" || b.Senderemail == "" {
		return errors.New("invalid sender configuration")
	}
	if b.ApiKey == "" {
		return errors.New("invalid API key")
	}
	if attachmentName == "" {
		return errors.New("invalid attachment name")
	}
	if len(attachmentBytes) == 0 {
		return errors.New("empty attachment")
	}

	tmpl, err := template.New("incident.html").Funcs(template.FuncMap{
		"toLower": strings.ToLower,
	}).ParseFS(incidentTemplateFS, "templates/incident.html")
	if err != nil {
		return fmt.Errorf("failed to parse HTML template: %w", err)
	}

	var body bytes.Buffer
	if err := tmpl.Execute(&body, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	htmlContent := body.String()
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

	attachment := brevo.SendSmtpEmailAttachment{
		Content: base64.StdEncoding.EncodeToString(attachmentBytes),
		Name:    attachmentName,
	}

	emailContent := brevo.SendSmtpEmail{
		Sender:      &sender,
		To:          recipients,
		Subject:     "Incident Report",
		HtmlContent: htmlContent,
		Attachment:  []brevo.SendSmtpEmailAttachment{attachment},
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

//go:embed templates/accepted_registration_form.html
var acceptedRegistrationFormTemplateFS embed.FS

func (b *BrevoConf) SendAcceptedRegistrationForm(ctx context.Context, to []string, data AcceptedRegitrationForm) error {
	if len(to) == 0 {
		return errors.New("no recipient addresses provided")
	}
	if b.SenderName == "" || b.Senderemail == "" {
		return errors.New("invalid sender configuration")
	}
	if b.ApiKey == "" {
		return errors.New("invalid API key")
	}

	tmpl, err := template.ParseFS(acceptedRegistrationFormTemplateFS, "templates/accepted_registration_form.html")
	if err != nil {
		return fmt.Errorf("failed to parse HTML template: %w", err)
	}

	var body bytes.Buffer
	if err := tmpl.Execute(&body, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	htmlContent := body.String()
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
		Subject:     "Accepted Registration Form",
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

//go:embed templates/client_contract_reminder.html
var clientContractReminderTemplateFS embed.FS

func (b *BrevoConf) SendClientContractReminder(ctx context.Context, to []string, data ClientContractReminder) error {
	if len(to) == 0 {
		return errors.New("no recipient addresses provided")
	}
	if b.SenderName == "" || b.Senderemail == "" {
		return errors.New("invalid sender configuration")
	}
	if b.ApiKey == "" {
		return errors.New("invalid API key")
	}

	tmpl, err := template.ParseFS(clientContractReminderTemplateFS, "templates/client_contract_reminder.html")
	if err != nil {
		return fmt.Errorf("failed to parse HTML template: %w", err)
	}

	var body bytes.Buffer
	if err := tmpl.Execute(&body, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	htmlContent := body.String()
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
		Subject:     "Client Contract Reminder",
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
