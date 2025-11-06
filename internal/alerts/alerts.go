package alerts

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/rgnote/TheneKunda/internal/config"
	"github.com/rgnote/TheneKunda/internal/logger"
)

// AlertManager handles sending alerts
type AlertManager struct {
	config config.AlertsConfig
	logger *logger.Logger
	client *http.Client
}

// New creates a new alert manager
func New(cfg config.AlertsConfig, log *logger.Logger) *AlertManager {
	return &AlertManager{
		config: cfg,
		logger: log,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Start starts the alert manager
func (a *AlertManager) Start(eventChan <-chan logger.Event) {
	go a.processEvents(eventChan)
}

func (a *AlertManager) processEvents(eventChan <-chan logger.Event) {
	for event := range eventChan {
		// Send webhook if enabled
		if a.config.Webhook.Enabled && a.config.Webhook.URL != "" {
			go a.sendWebhook(event)
		}

		// Console alerts are handled by the logger itself
	}
}

func (a *AlertManager) sendWebhook(event logger.Event) {
	payload, err := json.Marshal(event)
	if err != nil {
		a.logger.Error("Failed to marshal webhook payload: %v", err)
		return
	}

	req, err := http.NewRequest("POST", a.config.Webhook.URL, bytes.NewBuffer(payload))
	if err != nil {
		a.logger.Error("Failed to create webhook request: %v", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "TheneKunda-Honeypot/1.0")

	resp, err := a.client.Do(req)
	if err != nil {
		a.logger.Error("Failed to send webhook: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		a.logger.Warning("Webhook returned non-2xx status: %d", resp.StatusCode)
	}
}

// WebhookPayload represents the structure sent to webhooks
type WebhookPayload struct {
	Event     logger.Event `json:"event"`
	Honeypot  string       `json:"honeypot"`
	Timestamp string       `json:"timestamp"`
}

// FormatSlackMessage formats an event for Slack webhooks
func FormatSlackMessage(event logger.Event) map[string]interface{} {
	color := "warning"
	if event.EventType == "command" {
		color = "danger"
	}

	var text string
	switch event.EventType {
	case "connection":
		text = fmt.Sprintf("*[%s]* Connection attempt from `%s`", event.Service, event.SourceIP)
		if event.Username != "" {
			text += fmt.Sprintf("\n• Username: `%s`", event.Username)
		}
		if event.Password != "" {
			text += fmt.Sprintf("\n• Password: `%s`", event.Password)
		}
	case "command":
		text = fmt.Sprintf("*[%s]* Command executed from `%s`\n• Command: `%s`",
			event.Service, event.SourceIP, event.Command)
	default:
		text = fmt.Sprintf("*[%s]* %s from `%s`", event.Service, event.Message, event.SourceIP)
	}

	return map[string]interface{}{
		"attachments": []map[string]interface{}{
			{
				"color":      color,
				"text":       text,
				"footer":     "TheneKunda Honeypot",
				"footer_icon": "🍯",
				"ts":         time.Now().Unix(),
			},
		},
	}
}
