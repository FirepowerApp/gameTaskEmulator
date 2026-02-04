# Docker Installation Guide

This guide explains how to install and run the Game Task Emulator using Docker with scheduled execution.

## Overview

The Docker installation provides a containerized scheduled execution that runs the Game Task Emulator **every Monday at 5:00 AM**. This is an alternative to the systemd-based installation and is ideal for:

- **Local development** - Sends tasks to local task queue at `http://host.docker.internal:8080` (default behavior)
- Running on systems without systemd (e.g., macOS, Windows with Docker Desktop)
- Container-based infrastructure (Kubernetes, Docker Swarm, etc.)
- Cloud hosting environments
- Simplified deployment and isolation

**Default Configuration**: By default, the installation uses the `-local` flag to send tasks to a local task queue running on your machine at `http://host.docker.internal:8080`. This is the typical development workflow.

## Comparison: Systemd vs Docker

| Feature | Systemd (install.sh) | Docker (docker-install.sh) |
|---------|---------------------|----------------------------|
| Schedule | Daily at 6:00 AM | **Every Monday at 5:00 AM** |
| Platform | Linux with systemd | Any platform with Docker |
| Installation | System-wide | Containerized |
| Logs | journalctl | docker logs |
| Updates | Manual rebuild | Rebuild container |

## Prerequisites

1. **Docker** installed and running
   - Docker Desktop (macOS/Windows)
   - Docker Engine (Linux)
   - Verify: `docker --version`

2. **Docker Compose** (optional, for docker-compose.yml method)
   - Included with Docker Desktop
   - Linux: Install separately or use `docker compose`
   - Verify: `docker-compose --version` or `docker compose version`

## Quick Start

### Method 1: Automated Installation Script (Recommended)

The easiest way to get started:

```bash
# Basic installation (sends to local task queue, Dallas Stars)
./docker-install.sh

# Install for specific team (sends to local task queue)
./docker-install.sh --team CHI

# Install for multiple teams (sends to local task queue)
./docker-install.sh --team CHI,DAL,BOS

# Custom timezone (still sends to local task queue)
./docker-install.sh --team CHI --timezone America/New_York

# Production mode with GCP credentials (not local)
./docker-install.sh --team DAL --flags "-today -prod" --credentials ./gcp-key.json
```

**Note**: By default, all commands use the `-local` flag to send tasks to `http://host.docker.internal:8080` on your local machine. Make sure you have a local task queue running to receive these tasks.

### Method 2: Docker Compose

Using docker-compose.yml for deployment:

1. **Edit docker-compose.yml** to configure your settings:

```yaml
environment:
  - TZ=America/Chicago          # Your timezone
  - TEAM_CODE=CHI               # Your team(s)
  - ADDITIONAL_FLAGS=-today     # Application flags
```

2. **Start the container**:

```bash
# Build and start in detached mode
docker-compose up -d

# View logs
docker-compose logs -f

# Stop the container
docker-compose down
```

### Method 3: Manual Docker Commands

For complete control:

1. **Build the image**:

```bash
docker build -f Dockerfile.scheduled -t gametask-emulator:scheduled .
```

2. **Run the container**:

```bash
docker run -d \
  --name gametask-emulator-scheduled \
  --restart unless-stopped \
  -e TZ=America/Chicago \
  -e TEAM_CODE=CHI \
  -e ADDITIONAL_FLAGS="-today" \
  --add-host=host.docker.internal:host-gateway \
  gametask-emulator:scheduled
```

## Local Development Setup

The default installation is configured for **local development**, sending tasks to a local task queue running on your machine.

### Requirements for Local Development

1. **Local Task Queue**: You must have a task queue service running locally on port 8080
2. **Docker Networking**: The container uses `host.docker.internal:8080` to reach your local machine
3. **Default Flags**: The installation automatically uses `-local -today` flags

### How It Works

When you run:
```bash
./docker-install.sh --team CHI
```

The container will:
1. Run every Monday at 5:00 AM
2. Execute: `/app/gameTaskEmulator -local -today -teams CHI`
3. Send tasks to: `http://host.docker.internal:8080`

