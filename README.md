# Game Task Emulator

This program fetches NHL game schedules and creates Google Cloud Tasks for game tracking. It's designed to work with the CrashTheCrease backend system to automatically schedule game monitoring tasks.

## Features

- **Automatic Game Detection**: Fetches games for today or a specified date using the NHL API
- **Team Filtering**: Supports filtering games by specific teams using city codes (CHI, DAL) or numeric IDs (defaults to Dallas Stars)
- **Today's Upcoming Games**: New `-today` flag filters to only today's games that haven't started yet
- **Flexible Scheduling**: Can schedule tasks for future dates
- **Test Mode**: Includes a test mode with predefined game data for development
- **Production Support**: Configurable for both local development and production environments
- **Cloud Task Integration**: Creates Google Cloud Tasks that integrate with the existing game monitoring system
- **Discord Notifications**: Optional Discord webhook notifications with a summary of all scheduled games

## Usage

### Basic Usage

Run with default settings (Dallas Stars games for today, sending to local host):
```bash
./gameTaskEmulator -local
```

Or send to a custom host:
```bash
./gameTaskEmulator -host https://example.com/api
```

### Command Line Options

#### Required Flags (one must be specified)
- `-local`: Send requests to local host at `http://host.docker.internal:8080`
- `-host URL`: Custom host URL to send requests to (e.g., `https://example.com/api`)

**Note**: You must specify either `-local` or `-host <url>`. The application will exit with an error if neither is provided, or if both are specified.

#### Optional Flags
- `-date YYYY-MM-DD`: Specify a future date to query (default: today)
- `-teams ID1,ID2,ID3`: Comma-separated list of NHL team IDs or city codes to filter for (supports both formats)
- `-today`: Filter for today's upcoming games only (overrides -date)
- `-all`: Include all teams playing on the specified date
- `-test`: Run in test mode with predefined game data. Sets `ShouldNotify: false` in the payload (default: `ShouldNotify: true`)
- `-prod`: Send tasks to production queue instead of local emulator
- `-project PROJECT_ID`: GCP Project ID (default: "localproject")
- `-location LOCATION`: GCP Location (default: "us-south1")
- `-queue QUEUE_NAME`: Task Queue name (default: "gameschedule")
- `-discord-webhook URL`: Discord webhook URL for notifications (can also be set via `DISCORD_WEBHOOK_URL` environment variable)

### Examples

**Get Dallas Stars games for today to local host**:
```bash
./gameTaskEmulator -local
```

**Get today's upcoming games for Chicago Blackhawks to local host**:
```bash
./gameTaskEmulator -local -today -teams CHI
```

**Get today's upcoming games for multiple teams using city codes**:
```bash
./gameTaskEmulator -local -today -teams CHI,DAL,BOS
```

**Get games for specific teams on a future date (mixing city codes and IDs)**:
```bash
./gameTaskEmulator -local -date 2024-03-15 -teams CHI,25,1
```

**Get all games for tomorrow**:
```bash
./gameTaskEmulator -local -date 2024-03-16 -all
```

**Run in test mode (sets ShouldNotify: false in payload)**:
```bash
./gameTaskEmulator -local -test
```

**Send tasks to a custom host URL**:
```bash
./gameTaskEmulator -host https://example.com/api/tasks -today -teams DAL
```

**Send tasks to production cloud function**:
```bash
./gameTaskEmulator -host https://us-south1-myproject.cloudfunctions.net/watchGameUpdates -today
```

## NHL Team IDs and City Codes

The program now supports both numeric team IDs and city codes for team filtering:

### Common Teams (City Code - Team ID - Team Name):
- **CHI - 16** - Chicago Blackhawks
- **DAL - 25** - Dallas Stars
- **BOS - 1** - Boston Bruins
- **TOR - 10** - Toronto Maple Leafs
- **NYR - 3** - New York Rangers
- **DET - 17** - Detroit Red Wings

