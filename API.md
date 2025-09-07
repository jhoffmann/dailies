# Dailies API Documentation

This document describes the REST API endpoints for the Dailies task management application.

## Base URL

```
http://localhost:8080
```

## Content Type

All endpoints accept and return `application/json` unless otherwise specified.

**Note:** Examples use [HTTPie](https://httpie.io/) command-line tool. Install with `pip install httpie` or use your preferred HTTP client.

---

## Health Check

### Check Application Health

Get the health status of the application.

**Endpoint:** `GET /healthz`

**Example Request:**

```bash
http GET :8080/healthz
```

**Response (200 OK):**

```json
{
  "status": "UP"
}
```

---

## Tasks

### Get All Tasks

Retrieve all tasks with optional filtering and sorting.

**Endpoint:** `GET /tasks`

**Query Parameters:**

- `completed` (boolean, optional) - Filter by completion status
- `name` (string, optional) - Filter by task name (partial matching)
- `tag_ids` (string, optional) - Filter by tag IDs (comma-separated UUIDs, tasks must have ALL specified tags)
- `sort` (string, optional) - Sort field: `completed`, `priority`, `name` (default: `priority`)

**Example Requests:**

```bash
# Get all tasks
http GET :8080/tasks

# Get completed tasks sorted by name
http GET :8080/tasks completed==true sort==name

# Search for tasks containing "review"
http GET :8080/tasks name==review

# Get tasks with specific tag
http GET :8080/tasks tag_ids==b5948bc2-d918-5fe5-ca5a-e8e48f406cf8

# Get tasks that have both "work" and "urgent" tags (AND operation)
http GET :8080/tasks tag_ids==b5948bc2-d918-5fe5-ca5a-e8e48f406cf8,c6059cd3-ea29-6gf6-db6b-f9f59g517dg9

# Combine tag filtering with other filters
http GET :8080/tasks tag_ids==b5948bc2-d918-5fe5-ca5a-e8e48f406cf8 completed==false name==review sort==priority
```

**Response (200 OK):**

```json
[
  {
    "id": "a4837ac1-c807-4edd-ba49-d4e37f295be7",
    "name": "Review pull requests",
    "date_created": "2025-09-05T21:17:04.323385Z",
    "date_modified": "2025-09-05T21:17:04.323385Z",
    "completed": false,
    "priority": 3,
    "tags": [
      {
        "id": "b5948bc2-d918-5fe5-ca5a-e8e48f406cf8",
        "name": "work",
        "color": "#bfdbfe"
      },
      {
        "id": "c6059cd3-ea29-6gf6-db6b-f9f59g517dg9",
        "name": "urgent",
        "color": "#fca5a5"
      }
    ]
  }
]
```

### Get Single Task

Retrieve a specific task by its ID.

**Endpoint:** `GET /tasks/{id}`

**Path Parameters:**

- `id` (string, required) - Task UUID

**Example Request:**

```bash
http GET :8080/tasks/a4837ac1-c807-4edd-ba49-d4e37f295be7
```

**Response (200 OK):**

```json
{
  "id": "a4837ac1-c807-4edd-ba49-d4e37f295be7",
  "name": "Review pull requests",
  "date_created": "2025-09-05T21:17:04.323385Z",
  "date_modified": "2025-09-05T21:17:04.323385Z",
  "completed": false,
  "priority": 3,
  "frequency_id": "f1234567-ab12-34cd-56ef-123456789012",
  "frequency": {
    "id": "f1234567-ab12-34cd-56ef-123456789012",
    "name": "Daily Evening",
    "reset": "0 18 * * *"
  },
  "tags": [
    {
      "id": "b5948bc2-d918-5fe5-ca5a-e8e48f406cf8",
      "name": "work",
      "color": "#bfdbfe"
    }
  ]
}
```

**Error Response (404 Not Found):**

```json
{
  "error": "task not found"
}
```

### Create Task

Create a new task with optional tag associations and frequency assignment.

**Endpoint:** `POST /tasks`

**Request Body:**

```json
{
  "name": "Complete API documentation",
  "frequency_id": "f1234567-ab12-34cd-56ef-123456789012",
  "tag_ids": [
    "b5948bc2-d918-5fe5-ca5a-e8e48f406cf8",
    "c6059cd3-ea29-6gf6-db6b-f9f59g517dg9"
  ]
}
```

**Example Requests:**

```bash
# Create task without frequency
http POST :8080/tasks \
  name="Complete API documentation" \
  tag_ids:='["b5948bc2-d918-5fe5-ca5a-e8e48f406cf8"]'

# Create task with frequency
http POST :8080/tasks \
  name="Daily standup meeting" \
  frequency_id="f1234567-ab12-34cd-56ef-123456789012"

# Create task with both frequency and tags
http POST :8080/tasks \
  name="Review and deploy code" \
  frequency_id="f1234567-ab12-34cd-56ef-123456789012" \
  tag_ids:='["b5948bc2-d918-5fe5-ca5a-e8e48f406cf8"]'
```

> **Note:** Frequency assignment for tasks is currently handled through direct database operations. Future API versions will support `frequency_id` in task creation and updates.

**Response (201 Created):**

```json
{
  "id": "d7160de4-fb3a-7hg7-ec7c-g0g60h628eh0",
  "name": "Complete API documentation",
  "date_created": "2025-09-05T22:30:15.456789Z",
  "date_modified": "2025-09-05T22:30:15.456789Z",
  "completed": false,
  "priority": 3,
  "tags": [
    {
      "id": "b5948bc2-d918-5fe5-ca5a-e8e48f406cf8",
      "name": "work",
      "color": "#bfdbfe"
    }
  ]
}
```

**Error Response (400 Bad Request):**

```json
{
  "error": "task name is required"
}
```

### Update Task

Update an existing task by its ID, including completion status, priority, frequency assignment, and tag associations.

**Endpoint:** `PUT /tasks/{id}`

**Path Parameters:**

- `id` (string, required) - Task UUID

**Request Body:**

```json
{
  "name": "Updated task name",
  "completed": true,
  "priority": 1,
  "frequency_id": "f1234567-ab12-34cd-56ef-123456789012",
  "tag_ids": [
    "b5948bc2-d918-5fe5-ca5a-e8e48f406cf8",
    "c6059cd3-ea29-6gf6-db6b-f9f59g517dg9"
  ]
}
```

**Example Requests:**

```bash
# Update task completion status
http PUT :8080/tasks/a4837ac1-c807-4edd-ba49-d4e37f295be7 \
  completed:=true \
  priority:=1

# Add frequency to existing task
http PUT :8080/tasks/a4837ac1-c807-4edd-ba49-d4e37f295be7 \
  frequency_id="f1234567-ab12-34cd-56ef-123456789012"

# Remove frequency from task (set to null)
http PUT :8080/tasks/a4837ac1-c807-4edd-ba49-d4e37f295be7 \
  frequency_id:=null

# Change task frequency
http PUT :8080/tasks/a4837ac1-c807-4edd-ba49-d4e37f295be7 \
  frequency_id="f2345678-bc23-45de-67fa-234567890123"

# Update multiple fields including frequency
http PUT :8080/tasks/a4837ac1-c807-4edd-ba49-d4e37f295be7 \
  name="Updated task with new frequency" \
  completed:=false \
  priority:=2 \
  frequency_id="f1234567-ab12-34cd-56ef-123456789012"

# Update task tags (replaces all existing tags)
http PUT :8080/tasks/a4837ac1-c807-4edd-ba49-d4e37f295be7 \
  tag_ids:='["b5948bc2-d918-5fe5-ca5a-e8e48f406cf8", "c6059cd3-ea29-6gf6-db6b-f9f59g517dg9"]'

# Clear all tags from a task
http PUT :8080/tasks/a4837ac1-c807-4edd-ba49-d4e37f295be7 \
  tag_ids:='[]'

# Update task with both frequency and tags
http PUT :8080/tasks/a4837ac1-c807-4edd-ba49-d4e37f295be7 \
  name="Updated task" \
  frequency_id="f1234567-ab12-34cd-56ef-123456789012" \
  tag_ids:='["b5948bc2-d918-5fe5-ca5a-e8e48f406cf8"]'
```

**Example Requests:**

```bash
# Update completion status and priority
http PUT :8080/tasks/a4837ac1-c807-4edd-ba49-d4e37f295be7 \
  completed:=true \
  priority:=1

# Assign a frequency to an existing task
http PUT :8080/tasks/a4837ac1-c807-4edd-ba49-d4e37f295be7 \
  frequency_id="f1234567-ab12-34cd-56ef-123456789012"

# Remove frequency from a task (set to null)
http PUT :8080/tasks/a4837ac1-c807-4edd-ba49-d4e37f295be7 \
  frequency_id:=null

# Update multiple fields including frequency
http PUT :8080/tasks/a4837ac1-c807-4edd-ba49-d4e37f295be7 \
  name="Updated task name" \
  completed:=true \
  frequency_id="f2345678-bc23-45de-67fa-234567890123"
```

**Response (200 OK):**

```json
{
  "id": "a4837ac1-c807-4edd-ba49-d4e37f295be7",
  "name": "Updated task name",
  "date_created": "2025-09-05T21:17:04.323385Z",
  "date_modified": "2025-09-05T22:45:30.789012Z",
  "completed": true,
  "priority": 1,
  "frequency_id": "f2345678-bc23-45de-67fa-234567890123",
  "frequency": {
    "id": "f2345678-bc23-45de-67fa-234567890123",
    "name": "Weekly Monday",
    "reset": "0 23 * * 1"
  },
  "tags": [
    {
      "id": "b5948bc2-d918-5fe5-ca5a-e8e48f406cf8",
      "name": "work",
      "color": "#bfdbfe"
    }
  ]
}
```

### Delete Task

Delete a task by its ID.

**Endpoint:** `DELETE /tasks/{id}`

**Path Parameters:**

- `id` (string, required) - Task UUID

**Example Request:**

```bash
http DELETE :8080/tasks/a4837ac1-c807-4edd-ba49-d4e37f295be7
```

**Response (204 No Content):**

```
(empty response body)
```

**Error Response (404 Not Found):**

```json
{
  "error": "task not found"
}
```

---

## Tags

### Get All Tags

Retrieve all tags with optional name filtering.

**Endpoint:** `GET /tags`

**Query Parameters:**

- `name` (string, optional) - Filter by tag name (partial matching)

**Example Requests:**

```bash
# Get all tags
http GET :8080/tags

# Search for tags containing "work"
http GET :8080/tags name==work
```

**Response (200 OK):**

```json
[
  {
    "id": "b5948bc2-d918-5fe5-ca5a-e8e48f406cf8",
    "name": "work",
    "color": "#bfdbfe"
  },
  {
    "id": "c6059cd3-ea29-6gf6-db6b-f9f59g517dg9",
    "name": "personal",
    "color": "#bbf7d0"
  }
]
```

### Get Single Tag

Retrieve a specific tag by its ID.

**Endpoint:** `GET /tags/{id}`

**Path Parameters:**

- `id` (string, required) - Tag UUID

**Example Request:**

```bash
http GET :8080/tags/b5948bc2-d918-5fe5-ca5a-e8e48f406cf8
```

**Response (200 OK):**

```json
{
  "id": "b5948bc2-d918-5fe5-ca5a-e8e48f406cf8",
  "name": "work",
  "color": "#bfdbfe"
}
```

**Error Response (404 Not Found):**

```json
{
  "error": "tag not found"
}
```

### Create Tag

Create a new tag with an optional color.

**Endpoint:** `POST /tags`

**Request Body:**

```json
{
  "name": "urgent",
  "color": "#fca5a5"
}
```

**Example Request:**

```bash
http POST :8080/tags \
  name=urgent \
  color="#fca5a5"
```

**Response (201 Created):**

```json
{
  "id": "e8271ef5-gc4b-8ih8-fd8d-h1h71i739fi1",
  "name": "urgent",
  "color": "#fca5a5"
}
```

**Example Request (auto-generated color):**

```bash
http POST :8080/tags name=meeting
```

**Response (201 Created):**

```json
{
  "id": "f9382fg6-hd5c-9ji9-ge9e-i2i82j840gj2",
  "name": "meeting",
  "color": "#c7d2fe"
}
```

**Error Responses:**

```json
// 400 Bad Request - Missing name
{
  "error": "tag name is required"
}

// 400 Bad Request - Duplicate name
{
  "error": "tag name must be unique"
}
```

### Update Tag

Update an existing tag by its ID.

**Endpoint:** `PUT /tags/{id}`

**Path Parameters:**

- `id` (string, required) - Tag UUID

**Request Body:**

```json
{
  "name": "high-priority",
  "color": "#f87171"
}
```

**Example Request:**

```bash
http PUT :8080/tags/b5948bc2-d918-5fe5-ca5a-e8e48f406cf8 \
  name=high-priority \
  color="#f87171"
```

**Response (200 OK):**

```json
{
  "id": "b5948bc2-d918-5fe5-ca5a-e8e48f406cf8",
  "name": "high-priority",
  "color": "#f87171"
}
```

### Delete Tag

Delete a tag by its ID.

**Endpoint:** `DELETE /tags/{id}`

**Path Parameters:**

- `id` (string, required) - Tag UUID

**Example Request:**

```bash
http DELETE :8080/tags/b5948bc2-d918-5fe5-ca5a-e8e48f406cf8
```

**Response (204 No Content):**

```
(empty response body)
```

**Error Response (404 Not Found):**

```json
{
  "error": "tag not found"
}
```

---

## Frequencies

### Get All Frequencies

Retrieve all frequencies with optional name filtering.

**Endpoint:** `GET /frequencies`

**Query Parameters:**

- `name` (string, optional) - Filter by frequency name (partial matching)

**Example Requests:**

```bash
# Get all frequencies
http GET :8080/frequencies

# Search for frequencies containing "daily"
http GET :8080/frequencies name==daily
```

**Response (200 OK):**

```json
[
  {
    "id": "f1234567-ab12-34cd-56ef-123456789012",
    "name": "Daily Evening",
    "reset": "0 18 * * *"
  },
  {
    "id": "f2345678-bc23-45de-67fa-234567890123",
    "name": "Weekly Monday",
    "reset": "0 23 * * 1"
  }
]
```

### Get Single Frequency

Retrieve a specific frequency by its ID.

**Endpoint:** `GET /frequencies/{id}`

**Path Parameters:**

- `id` (string, required) - Frequency UUID

**Example Request:**

```bash
http GET :8080/frequencies/f1234567-ab12-34cd-56ef-123456789012
```

**Response (200 OK):**

```json
{
  "id": "f1234567-ab12-34cd-56ef-123456789012",
  "name": "Daily Evening",
  "reset": "0 18 * * *"
}
```

**Error Response (404 Not Found):**

```json
{
  "error": "frequency not found"
}
```

### Create Frequency

Create a new frequency with a cron expression for reset timing.

**Endpoint:** `POST /frequencies`

**Request Body:**

```json
{
  "name": "Daily Morning",
  "reset": "0 6 * * *"
}
```

**Example Requests:**

```bash
# Daily at 6 AM UTC
http POST :8080/frequencies \
  name="Daily Morning" \
  reset="0 6 * * *"

# Weekly on Mondays at 11 PM UTC
http POST :8080/frequencies \
  name="Weekly Monday" \
  reset="0 23 * * 1"

# Monthly on the 15th at midnight UTC
http POST :8080/frequencies \
  name="Monthly 15th" \
  reset="0 0 15 * *"
```

**Response (201 Created):**

```json
{
  "id": "f3456789-cd34-56ef-78ab-345678901234",
  "name": "Daily Morning",
  "reset": "0 6 * * *"
}
```

**Error Responses:**

```json
// 400 Bad Request - Missing name
{
  "error": "frequency name is required"
}

// 400 Bad Request - Missing reset
{
  "error": "frequency reset is required"
}

// 400 Bad Request - Duplicate name
{
  "error": "frequency name must be unique"
}
```

### Update Frequency

Update an existing frequency by its ID.

**Endpoint:** `PUT /frequencies/{id}`

**Path Parameters:**

- `id` (string, required) - Frequency UUID

**Request Body:**

```json
{
  "name": "Updated Frequency Name",
  "reset": "0 12 * * *"
}
```

**Example Request:**

```bash
http PUT :8080/frequencies/f1234567-ab12-34cd-56ef-123456789012 \
  name="Daily Noon" \
  reset="0 12 * * *"
```

**Response (200 OK):**

```json
{
  "id": "f1234567-ab12-34cd-56ef-123456789012",
  "name": "Daily Noon",
  "reset": "0 12 * * *"
}
```

### Delete Frequency

Delete a frequency by its ID.

**Endpoint:** `DELETE /frequencies/{id}`

**Path Parameters:**

- `id` (string, required) - Frequency UUID

**Example Request:**

```bash
http DELETE :8080/frequencies/f1234567-ab12-34cd-56ef-123456789012
```

**Response (204 No Content):**

```
(empty response body)
```

**Error Response (404 Not Found):**

```json
{
  "error": "frequency not found"
}
```

---

## Sample Data Population

### Populate Sample Data

Populate the database with sample tasks, tags, and frequencies for testing or development purposes. This endpoint clears all existing data and creates new sample data.

**Endpoint:** `POST /populateSampleData`

**Query Parameters:**

- `count` (integer, optional) - Number of sample tasks to create (default: 10)

**Example Requests:**

```bash
# Create 10 sample tasks (default)
http POST :8080/populateSampleData

# Create 25 sample tasks
http POST :8080/populateSampleData count==25
```

**Response (200 OK):**

```json
{
  "success": true,
  "message": "Successfully created 10 sample tasks",
  "count": 10
}
```

**Error Responses:**

```json
// 400 Bad Request - Invalid count
{
  "error": "Invalid count parameter"
}

// 405 Method Not Allowed
{
  "error": "Method not allowed"
}

// 500 Internal Server Error
{
  "error": "Failed to populate database: [error details]"
}
```

> **Warning:** This endpoint will **permanently delete** all existing tasks, tags, and frequencies before creating new sample data. Use with caution in production environments.

---

## Error Responses

All endpoints may return these common error responses:

### 400 Bad Request

```json
{
  "error": "Invalid JSON"
}
```

### 500 Internal Server Error

```json
{
  "error": "Database connection failed"
}
```

---

## Data Models

### Task

```json
{
  "id": "uuid",
  "name": "string",
  "date_created": "ISO 8601 timestamp",
  "date_modified": "ISO 8601 timestamp",
  "completed": "boolean",
  "priority": "integer (1-5, default: 3)",
  "frequency_id": "uuid (optional)",
  "frequency": "Frequency object (optional)",
  "tags": "array of Tag objects (optional)"
}
```

### Tag

```json
{
  "id": "uuid",
  "name": "string (unique)",
  "color": "string (hex color code)"
}
```

### Frequency

```json
{
  "id": "uuid",
  "name": "string (unique)",
  "reset": "string (cron expression)"
}
```

---

## Notes

- All timestamps are in ISO 8601 format (UTC)
- UUIDs are version 4 format
- Tag colors are automatically assigned from a predefined palette if not specified
- Priority ranges from 1 (highest) to 5 (lowest), defaults to 3
- Tasks are sorted by completion status first (incomplete tasks first), then by the specified sort field
- Tag filtering uses AND logic: when multiple tag IDs are specified, tasks must have ALL specified tags to be returned
- When updating task tags using `tag_ids`, the provided array completely replaces existing tag associations. To clear all tags, provide an empty array `[]`
- Tag IDs in the `tag_ids` parameter should be comma-separated with no spaces (e.g., `uuid1,uuid2,uuid3`)
- Tasks can optionally be associated with a frequency for automatic reset scheduling
- Frequency `reset` field uses standard cron expression format (minute hour day month day-of-week)
- Frequency names must be unique across the system
- Common cron expression examples:
  - `0 18 * * *` - Daily at 6:00 PM UTC
  - `0 23 * * 1` - Weekly on Mondays at 11:00 PM UTC
  - `0 0 15 * *` - Monthly on the 15th at midnight UTC
  - `0 9 * * 1-5` - Weekdays at 9:00 AM UTC
- When a task has a frequency, the `frequency` object will be included in GET responses if the relationship is preloaded
