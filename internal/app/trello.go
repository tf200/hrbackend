package app

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"hrbackend/config"
	"hrbackend/internal/domain"
)

const trelloCardsEndpoint = "https://api.trello.com/1/cards"

type trelloBugReportPublisher struct {
	apiKey string
	token  string
	listID string
	client *http.Client
}

type trelloCreateCardResponse struct {
	ID       string `json:"id"`
	URL      string `json:"url"`
	ShortURL string `json:"shortUrl"`
}

func newTrelloBugReportPublisher(cfg config.Config) domain.BugReportCardPublisher {
	apiKey := strings.TrimSpace(cfg.TrelloAPIKey)
	token := strings.TrimSpace(cfg.TrelloToken)
	listID := strings.TrimSpace(cfg.TrelloListID)
	if apiKey == "" || token == "" || listID == "" {
		return nil
	}

	return &trelloBugReportPublisher{
		apiKey: apiKey,
		token:  token,
		listID: listID,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (p *trelloBugReportPublisher) CreateBugReportCard(
	ctx context.Context,
	report domain.BugReport,
) (*domain.BugReportCard, error) {
	values := url.Values{}
	values.Set("key", p.apiKey)
	values.Set("token", p.token)
	values.Set("idList", p.listID)
	values.Set("name", trelloBugReportCardName(report))
	values.Set("desc", trelloBugReportCardDescription(report))

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		trelloCardsEndpoint,
		strings.NewReader(values.Encode()),
	)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("trello create card failed: status=%d body=%s", resp.StatusCode, string(body))
	}

	var result trelloCreateCardResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	cardURL := result.URL
	if cardURL == "" {
		cardURL = result.ShortURL
	}
	return &domain.BugReportCard{ID: result.ID, URL: cardURL}, nil
}

func trelloBugReportCardName(report domain.BugReport) string {
	return fmt.Sprintf("[%s] %s", report.Severity, report.Subject)
}

func trelloBugReportCardDescription(report domain.BugReport) string {
	var builder strings.Builder
	fmt.Fprintf(&builder, "Bug Report ID: %s\n", report.ID)
	fmt.Fprintf(&builder, "User ID: %s\n", report.UserID)
	fmt.Fprintf(&builder, "Category: %s\n", report.Category)
	fmt.Fprintf(&builder, "Severity: %s\n", report.Severity)
	fmt.Fprintf(&builder, "Status: %s\n", report.Status)
	fmt.Fprintf(&builder, "Created At: %s\n\n", report.CreatedAt.Format(time.RFC3339))
	fmt.Fprintf(&builder, "Description:\n%s\n", report.Description)

	if report.Steps != nil && strings.TrimSpace(*report.Steps) != "" {
		fmt.Fprintf(&builder, "\nSteps:\n%s\n", *report.Steps)
	}

	debugInfo := strings.TrimSpace(formatBugReportDebugInfo(report.DebugInfo))
	if debugInfo != "" {
		fmt.Fprintf(&builder, "\nDebug Info:\n```json\n%s\n```\n", debugInfo)
	}

	return builder.String()
}

func formatBugReportDebugInfo(raw json.RawMessage) string {
	if len(raw) == 0 {
		return "{}"
	}

	var output bytes.Buffer
	if err := json.Indent(&output, raw, "", "  "); err != nil {
		return string(raw)
	}
	return output.String()
}
