# Dailies

A task management application with scheduled recurring tasks, tags, and frequencies. Built with Go backend and Angular frontend.

![screenshot](https://github.com/jhoffmann/dailies/blob/main/screenshot.png?raw=true)

## Features

- Task management with CRUD operations
- Recurring tasks with customizable frequencies
- Tag-based organization
- Real-time updates via WebSocket
- RESTful API
- SQLite database
- Docker support

## Tech Stack

**Backend:**

- Go 1.25.1
- Gin web framework
- GORM ORM
- SQLite database
- WebSocket support
- Cron-based task scheduling

**Frontend:**

- Angular 20+

## Getting Started

### Prerequisites

- Go 1.25.1+
- Node.js and npm (for frontend)
- Docker and Docker Compose (recommended for deployment)

### Local Development

**Backend:**

```bash
# Install dependencies
go mod download

# Run with hot reload (requires air)
air

# Or run directly
go run main.go

# Build
go build -o build/api

# Run tests
go test ./...
```

**Frontend:**

```bash
cd frontends/web
npm install
npm start
```

The API server runs on port 8080 by default, and the frontend dev server runs on port 4200.

### Docker Deployment

```bash
docker-compose up -d
```

- API: http://localhost:9002
- Web: http://localhost:9003

## Configuration

Environment variables:

- `DB_PATH`: Path to SQLite database (default: `./tasks.db`)
- `DB_TIMEZONE`: Timezone for scheduled tasks (default: `MST7MDT`)
- `GIN_MODE`: Gin mode (`debug` or `release`)
- `PORT`: Server port (default: `8080`)

## API Endpoints

### Tasks

- `GET /api/tasks` - List all tasks
- `GET /api/tasks/:id` - Get task by ID
- `POST /api/tasks` - Create task
- `PUT /api/tasks/:id` - Update task
- `DELETE /api/tasks/:id` - Delete task

### Frequencies

- `GET /api/frequencies` - List all frequencies
- `GET /api/frequencies/:id` - Get frequency by ID
- `GET /api/frequencies/timers` - Get frequency timers
- `POST /api/frequencies` - Create frequency
- `PUT /api/frequencies/:id` - Update frequency
- `DELETE /api/frequencies/:id` - Delete frequency

### Tags

- `GET /api/tags` - List all tags
- `GET /api/tags/:id` - Get tag by ID
- `POST /api/tags` - Create tag
- `PUT /api/tags/:id` - Update tag
- `DELETE /api/tags/:id` - Delete tag

### Other

- `GET /health` - Health check
- `GET /ws` - WebSocket connection
- `GET /api/timezone` - Get server timezone info
