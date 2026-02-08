package notification

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// --- NoOpSender tests ---

func TestNoOpSender_IsEnabled(t *testing.T) {
	s := NewNoOpSender()
	if s.IsEnabled() {
		t.Error("NoOpSender.IsEnabled() = true, want false")
	}
}

func TestNoOpSender_Send(t *testing.T) {
	s := NewNoOpSender()
	if err := s.Send("test"); err != nil {
		t.Errorf("NoOpSender.Send() returned error: %v", err)
	}
}

func TestNoOpSender_SendScheduleSummary(t *testing.T) {
	s := NewNoOpSender()
	games := []GameInfo{{ID: "1", HomeTeam: "BOS", AwayTeam: "DAL"}}
	if err := s.SendScheduleSummary(games); err != nil {
		t.Errorf("NoOpSender.SendScheduleSummary() returned error: %v", err)
	}
}

// --- NewDiscordSender constructor tests ---

func TestNewDiscordSender_EmptyURL(t *testing.T) {
	s := NewDiscordSender("")
	if s.IsEnabled() {
		t.Error("NewDiscordSender(\"\").IsEnabled() = true, want false (should return NoOpSender)")
	}
	// Verify it's actually a *NoOpSender
	if _, ok := s.(*NoOpSender); !ok {
		t.Errorf("NewDiscordSender(\"\") returned %T, want *NoOpSender", s)
	}
}

func TestNewDiscordSender_WithURL(t *testing.T) {
	s := NewDiscordSender("https://discord.com/api/webhooks/test")
	if !s.IsEnabled() {
		t.Error("NewDiscordSender(url).IsEnabled() = false, want true")
	}
	if _, ok := s.(*DiscordSender); !ok {
		t.Errorf("NewDiscordSender(url) returned %T, want *DiscordSender", s)
	}
}

// --- Discord Send tests ---

func TestDiscordSender_Send(t *testing.T) {
	var received discordMessage

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("Content-Type = %q, want %q", ct, "application/json")
		}
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	s := NewDiscordSender(server.URL)
	if err := s.Send("hello world"); err != nil {
		t.Fatalf("Send() returned error: %v", err)
	}

	if received.Content != "hello world" {
		t.Errorf("payload content = %q, want %q", received.Content, "hello world")
	}
	if len(received.Embeds) != 0 {
		t.Errorf("payload embeds count = %d, want 0", len(received.Embeds))
	}
}

// --- SendScheduleSummary tests ---

func TestDiscordSender_SendScheduleSummary_NoGames(t *testing.T) {
	var received discordMessage

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	s := NewDiscordSender(server.URL)
	if err := s.SendScheduleSummary(nil); err != nil {
		t.Fatalf("SendScheduleSummary(nil) returned error: %v", err)
	}

	if len(received.Embeds) != 1 {
		t.Fatalf("embed count = %d, want 1", len(received.Embeds))
	}

	embed := received.Embeds[0]
	if embed.Title != "NHL Game Schedule" {
		t.Errorf("title = %q, want %q", embed.Title, "NHL Game Schedule")
	}
	if embed.Description != "No games were identified to schedule." {
		t.Errorf("description = %q, want %q", embed.Description, "No games were identified to schedule.")
	}
	if embed.Color != 9807270 {
		t.Errorf("color = %d, want %d (gray)", embed.Color, 9807270)
	}
	if embed.Timestamp == "" {
		t.Error("timestamp is empty, want RFC3339 value")
	}
}

func TestDiscordSender_SendScheduleSummary_EmptySlice(t *testing.T) {
	var received discordMessage

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&received)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	s := NewDiscordSender(server.URL)
	if err := s.SendScheduleSummary([]GameInfo{}); err != nil {
		t.Fatalf("SendScheduleSummary([]) returned error: %v", err)
	}

	if len(received.Embeds) != 1 {
		t.Fatalf("embed count = %d, want 1", len(received.Embeds))
	}
	if received.Embeds[0].Description != "No games were identified to schedule." {
		t.Errorf("description = %q, want no-games message", received.Embeds[0].Description)
	}
}

