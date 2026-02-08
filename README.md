# Game Task Emulator

Fetches NHL game schedules and creates Google Cloud Tasks for game tracking. Integrates with the CrashTheCrease backend to automatically schedule game monitoring tasks.

## Quick Start

```bash
# Build
make build

# Run for today's Dallas Stars games (default team)
./bin/gameTaskEmulator -local

# Run for specific team
./bin/gameTaskEmulator -local -today -teams CHI

# Run for multiple teams
./bin/gameTaskEmulator -local -today -teams CHI,DAL,BOS
```

## Command Line Flags

### Required (one must be specified)

| Flag | Description |
|------|-------------|
| `-local` | Send tasks to `http://host.docker.internal:8080` |
| `-host URL` | Send tasks to custom URL |

### Optional

| Flag | Description | Default |
|------|-------------|---------|
| `-teams IDs` | Comma-separated team codes or IDs | `DAL` (25) |
| `-date YYYY-MM-DD` | Date to query | today |
| `-today` | Filter to upcoming games only | false |
| `-all` | Include all teams | false |
| `-test` | Test mode with mock data | false |
| `-shootout` | Use shootout test game | false |
| `-prod` | Use production Cloud Tasks | false |
| `-emulator HOST` | Cloud Tasks emulator host | `localhost:8123` |
| `-project ID` | GCP Project ID | `localproject` |
| `-location LOC` | GCP Location | `us-south1` |
| `-queue NAME` | Task Queue name | `gameschedule` |
| `-discord-webhook URL` | Discord notification webhook | - |

## Team Codes

Use standard 3-letter NHL abbreviations: `CHI`, `DAL`, `BOS`, `NYR`, `TOR`, etc.

You can also use numeric team IDs or mix formats: `-teams CHI,25,BOS`

<details>
<summary>Full team code list</summary>

| Code | Team | ID |
|------|------|-----|
| ANA | Anaheim Ducks | 24 |
| ARI | Arizona Coyotes | 53 |
| BOS | Boston Bruins | 1 |
| BUF | Buffalo Sabres | 7 |
| CAR | Carolina Hurricanes | 12 |
| CBJ | Columbus Blue Jackets | 29 |
| CGY | Calgary Flames | 20 |
| CHI | Chicago Blackhawks | 16 |
| COL | Colorado Avalanche | 21 |
| DAL | Dallas Stars | 25 |
| DET | Detroit Red Wings | 17 |
| EDM | Edmonton Oilers | 22 |
| FLA | Florida Panthers | 13 |
| LAK | Los Angeles Kings | 26 |
| MIN | Minnesota Wild | 30 |
| MTL | Montreal Canadiens | 8 |
| NJD | New Jersey Devils | 6 |
| NSH | Nashville Predators | 18 |
| NYI | New York Islanders | 2 |
| NYR | New York Rangers | 3 |
| OTT | Ottawa Senators | 9 |
| PHI | Philadelphia Flyers | 4 |
| PIT | Pittsburgh Penguins | 5 |
| SEA | Seattle Kraken | 55 |
| SJS | San Jose Sharks | 28 |
| STL | St. Louis Blues | 19 |
| TBL | Tampa Bay Lightning | 14 |
| TOR | Toronto Maple Leafs | 10 |
| VAN | Vancouver Canucks | 23 |
| VGK | Vegas Golden Knights | 54 |
| WPG | Winnipeg Jets | 52 |
| WSH | Washington Capitals | 15 |

</details>

## Environment Variables

| Variable | Description |
|----------|-------------|
| `CLOUD_TASKS_EMULATOR` | Emulator host (alternative to `-emulator` flag) |
| `DISCORD_WEBHOOK_URL` | Discord webhook (alternative to `-discord-webhook` flag) |
| `GOOGLE_APPLICATION_CREDENTIALS` | GCP service account key path |

## Development

```bash
make build    # Build binary to ./bin/
make test     # Run tests
make clean    # Remove build artifacts
```

### Requirements

- Go 1.21+
- Cloud Tasks emulator (for local development)
- Internet access (NHL API)

## Deployment

### Docker (Recommended)

```bash
# Run directly
docker run --rm blnelson/firepowergametaskemulator:latest -local -today -teams CHI

# Or use the run script
./run.sh -local -today -teams CHI
```

### Scheduled Execution

**Docker** (any platform) - runs weekly on Mondays at 5 AM:
```bash
./docker-install.sh --team CHI
```
See [DOCKER_INSTALL.md](DOCKER_INSTALL.md) for full documentation.

**Systemd** (Linux) - runs daily at 6 AM:
```bash
sudo ./install.sh --team CHI --flags "-local -today"
```

### Container Image

Published to Docker Hub on merge to main:
```
blnelson/firepowergametaskemulator:latest
```

## How It Works

1. Fetches game schedule from NHL API for the specified date
2. Filters games by team selection
3. Creates a Cloud Task for each game, scheduled 5 minutes before start
4. Task payload includes game info and 4-hour execution window
5. Optionally sends Discord summary notification
