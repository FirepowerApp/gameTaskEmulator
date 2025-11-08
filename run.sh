#!/bin/bash
# Base execution script for gameTaskEmulator
# This script pulls the latest container image and runs it with passed-through flags

set -e

# Configuration
IMAGE="${DOCKER_IMAGE:-blnelson/firepowergametaskemulator:latest}"

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

log_info "Pulling latest container image: ${IMAGE}"

# Pull the latest image
if docker pull "${IMAGE}"; then
    log_info "Successfully pulled image: ${IMAGE}"
else
    log_error "Failed to pull image: ${IMAGE}"
    log_warn "Make sure you are authenticated to Docker Hub:"
    log_warn "  docker login"
    exit 1
fi

log_info "Running container with provided flags: $*"

# Run the container with all passed-through arguments
# Mount credentials if they exist
DOCKER_ARGS=()

# Check for Google Cloud credentials
if [ -n "${GOOGLE_APPLICATION_CREDENTIALS}" ]; then
    if [ -f "${GOOGLE_APPLICATION_CREDENTIALS}" ]; then
        DOCKER_ARGS+=(-v "${GOOGLE_APPLICATION_CREDENTIALS}:/secrets/gcp-key.json:ro")
        DOCKER_ARGS+=(-e "GOOGLE_APPLICATION_CREDENTIALS=/secrets/gcp-key.json")
        log_info "Mounting Google Cloud credentials from ${GOOGLE_APPLICATION_CREDENTIALS}"
    fi
fi

# Set timezone to host timezone
DOCKER_ARGS+=(-e "TZ=$(cat /etc/timezone 2>/dev/null || echo 'UTC')")

# Network mode for local development (allows access to host.docker.internal)
if [[ "$*" != *"-prod"* ]]; then
    DOCKER_ARGS+=(--add-host=host.docker.internal:host-gateway)
fi

# Run the container
docker run --rm \
    "${DOCKER_ARGS[@]}" \
    "${IMAGE}" \
    "$@"

log_info "Container execution completed"
