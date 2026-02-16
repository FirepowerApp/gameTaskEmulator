.PHONY: help docker-profiles docker-scheduled docker-oneshot docker-dev docker-prod docker-stop docker-logs docker-clean

# Default target - show help
help:
	@echo ""
	@echo "╔════════════════════════════════════════════════════════════════════╗"
	@echo "║        Game Task Emulator - Docker Compose Commands               ║"
	@echo "╚════════════════════════════════════════════════════════════════════╝"
	@echo ""
	@echo "Available profiles and commands:"
	@echo ""
	@echo "  make docker-profiles    - List all available Docker Compose profiles"
	@echo ""
	@echo "Profile-specific commands:"
	@echo "  make docker-scheduled   - Run scheduled/cronjob deployment (background)"
	@echo "  make docker-oneshot     - Run task once and exit (foreground)"
	@echo "  make docker-dev         - Run development/interactive mode (shell)"
	@echo "  make docker-prod        - Run production deployment (background)"
	@echo ""
	@echo "Management commands:"
	@echo "  make docker-stop        - Stop all running containers"
	@echo "  make docker-logs        - View logs from running containers"
	@echo "  make docker-clean       - Stop and remove all containers"
	@echo ""
	@echo "For more information, run: make docker-profiles"
	@echo ""

# List all available profiles with detailed descriptions
docker-profiles:
	@echo ""
	@echo "╔════════════════════════════════════════════════════════════════════╗"
	@echo "║           Docker Compose Profiles - Game Task Emulator            ║"
	@echo "╚════════════════════════════════════════════════════════════════════╝"
	@echo ""
	@echo "📋 Available Profiles:"
	@echo ""
	@echo "┌─────────────────────────────────────────────────────────────────────┐"
	@echo "│ 1. SCHEDULED                                                        │"
	@echo "│    Description: Runs tasks on a cron schedule (Monday 5:00 AM)     │"
	@echo "│    Use Case:    Automated weekly game scheduling                   │"
	@echo "│    Container:   Runs continuously in the background with cron      │"
	@echo "│                                                                     │"
	@echo "│    Quick Start:                                                     │"
	@echo "│      make docker-scheduled                                          │"
	@echo "│                                                                     │"
	@echo "│    Manual:                                                          │"
	@echo "│      docker compose --profile scheduled up -d                       │"
	@echo "└─────────────────────────────────────────────────────────────────────┘"
	@echo ""
	@echo "┌─────────────────────────────────────────────────────────────────────┐"
	@echo "│ 2. ONESHOT                                                          │"
	@echo "│    Description: Runs task once and exits                           │"
	@echo "│    Use Case:    Manual execution, testing, or triggered runs       │"
	@echo "│    Container:   Exits after completion                             │"
	@echo "│                                                                     │"
	@echo "│    Quick Start:                                                     │"
	@echo "│      make docker-oneshot                                            │"
	@echo "│                                                                     │"
	@echo "│    Manual:                                                          │"
	@echo "│      docker compose --profile oneshot up                            │"
	@echo "│                                                                     │"
	@echo "│    Custom Args:                                                     │"
	@echo "│      ONESHOT_ARGS=\"-local -today -teams CHI\" make docker-oneshot    │"
	@echo "└─────────────────────────────────────────────────────────────────────┘"
	@echo ""
	@echo "┌─────────────────────────────────────────────────────────────────────┐"
	@echo "│ 3. DEV                                                              │"
	@echo "│    Description: Interactive container for development/debugging    │"
	@echo "│    Use Case:    Development, testing, debugging                    │"
	@echo "│    Container:   Provides shell access for manual commands          │"
	@echo "│                                                                     │"
	@echo "│    Quick Start:                                                     │"
	@echo "│      make docker-dev                                                │"
	@echo "│                                                                     │"
	@echo "│    Manual:                                                          │"
	@echo "│      docker compose --profile dev run --rm app-dev                  │"
	@echo "│                                                                     │"
	@echo "│    Inside container, run:                                           │"
	@echo "│      /app/gameTaskEmulator -local -today                            │"
	@echo "└─────────────────────────────────────────────────────────────────────┘"
	@echo ""
	@echo "┌─────────────────────────────────────────────────────────────────────┐"
	@echo "│ 4. PROD                                                             │"
	@echo "│    Description: Production scheduled deployment with GCP creds     │"
	@echo "│    Use Case:    Production environment with Google Cloud Tasks     │"
	@echo "│    Container:   Runs continuously with cron + GCP authentication   │"
	@echo "│                                                                     │"
	@echo "│    Prerequisites:                                                   │"
	@echo "│      - Set GOOGLE_APPLICATION_CREDENTIALS in .env file             │"
	@echo "│      - Ensure GCP key file exists at specified path                │"
	@echo "│                                                                     │"
	@echo "│    Quick Start:                                                     │"
	@echo "│      make docker-prod                                               │"
	@echo "│                                                                     │"
	@echo "│    Manual:                                                          │"
	@echo "│      docker compose --profile prod up -d                            │"
	@echo "└─────────────────────────────────────────────────────────────────────┘"
	@echo ""
	@echo "📝 Environment Configuration:"
	@echo ""
	@echo "   Create a .env file in the project root with:"
	@echo ""
	@echo "   # Basic settings"
	@echo "   TZ=America/Chicago"
	@echo "   TEAM_CODE=CHI,DAL"
	@echo ""
	@echo "   # Scheduled/Local mode"
	@echo "   ADDITIONAL_FLAGS=-local -today"
	@echo ""
	@echo "   # One-shot mode"
	@echo "   ONESHOT_ARGS=-local -today -teams CHI"
	@echo ""
	@echo "   # Production mode"
	@echo "   GOOGLE_APPLICATION_CREDENTIALS=./path/to/gcp-key.json"
	@echo ""
	@echo "For more details, see DOCKER_INSTALL.md"
	@echo ""

