#!/bin/bash
# Entrypoint script for scheduled gameTaskEmulator container
# Sets up cron and starts the cron daemon in foreground mode

set -e

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_info "Starting gameTaskEmulator scheduled container"
log_info "Timezone: ${TZ:-UTC}"
log_info "Schedule: Every Monday at 5:00 AM"

# Display configuration
if [ -n "${TEAM_CODE}" ]; then
    log_info "Team Code: ${TEAM_CODE}"
else
    log_warn "TEAM_CODE not set - will use default team"
fi

if [ -n "${ADDITIONAL_FLAGS}" ]; then
    log_info "Additional Flags: ${ADDITIONAL_FLAGS}"
else
    log_info "Additional Flags: (none)"
fi

# Make sure run-task.sh is executable
chmod +x /app/run-task.sh

# Set up cron job from crontab file
log_info "Setting up cron schedule..."

# Copy crontab for the current user
crontab /app/docker/crontab

log_info "Cron schedule installed:"
crontab -l

# Create log file if it doesn't exist
touch /var/log/cron.log
chmod 666 /var/log/cron.log

log_info "Container is ready and cron daemon is starting..."
log_info "Logs will be written to /var/log/cron.log and stdout"
log_info ""
log_info "To view logs in real-time: docker logs -f <container-name>"
log_info ""

# Tail the cron log file in the background to output to stdout
# The -F flag follows the file and retries if it's temporarily unavailable
tail -F /var/log/cron.log &

# Start crond in foreground mode
# -f: foreground mode
# -l 2: log level (2 = default, 8 = debug)
exec crond -f -l 2