This is equivalent to running on your local machine:
```bash
./bin/gameTaskEmulator -local -today -teams CHI
```

### Testing Your Local Setup

To manually trigger a task and verify your local queue is receiving tasks:

```bash
# Run the task immediately (without waiting for the scheduled time)
docker exec gametask-emulator-scheduled /app/run-task.sh

# Watch the logs
docker logs -f gametask-emulator-scheduled
```

## Configuration

### Environment Variables

| Variable | Description | Default | Example |
|----------|-------------|---------|---------|
| `TZ` | Container timezone | UTC | `America/Chicago` |
| `TEAM_CODE` | NHL team code(s), comma-separated | (none) | `CHI`, `CHI,DAL,BOS` |
| `ADDITIONAL_FLAGS` | Flags passed to gameTaskEmulator | `-local -today` | `-local -today`, `-today -prod` |
| `GOOGLE_APPLICATION_CREDENTIALS` | Path to GCP credentials inside container | (none) | `/secrets/gcp-key.json` |

### Available Application Flags

- `-local` - **[DEFAULT]** Send to local task queue at `http://host.docker.internal:8080`
- `-today` - Filter to today's upcoming games only
- `-all` - Include all teams
- `-prod` - Use production Cloud Tasks queue (requires GCP credentials)
- `-host URL` - Custom host URL for task delivery
- `-date YYYY-MM-DD` - Specific date (default: today)
- `-teams ID1,ID2` - Team IDs or city codes (set via TEAM_CODE env var)
- `-test` - Test mode with predefined data
- `-project PROJECT_ID` - GCP Project ID (default: localproject)
- `-location LOCATION` - GCP Location (default: us-south1)
- `-queue QUEUE_NAME` - Task Queue name (default: gameschedule)

**Important**: Either `-local` or `-host` must be specified. The default installation uses `-local` to send tasks to a local task queue.

### Team Codes

Common NHL team codes:
- Chicago Blackhawks: `CHI`
- Dallas Stars: `DAL`
- Boston Bruins: `BOS`
- New York Rangers: `NYR`
- Toronto Maple Leafs: `TOR`

