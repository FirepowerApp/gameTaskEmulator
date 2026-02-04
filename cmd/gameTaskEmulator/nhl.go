package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// Team represents team information from the NHL API, also used in task payloads.
type Team struct {
	ID                       int               `json:"id"`
	CommonName               map[string]string `json:"commonName"`
	PlaceName                map[string]string `json:"placeName"`
	PlaceNameWithPreposition map[string]string `json:"placeNameWithPreposition"`
	Abbrev                   string            `json:"abbrev"`
}

// Game represents a single NHL game with relevant information
type Game struct {
	ID        int    `json:"id"`
	GameDate  string `json:"gameDate"`
	StartTime string `json:"startTimeUTC"`
	AwayTeam  Team   `json:"awayTeam"`
	HomeTeam  Team   `json:"homeTeam"`
}

// ScheduleResponse represents the NHL API schedule response
type ScheduleResponse struct {
	GameWeek []struct {
		Date  string `json:"date"`
		Games []Game `json:"games"`
	} `json:"gameWeek"`
}

// fetchGamesForDate retrieves games for a specific date from the NHL API
func fetchGamesForDate(date string) ([]Game, error) {
	url := fmt.Sprintf("%s/schedule/%s", NHLAPIBaseURL, date)

	log.Printf("Fetching games from NHL API: %s", url)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch schedule: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("NHL API returned status: %d", resp.StatusCode)
	}

	var schedule ScheduleResponse
	if err := json.NewDecoder(resp.Body).Decode(&schedule); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	var games []Game
	for _, week := range schedule.GameWeek {
		for _, game := range week.Games {
			games = append(games, game)
		}
	}

	log.Printf("Found %d games for date %s", len(games), date)
	return games, nil
}

// filterGamesForTeams filters games to include only those involving specified teams
func filterGamesForTeams(games []Game, teams []int) []Game {
	if len(teams) == 0 {
		// Return all games if no specific teams are specified
		return games
	}

	teamMap := make(map[int]bool)
	for _, teamID := range teams {
		teamMap[teamID] = true
	}

	var filteredGames []Game
	for _, game := range games {
		if teamMap[game.HomeTeam.ID] || teamMap[game.AwayTeam.ID] {
			filteredGames = append(filteredGames, game)
		}
	}

	log.Printf("Filtered to %d games involving specified teams", len(filteredGames))
	return filteredGames
}

// filterUpcomingGames filters games to include only those that haven't started yet
func filterUpcomingGames(games []Game) []Game {
	now := time.Now()
	var upcomingGames []Game

	for _, game := range games {
		startTime, err := time.Parse(time.RFC3339, game.StartTime)
		if err != nil {
			log.Printf("Warning: Could not parse start time for game %d: %v", game.ID, err)
			continue
		}

		// Include games that haven't started yet
		if startTime.After(now) {
			upcomingGames = append(upcomingGames, game)
		}
	}

	log.Printf("Filtered to %d upcoming games", len(upcomingGames))
	return upcomingGames
}

// createTestGame creates a test game with predefined data for testing purposes
func createTestGame(shootout bool) Game {
	gameID := 2024030411
	if shootout {
		gameID = 2024030412
	}

	return Game{
		ID:        gameID,
		GameDate:  time.Now().Format("2006-01-02"),
		StartTime: time.Now().Format(time.RFC3339),
		AwayTeam: Team{
			ID:                       DefaultTeamID,
			CommonName:               map[string]string{"default": "Stars"},
			PlaceName:                map[string]string{"default": "Dallas"},
			PlaceNameWithPreposition: map[string]string{"default": "Dallas"},
			Abbrev:                   "DAL",
		},
		HomeTeam: Team{
			ID:                       1, // Boston Bruins
			CommonName:               map[string]string{"default": "Bruins"},
			PlaceName:                map[string]string{"default": "Boston"},
			PlaceNameWithPreposition: map[string]string{"default": "Boston"},
			Abbrev:                   "BOS",
		},
	}
}
