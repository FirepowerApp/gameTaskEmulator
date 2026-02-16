# Docker Compose Guide - Game Task Emulator

This guide explains how to use Docker Compose with multiple profiles to run the Game Task Emulator in different scenarios.

## Quick Start

```bash
# List all available configurations
make docker-profiles

# Run scheduled deployment (cronjob)
make docker-scheduled

# Run once and exit
make docker-oneshot

# Interactive development mode
make docker-dev
```

## Overview

This project uses **Docker Compose Profiles** to support multiple deployment scenarios from a single `docker-compose.yml` file. Profiles allow you to define different sets of services that can be activated selectively.

## Available Profiles

### 1. 🕐 Scheduled (Cronjob Deployment)

**Purpose:** Runs gameTaskEmulator on a schedule using cron (every Monday at 5:00 AM)

**When to use:**
- Automated weekly game scheduling
- Production or staging environments
- Set-it-and-forget-it deployments

**How to run:**
```bash
# Using Makefile
make docker-scheduled

# Using docker compose directly
docker compose --profile scheduled up -d

# View logs
docker compose logs -f app-scheduled

# Stop
docker compose --profile scheduled down
```

**Configuration:**
- Uses `Dockerfile.scheduled` (includes cron daemon)
- Runs continuously in the background
- Container restarts automatically unless stopped
- Schedule defined in `docker/crontab` (Monday 5:00 AM)
- Includes health check for cron process

**Environment variables:**
```bash
TZ=America/Chicago              # Your timezone
TEAM_CODE=CHI,DAL              # Comma-separated team codes
ADDITIONAL_FLAGS=-local -today  # Flags passed to gameTaskEmulator
```

### 2. 🎯 One-Shot (Single Execution)

**Purpose:** Runs gameTaskEmulator once and exits

**When to use:**
- Manual execution
- Testing changes
- Triggered runs (CI/CD, webhooks, etc.)
- Quick validation

**How to run:**
```bash
# Using Makefile (shows output)
make docker-oneshot

# With custom arguments
ONESHOT_ARGS="-local -today -teams CHI,DAL" make docker-oneshot

# Using docker compose directly
docker compose --profile oneshot up

# Run in background (less common for one-shot)
docker compose --profile oneshot up -d
```

**Configuration:**
- Uses `Dockerfile` (lightweight, no cron)
- Exits after completion
- Default args: `-local -today`
- Customize with `ONESHOT_ARGS` environment variable

**Examples:**
```bash
# Run for specific teams
ONESHOT_ARGS="-local -today -teams CHI,DAL,BOS" make docker-oneshot

# Run for all teams
ONESHOT_ARGS="-local -today -all" make docker-oneshot

# Run for a specific date
ONESHOT_ARGS="-local -date 2026-03-15" make docker-oneshot

# Test mode
ONESHOT_ARGS="-local -today -test" make docker-oneshot
```

### 3. 🔧 Dev (Development/Interactive Mode)

**Purpose:** Interactive container with shell access for development and debugging

**When to use:**
- Testing code changes
- Debugging issues
- Exploring the application behavior
- Running manual commands

**How to run:**
```bash
# Using Makefile
make docker-dev

# Using docker compose directly
docker compose --profile dev run --rm app-dev

# Inside the container, run commands like:
/app/gameTaskEmulator -local -today
/app/gameTaskEmulator -local -date 2026-03-15 -teams CHI
```

**Configuration:**
- Uses `Dockerfile` (lightweight)
- Provides interactive shell (`/bin/sh`)
- Container removed after exit (`--rm`)
- Optional: Mount source code for live development (commented out by default)

**Tips:**
- You can run the application multiple times with different flags
- Test different configurations interactively
- Examine logs and output in real-time
- Exit with `exit` or `Ctrl+D`

### 4. 🚀 Prod (Production Deployment)

**Purpose:** Production scheduled deployment with Google Cloud credentials

