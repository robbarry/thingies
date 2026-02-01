# thingies

CLI and REST API for Things 3 task management on macOS.

## Installation

```bash
# From source
git clone <repo> && cd thingies
make install
```

Binary installs to `/usr/local/bin/thingies`.

## Quick Start

```bash
# View tasks
thingies today                    # Today's tasks
thingies snapshot                 # Full hierarchy

# Create task
thingies tasks create "Buy groceries" --when today

# Complete task
thingies tasks complete <uuid>

# Start REST API
thingies serve -p 8484
```

## CLI Commands

### Views

```bash
thingies today                    # Today's tasks
thingies inbox                    # Inbox (no project, not scheduled)
thingies upcoming                 # Future scheduled
thingies someday                  # Deferred tasks
thingies anytime                  # Available but not scheduled
thingies logbook -n 50            # Completed tasks (default 50)
thingies snapshot                 # Full hierarchy as text
thingies snapshot --json          # Full hierarchy as JSON
thingies search <query>           # Search by title
thingies search <query> --in-notes
```

### Tasks

```bash
thingies tasks list
thingies tasks list --status all|completed|incomplete
thingies tasks list --today
thingies tasks list --area "Work"
thingies tasks list --project "Project Name"
thingies tasks list --tag "urgent"
thingies tasks show <uuid>
thingies tasks create "Title"
thingies tasks create "Title" --when today|tomorrow|evening|anytime|someday|YYYY-MM-DD
thingies tasks create "Title" --deadline YYYY-MM-DD --tags "a,b" --list "Project" --heading "Section"
thingies tasks update <uuid> --title "New" --notes "..." --when tomorrow
thingies tasks complete <uuid>
thingies tasks cancel <uuid>
thingies tasks delete <uuid>
thingies tasks delete <uuid> -f   # Skip confirmation
```

### Projects

```bash
thingies projects list
thingies projects list --include-completed
thingies projects show <uuid>
thingies projects show <uuid> --include-completed
thingies projects create "Title"
thingies projects create "Title" --area "Work" --todos "Task 1\nTask 2"
thingies projects update <uuid> --title "New" --notes "..."
thingies projects complete <uuid>
thingies projects delete <uuid>
```

### Areas

```bash
thingies areas list
thingies areas show <uuid>
thingies areas show <uuid> --include-completed
thingies areas create "Name"
thingies areas update <uuid> --title "New"
thingies areas delete <uuid>
```

### Tags

```bash
thingies tags list                # Shows usage counts
thingies tags create "Name"
thingies tags create "Name" --parent <uuid>
thingies tags update <uuid> --title "New"
thingies tags delete <uuid>
```

### Global Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--json` | `-j` | Output as JSON |
| `--db` | `-d` | Path to Things database (default: auto-detect) |
| `--no-color` | | Disable colors |
| `--verbose` | | Verbose output |

### Name Resolution

Commands accept either UUID or name for areas/projects:

```bash
thingies tasks list --area "Work"           # By name
thingies tasks list --area "ABC123-..."     # By UUID
```

If multiple items match a name, use UUID instead.

## REST API

### Start Server

```bash
thingies serve                    # 0.0.0.0:8484
thingies serve -p 3000            # Custom port
thingies serve --host 127.0.0.1   # Localhost only
```

### Endpoints

#### Health

```
GET /health
```

Response:
```json
{"status": "ok", "time": "2026-01-31T12:00:00Z"}
```

#### Views

```
GET /today
GET /inbox
GET /upcoming
GET /someday
GET /anytime
GET /logbook
GET /deadlines
GET /snapshot
```

All return `TaskJSON[]` except `/snapshot` which returns `{"snapshot": "..."}`.

#### Tasks