# Run scheduled profile (cronjob deployment)
docker-scheduled:
	@echo "Starting scheduled deployment (cronjob)..."
	docker compose --profile scheduled up -d
	@echo ""
	@echo "✓ Scheduled container is running in the background"
	@echo "  View logs: make docker-logs"
	@echo "  Stop:      make docker-stop"

# Run one-shot profile (execute once and exit)
docker-oneshot:
	@echo "Running one-shot execution..."
	@echo "Args: ${ONESHOT_ARGS}"
	docker compose --profile oneshot up
	@echo ""
	@echo "✓ One-shot execution completed"

# Run dev profile (interactive mode)
docker-dev:
	@echo "Starting development/interactive mode..."
	@echo ""
	@echo "You are now in an interactive shell inside the container."
	@echo "To run the application, use:"
	@echo "  /app/gameTaskEmulator -local -today"
	@echo ""
	docker compose --profile dev run --rm app-dev

# Run production profile
docker-prod:
	@echo "Starting production deployment..."
	@if [ -z "$$GOOGLE_APPLICATION_CREDENTIALS" ]; then \
		echo "ERROR: GOOGLE_APPLICATION_CREDENTIALS not set"; \
		echo "Please set it in your .env file or environment"; \
		exit 1; \
	fi
	docker compose --profile prod up -d
	@echo ""
	@echo "✓ Production container is running in the background"
	@echo "  View logs: make docker-logs"
	@echo "  Stop:      make docker-stop"

# Stop all running containers
docker-stop:
	@echo "Stopping all containers..."
	docker compose --profile scheduled --profile oneshot --profile dev --profile prod down
	@echo "✓ All containers stopped"

# View logs from running containers
docker-logs:
	@echo "Viewing logs (Ctrl+C to exit)..."
	docker compose logs -f

# Clean up - stop and remove all containers
docker-clean:
	@echo "Cleaning up all containers and networks..."
	docker compose --profile scheduled --profile oneshot --profile dev --profile prod down -v
	@echo "✓ All containers, networks, and volumes removed"