func TestDiscordSender_SendScheduleSummary_OneGame(t *testing.T) {
	var received discordMessage

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&received)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	games := []GameInfo{
		{
			ID:        "2024020001",
			GameDate:  "2024-11-15",
			StartTime: "2024-11-15T00:00:00Z",
			HomeTeam:  "BOS",
			AwayTeam:  "DAL",
		},
	}

	s := NewDiscordSender(server.URL)
	if err := s.SendScheduleSummary(games); err != nil {
		t.Fatalf("SendScheduleSummary() returned error: %v", err)
	}

	if len(received.Embeds) != 1 {
		t.Fatalf("embed count = %d, want 1", len(received.Embeds))
	}

	embed := received.Embeds[0]

	// Singular "game" for 1 game
	expectedTitle := "NHL Game Schedule (1 game scheduled)"
	if embed.Title != expectedTitle {
		t.Errorf("title = %q, want %q", embed.Title, expectedTitle)
	}

	if embed.Color != 3066993 {
		t.Errorf("color = %d, want %d (green)", embed.Color, 3066993)
	}

	// Verify game details appear in description
	if !strings.Contains(embed.Description, "DAL @ BOS") {
		t.Errorf("description missing matchup, got: %q", embed.Description)
	}
	if !strings.Contains(embed.Description, "2024-11-15") {
		t.Errorf("description missing game date, got: %q", embed.Description)
	}
	if !strings.Contains(embed.Description, "2024-11-15T00:00:00Z") {
		t.Errorf("description missing start time, got: %q", embed.Description)
	}
	if embed.Timestamp == "" {
		t.Error("timestamp is empty, want RFC3339 value")
	}
}

func TestDiscordSender_SendScheduleSummary_MultipleGames(t *testing.T) {
	var received discordMessage

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&received)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	games := []GameInfo{
		{
			ID:        "2024020001",
			GameDate:  "2024-11-15",
			StartTime: "2024-11-15T19:00:00Z",
			HomeTeam:  "BOS",
			AwayTeam:  "DAL",
		},
		{
			ID:        "2024020002",
			GameDate:  "2024-11-15",
			StartTime: "2024-11-15T20:00:00Z",
			HomeTeam:  "NYR",
			AwayTeam:  "CHI",
		},
		{
			ID:        "2024020003",
			GameDate:  "2024-11-15",
			StartTime: "2024-11-15T22:30:00Z",
			HomeTeam:  "LAK",
			AwayTeam:  "SEA",
		},
	}

	s := NewDiscordSender(server.URL)
	if err := s.SendScheduleSummary(games); err != nil {
		t.Fatalf("SendScheduleSummary() returned error: %v", err)
	}

	embed := received.Embeds[0]

	// Plural "games" for multiple
	expectedTitle := "NHL Game Schedule (3 games scheduled)"
	if embed.Title != expectedTitle {
		t.Errorf("title = %q, want %q", embed.Title, expectedTitle)
	}

	// All matchups present
	for _, matchup := range []string{"DAL @ BOS", "CHI @ NYR", "SEA @ LAK"} {
		if !strings.Contains(embed.Description, matchup) {
			t.Errorf("description missing matchup %q, got:\n%s", matchup, embed.Description)
		}
	}

	// All start times present
	for _, startTime := range []string{"19:00:00Z", "20:00:00Z", "22:30:00Z"} {
		if !strings.Contains(embed.Description, startTime) {
			t.Errorf("description missing start time %q, got:\n%s", startTime, embed.Description)
		}
	}
}

// --- HTTP response handling tests ---

func TestDiscordSender_Send_HTTP200(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	s := NewDiscordSender(server.URL)
	if err := s.Send("test"); err != nil {
		t.Errorf("Send() with HTTP 200 returned error: %v", err)
	}
}