```
GET /tasks
GET /tasks?status=all|completed|incomplete
GET /tasks?today=true
GET /tasks?area=<name-or-uuid>
GET /tasks?project=<name-or-uuid>
GET /tasks?tag=<name>
GET /tasks?include-future=true
GET /tasks/search?q=<query>&in-notes=true&include-future=true
GET /tasks/{uuid}
POST /tasks
PATCH /tasks/{uuid}
DELETE /tasks/{uuid}
POST /tasks/{uuid}/complete
POST /tasks/{uuid}/cancel
POST /tasks/{uuid}/move-to-today
POST /tasks/{uuid}/move-to-someday
```

**Create task (POST /tasks):**
```json
{
  "title": "Task title",
  "notes": "Optional notes",
  "when": "today|tomorrow|evening|anytime|someday|YYYY-MM-DD",
  "deadline": "YYYY-MM-DD",
  "tags": "tag1,tag2",
  "list": "Project name",
  "heading": "Section name"
}
```

**Update task (PATCH /tasks/{uuid}):**
```json
{
  "title": "New title",
  "notes": "New notes",
  "when": "tomorrow",
  "deadline": "2026-02-15",
  "tags": "new,tags"
}
```

**Success response:**
```json
{"success": true, "message": "task created"}
```

**Error response:**
```json
{"success": false, "message": "error description"}
```

#### Projects

```
GET /projects
GET /projects?include-completed=true
GET /projects/{uuid}
GET /projects/{uuid}/tasks
GET /projects/{uuid}/tasks?include-completed=true
GET /projects/{uuid}/headings
```

#### Areas

```
GET /areas
GET /areas/{uuid}
GET /areas/{uuid}/tasks
GET /areas/{uuid}/tasks?include_completed=true
GET /areas/{uuid}/projects
GET /areas/{uuid}/projects?include_completed=true
```

#### Tags

```
GET /tags
GET /tags/{name}/tasks
```

Tag names are URL-encoded: `/tags/my%20tag/tasks`

#### Headings

```
PATCH /headings/{uuid}
DELETE /headings/{uuid}
```

### Response Types

**TaskJSON:**
```json
{
  "uuid": "ABC123...",
  "title": "Task title",
  "notes": "Optional notes",
  "status": "incomplete|completed|canceled",
  "type": "task|project|heading",
  "created": "2026-01-31T12:00:00Z",
  "modified": "2026-01-31T12:00:00Z",
  "scheduled": "2026-02-01T00:00:00Z",
  "due": "2026-02-15T00:00:00Z",
  "completed": "",
  "area_name": "Work",
  "project_uuid": "DEF456...",
  "project_name": "My Project",
  "heading_uuid": "",
  "heading_name": "",
  "tags": "tag1,tag2",
  "is_repeating": false,
  "checklist_items": [
    {"uuid": "...", "title": "Item", "completed": false, "index": 0}
  ]
}
```

## Common Patterns

### Get today's tasks as JSON

```bash
thingies today --json
```

```bash
curl http://localhost:8484/today
```

### Create task in project

```bash
thingies tasks create "Review PR" --list "Work" --when today
```

```bash
curl -X POST http://localhost:8484/tasks \
  -H "Content-Type: application/json" \
  -d '{"title": "Review PR", "list": "Work", "when": "today"}'
```

### Complete a task

```bash
thingies tasks complete ABC123-DEF456-...
```

```bash
curl -X POST http://localhost:8484/tasks/ABC123-DEF456-.../complete
```

### Get full snapshot

```bash
thingies snapshot --json | jq
```

```bash
curl http://localhost:8484/snapshot | jq -r '.snapshot'
```

## Gotchas

- **macOS only**: Requires Things 3 installed
- **Read operations**: Direct SQLite (fast, doesn't launch Things)
- **Write operations**: Via AppleScript (may briefly activate Things)
- **Create operations**: Via URL scheme (launches Things briefly)
- **Repeating tasks**: Filtered by default; use `--include-future` to show
- **API writes**: Return immediately; Things processes asynchronously