**When to use:**
- Production environment
- Integration with Google Cloud Tasks
- Scheduled production runs with authentication

**Prerequisites:**
1. Google Cloud service account key file
2. `GOOGLE_APPLICATION_CREDENTIALS` environment variable set
3. Proper GCP permissions configured

**How to run:**
```bash
# Create .env file first:
echo 'GOOGLE_APPLICATION_CREDENTIALS=./path/to/gcp-key.json' >> .env

# Using Makefile
make docker-prod

# Using docker compose directly
docker compose --profile prod up -d

# View logs
docker compose logs -f app-prod

# Stop
docker compose --profile prod down
```

**Configuration:**
- Uses `Dockerfile.scheduled` (includes cron)
- Mounts GCP credentials into container
- Uses `-prod` flag (sends to real Google Cloud Tasks)
- Runs continuously with restart policy
- Includes health check for cron process

**Environment variables:**
```bash
TZ=America/Chicago
TEAM_CODE=CHI,DAL
ADDITIONAL_FLAGS=-today -prod
GOOGLE_APPLICATION_CREDENTIALS=./path/to/gcp-key.json
```

## Environment Configuration

Create a `.env` file in the project root to configure all profiles:

```bash
# General settings
TZ=America/Chicago
TEAM_CODE=CHI,DAL

# Scheduled profile settings
ADDITIONAL_FLAGS=-local -today

# One-shot profile settings
ONESHOT_ARGS=-local -today -teams CHI

# Production profile settings
GOOGLE_APPLICATION_CREDENTIALS=./secrets/gcp-key.json
```

### Common Environment Variables

| Variable | Description | Default | Example |
|----------|-------------|---------|---------|
| `TZ` | Container timezone | `America/Chicago` | `America/New_York` |
| `TEAM_CODE` | Comma-separated team codes | _(none)_ | `CHI,DAL,BOS` |
| `ADDITIONAL_FLAGS` | Flags for scheduled profile | `-local -today` | `-local -today -all` |
| `ONESHOT_ARGS` | Flags for one-shot profile | `-local -today` | `-local -teams CHI -today` |
| `GOOGLE_APPLICATION_CREDENTIALS` | Path to GCP key file | _(none)_ | `./secrets/gcp-key.json` |

### Application Flags

| Flag | Description |
|------|-------------|
| `-local` | Send to local task queue at http://host.docker.internal:8080 |
| `-host URL` | Custom host URL to send requests to |
| `-today` | Filter for today's upcoming games only |
| `-date YYYY-MM-DD` | Specify a specific date to query |
| `-teams CODE1,CODE2` | Comma-separated team codes (e.g., CHI,DAL) |
| `-all` | Include all teams playing on the specified date |
| `-test` | Run in test mode with predefined game data |
| `-prod` | Send tasks to production Google Cloud Tasks queue |

## Listing Available Profiles

### Using Makefile (Recommended)

```bash
make docker-profiles
```

This displays a formatted list of all profiles with descriptions, use cases, and examples.

### Using Docker Compose

```bash
# List profile names only
docker compose config --profiles

# View full configuration for a specific profile
docker compose --profile scheduled config

# View all services with their profiles
docker compose config --services
```

## Running Multiple Profiles

You can run multiple profiles simultaneously:

```bash
# Run both scheduled and one-shot
docker compose --profile scheduled --profile oneshot up -d

# Stop specific profiles
docker compose --profile scheduled down
docker compose --profile oneshot down

# Stop all profiles
make docker-stop
```

## Management Commands

```bash
# Stop all containers
make docker-stop
# or
docker compose down --remove-orphans

# View logs from all running containers
make docker-logs
# or
docker compose logs -f

# Clean up everything (containers, networks, volumes)
make docker-clean
# or
docker compose down -v --remove-orphans

# Rebuild containers after code changes
docker compose build
docker compose --profile scheduled up -d --force-recreate
```

## Checking Container Status

