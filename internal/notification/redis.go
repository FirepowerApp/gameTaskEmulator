package notification

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisSender sends notifications to a Redis queue.
type RedisSender struct {
	client    *redis.Client
	queueName string
}

// RedisMessage represents a message structure for Redis queue.
type RedisMessage struct {
	Type      string     `json:"type"`
	Message   string     `json:"message,omitempty"`
	Games     []GameInfo `json:"games,omitempty"`
	Timestamp string     `json:"timestamp"`
}

// NewRedisSender creates a new Redis notification sender.
// Returns a NoOpSender if the Redis URL is empty.
func NewRedisSender(redisURL, queueName string) Sender {
	if redisURL == "" {
		return NewNoOpSender()
	}

	if queueName == "" {
		queueName = "game-notifications"
	}

	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		fmt.Printf("Warning: Failed to parse Redis URL, notifications disabled: %v\n", err)
		return NewNoOpSender()
	}

	client := redis.NewClient(opt)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		fmt.Printf("Warning: Failed to connect to Redis, notifications disabled: %v\n", err)
		return NewNoOpSender()
	}

	return &RedisSender{
		client:    client,
		queueName: queueName,
	}
}

// Send sends a simple text message to Redis queue.
func (r *RedisSender) Send(message string) error {
	msg := RedisMessage{
		Type:      "simple",
		Message:   message,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	return r.pushToQueue(msg)
}

// SendScheduleSummary sends a summary of all scheduled games to Redis queue.
// If no games were scheduled, sends a message indicating that.
func (r *RedisSender) SendScheduleSummary(games []GameInfo) error {
	var messageType string
	if len(games) == 0 {
		messageType = "no_games"
	} else {
		messageType = "schedule_summary"
	}

	msg := RedisMessage{
		Type:      messageType,
		Games:     games,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	return r.pushToQueue(msg)
}

// IsEnabled returns true if the Redis sender has a configured client.
func (r *RedisSender) IsEnabled() bool {
	return r.client != nil
}

// pushToQueue pushes a message to the Redis queue using RPUSH.
func (r *RedisSender) pushToQueue(msg RedisMessage) error {
	jsonPayload, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal Redis message: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := r.client.RPush(ctx, r.queueName, jsonPayload).Err(); err != nil {
		return fmt.Errorf("failed to push message to Redis queue: %w", err)
	}

	return nil
}

// Close closes the Redis connection.
func (r *RedisSender) Close() error {
	if r.client != nil {
		return r.client.Close()
	}
	return nil
}
