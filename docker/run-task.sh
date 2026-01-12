#!/bin/bash
# Task execution script for cron scheduled runs
# This script is called by cron and runs the gameTaskEmulator with configured flags

set -e

# Log execution start
echo "[$(date '+%Y-%m-%d %H:%M:%S')] Starting gameTaskEmulator scheduled run"

# Build command with flags from environment variables
CMD="/app/gameTaskEmulator"
ARGS=""

# Add flags from ADDITIONAL_FLAGS environment variable
if [ -n "${ADDITIONAL_FLAGS}" ]; then
    ARGS="${ADDITIONAL_FLAGS}"
fi

# Add team flag if TEAM_CODE is set
if [ -n "${TEAM_CODE}" ]; then
    ARGS="${ARGS} -teams ${TEAM_CODE}"
fi

# Log the command being executed
echo "[$(date '+%Y-%m-%d %H:%M:%S')] Executing: ${CMD} ${ARGS}"

# Execute the command
${CMD} ${ARGS}

EXIT_CODE=$?

# Log completion
if [ $EXIT_CODE -eq 0 ]; then
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] gameTaskEmulator completed successfully"
else
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] gameTaskEmulator failed with exit code ${EXIT_CODE}"
fi

exit $EXIT_CODE
