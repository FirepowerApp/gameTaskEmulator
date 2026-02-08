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
	userID     string
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

// DiscordOption configures a DiscordSender.
type DiscordOption func(*DiscordSender)

// WithUserID sets the Discord user ID for @ mentions in notifications.
func WithUserID(userID string) DiscordOption {
	return func(d *DiscordSender) {
		d.userID = userID
	}
}

// NewDiscordSender creates a new Discord notification sender.
// Returns a NoOpSender if the webhook URL is empty.
// Use WithUserID option to enable @ mentions in notifications.
func NewDiscordSender(webhookURL string, opts ...DiscordOption) Sender {
	if webhookURL == "" {
		return NewNoOpSender()
	}

	d := &DiscordSender{
		webhookURL: webhookURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}

	for _, opt := range opts {
		opt(d)
	}

	return d
}

// Send sends a simple text message to Discord.
func (d *DiscordSender) Send(message string) error {
	payload := discordMessage{
		Content: message,
	}

	return d.sendPayload(payload)
}

// SendScheduleSummary sends a summary of all scheduled games to Discord.
// If no games were scheduled, sends a message indicating that.
func (d *DiscordSender) SendScheduleSummary(games []GameInfo) error {
	var embed discordEmbed

	if len(games) == 0 {
		embed = discordEmbed{
			Title:       "NHL Game Schedule",
			Description: "No games were identified to schedule.",
			Color:       9807270, // Gray
			Timestamp:   time.Now().UTC().Format(time.RFC3339),
		}
	} else {
		// Build description with all games
		var description string
		for _, game := range games {
			description += fmt.Sprintf("**%s @ %s**\n%s at %s\n\n",
				game.AwayTeam, game.HomeTeam, game.GameDate, game.StartTime)
		}

		// Add user mention at the end of description if configured
		if d.userID != "" {
			description += fmt.Sprintf("<@%s>", d.userID)
		}

		title := fmt.Sprintf("NHL Game Schedule (%d game", len(games))
		if len(games) != 1 {
			title += "s"
		}
		title += " scheduled)"

		embed = discordEmbed{
			Title:       title,
			Description: description,
			Color:       3066993, // Green
			Timestamp:   time.Now().UTC().Format(time.RFC3339),
		}
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
