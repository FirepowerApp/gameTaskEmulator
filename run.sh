#!/bin/bash
# Base execution script for gameTaskEmulator
# This script pulls the latest container image and runs it with passed-through flags

set -e

# Configuration
IMAGE="${DOCKER_IMAGE:-blnelson/firepowergametaskemulator:latest}"
FORCE_PULL=false

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
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

# Parse script-specific flags
CONTAINER_ARGS=()
while [[ $# -gt 0 ]]; do
    case $1 in
        --force-pull)
            FORCE_PULL=true
            shift
            ;;
        *)
            CONTAINER_ARGS+=("$1")
            shift
            ;;
    esac
done

# Check if Docker is installed
if ! command -v docker &> /dev/null; then
    log_error "Docker is not installed. Please install Docker to use this script."
    exit 1
fi

# Check if Docker daemon is running
if ! docker info &> /dev/null; then
    log_error "Docker daemon is not running. Please start Docker."
    exit 1
fi

# Function to check if image exists locally
check_local_image() {
    docker image inspect "${IMAGE}" &> /dev/null
}

log_info "Pulling latest container image: ${IMAGE}"

# Pull the latest image
PULL_SUCCESS=false
if docker pull "${IMAGE}" 2>/dev/null; then
    log_info "Successfully pulled image: ${IMAGE}"
    PULL_SUCCESS=true
else
    log_warn "Failed to pull image from registry"

    if [ "$FORCE_PULL" = true ]; then
        log_error "Force pull mode enabled - failing due to registry pull failure"
        log_warn "Make sure you are authenticated to Docker Hub:"
        log_warn "  docker login"
        exit 1
    fi

    # Try to use local image
    if check_local_image; then
        log_warn "Using locally cached image: ${IMAGE}"
        log_warn "This may not be the latest version"
    else
        log_error "No local image found: ${IMAGE}"
        log_error "Unable to pull from registry and no cached image available"
        log_warn "Make sure you are authenticated to Docker Hub:"
        log_warn "  docker login"
        exit 1
    fi
fi

log_info "Running container with provided flags: ${CONTAINER_ARGS[*]}"

# Run the container with all passed-through arguments
# Mount credentials if they exist
RUN_ARGS=()

# Check for Google Cloud credentials
if [ -n "${GOOGLE_APPLICATION_CREDENTIALS}" ]; then
    if [ -f "${GOOGLE_APPLICATION_CREDENTIALS}" ]; then
        RUN_ARGS+=(-v "${GOOGLE_APPLICATION_CREDENTIALS}:/secrets/gcp-key.json:ro")
        RUN_ARGS+=(-e "GOOGLE_APPLICATION_CREDENTIALS=/secrets/gcp-key.json")
        log_info "Mounting Google Cloud credentials from ${GOOGLE_APPLICATION_CREDENTIALS}"
    fi
fi

# Set timezone to host timezone
RUN_ARGS+=(-e "TZ=$(cat /etc/timezone 2>/dev/null || echo 'UTC')")

# Network mode for local development (allows access to host.docker.internal)
if [[ "${CONTAINER_ARGS[*]}" != *"-prod"* ]]; then
    RUN_ARGS+=(--add-host=host.docker.internal:host-gateway)
fi

# Run the container
docker run --rm \
    "${RUN_ARGS[@]}" \
    "${IMAGE}" \
    "${CONTAINER_ARGS[@]}"

log_info "Container execution completed"