### All Supported City Codes:
- **ANA** (24) - Anaheim Ducks
- **ARI** (53) - Arizona Coyotes
- **BOS** (1) - Boston Bruins
- **BUF** (7) - Buffalo Sabres
- **CAR** (12) - Carolina Hurricanes
- **CBJ** (29) - Columbus Blue Jackets
- **CGY** (20) - Calgary Flames
- **CHI** (16) - Chicago Blackhawks
- **COL** (21) - Colorado Avalanche
- **DAL** (25) - Dallas Stars
- **DET** (17) - Detroit Red Wings
- **EDM** (22) - Edmonton Oilers
- **FLA** (13) - Florida Panthers
- **LAK** (26) - Los Angeles Kings
- **MIN** (30) - Minnesota Wild
- **MTL** (8) - Montreal Canadiens
- **NJD** (6) - New Jersey Devils
- **NSH** (18) - Nashville Predators
- **NYI** (2) - New York Islanders
- **NYR** (3) - New York Rangers
- **OTT** (9) - Ottawa Senators
- **PHI** (4) - Philadelphia Flyers
- **PIT** (5) - Pittsburgh Penguins
- **SEA** (55) - Seattle Kraken
- **SJS** (28) - San Jose Sharks
- **STL** (19) - St. Louis Blues
- **TBL** (14) - Tampa Bay Lightning
- **TOR** (10) - Toronto Maple Leafs
- **VAN** (23) - Vancouver Canucks
- **VGK** (54) - Vegas Golden Knights
- **WPG** (52) - Winnipeg Jets
- **WSH** (15) - Washington Capitals

You can use either format: `-teams CHI,DAL` or `-teams 16,25` or mix them: `-teams CHI,25,BOS`

## Task Scheduling

The program schedules Google Cloud Tasks to run 5 minutes before each game's start time. Each task contains:
- Game information (ID, date, start time, home team, away team)
- Execution end time (game start + 4 hours)
- ShouldNotify flag (false when `-test` flag is used, true otherwise)

The target URL for tasks is determined by the `-local` or `-host` flags:
- `-local`: Sends to `http://host.docker.internal:8080`
- `-host <url>`: Sends to the specified URL

These tasks are consumed by the existing `watchGameUpdates` service in the CrashTheCrease backend.

## Development

### Building

This project uses a custom build system (based on the CrashTheCrease backend build system) that compiles binaries and saves them to the `bin/` directory.

#### Build System Usage

The build system supports three main commands:

**Build a specific target**:
```bash
go run build.go -target gameTaskEmulator
```

**Build all available targets**:
```bash
go run build.go -all
```

**List available build targets**:
```bash
go run build.go -list
```

#### Build Output

All binaries are compiled with `CGO_ENABLED=0` for static linking and saved to the `./bin/` directory. The `bin/` directory is excluded from version control via `.gitignore` to prevent binaries from being committed to the repository.

After building, you can run the binary directly:
```bash
./bin/gameTaskEmulator [options]
```

#### Available Build Targets

- **gameTaskEmulator**: NHL game tracker scheduler that creates Cloud Tasks for game monitoring
  - Source: `./cmd/gameTaskEmulator`
  - Binary: `./bin/gameTaskEmulator`

### Dependencies

The program requires:
- Go 1.21+
- Google Cloud Tasks API access
- Internet connectivity for NHL API access

### Testing

Test mode can be used for development without making actual API calls. The `-test` flag uses predefined game data and sets `ShouldNotify: false` in the task payload to prevent notifications:
```bash
./gameTaskEmulator -local -test
```

You can also test with a custom host:
```bash
./gameTaskEmulator -host https://test.example.com/api -test
```

### Local Development

For local development, ensure the local Cloud Tasks emulator is running and use the `-local` flag to send tasks to `http://host.docker.internal:8080`:
```bash
./gameTaskEmulator -local -today -teams DAL
```

## Configuration

### Environment Variables

While the program primarily uses command-line flags, the following environment variables are supported:

- `GOOGLE_APPLICATION_CREDENTIALS`: Path to GCP service account key (required for production mode)
- `DISCORD_WEBHOOK_URL`: Discord webhook URL for notifications (optional, can also be set via `-discord-webhook` flag)