func TestDiscordSender_Send_HTTP204(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	s := NewDiscordSender(server.URL)
	if err := s.Send("test"); err != nil {
		t.Errorf("Send() with HTTP 204 returned error: %v", err)
	}
}

func TestDiscordSender_Send_HTTP500(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	s := NewDiscordSender(server.URL)
	err := s.Send("test")
	if err == nil {
		t.Fatal("Send() with HTTP 500 returned nil error, want error")
	}
	if !strings.Contains(err.Error(), "500") {
		t.Errorf("error = %q, want it to contain status code 500", err.Error())
	}
}

func TestDiscordSender_Send_HTTP403(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer server.Close()

	s := NewDiscordSender(server.URL)
	err := s.Send("test")
	if err == nil {
		t.Fatal("Send() with HTTP 403 returned nil error, want error")
	}
	if !strings.Contains(err.Error(), "403") {
		t.Errorf("error = %q, want it to contain status code 403", err.Error())
	}
}

func TestDiscordSender_Send_HTTP429(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	s := NewDiscordSender(server.URL)
	err := s.Send("test")
	if err == nil {
		t.Fatal("Send() with HTTP 429 returned nil error, want error")
	}
	if !strings.Contains(err.Error(), "429") {
		t.Errorf("error = %q, want it to contain status code 429", err.Error())
	}
}

func TestDiscordSender_Send_ConnectionRefused(t *testing.T) {
	// Use a URL with a port that's definitely not listening
	s := NewDiscordSender("http://127.0.0.1:1")
	err := s.Send("test")
	if err == nil {
		t.Fatal("Send() to unreachable server returned nil error, want error")
	}
	if !strings.Contains(err.Error(), "failed to send Discord notification") {
		t.Errorf("error = %q, want it to wrap with 'failed to send Discord notification'", err.Error())
	}
}

func TestDiscordSender_SendScheduleSummary_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	s := NewDiscordSender(server.URL)
	games := []GameInfo{{ID: "1", GameDate: "2024-01-01", StartTime: "2024-01-01T19:00:00Z", HomeTeam: "BOS", AwayTeam: "DAL"}}
	err := s.SendScheduleSummary(games)
	if err == nil {
		t.Fatal("SendScheduleSummary() with server error returned nil, want error")
	}
}

// --- Payload structure validation ---

func TestDiscordSender_SendScheduleSummary_PayloadIsValidJSON(t *testing.T) {
	var rawBody []byte

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf := make([]byte, r.ContentLength)
		r.Body.Read(buf)
		rawBody = buf
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	games := []GameInfo{
		{ID: "1", GameDate: "2024-01-01", StartTime: "2024-01-01T19:00:00Z", HomeTeam: "BOS", AwayTeam: "DAL"},
	}

	s := NewDiscordSender(server.URL)
	s.SendScheduleSummary(games)

	// Verify the raw body is valid JSON
	if !json.Valid(rawBody) {
		t.Errorf("payload is not valid JSON: %s", string(rawBody))
	}

	// Verify it round-trips cleanly through discordMessage
	var msg discordMessage
	if err := json.Unmarshal(rawBody, &msg); err != nil {
		t.Fatalf("failed to unmarshal payload: %v", err)
	}
	if len(msg.Embeds) != 1 {
		t.Errorf("expected 1 embed, got %d", len(msg.Embeds))
	}
}

func TestDiscordSender_SendScheduleSummary_NoContentWithoutUserID(t *testing.T) {
	var received discordMessage

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&received)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	s := NewDiscordSender(server.URL)
	s.SendScheduleSummary([]GameInfo{{ID: "1", HomeTeam: "BOS", AwayTeam: "DAL"}})

	// Without a user ID, content should be empty
	if received.Content != "" {
		t.Errorf("SendScheduleSummary set content field = %q, want empty", received.Content)
	}
}

