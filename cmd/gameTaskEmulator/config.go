package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	// DefaultTeamID represents the Dallas Stars team ID in the NHL API
	DefaultTeamID = 25
	// NHLAPIBaseURL is the base URL for NHL API endpoints
	NHLAPIBaseURL = "https://api-web.nhle.com/v1"
)

// Config holds the configuration for the application
type Config struct {
	Date              string // Date to query games for (YYYY-MM-DD format)
	Teams             []int  // Team IDs to filter games for
	TestMode          bool   // Whether to run in test mode
	AllTeams          bool   // Whether to include all teams
	Today             bool   // Whether to filter for today's upcoming games only
	Production        bool   // Whether to use production task queue
	Shootout          bool   // Whether to use shootout game ID (2024030412)
	ProjectID         string // GCP Project ID
	Location          string // GCP Location
	QueueName         string // Task Queue name
	LocalMode         bool   // Whether to send requests to local host
	HostURL           string // Custom host URL for sending requests
	DiscordWebhookURL string // Discord webhook URL for notifications
	EmulatorHost      string // Cloud Tasks emulator host (default: localhost:8123)
}

// cityCodeToTeamID maps NHL team city codes to their corresponding team IDs
var cityCodeToTeamID = map[string]int{
	"ANA": 24, // Anaheim Ducks
	"ARI": 53, // Arizona Coyotes
	"BOS": 1,  // Boston Bruins
	"BUF": 7,  // Buffalo Sabres
	"CAR": 12, // Carolina Hurricanes
	"CBJ": 29, // Columbus Blue Jackets
	"CGY": 20, // Calgary Flames
	"CHI": 16, // Chicago Blackhawks
	"COL": 21, // Colorado Avalanche
	"DAL": 25, // Dallas Stars
	"DET": 17, // Detroit Red Wings
	"EDM": 22, // Edmonton Oilers
	"FLA": 13, // Florida Panthers
	"LAK": 26, // Los Angeles Kings
	"MIN": 30, // Minnesota Wild
	"MTL": 8,  // Montreal Canadiens
	"NJD": 6,  // New Jersey Devils
	"NSH": 18, // Nashville Predators
	"NYI": 2,  // New York Islanders
	"NYR": 3,  // New York Rangers
	"OTT": 9,  // Ottawa Senators
	"PHI": 4,  // Philadelphia Flyers
	"PIT": 5,  // Pittsburgh Penguins
	"SEA": 55, // Seattle Kraken
	"SJS": 28, // San Jose Sharks
	"STL": 19, // St. Louis Blues
	"TBL": 14, // Tampa Bay Lightning
	"TOR": 10, // Toronto Maple Leafs
	"VAN": 23, // Vancouver Canucks
	"VGK": 54, // Vegas Golden Knights
	"WPG": 52, // Winnipeg Jets
	"WSH": 15, // Washington Capitals
}

// parseTeamIdentifier converts a team identifier (city code or numeric ID) to a team ID
func parseTeamIdentifier(identifier string) (int, error) {
	identifier = strings.TrimSpace(strings.ToUpper(identifier))

	// Try to parse as city code first
	if teamID, exists := cityCodeToTeamID[identifier]; exists {
		return teamID, nil
	}

	// Try to parse as numeric ID
	teamID, err := strconv.Atoi(identifier)
	if err != nil {
		return 0, fmt.Errorf("invalid team identifier: %s (use city code like CHI or numeric ID like 16)", identifier)
	}

	return teamID, nil
}

// parseFlags parses and validates command-line flags
func parseFlags() *Config {
	config := &Config{}

	var teamsStr string
	var emulatorHost string
	flag.StringVar(&config.Date, "date", "", "Specific date to query (YYYY-MM-DD format). Defaults to today.")
	flag.StringVar(&teamsStr, "teams", "", "Comma-separated list of team IDs or city codes (e.g., '25,CHI,DAL'). Defaults to Dallas Stars (25).")
	flag.BoolVar(&config.TestMode, "test", false, "Run in test mode with predefined game ID")
	flag.BoolVar(&config.AllTeams, "all", false, "Include all teams playing on the specified date")
	flag.BoolVar(&config.Today, "today", false, "Filter for today's upcoming games only (overrides -date)")
	flag.BoolVar(&config.Production, "prod", false, "Send tasks to production queue instead of local emulator")
	flag.BoolVar(&config.Shootout, "shootout", false, "Use shootout game ID (2024030412) instead of default (2024030411)")
	flag.StringVar(&config.ProjectID, "project", "localproject", "GCP Project ID")
	flag.StringVar(&config.Location, "location", "us-south1", "GCP Location")
	flag.StringVar(&config.QueueName, "queue", "gameschedule", "Task Queue name")
	flag.BoolVar(&config.LocalMode, "local", false, "Send requests to local host (http://host.docker.internal:8080)")
	flag.StringVar(&config.HostURL, "host", "", "Custom host URL to send requests to")
	flag.StringVar(&config.DiscordWebhookURL, "discord-webhook", "", "Discord webhook URL for notifications (can also be set via DISCORD_WEBHOOK_URL env var)")
	flag.StringVar(&emulatorHost, "emulator", "", "Cloud Tasks emulator host (default: localhost:8123 or CLOUD_TASKS_EMULATOR env var)")

	flag.Parse()

	// Check for Discord webhook URL from environment variable if not set via flag
	if config.DiscordWebhookURL == "" {
		config.DiscordWebhookURL = os.Getenv("DISCORD_WEBHOOK_URL")
	}

	// Set emulator host from flag, environment variable, or default
	if emulatorHost != "" {
		config.EmulatorHost = emulatorHost
	} else if envHost := os.Getenv("CLOUD_TASKS_EMULATOR"); envHost != "" {
		config.EmulatorHost = envHost
	} else {
		config.EmulatorHost = "localhost:8123"
	}

	// Validate that either -local or -host is provided
	if !config.LocalMode && config.HostURL == "" {
		log.Fatalf("Error: Either -local or -host <url> must be provided")
	}

	// Validate that both -local and -host are not provided at the same time
	if config.LocalMode && config.HostURL != "" {
		log.Fatalf("Error: Cannot specify both -local and -host flags")
	}

	// Handle today flag - overrides date setting
	if config.Today {
		config.Date = time.Now().Format("2006-01-02")
	} else if config.Date == "" {
		config.Date = time.Now().Format("2006-01-02")
	}

	// Parse team IDs
	if config.AllTeams {
		config.Teams = []int{} // Empty slice means all teams
	} else if teamsStr != "" {
		teamStrs := strings.Split(teamsStr, ",")
		config.Teams = make([]int, len(teamStrs))
		for i, teamStr := range teamStrs {
			teamID, err := parseTeamIdentifier(teamStr)
			if err != nil {
				log.Fatalf("Invalid team identifier: %s", err)
			}
			config.Teams[i] = teamID
		}
	} else {
		config.Teams = []int{DefaultTeamID} // Default to Dallas Stars
	}

	return config
}
