package notification

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// DiscordSender sends notifications via Discord webhooks.
type DiscordSender struct {
	webhookURL string
	httpClient *http.Client
}

// discordMessage represents the payload structure for Discord webhook messages.
type discordMessage struct {
	Content string         `json:"content,omitempty"`
	Embeds  []discordEmbed `json:"embeds,omitempty"`
}

// discordEmbed represents an embed in a Discord message.
type discordEmbed struct {
	Title       string              `json:"title,omitempty"`
	Description string              `json:"description,omitempty"`
	Color       int                 `json:"color,omitempty"`
	Fields      []discordEmbedField `json:"fields,omitempty"`
	Timestamp   string              `json:"timestamp,omitempty"`
}

// discordEmbedField represents a field in a Discord embed.
type discordEmbedField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline,omitempty"`
}

// NewDiscordSender creates a new Discord notification sender.
// Returns a NoOpSender if the webhook URL is empty.
func NewDiscordSender(webhookURL string) Sender {
	if webhookURL == "" {
		return NewNoOpSender()
	}

	return &DiscordSender{
		webhookURL: webhookURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Send sends a simple text message to Discord.
func (d *DiscordSender) Send(message string) error {
	payload := discordMessage{
		Content: message,
	}

	return d.sendPayload(payload)
}

// SendGameNotification sends a formatted game notification to Discord.
func (d *DiscordSender) SendGameNotification(game GameInfo, eventType string) error {
	// Color codes for Discord embeds (decimal)
	// Green for game scheduled, Blue for game started, etc.
	colorMap := map[string]int{
		"scheduled": 3066993,  // Green
		"started":   3447003,  // Blue
		"ended":     15158332, // Red
		"reminder":  16776960, // Yellow
	}

	color, ok := colorMap[eventType]
	if !ok {
		color = 9807270 // Gray for unknown event types
	}

	embed := discordEmbed{
		Title:       fmt.Sprintf("NHL Game %s", capitalizeFirst(eventType)),
		Description: fmt.Sprintf("%s vs %s", game.AwayTeam, game.HomeTeam),
		Color:       color,
		Fields: []discordEmbedField{
			{
				Name:   "Game ID",
				Value:  game.ID,
				Inline: true,
			},
			{
				Name:   "Date",
				Value:  game.GameDate,
				Inline: true,
			},
			{
				Name:   "Start Time",
				Value:  game.StartTime,
				Inline: true,
			},
		},
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	payload := discordMessage{
		Embeds: []discordEmbed{embed},
	}

	return d.sendPayload(payload)
}

// IsEnabled returns true if the Discord sender has a configured webhook URL.
func (d *DiscordSender) IsEnabled() bool {
	return d.webhookURL != ""
}

// sendPayload sends a Discord message payload to the webhook URL.
func (d *DiscordSender) sendPayload(payload discordMessage) error {
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal Discord payload: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, d.webhookURL, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to create Discord request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send Discord notification: %w", err)
	}
	defer resp.Body.Close()

	// Discord returns 204 No Content on success
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("Discord webhook returned status %d", resp.StatusCode)
	}

	return nil
}

// capitalizeFirst capitalizes the first letter of a string.
func capitalizeFirst(s string) string {
	if s == "" {
		return s
	}
	if len(s) == 1 {
		return string(s[0]-32) // Convert to uppercase for ASCII letters
	}
	if s[0] >= 'a' && s[0] <= 'z' {
		return string(s[0]-32) + s[1:]
	}
	return s
}