func TestDiscordSender_SendScheduleSummary_MentionWithUserID(t *testing.T) {
	var received discordMessage

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&received)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	s := NewDiscordSender(server.URL, WithUserID("417487003588755480"))
	s.SendScheduleSummary([]GameInfo{{ID: "1", GameDate: "2024-01-01", StartTime: "2024-01-01T19:00:00Z", HomeTeam: "BOS", AwayTeam: "DAL"}})

	// With a user ID, the mention should be at the end of the embed description
	expectedMention := "<@417487003588755480>"
	if !strings.HasSuffix(received.Embeds[0].Description, expectedMention) {
		t.Errorf("SendScheduleSummary description should end with mention, got: %q", received.Embeds[0].Description)
	}

	// Content should be empty (mention is in embed, not content)
	if received.Content != "" {
		t.Errorf("SendScheduleSummary content = %q, want empty", received.Content)
	}
}

// --- Interface compliance ---

func TestSenderInterfaceCompliance(t *testing.T) {
	// Compile-time check that both types satisfy the Sender interface
	var _ Sender = &NoOpSender{}
	var _ Sender = &DiscordSender{}
}

// --- Two-game plural boundary ---

func TestDiscordSender_SendScheduleSummary_TwoGames(t *testing.T) {
	var received discordMessage

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&received)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	games := []GameInfo{
		{ID: "1", GameDate: "2024-01-01", StartTime: "2024-01-01T19:00:00Z", HomeTeam: "BOS", AwayTeam: "DAL"},
		{ID: "2", GameDate: "2024-01-01", StartTime: "2024-01-01T20:00:00Z", HomeTeam: "NYR", AwayTeam: "CHI"},
	}

	s := NewDiscordSender(server.URL)
	s.SendScheduleSummary(games)

	expected := "NHL Game Schedule (2 games scheduled)"
	if received.Embeds[0].Title != expected {
		t.Errorf("title = %q, want %q", received.Embeds[0].Title, expected)
	}
}

// --- Comprehensive end-to-end webhook payload validation ---

