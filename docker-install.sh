#!/bin/bash
# Docker installation script for Game Task Emulator
# This script sets up a Docker container that runs the emulator every Monday at 5:00 AM

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_step() {
    echo -e "${BLUE}[STEP]${NC} $1"
}

# Show usage
usage() {
    cat << EOF
Usage: $0 [OPTIONS]

Install Game Task Emulator as a Docker container with weekly scheduled execution.
The container will run every Monday at 5:00 AM.

OPTIONS:
    -t, --team TEAM_CODE    Team city code (e.g., DAL, CHI, BOS)
                           Can specify multiple teams: -t DAL,CHI,BOS
                           Default: no team (Dallas Stars)

    -f, --flags FLAGS      Additional flags to pass to the application
                           Default: -today
                           Example: -f "-today -prod"

    -z, --timezone TZ      Timezone for the container
                           Default: America/Chicago
                           Example: -z America/New_York

    -c, --credentials FILE Path to Google Cloud credentials JSON file
                           Required if using -prod flag
                           Example: -c /path/to/gcp-key.json

    --build-only          Build the Docker image without starting the container

    --stop                Stop and remove existing container

    -h, --help            Show this help message

EXAMPLES:
    # Install with default settings (Dallas Stars, -today flag)
    $0

    # Install for Chicago Blackhawks
    $0 --team CHI

    # Install for multiple teams
    $0 --team CHI,DAL,BOS

    # Install with production flag and GCP credentials
    $0 --team DAL --flags "-today -prod" --credentials ./gcp-key.json

    # Set custom timezone
    $0 --team CHI --timezone America/New_York

    # Build only without starting
    $0 --build-only

    # Stop existing container
    $0 --stop

EOF
    exit 1
}

# Check dependencies
check_dependencies() {
    log_step "Checking dependencies..."

    local missing_deps=()

    if ! command -v docker &> /dev/null; then
        missing_deps+=("docker")
    fi

    if ! command -v docker-compose &> /dev/null; then
        if ! docker compose version &> /dev/null; then
            log_warn "Neither docker-compose nor 'docker compose' found"
            log_warn "Will use docker commands directly"
        fi
    fi

    if [ ${#missing_deps[@]} -ne 0 ]; then
        log_error "Missing required dependencies: ${missing_deps[*]}"
        log_error "Please install the missing dependencies and try again"
        exit 1
    fi

    # Check if Docker daemon is running
    if ! docker info &> /dev/null; then
        log_error "Docker daemon is not running. Please start Docker."
        exit 1
    fi

    log_info "All dependencies are installed"
}

# Parse command line arguments
TEAM_CODE=""
ADDITIONAL_FLAGS="-today"
TIMEZONE="America/Chicago"
GCP_CREDENTIALS=""
BUILD_ONLY=false
STOP_CONTAINER=false

while [[ $# -gt 0 ]]; do
    case $1 in
        -t|--team)
            TEAM_CODE="$2"
            shift 2
            ;;
        -f|--flags)
            ADDITIONAL_FLAGS="$2"
            shift 2
            ;;
        -z|--timezone)
            TIMEZONE="$2"
            shift 2
            ;;
        -c|--credentials)
            GCP_CREDENTIALS="$2"
            shift 2
            ;;
        --build-only)
            BUILD_ONLY=true
            shift
            ;;
        --stop)
            STOP_CONTAINER=true
            shift
            ;;
        -h|--help)
            usage
            ;;
        *)
            log_error "Unknown option: $1"
            usage
            ;;
    esac
done

# Container name
CONTAINER_NAME="gametask-emulator-scheduled"

# Stop container if requested
if [ "$STOP_CONTAINER" = true ]; then
    log_step "Stopping and removing existing container..."
    if docker ps -a --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"; then
        docker stop "${CONTAINER_NAME}" || true
        docker rm "${CONTAINER_NAME}" || true
        log_info "Container stopped and removed"
    else
        log_info "No existing container found"
    fi
    exit 0
fi

# Main installation
main() {
    check_dependencies

    log_info "Installing Game Task Emulator (Docker Scheduled)"
    log_info "Schedule: Every Monday at 5:00 AM"
    log_info "Timezone: ${TIMEZONE}"
    log_info "Team Code: ${TEAM_CODE:-none (default: Dallas Stars)}"
    log_info "Additional Flags: ${ADDITIONAL_FLAGS}"
    echo

    # Verify credentials file exists if specified
    if [ -n "${GCP_CREDENTIALS}" ]; then
        if [ ! -f "${GCP_CREDENTIALS}" ]; then
            log_error "Credentials file not found: ${GCP_CREDENTIALS}"
            exit 1
        fi
        GCP_CREDENTIALS="$(realpath "${GCP_CREDENTIALS}")"
        log_info "Using GCP credentials: ${GCP_CREDENTIALS}"
    fi

    # Build the Docker image
    log_step "Building Docker image..."
    docker build -f Dockerfile.scheduled -t gametask-emulator:scheduled .
    log_info "Docker image built successfully"

    if [ "$BUILD_ONLY" = true ]; then
        log_info "Build-only mode: Container not started"
        exit 0
    fi

    # Stop existing container if running
    if docker ps -a --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"; then
        log_step "Stopping existing container..."
        docker stop "${CONTAINER_NAME}" || true
        docker rm "${CONTAINER_NAME}" || true
    fi

    # Prepare docker run command
    log_step "Starting container..."

    DOCKER_ARGS=(
        --name "${CONTAINER_NAME}"
        --restart unless-stopped
        -e "TZ=${TIMEZONE}"
        -e "TEAM_CODE=${TEAM_CODE}"
        -e "ADDITIONAL_FLAGS=${ADDITIONAL_FLAGS}"
        --add-host=host.docker.internal:host-gateway
    )

    # Add credentials mount if specified
    if [ -n "${GCP_CREDENTIALS}" ]; then
        DOCKER_ARGS+=(
            -v "${GCP_CREDENTIALS}:/secrets/gcp-key.json:ro"
            -e "GOOGLE_APPLICATION_CREDENTIALS=/secrets/gcp-key.json"
        )
    fi

    # Run the container in detached mode
    docker run -d "${DOCKER_ARGS[@]}" gametask-emulator:scheduled

    log_info "Container started successfully"
    echo

    # Show status
    log_info "Container Status:"
    docker ps --filter "name=${CONTAINER_NAME}" --format "table {{.Names}}\t{{.Status}}\t{{.CreatedAt}}"
    echo

    log_info "Installation complete!"
    echo
    log_info "Useful commands:"
    echo "  View container logs:           docker logs -f ${CONTAINER_NAME}"
    echo "  View cron logs:                docker exec ${CONTAINER_NAME} tail -f /var/log/cron.log"
    echo "  Check cron schedule:           docker exec ${CONTAINER_NAME} crontab -l"
    echo "  Stop container:                docker stop ${CONTAINER_NAME}"
    echo "  Start container:               docker start ${CONTAINER_NAME}"
    echo "  Remove container:              docker rm -f ${CONTAINER_NAME}"
    echo "  Rebuild and restart:           $0 --team ${TEAM_CODE} --flags \"${ADDITIONAL_FLAGS}\""
    echo "  Run task manually now:         docker exec ${CONTAINER_NAME} /app/run-task.sh"
    echo
}

# Run main installation
main
