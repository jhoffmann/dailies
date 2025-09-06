# Dailies

A web-based task tracking application designed for managing daily routines and recurring tasks with automated reset capabilities.

![Dailies Screenshot](assets/screenshot.webp)

## Overview

Dailies is a modern task management system that helps users track their daily activities, habits, and recurring tasks. The application features automatic task reset functionality based on customizable frequency schedules, allowing users to maintain consistent daily routines without manual intervention.

Key features include:

- **Task Management**: Create, edit, and track completion of daily tasks
- **Smart Scheduling**: Automatic task reset based on cron-style frequency schedules
- **Organization Tools**: Tag-based categorization and priority-based sorting
- **Real-time Updates**: WebSocket integration for live task list updates
- **Responsive Interface**: HTMX-powered frontend for seamless user interactions

## Use Cases

- **MMORPG Activities**: Keep track of daily/weekly/monthly in-game tasks, quests, or resource gathering routines
- **Daily Routines**: Track morning routines, exercise, meditation, or reading habits
- **Work Tasks**: Manage recurring work activities like code reviews, standup preparation, or report generation
- **Health & Wellness**: Monitor daily health activities, medication schedules, or wellness check-ins
- **Personal Development**: Track learning goals, skill practice, or creative activities
- **Household Management**: Organize daily chores, maintenance tasks, or family activities
- **Team Coordination**: Share recurring team tasks with automatic reset cycles

## Technologies

### Backend

- **Go 1.24.6**: Core web services and API
- **GORM**: Database ORM with SQLite support
- **Cron Scheduler**: Automated task reset functionality
- **WebSocket**: Real-time notifications and updates

### Frontend

- **HTMX**: Dynamic HTML interactions without JavaScript frameworks
- **Hyperscript**: Animation and interactivity enhancements
- **HTML Templates**: Server-side rendered components

### Development & Build

- **Webpack**: Asset bundling and optimization
- **Bun**: Package management and frontend tooling
- **Air**: Hot reload development server
- **Docker**: Containerization support
- **Mise-en-place**: Development tasks

### Infrastructure

- **SQLite**: Lightweight embedded database
- **UUID**: Distributed identifier system
- **JSON API**: RESTful API design
- **Embedded Assets**: Self-contained binary deployment

## Getting Started

### Development

```bash
# Install dependencies
bun install
mise run dev

# Start development server with hot reload
bun run serve

# Build frontend assets
bun run build
```

### Production

```bash
# Build the application
mise run build

# Run the server
./server --address :9001 --db production.db
```

The application will be available at `http://localhost:9001`.