// TestDiscordWebhook_FullPayloadValidation_WithGamesAndMention validates the complete
// Discord webhook payload structure when games are scheduled and a user ID is configured.
// This test simulates a real Discord webhook server and verifies every field.
func TestDiscordWebhook_FullPayloadValidation_WithGamesAndMention(t *testing.T) {
	var received discordMessage
	var rawBody []byte
	var requestMethod string
	var contentType string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestMethod = r.Method
		contentType = r.Header.Get("Content-Type")

		// Capture raw body for JSON validation
		buf := make([]byte, r.ContentLength)
		r.Body.Read(buf)
		rawBody = buf

		// Also decode into struct
		json.Unmarshal(rawBody, &received)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	games := []GameInfo{
		{
			ID:        "2024020100",
			GameDate:  "2024-12-25",
			StartTime: "2024-12-25T18:00:00Z",
			HomeTeam:  "BOS",
			AwayTeam:  "CHI",
		},
		{
			ID:        "2024020101",
			GameDate:  "2024-12-25",
			StartTime: "2024-12-25T20:30:00Z",
			HomeTeam:  "NYR",
			AwayTeam:  "DAL",
		},
	}
	sender := NewDiscordSender(server.URL, WithUserID("417487003588755480"))
	err := sender.SendScheduleSummary(games)

	// --- Request validation ---
	if err != nil {
		t.Fatalf("SendScheduleSummary returned error: %v", err)
	}
	if requestMethod != http.MethodPost {
		t.Errorf("HTTP method = %q, want POST", requestMethod)
	}
	if contentType != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", contentType)
	}

	// --- JSON validity ---
	if !json.Valid(rawBody) {
		t.Fatalf("payload is not valid JSON: %s", string(rawBody))
	}

	// --- Top-level structure ---
	if received.Content != "" {
		t.Errorf("content = %q, want empty (mention should be in embed description)", received.Content)
	}
	if len(received.Embeds) != 1 {
		t.Fatalf("embed count = %d, want 1", len(received.Embeds))
	}

	embed := received.Embeds[0]

	// --- Embed title ---
	expectedTitle := "NHL Game Schedule (2 games scheduled)"
	if embed.Title != expectedTitle {
		t.Errorf("embed.Title = %q, want %q", embed.Title, expectedTitle)
	}

	// --- Embed color (green for games scheduled) ---
	if embed.Color != 3066993 {
		t.Errorf("embed.Color = %d, want 3066993 (green)", embed.Color)
	}

	// --- Embed timestamp (must be valid RFC3339) ---
	if embed.Timestamp == "" {
		t.Error("embed.Timestamp is empty, want RFC3339 value")
	} else {
		if _, err := time.Parse(time.RFC3339, embed.Timestamp); err != nil {
			t.Errorf("embed.Timestamp = %q is not valid RFC3339: %v", embed.Timestamp, err)
		}
	}

	// --- Embed description: verify all game details ---
	desc := embed.Description

	// Game 1: CHI @ BOS
	if !strings.Contains(desc, "**CHI @ BOS**") {
		t.Errorf("description missing '**CHI @ BOS**' matchup formatting, got:\n%s", desc)
	}
	if !strings.Contains(desc, "2024-12-25") {
		t.Errorf("description missing game date '2024-12-25', got:\n%s", desc)
	}
	if !strings.Contains(desc, "2024-12-25T18:00:00Z") {
		t.Errorf("description missing start time '2024-12-25T18:00:00Z', got:\n%s", desc)
	}

	// Game 2: DAL @ NYR
	if !strings.Contains(desc, "**DAL @ NYR**") {
		t.Errorf("description missing '**DAL @ NYR**' matchup formatting, got:\n%s", desc)
	}
	if !strings.Contains(desc, "2024-12-25T20:30:00Z") {
		t.Errorf("description missing start time '2024-12-25T20:30:00Z', got:\n%s", desc)
	}

	// User mention at end
	expectedMention := "<@417487003588755480>"
	if !strings.HasSuffix(desc, expectedMention) {
		t.Errorf("description should end with mention %q, got:\n%s", expectedMention, desc)
	}

	// Verify mention appears only once
	if strings.Count(desc, expectedMention) != 1 {
		t.Errorf("mention should appear exactly once, found %d times in:\n%s", strings.Count(desc, expectedMention), desc)
	}
}

// TestDiscordWebhook_FullPayloadValidation_NoGames validates the complete
// Discord webhook payload structure when no games are scheduled.
func TestDiscordWebhook_FullPayloadValidation_NoGames(t *testing.T) {
	var received discordMessage
	var rawBody []byte

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf := make([]byte, r.ContentLength)
		r.Body.Read(buf)
		rawBody = buf
		json.Unmarshal(rawBody, &received)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	// Test with user ID - should NOT add mention when no games
	sender := NewDiscordSender(server.URL, WithUserID("417487003588755480"))
	err := sender.SendScheduleSummary([]GameInfo{})

	if err != nil {
		t.Fatalf("SendScheduleSummary returned error: %v", err)
	}

	// --- JSON validity ---
	if !json.Valid(rawBody) {
		t.Fatalf("payload is not valid JSON: %s", string(rawBody))
	}

	// --- Top-level structure ---
	if received.Content != "" {
		t.Errorf("content = %q, want empty", received.Content)
	}
	if len(received.Embeds) != 1 {
		t.Fatalf("embed count = %d, want 1", len(received.Embeds))
	}

	embed := received.Embeds[0]

	// --- Embed title (no game count) ---
	if embed.Title != "NHL Game Schedule" {
		t.Errorf("embed.Title = %q, want 'NHL Game Schedule'", embed.Title)
	}

	// --- Embed color (gray for no games) ---
	if embed.Color != 9807270 {
		t.Errorf("embed.Color = %d, want 9807270 (gray)", embed.Color)
	}

	// --- Embed description ---
	expectedDesc := "No games were identified to schedule."
	if embed.Description != expectedDesc {
		t.Errorf("embed.Description = %q, want %q", embed.Description, expectedDesc)
	}

	// --- Embed timestamp ---
	if embed.Timestamp == "" {
		t.Error("embed.Timestamp is empty, want RFC3339 value")
	}
}