```bash
export GOOGLE_APPLICATION_CREDENTIALS="path/to/service-account-key.json"
export DISCORD_WEBHOOK_URL="https://discord.com/api/webhooks/YOUR_WEBHOOK_ID/YOUR_WEBHOOK_TOKEN"
```

When a Discord webhook URL is configured, the application sends a single summary notification after all games have been processed. The notification includes the date, time, and opponents for each scheduled game, or a message indicating that no games were identified.

### Production Configuration

When using `-host` flag with a production URL, ensure:
1. GCP credentials are properly configured (if required by the endpoint)
2. The target endpoint is deployed and accessible
3. The task queue exists in the specified project/location

Example production usage:
```bash
./gameTaskEmulator -host https://us-south1-myproject.cloudfunctions.net/watchGameUpdates -today -teams DAL
```

## Error Handling

The program includes comprehensive error handling for:
- NHL API connectivity issues
- Invalid team IDs
- Google Cloud Tasks creation failures
- Date parsing errors

Failed task creations are logged but don't stop processing of other games.

## Integration

This program integrates with:
- **NHL API**: For fetching game schedules
- **Google Cloud Tasks**: For task scheduling
- **CrashTheCrease Backend**: Via the `watchGameUpdates` handler
- **Discord Webhooks**: For optional schedule summary notifications

## Troubleshooting

### Common Issues

1. **NHL API Errors**: Check internet connectivity and try again
2. **Cloud Tasks Errors**: Verify GCP credentials and project configuration
3. **Invalid Team IDs**: Refer to NHL API documentation for correct team IDs
4. **Date Format Errors**: Use YYYY-MM-DD format for dates

### Logging

The program provides detailed logging of:
- Configuration settings
- API requests and responses
- Task creation results
- Error conditions

## Deployment

### Container Deployment

The application is available as a Docker container and automatically built via GitHub Actions.

#### Container Image

The container image is published to Docker Hub when pull requests are merged to main:

```
blnelson/firepowergametaskemulator:latest
```

#### Running with Docker

You can run the application directly using Docker:

```bash
docker pull blnelson/firepowergametaskemulator:latest
docker run --rm blnelson/firepowergametaskemulator:latest -local -today -teams CHI
```

#### Using the Run Script

The repository includes a `run.sh` script that simplifies container execution:

```bash
# Run with default settings to local host (Dallas Stars, today's games)
./run.sh -local

# Run with specific team to local host
./run.sh -local -today -teams CHI

# Run with multiple teams to local host
./run.sh -local -today -teams CHI,DAL,BOS

# Run with custom host URL
./run.sh -host https://example.com/api -today -teams DAL

# Force pull from registry (fail if network unavailable)
./run.sh --force-pull -local -today -teams CHI
```

The script automatically:
- Pulls the latest container image from Docker Hub
- Falls back to locally cached image if registry pull fails
- Mounts Google Cloud credentials if available
- Passes through all command-line flags to the container

**Run Script Options:**
- `--force-pull`: Fail if unable to pull from registry (no fallback to local cache)

**Default Behavior:**
By default, if the script encounters a **network error** when pulling from Docker Hub (connection timeout, DNS failure, network unreachable), it will fall back to the last successfully pulled local image. This ensures the application can run even without internet connectivity.

**Error Handling:**
- **Network errors** (timeout, connection refused, DNS failure): Falls back to local cached image
- **Authentication errors**: Always fails immediately (requires `docker login`)
- **Image not found errors**: Always fails immediately (check image name)
- **Force pull mode** (`--force-pull`): Always fails on any pull error (no fallback)

### Docker Scheduled Installation (Recommended for Local Development)

For automated **weekly execution every Monday at 5:00 AM** using Docker, see the [Docker Installation Guide](DOCKER_INSTALL.md).