```bash
# List running containers
docker compose ps

# Check specific profile status
docker compose --profile scheduled ps

# View logs
docker compose logs app-scheduled
docker compose logs -f app-oneshot  # Follow logs

# Execute commands in running container
docker compose exec app-scheduled /bin/sh
docker compose exec app-scheduled crontab -l
```

## Customizing the Cron Schedule

To change when the scheduled task runs, edit `docker/crontab`:

```bash
# Current: Every Monday at 5:00 AM
0 5 * * 1 /app/run-task.sh >> /var/log/cron.log 2>&1

# Every day at 3:00 AM
0 3 * * * /app/run-task.sh >> /var/log/cron.log 2>&1

# Every 6 hours
0 */6 * * * /app/run-task.sh >> /var/log/cron.log 2>&1

# Twice a week (Monday and Thursday at 5:00 AM)
0 5 * * 1,4 /app/run-task.sh >> /var/log/cron.log 2>&1
```

After changing the schedule, rebuild and restart:
```bash
docker compose --profile scheduled down
docker compose build
docker compose --profile scheduled up -d
```

## Troubleshooting

### Container won't start

```bash
# Check logs
docker compose logs app-scheduled

# Validate compose file
docker compose config

# Check environment variables
docker compose config | grep -A 5 environment
```

### Cron not executing tasks

```bash
# Check if cron is running
docker compose exec app-scheduled pgrep crond

# View cron logs
docker compose exec app-scheduled cat /var/log/cron.log

# List installed crontab
docker compose exec app-scheduled crontab -l

# Manually trigger the task
docker compose exec app-scheduled /app/run-task.sh
```

### Can't connect to host.docker.internal

This is needed for the `-local` flag. Ensure:
1. Your local task queue is running on port 8080
2. Docker Desktop is running (for Docker Desktop users)
3. The `extra_hosts` configuration is present in docker-compose.yml

### GCP credentials not working (prod profile)

```bash
# Verify the file exists
ls -la ${GOOGLE_APPLICATION_CREDENTIALS}

# Check the path in the container
docker compose exec app-prod ls -la /secrets/

# Verify environment variable is set
docker compose exec app-prod env | grep GOOGLE
```

## Best Practices

1. **Use .env file** - Don't hardcode values in docker-compose.yml
2. **Version control** - Add `.env` to `.gitignore`, commit `.env.example`
3. **Security** - Never commit GCP credentials or sensitive data
4. **Logging** - Use `docker compose logs` to monitor execution
5. **Testing** - Test with `oneshot` profile before deploying `scheduled`
6. **Health checks** - Monitor container health with `docker compose ps`

## Examples

### Local Development Workflow

```bash
# 1. Test with one-shot
make docker-oneshot

# 2. If successful, run scheduled
make docker-scheduled

# 3. Monitor logs
make docker-logs

# 4. When done
make docker-stop
```

### Production Deployment

```bash
# 1. Set up .env file
cat > .env << EOF
TZ=America/Chicago
TEAM_CODE=CHI,DAL
GOOGLE_APPLICATION_CREDENTIALS=./secrets/gcp-key.json
EOF

# 2. Verify credentials exist
ls -la ./secrets/gcp-key.json

# 3. Deploy
make docker-prod

# 4. Verify it's running
docker compose ps
docker compose logs -f app-prod
```

### Testing Multiple Teams

```bash
# Test for specific teams
ONESHOT_ARGS="-local -today -teams CHI,DAL,BOS,NYR" make docker-oneshot

# Test for all teams
ONESHOT_ARGS="-local -today -all" make docker-oneshot

# Test for specific date
ONESHOT_ARGS="-local -date 2026-03-20 -teams CHI" make docker-oneshot
```

## Additional Resources

- [Docker Compose Profiles Documentation](https://docs.docker.com/compose/profiles/)
- [Cron Expression Syntax](https://crontab.guru/)
- Main README: `README.md`
- Docker installation guide: `DOCKER_INSTALL.md`
