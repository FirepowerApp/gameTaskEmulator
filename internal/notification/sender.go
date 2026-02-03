// Package notification provides interfaces and implementations for sending notifications.
package notification

// GameInfo contains information about a game for notifications.
type GameInfo struct {
	ID        string
	GameDate  string
	StartTime string
	HomeTeam  string
	AwayTeam  string
}

// Sender defines the interface for sending notifications.
// Implementations of this interface can send notifications via different channels
// such as Discord, Slack, email, etc.
type Sender interface {
	// Send sends a notification message.
	// Returns an error if the notification could not be sent.
	Send(message string) error

	// SendScheduleSummary sends a summary notification of all scheduled games.
	// If games is empty, sends a message indicating no games were scheduled.
	// Returns an error if the notification could not be sent.
	SendScheduleSummary(games []GameInfo) error

	// IsEnabled returns whether the notification sender is configured and enabled.
	IsEnabled() bool
}

// NoOpSender is a notification sender that does nothing.
// It is used when notifications are disabled.
type NoOpSender struct{}

// Send does nothing and returns nil.
func (n *NoOpSender) Send(message string) error {
	return nil
}

// SendScheduleSummary does nothing and returns nil.
func (n *NoOpSender) SendScheduleSummary(games []GameInfo) error {
	return nil
}

// IsEnabled always returns false for the no-op sender.
func (n *NoOpSender) IsEnabled() bool {
	return false
}

// NewNoOpSender creates a new no-op notification sender.
func NewNoOpSender() Sender {
	return &NoOpSender{}
}
