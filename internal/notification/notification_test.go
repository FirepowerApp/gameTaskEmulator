package notification

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
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

func TestDiscordSender_SendScheduleSummary_NoExtraContentField(t *testing.T) {
	var received discordMessage

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&received)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	s := NewDiscordSender(server.URL)
	s.SendScheduleSummary([]GameInfo{{ID: "1", HomeTeam: "BOS", AwayTeam: "DAL"}})

	// SendScheduleSummary should use embeds, not content
	if received.Content != "" {
		t.Errorf("SendScheduleSummary set content field = %q, want empty", received.Content)
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

// --- Exact payload formatting tests ---
// These verify the precise description string to catch formatting regressions
// (ordering, newlines, separators, markdown).

func TestDiscordSender_SendScheduleSummary_ExactDescription_OneGame(t *testing.T) {
	var received discordMessage

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&received)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	games := []GameInfo{
		{ID: "2024020415", GameDate: "2025-01-15", StartTime: "2025-01-16T00:00:00Z", HomeTeam: "BOS", AwayTeam: "DAL"},
	}

	s := NewDiscordSender(server.URL)
	if err := s.SendScheduleSummary(games); err != nil {
		t.Fatalf("SendScheduleSummary() returned error: %v", err)
	}

	wantDescription := "**DAL @ BOS**\n2025-01-15 at 2025-01-16T00:00:00Z\n\n"
	if received.Embeds[0].Description != wantDescription {
		t.Errorf("description mismatch:\n got: %q\nwant: %q", received.Embeds[0].Description, wantDescription)
	}
}

func TestDiscordSender_SendScheduleSummary_ExactDescription_MultipleGames(t *testing.T) {
	var received discordMessage

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&received)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	games := []GameInfo{
		{ID: "2024020415", GameDate: "2025-01-15", StartTime: "2025-01-16T00:00:00Z", HomeTeam: "BOS", AwayTeam: "DAL"},
		{ID: "2024020416", GameDate: "2025-01-15", StartTime: "2025-01-16T01:00:00Z", HomeTeam: "NYR", AwayTeam: "CHI"},
		{ID: "2024020417", GameDate: "2025-01-15", StartTime: "2025-01-16T03:30:00Z", HomeTeam: "LAK", AwayTeam: "SEA"},
	}

	s := NewDiscordSender(server.URL)
	if err := s.SendScheduleSummary(games); err != nil {
		t.Fatalf("SendScheduleSummary() returned error: %v", err)
	}

	wantDescription := "" +
		"**DAL @ BOS**\n2025-01-15 at 2025-01-16T00:00:00Z\n\n" +
		"**CHI @ NYR**\n2025-01-15 at 2025-01-16T01:00:00Z\n\n" +
		"**SEA @ LAK**\n2025-01-15 at 2025-01-16T03:30:00Z\n\n"
	if received.Embeds[0].Description != wantDescription {
		t.Errorf("description mismatch:\n got: %q\nwant: %q", received.Embeds[0].Description, wantDescription)
	}
}

func TestDiscordSender_SendScheduleSummary_ExactPayload_NoGames(t *testing.T) {
	var received discordMessage

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&received)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	s := NewDiscordSender(server.URL)
	if err := s.SendScheduleSummary(nil); err != nil {
		t.Fatalf("SendScheduleSummary(nil) returned error: %v", err)
	}

	if received.Content != "" {
		t.Errorf("content = %q, want empty", received.Content)
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
		t.Errorf("color = %d, want 9807270 (gray)", embed.Color)
	}
	if len(embed.Fields) != 0 {
		t.Errorf("fields count = %d, want 0", len(embed.Fields))
	}
}

func TestDiscordSender_Send_ExactPayload(t *testing.T) {
	var received discordMessage

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&received)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	s := NewDiscordSender(server.URL)
	if err := s.Send("Game Task Emulator run complete."); err != nil {
		t.Fatalf("Send() returned error: %v", err)
	}

	if received.Content != "Game Task Emulator run complete." {
		t.Errorf("content = %q, want %q", received.Content, "Game Task Emulator run complete.")
	}
	if len(received.Embeds) != 0 {
		t.Errorf("embeds count = %d, want 0", len(received.Embeds))
	}
}

// --- Game ordering preservation ---

func TestDiscordSender_SendScheduleSummary_PreservesGameOrder(t *testing.T) {
	var received discordMessage

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&received)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	// Deliberately pass games in non-chronological order
	games := []GameInfo{
		{ID: "3", GameDate: "2025-01-15", StartTime: "2025-01-16T03:00:00Z", HomeTeam: "LAK", AwayTeam: "SEA"},
		{ID: "1", GameDate: "2025-01-15", StartTime: "2025-01-16T00:00:00Z", HomeTeam: "BOS", AwayTeam: "DAL"},
		{ID: "2", GameDate: "2025-01-15", StartTime: "2025-01-16T01:00:00Z", HomeTeam: "NYR", AwayTeam: "CHI"},
	}

	s := NewDiscordSender(server.URL)
	s.SendScheduleSummary(games)

	desc := received.Embeds[0].Description
	// The first game in the input should appear first in the output
	lakIdx := strings.Index(desc, "SEA @ LAK")
	bosIdx := strings.Index(desc, "DAL @ BOS")
	nyrIdx := strings.Index(desc, "CHI @ NYR")

	if lakIdx == -1 || bosIdx == -1 || nyrIdx == -1 {
		t.Fatalf("missing matchups in description: %q", desc)
	}
	if lakIdx > bosIdx || bosIdx > nyrIdx {
		t.Errorf("games not in input order: LAK@%d BOS@%d NYR@%d\ndescription: %q", lakIdx, bosIdx, nyrIdx, desc)
	}
}
