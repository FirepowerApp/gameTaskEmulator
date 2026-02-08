package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	taskspb "google.golang.org/genproto/googleapis/cloud/tasks/v2"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// GameInfo represents game information for the task payload
type GameInfo struct {
	ID        string `json:"id"`
	GameDate  string `json:"gameDate"`
	StartTime string `json:"startTimeUTC"`
	HomeTeam  Team   `json:"homeTeam"`
	AwayTeam  Team   `json:"awayTeam"`
}

// TaskPayload represents the payload structure for cloud tasks, matching new system
type TaskPayload struct {
	Game         GameInfo `json:"game"`
	ExecutionEnd *string  `json:"execution_end,omitempty"`
	ShouldNotify bool     `json:"ShouldNotify"`
}

// connectToTasksService connects to Cloud Tasks service (emulator or production)
func connectToTasksService(ctx context.Context, config *Config) (taskspb.CloudTasksClient, *grpc.ClientConn, error) {
	if !config.Production {
		// Connect to local emulator using direct GRPC (like localCloudTasksTest)
		endpoint := config.EmulatorHost
		log.Printf("Connecting to local Cloud Tasks emulator at %s", endpoint)

		conn, err := grpc.DialContext(ctx, endpoint, grpc.WithInsecure(), grpc.WithBlock())
		if err != nil {
			return nil, nil, fmt.Errorf("failed to connect to local Cloud Tasks emulator at %s - ensure the emulator is running: %w", endpoint, err)
		}

		client := taskspb.NewCloudTasksClient(conn)
		return client, conn, nil
	} else {
		// For production mode, we would need to implement the official client approach
		// This is a placeholder - in practice you'd use the official Cloud Tasks client
		return nil, nil, fmt.Errorf("production mode not implemented in this version")
	}
}

// createQueue creates a task queue if it doesn't exist
func createQueue(client taskspb.CloudTasksClient, ctx context.Context, config *Config) error {
	// projects/localproject/locations/us-south1/queues/gameschedule
	queuePath := fmt.Sprintf("projects/%s/locations/%s/queues/%s", config.ProjectID, config.Location, config.QueueName)
	parentPath := fmt.Sprintf("projects/%s/locations/%s", config.ProjectID, config.Location)

	req := &taskspb.CreateQueueRequest{
		Parent: parentPath,
		Queue: &taskspb.Queue{
			Name: queuePath,
		},
	}
	_, err := client.CreateQueue(ctx, req)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") || strings.Contains(err.Error(), "AlreadyExists") {
			log.Printf("Queue %s already exists, skipping creation", config.QueueName)
			return nil
		}
		return fmt.Errorf("failed to create queue: %w", err)
	}
	log.Printf("Created queue: %s", queuePath)
	return nil
}

// createCloudTask creates a Google Cloud Task for a given game using direct GRPC
func createCloudTask(ctx context.Context, client taskspb.CloudTasksClient, config *Config, game Game) error {
	// Create execution end time (game start time + 4 hours for typical game duration)
	startTime, err := time.Parse(time.RFC3339, game.StartTime)
	if err != nil {
		return fmt.Errorf("failed to parse start time: %w", err)
	}

	executionEnd := startTime.Add(4 * time.Hour).Format(time.RFC3339)

	// Prepare the task payload with full game context
	payload := TaskPayload{
		Game: GameInfo{
			ID:        strconv.Itoa(game.ID),
			GameDate:  game.GameDate,
			StartTime: game.StartTime,
			HomeTeam:  game.HomeTeam,
			AwayTeam:  game.AwayTeam,
		},
		ExecutionEnd: &executionEnd,
		ShouldNotify: !config.TestMode,
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Determine the target URL based on host configuration
	var targetURL string
	if config.LocalMode {
		targetURL = "http://host.docker.internal:8080"
	} else {
		targetURL = config.HostURL
	}

	// Schedule task to run 5 minutes before game start
	scheduleTime := startTime.Add(-5 * time.Minute)

	// Create the task request using taskspb format (works for emulator)
	queuePath := fmt.Sprintf("projects/%s/locations/%s/queues/%s", config.ProjectID, config.Location, config.QueueName)

	req := &taskspb.CreateTaskRequest{
		Parent: queuePath,
		Task: &taskspb.Task{
			MessageType: &taskspb.Task_HttpRequest{
				HttpRequest: &taskspb.HttpRequest{
					HttpMethod: taskspb.HttpMethod_POST,
					Url:        targetURL,
					Headers: map[string]string{
						"Content-Type": "application/json",
					},
					Body: payloadJSON,
				},
			},
			ScheduleTime: timestamppb.New(scheduleTime),
		},
	}

	// Create the task
	task, err := client.CreateTask(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to create task: %w", err)
	}

	log.Printf("Created task %s for game %d, scheduled for %s", task.Name, game.ID, scheduleTime.Format(time.RFC3339))
	return nil
}

// processGames processes a list of games and creates cloud tasks for each
func processGames(ctx context.Context, client taskspb.CloudTasksClient, config *Config, games []Game) error {
	if len(games) == 0 {
		log.Println("No games found to process")
		return nil
	}

	// Create queue if it doesn't exist
	if err := createQueue(client, ctx, config); err != nil {
		log.Printf("Warning: Failed to create queue: %v", err)
	}

	log.Printf("Processing %d games", len(games))

	for _, game := range games {
		log.Printf("Processing game %d: %s", game.ID, game.StartTime)

		if err := createCloudTask(ctx, client, config, game); err != nil {
			log.Printf("Failed to create task for game %d: %v", game.ID, err)
			continue
		}
	}

	return nil
}