See [NHL API documentation](https://github.com/FirepowerApp/gameTaskEmulator#team-codes) for complete list.

## Cron Schedule

The container uses cron to schedule execution. The default schedule is:

```cron
0 5 * * 1  # Every Monday at 5:00 AM
```

To customize the schedule:

1. Edit `docker/crontab`
2. Rebuild the Docker image
3. Restart the container

Cron format: `minute hour day month weekday`
- `0 5 * * 1` = Every Monday at 5:00 AM
- `0 6 * * *` = Every day at 6:00 AM
- `0 5 * * 1-5` = Every weekday at 5:00 AM

## Production Deployment with Google Cloud

To use production Google Cloud Tasks:

1. **Obtain GCP credentials**:
   - Create a service account in GCP Console
   - Grant Cloud Tasks permissions
   - Download the JSON key file

2. **Install with credentials**:

```bash
./docker-install.sh \
  --team CHI \
  --flags "-today -prod" \
  --credentials ./path/to/gcp-key.json
```

Or with Docker Compose:

```yaml
services:
  gametask-emulator-scheduled:
    environment:
      - ADDITIONAL_FLAGS=-today -prod
      - GOOGLE_APPLICATION_CREDENTIALS=/secrets/gcp-key.json
    volumes:
      - ./path/to/gcp-key.json:/secrets/gcp-key.json:ro
```

## Managing the Container

### View Logs

```bash
# All container logs (including startup and cron daemon)
docker logs -f gametask-emulator-scheduled

# Only cron job execution logs
docker exec gametask-emulator-scheduled tail -f /var/log/cron.log
```

### Check Cron Schedule

```bash
docker exec gametask-emulator-scheduled crontab -l
```

### Run Task Manually

To test without waiting for the scheduled time:

```bash
docker exec gametask-emulator-scheduled /app/run-task.sh
```

### Stop/Start Container

```bash
# Stop
docker stop gametask-emulator-scheduled

# Start
docker start gametask-emulator-scheduled

# Restart
docker restart gametask-emulator-scheduled
```

### Remove Container

```bash
# Stop and remove
docker rm -f gametask-emulator-scheduled

# Or use the install script
./docker-install.sh --stop
```

### Update and Rebuild

When you update the code:

```bash
# Pull latest changes
git pull

# Rebuild and restart
./docker-install.sh --team CHI --flags "-today"
```

## Troubleshooting

### Container Won't Start

1. Check Docker daemon is running:
   ```bash
   docker info
   ```

2. Check container logs:
   ```bash
   docker logs gametask-emulator-scheduled
   ```

3. Verify timezone is valid:
   ```bash
   docker run --rm alpine cat /usr/share/zoneinfo/America/Chicago
   ```

### Cron Not Executing

1. Verify cron is running:
   ```bash
   docker exec gametask-emulator-scheduled pgrep crond
   ```

2. Check cron logs:
   ```bash
   docker exec gametask-emulator-scheduled cat /var/log/cron.log
   ```

3. Test manual execution:
   ```bash
   docker exec gametask-emulator-scheduled /app/run-task.sh
   ```

### GCP Authentication Errors

1. Verify credentials are mounted:
   ```bash
   docker exec gametask-emulator-scheduled ls -l /secrets/gcp-key.json
   ```

2. Check environment variable:
   ```bash
   docker exec gametask-emulator-scheduled env | grep GOOGLE_APPLICATION_CREDENTIALS
   ```

3. Test credentials:
   ```bash
   docker exec gametask-emulator-scheduled cat $GOOGLE_APPLICATION_CREDENTIALS
   ```

### Time Zone Issues

The container uses the `TZ` environment variable for timezone. Common values:
- `America/New_York` (Eastern)
- `America/Chicago` (Central)
- `America/Denver` (Mountain)
- `America/Los_Angeles` (Pacific)
- `UTC` (Coordinated Universal Time)

View available timezones:
```bash
docker run --rm alpine ls /usr/share/zoneinfo
```

## Advanced Usage

### Running in Kubernetes

Example Kubernetes CronJob:

```yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: gametask-emulator
spec:
  schedule: "0 5 * * 1"  # Every Monday at 5 AM
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: gametask-emulator
            image: gametask-emulator:scheduled
            env:
            - name: TZ
              value: "America/Chicago"
            - name: TEAM_CODE
              value: "CHI"
            - name: ADDITIONAL_FLAGS
              value: "-today -prod"
          restartPolicy: OnFailure
```

### Custom Dockerfile Modifications

To modify the Dockerfile for your needs:

1. Edit `Dockerfile.scheduled`
2. Rebuild:
   ```bash
   docker build -f Dockerfile.scheduled -t gametask-emulator:scheduled .
   ```

### Using Pre-built Images

If you publish to a container registry:

```bash
# Build and tag
docker build -f Dockerfile.scheduled -t your-registry/gametask-emulator:scheduled .

# Push
docker push your-registry/gametask-emulator:scheduled

# Pull and run on another machine
docker pull your-registry/gametask-emulator:scheduled
docker run -d --name gametask-emulator-scheduled your-registry/gametask-emulator:scheduled
```

## Migration from Systemd

To migrate from systemd installation to Docker:

1. **Stop systemd service**:
   ```bash
   sudo systemctl stop gametask-emulator.timer
   sudo systemctl disable gametask-emulator.timer
   ```

2. **Note your configuration**:
   ```bash
   cat /etc/default/gametask-emulator
   ```

3. **Install Docker version**:
   ```bash
   ./docker-install.sh --team YOUR_TEAM --flags "YOUR_FLAGS"
   ```

4. **Verify it works**:
   ```bash
   docker logs gametask-emulator-scheduled
   ```

## Support

For issues, questions, or contributions:
- GitHub Issues: https://github.com/FirepowerApp/gameTaskEmulator/issues
- Documentation: https://github.com/FirepowerApp/gameTaskEmulator

## License

See [LICENSE](LICENSE) file in the repository.
