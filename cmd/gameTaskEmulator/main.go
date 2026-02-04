// Package main implements a program that fetches NHL game schedules and creates
// Google Cloud Tasks for game tracking. It supports various command-line options
// for customizing team selection, date ranges, and task queue destinations.
package main

import (
	"context"
	"log"
	"strconv"

	"github.com/CrashTheCrease/backend/gameTaskEmulator/internal/notification"
)

func main() {
	config := parseFlags()

	log.Printf("Starting NHL Game Tracker Scheduler")
	log.Printf("Configuration: Date=%s, Teams=%v, TestMode=%t, AllTeams=%t, Today=%t, Production=%t",
		config.Date, config.Teams, config.TestMode, config.AllTeams, config.Today, config.Production)

	ctx := context.Background()

	// Initialize notification sender (dependency injection)
	var notifier notification.Sender = notification.NewDiscordSender(config.DiscordWebhookURL)
	if notifier.IsEnabled() {
		log.Printf("Discord notifications enabled")
	} else {
		log.Printf("Discord notifications disabled (no webhook URL configured)")
	}

	// Connect to Cloud Tasks service (emulator or production)
	client, conn, err := connectToTasksService(ctx, config)
	if err != nil {
		log.Fatalf("Failed to connect to tasks service: %v", err)
	}
	defer conn.Close()

	var games []Game

	if config.TestMode {
		gameID := 2024030411
		if config.Shootout {
			gameID = 2024030412
		}
		log.Printf("Running in test mode with predefined game ID: %d", gameID)
		games = []Game{createTestGame(config.Shootout)}
	} else {
		// Fetch games from NHL API
		fetchedGames, err := fetchGamesForDate(config.Date)
		if err != nil {
			log.Fatalf("Failed to fetch games: %v", err)
		}

		// Filter games based on team selection
		games = filterGamesForTeams(fetchedGames, config.Teams)

		// If today flag is set, filter to only upcoming games
		if config.Today {
			games = filterUpcomingGames(games)
		}
	}

	// Process games and create tasks
	if err := processGames(ctx, client, config, games); err != nil {
		log.Fatalf("Failed to process games: %v", err)
	}

	log.Printf("Successfully processed %d games", len(games))

	// Send summary notification after all games have been processed
	if notifier.IsEnabled() {
		var gameInfos []notification.GameInfo
		for _, game := range games {
			gameInfos = append(gameInfos, notification.GameInfo{
				ID:        strconv.Itoa(game.ID),
				GameDate:  game.GameDate,
				StartTime: game.StartTime,
				HomeTeam:  game.HomeTeam.Abbrev,
				AwayTeam:  game.AwayTeam.Abbrev,
			})
		}
		if err := notifier.SendScheduleSummary(gameInfos); err != nil {
			log.Printf("Warning: Failed to send schedule summary notification: %v", err)
		}
	}
}