// TestDiscordWebhook_FullPayloadValidation_SingleGame validates singular grammar
// and complete payload for a single game.
func TestDiscordWebhook_FullPayloadValidation_SingleGame(t *testing.T) {
	var received discordMessage

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&received)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	games := []GameInfo{
		{
			ID:        "2024020999",
			GameDate:  "2024-04-15",
			StartTime: "2024-04-15T19:30:00Z",
			HomeTeam:  "TOR",
			AwayTeam:  "MTL",
		},
	}

	sender := NewDiscordSender(server.URL)
	err := sender.SendScheduleSummary(games)

	if err != nil {
		t.Fatalf("SendScheduleSummary returned error: %v", err)
	}

	embed := received.Embeds[0]

	// --- Singular "game" not "games" ---
	expectedTitle := "NHL Game Schedule (1 game scheduled)"
	if embed.Title != expectedTitle {
		t.Errorf("embed.Title = %q, want %q (singular 'game')", embed.Title, expectedTitle)
	}

	// --- Green color for scheduled games ---
	if embed.Color != 3066993 {
		t.Errorf("embed.Color = %d, want 3066993 (green)", embed.Color)
	}

	// --- Description has all required parts ---
	desc := embed.Description
	requiredParts := []string{
		"**MTL @ TOR**",      // Matchup with bold formatting
		"2024-04-15",         // Game date
		"2024-04-15T19:30:00Z", // Start time
	}
	for _, part := range requiredParts {
		if !strings.Contains(desc, part) {
			t.Errorf("description missing %q, got:\n%s", part, desc)
		}
	}

	// --- No mention without user ID ---
	if strings.Contains(desc, "<@") {
		t.Errorf("description should not contain mention without user ID, got:\n%s", desc)
	}
}

// TestDiscordWebhook_DescriptionFormat validates the exact formatting of game entries
func TestDiscordWebhook_DescriptionFormat(t *testing.T) {
	var received discordMessage

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&received)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	games := []GameInfo{
		{
			ID:        "1",
			GameDate:  "2024-01-15",
			StartTime: "7:00 PM ET",
			HomeTeam:  "BOS",
			AwayTeam:  "NYR",
		},
	}

	sender := NewDiscordSender(server.URL)
	sender.SendScheduleSummary(games)

	desc := received.Embeds[0].Description

	// Verify exact format: "**AWAY @ HOME**\nDATE at TIME\n\n"
	expectedFormat := "**NYR @ BOS**\n2024-01-15 at 7:00 PM ET\n\n"
	if desc != expectedFormat {
		t.Errorf("description format incorrect.\ngot:\n%q\nwant:\n%q", desc, expectedFormat)
	}
}

// TestDiscordWebhook_MentionNotAddedToNoGamesMessage verifies mention is only added
// when there are games scheduled
func TestDiscordWebhook_MentionNotAddedToNoGamesMessage(t *testing.T) {
	var received discordMessage

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&received)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	// Even with user ID configured, no mention for empty games
	sender := NewDiscordSender(server.URL, WithUserID("123456789"))
	sender.SendScheduleSummary(nil)

	desc := received.Embeds[0].Description

	if strings.Contains(desc, "<@") {
		t.Errorf("no-games message should not contain mention, got:\n%s", desc)
	}
	if desc != "No games were identified to schedule." {
		t.Errorf("description = %q, want exact no-games message", desc)
	}
}
