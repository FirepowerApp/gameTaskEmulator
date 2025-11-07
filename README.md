# Schedule Game Trackers

This program fetches NHL game schedules and creates Google Cloud Tasks for game tracking. It's designed to work with the CrashTheCrease backend system to automatically schedule game monitoring tasks.

## Features

- **Automatic Game Detection**: Fetches games for today or a specified date using the NHL API
- **Team Filtering**: Supports filtering games by specific teams using city codes (CHI, DAL) or numeric IDs (defaults to Dallas Stars)
- **Today's Upcoming Games**: New `-today` flag filters to only today's games that haven't started yet
- **Flexible Scheduling**: Can schedule tasks for future dates
- **Test Mode**: Includes a test mode with predefined game data for development
- **Production Support**: Configurable for both local development and production environments
- **Cloud Task Integration**: Creates Google Cloud Tasks that integrate with the existing game monitoring system
