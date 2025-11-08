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

## Usage

### Basic Usage

Run with default settings (Dallas Stars games for today, local task queue):
```bash
./gameTaskEmulator
```

### Command Line Options

- `-date YYYY-MM-DD`: Specify a future date to query (default: today)
- `-teams ID1,ID2,ID3`: Comma-separated list of NHL team IDs or city codes to filter for (supports both formats)
- `-today`: Filter for today's upcoming games only (overrides -date)
- `-all`: Include all teams playing on the specified date
- `-test`: Run in test mode with predefined game data
- `-prod`: Send tasks to production queue instead of local emulator
- `-project PROJECT_ID`: GCP Project ID (default: "crash-the-crease")
- `-location LOCATION`: GCP Location (default: "us-central1")
- `-queue QUEUE_NAME`: Task Queue name (default: "game-updates")

### Examples

**Get Dallas Stars games for today (default)**:
```bash
./gameTaskEmulator
```

**Get today's upcoming games for Chicago Blackhawks**:
```bash
./gameTaskEmulator -today -teams CHI
```

**Get today's upcoming games for multiple teams using city codes**:
```bash
./gameTaskEmulator -today -teams CHI,DAL,BOS
```

**Get games for specific teams on a future date (mixing city codes and IDs)**:
```bash
./gameTaskEmulator -date 2024-03-15 -teams CHI,25,1
```

**Get all games for tomorrow**:
```bash
./gameTaskEmulator -date 2024-03-16 -all
```

**Run in test mode**:
```bash
./gameTaskEmulator -test
```

**Send tasks to production**:
```bash
./gameTaskEmulator -prod -date 2024-03-20
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

The program schedules Google Cloud Tasks to run 30 minutes before each game's start time. Each task contains:
- Game ID
- Game execution end time (game start + 4 hours)

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

Test mode can be used for development without making actual API calls:
```bash
./gameTaskEmulator -test
```

### Local Development

For local development, ensure the local Cloud Tasks emulator is running and use the default settings (without `-prod` flag).

## Configuration

### Environment Variables

While the program primarily uses command-line flags, you may need to set up Google Cloud credentials:

```bash
export GOOGLE_APPLICATION_CREDENTIALS="path/to/service-account-key.json"
```

### Production Configuration

When using `-prod` flag, ensure:
1. GCP credentials are properly configured
2. The target Cloud Function is deployed
3. The task queue exists in the specified project/location

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
docker run --rm blnelson/firepowergametaskemulator:latest -today -teams CHI
```

#### Using the Run Script

The repository includes a `run.sh` script that simplifies container execution:

```bash
# Run with default settings (Dallas Stars, today's games)
./run.sh

# Run with specific team
./run.sh -today -teams CHI

# Run with multiple teams
./run.sh -today -teams CHI,DAL,BOS

# Run in production mode
./run.sh -prod -today -teams DAL

# Force pull from registry (fail if network unavailable)
./run.sh --force-pull -today -teams CHI
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

### Systemd Installation (Linux)

For automated daily execution on Linux systems, use the installation script:

#### Quick Install

Install with default settings (Dallas Stars, runs daily at 6:00 AM):

```bash
sudo ./install.sh
```

#### Custom Installation

Install for specific team(s):

```bash
# Single team
sudo ./install.sh --team CHI

# Multiple teams
sudo ./install.sh --team CHI,DAL,BOS

# With production mode
sudo ./install.sh --team DAL --flags "-today -prod"

# As specific user
sudo ./install.sh --user myuser --team CHI
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

# Additional flags
ADDITIONAL_FLAGS=-today -prod

# Google Cloud credentials (if using production mode)
GOOGLE_APPLICATION_CREDENTIALS=/path/to/service-account-key.json
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
