#!/bin/bash
# Base execution script for gameTaskEmulator
# This script pulls the latest container image and runs it with passed-through flags

set -e

# Configuration
export DOCKER_IMAGE="${DOCKER_IMAGE:-blnelson/firepowergametaskemulator:latest}"
FALLBACK_MODE="smart"  # smart (default), none (--no-fallback), lenient (--lenient)

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
        --no-fallback)
            FALLBACK_MODE="none"
            shift
            ;;
        --lenient)
            FALLBACK_MODE="lenient"
            shift
            ;;
        --force-pull)
            # Keep for backwards compatibility
            log_warn "The --force-pull flag is deprecated. Use --no-fallback instead."
            FALLBACK_MODE="none"
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
    docker image inspect "${DOCKER_IMAGE}" &> /dev/null
}

# Export environment variables for docker-compose
export TZ=$(cat /etc/timezone 2>/dev/null || echo 'UTC')
if [ -n "${GOOGLE_APPLICATION_CREDENTIALS}" ] && [ -f "${GOOGLE_APPLICATION_CREDENTIALS}" ]; then
    export GOOGLE_APPLICATION_CREDENTIALS
    log_info "Mounting Google Cloud credentials from ${GOOGLE_APPLICATION_CREDENTIALS}"
fi

# Check generic internet connectivity before attempting pull
check_internet() {
    # Try multiple common DNS servers with short timeout
    if timeout 2 ping -c 1 8.8.8.8 &>/dev/null || timeout 2 ping -c 1 1.1.1.1 &>/dev/null; then
        return 0
    fi
    return 1
}

INTERNET_AVAILABLE=false
if check_internet; then
    INTERNET_AVAILABLE=true
    log_info "Internet connection detected"
else
    log_warn "No internet connection detected"
fi

# Determine pull strategy based on internet availability and fallback mode
if [ "$INTERNET_AVAILABLE" = false ]; then
    # No internet connection
    if [ "$FALLBACK_MODE" = "none" ]; then
        log_error "No internet connection and --no-fallback mode enabled"
        log_error "Cannot pull image from registry"
        exit 1
    fi

    # Try to use cached image
    if check_local_image; then
        log_info "Using locally cached image: ${DOCKER_IMAGE}"
        log_warn "Internet unavailable - cannot check for updates"
    else
        log_error "No local image found: ${DOCKER_IMAGE}"
        log_error "Cannot pull from registry (no internet) and no cached image available"
        exit 1
    fi
else
    # Internet is available, attempt to pull
    log_info "Pulling latest container image: ${DOCKER_IMAGE}"

    PULL_ERROR=$(docker compose pull 2>&1)
    PULL_EXIT_CODE=$?

    if [ $PULL_EXIT_CODE -eq 0 ]; then
        log_info "Successfully pulled image: ${DOCKER_IMAGE}"
    else
        # Pull failed - analyze the error
        IS_AUTH_ERROR=false
        IS_NOT_FOUND_ERROR=false
        IS_NETWORK_ERROR=false

        if echo "$PULL_ERROR" | grep -qiE '(unauthorized|authentication required|denied|forbidden)'; then
            IS_AUTH_ERROR=true
        elif echo "$PULL_ERROR" | grep -qiE '(not found|manifest unknown)'; then
            IS_NOT_FOUND_ERROR=true
        elif echo "$PULL_ERROR" | grep -qiE '(dial tcp|connection refused|timeout|network unreachable|no route to host|temporary failure|name resolution|failed to resolve)'; then
            IS_NETWORK_ERROR=true
        fi

        # Handle errors based on fallback mode
        if [ "$FALLBACK_MODE" = "none" ]; then
            # No fallback mode - fail on any error
            log_error "Pull failed and --no-fallback mode enabled"
            if [ "$IS_AUTH_ERROR" = true ]; then
                log_error "Authentication failed. Run: docker login"
            elif [ "$IS_NOT_FOUND_ERROR" = true ]; then
                log_error "Image not found: ${DOCKER_IMAGE}"
            elif [ "$IS_NETWORK_ERROR" = true ]; then
                log_error "Network error accessing Docker Hub"
            else
                log_error "Error: ${PULL_ERROR}"
            fi
            exit 1
        elif [ "$FALLBACK_MODE" = "lenient" ]; then
            # Lenient mode - try to fall back on any error
            log_warn "Pull failed, attempting to use cached image (--lenient mode)"
            if [ "$IS_AUTH_ERROR" = true ]; then
                log_warn "Authentication error: ${PULL_ERROR}"
            elif [ "$IS_NOT_FOUND_ERROR" = true ]; then
                log_warn "Image not found: ${PULL_ERROR}"
            elif [ "$IS_NETWORK_ERROR" = true ]; then
                log_warn "Network error: ${PULL_ERROR}"
            else
                log_warn "Error: ${PULL_ERROR}"
            fi

            if check_local_image; then
                log_warn "Using locally cached image: ${DOCKER_IMAGE}"
                log_warn "This may not be the latest version"
            else
                log_error "No local image found: ${DOCKER_IMAGE}"
                log_error "Cannot use cached image - none available"
                exit 1
            fi
        else
            # Smart mode (default) - only fall back on network errors
            if [ "$IS_AUTH_ERROR" = true ]; then
                log_error "Authentication failed when pulling image"
                log_error "Make sure you are authenticated to Docker Hub:"
                log_error "  docker login"
                exit 1
            elif [ "$IS_NOT_FOUND_ERROR" = true ]; then
                log_error "Image not found in registry: ${DOCKER_IMAGE}"
                log_error "Please verify the image name is correct"
                exit 1
            elif [ "$IS_NETWORK_ERROR" = true ]; then
                log_warn "Network error while accessing Docker Hub"
                log_warn "Error: ${PULL_ERROR}"

                if check_local_image; then
                    log_warn "Using locally cached image: ${DOCKER_IMAGE}"
                    log_warn "This may not be the latest version"
                else
                    log_error "No local image found: ${DOCKER_IMAGE}"
                    log_error "Unable to pull from registry and no cached image available"
                    exit 1
                fi
            else
                # Unknown error - fail in smart mode
                log_error "Unknown error while pulling image"
                log_error "Error: ${PULL_ERROR}"
                exit 1
            fi
        fi
    fi
fi

log_info "Running container with provided flags: ${CONTAINER_ARGS[*]}"

# Run the container using docker compose
docker compose run --rm game-task-emulator "${CONTAINER_ARGS[@]}"

log_info "Container execution completed"