This method works on any platform (Linux, macOS, Windows) and provides:
- **Local development** - Sends to local task queue at `http://host.docker.internal:8080` by default
- Containerized, isolated execution
- Weekly schedule (Monday 5:00 AM)
- Easy configuration and management
- No systemd dependency

**Quick Start:**

```bash
# Basic installation (sends to local task queue)
./docker-install.sh

# Install for specific team (sends to local task queue)
./docker-install.sh --team CHI

# Install for multiple teams (sends to local task queue)
./docker-install.sh --team CHI,DAL,BOS

# With production mode (not local)
./docker-install.sh --team DAL --flags "-today -prod" --credentials ./gcp-key.json
```

**Default behavior**: Uses `-local -today` flags to send tasks to your local task queue at `http://host.docker.internal:8080`. This matches the standard workflow of running `./bin/gameTaskEmulator -local -today -teams CHI`.

For complete documentation, see [DOCKER_INSTALL.md](DOCKER_INSTALL.md).

### Systemd Installation (Linux - Daily Execution)

For automated **daily execution at 6:00 AM** on Linux systems, use the systemd installation script:

#### Quick Install

Install with default settings (Dallas Stars, runs daily at 6:00 AM):

```bash
sudo ./install.sh
```

#### Custom Installation

Install for specific team(s):

```bash
# Single team with local host
sudo ./install.sh --team CHI --flags "-local -today"

# Multiple teams with local host
sudo ./install.sh --team CHI,DAL,BOS --flags "-local -today"

# With custom host URL
sudo ./install.sh --team DAL --flags "-host https://example.com/api -today"

# As specific user
sudo ./install.sh --user myuser --team CHI --flags "-local -today"
```

#### What Gets Installed

The installation script:
1. Copies files to `/opt/gameTaskEmulator`
2. Creates systemd service and timer files
3. Configures the service with your team preferences
4. Enables daily execution at 6:00 AM (configurable)
5. Sets up logging via systemd journal

#### Managing the Service

```bash
# View timer status
systemctl status gametask-emulator.timer

# View service status
systemctl status gametask-emulator.service

# View logs
journalctl -u gametask-emulator.service -f

# Run manually now
sudo systemctl start gametask-emulator.service

# Edit configuration
sudo nano /etc/default/gametask-emulator

# Restart timer after config changes
sudo systemctl restart gametask-emulator.timer

# Disable automatic execution
sudo systemctl disable gametask-emulator.timer
```

#### Configuration

After installation, you can modify the configuration at `/etc/default/gametask-emulator`:

```bash
# Team city code (leave empty for Dallas Stars)
TEAM_CODE=CHI

# Additional flags (must include either -local or -host)
ADDITIONAL_FLAGS=-local -today

# Or with custom host:
# ADDITIONAL_FLAGS=-host https://example.com/api -today

# Google Cloud credentials (if using custom host)
GOOGLE_APPLICATION_CREDENTIALS=/path/to/service-account-key.json

# Discord webhook URL for notifications (optional)
DISCORD_WEBHOOK_URL=https://discord.com/api/webhooks/YOUR_WEBHOOK_ID/YOUR_WEBHOOK_TOKEN
```

### GitHub Actions

The repository includes a GitHub Action workflow (`.github/workflows/docker-publish.yml`) that:
- Builds the Docker image when pull requests are merged to main
- Publishes to Docker Hub (requires `DOCKERHUB_USERNAME` and `DOCKERHUB_TOKEN` secrets)
- Supports multi-architecture builds (amd64, arm64)
- Tags images with version numbers and commit SHAs

#### Setting Up Docker Hub Secrets

To enable automated builds, add the following secrets to your GitHub repository:

1. Go to your repository Settings → Secrets and variables → Actions
2. Add two secrets:
   - `DOCKERHUB_USERNAME`: Your Docker Hub username
   - `DOCKERHUB_TOKEN`: Your Docker Hub access token (create one at https://hub.docker.com/settings/security)

## Future Enhancements

Potential improvements:
- Support for playoff schedules
- Retry mechanisms for failed API calls
- Configuration file support
- Multiple date range support
- Team name to ID resolution
