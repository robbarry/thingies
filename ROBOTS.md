# thingies

Go CLI and REST API for Things 3 task management on macOS.

## Requirements

- macOS with Things 3 installed
- Go 1.21+ (for building from source)
- Things 3 app must be running for write operations

## Installation

```bash
git clone <repo> && cd thingies
make build      # builds to bin/thingies
make install    # copies to /usr/local/bin/thingies
```

## Architecture

| Operation | Method | Notes |
|-----------|--------|-------|
| Read | SQLite | Direct database access, fast, no app launch |
| Create | URL scheme | `things:///add`, launches Things briefly |
| Update | AppleScript | Via osascript, may activate Things |
| Delete | AppleScript | Via osascript, may activate Things |
| Complete | AppleScript | Via osascript, may activate Things |

Database path: `~/Library/Group Containers/JLMPQHK86H.com.culturedcode.ThingsMac/ThingsData-*/Things Database.thingsdatabase/main.sqlite`

## CLI Reference

### Global Flags

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--json` | `-j` | bool | false | Output as JSON |
| `--db` | `-d` | string | auto-detect | Path to Things database |
| `--no-color` | | bool | false | Disable colored output |
| `--verbose` | `-v` | bool | false | Verbose output |

### View Commands

```bash
thingies today                    # Today's tasks
thingies inbox                    # Inbox (no project, not scheduled)
thingies upcoming                 # Future scheduled tasks
thingies someday                  # Deferred to someday
thingies anytime                  # Available but not scheduled
thingies logbook                  # Completed tasks (default 50)
thingies logbook -n 100           # Completed tasks (custom limit)
thingies snapshot                 # Hierarchical text view
thingies snapshot --json          # Hierarchical JSON view
thingies search <query>           # Search by title
thingies search <query> --in-notes  # Search title and notes
```

### Task Commands

```bash
# List
thingies tasks list
thingies tasks list --status all              # all|completed|incomplete
thingies tasks list --today
thingies tasks list --area "Work"             # filter by area (name or UUID)
thingies tasks list --project "Project Name"  # filter by project (name or UUID)
thingies tasks list --tag "urgent"            # filter by tag

# Show
thingies tasks show <uuid>

# Create
thingies tasks create "Title"
thingies tasks create "Title" --when today
thingies tasks create "Title" --when tomorrow
thingies tasks create "Title" --when evening
thingies tasks create "Title" --when someday
thingies tasks create "Title" --when 2026-02-15
thingies tasks create "Title" --deadline 2026-02-15
thingies tasks create "Title" --tags "a,b,c"
thingies tasks create "Title" --list "Project Name"
thingies tasks create "Title" --list "Project" --heading "Section"

# Update
thingies tasks update <uuid> --title "New Title"
thingies tasks update <uuid> --notes "New notes"
thingies tasks update <uuid> --when tomorrow
thingies tasks update <uuid> --deadline 2026-02-15

# State changes
thingies tasks complete <uuid>
thingies tasks cancel <uuid>
thingies tasks delete <uuid>
thingies tasks delete <uuid> -f   # skip confirmation
```

### Project Commands

```bash
thingies projects list
thingies projects list --include-completed
thingies projects show <uuid>
thingies projects show <uuid> --include-completed
thingies projects create "Title"
thingies projects create "Title" --area "Work"
thingies projects create "Title" --todos "Task 1\nTask 2"
thingies projects update <uuid> --title "New Title"
thingies projects update <uuid> --notes "New notes"
thingies projects complete <uuid>
thingies projects delete <uuid>
```

### Area Commands

```bash
thingies areas list
thingies areas show <uuid>
thingies areas show <uuid> --include-completed
thingies areas create "Name"
thingies areas update <uuid> --title "New Name"
thingies areas delete <uuid>
```

### Tag Commands

```bash
thingies tags list                          # shows usage counts
thingies tags create "Name"
thingies tags create "Name" --parent <uuid>  # nested tag
thingies tags update <uuid> --title "New"
thingies tags delete <uuid>
```

### Server Command

```bash
thingies serve                    # 0.0.0.0:8484
thingies serve -p 3000            # custom port
thingies serve --host 127.0.0.1   # localhost only
```

## REST API Reference

Default: `http://localhost:8484`

### Health Check

```
GET /health
```

Response:
```json
{"status": "ok", "time": "2026-01-31T12:00:00Z"}
```

### View Endpoints

| Endpoint | Description |
|----------|-------------|
| `GET /today` | Today's tasks |
| `GET /inbox` | Inbox tasks |
| `GET /upcoming` | Future scheduled tasks |
| `GET /someday` | Someday tasks |
| `GET /anytime` | Anytime tasks |
| `GET /logbook?limit=50` | Completed tasks |
| `GET /deadlines?days=7` | Tasks with deadlines |
| `GET /snapshot` | Hierarchical text snapshot |

All view endpoints return `TaskJSON[]` except `/snapshot` which returns `{"snapshot": "..."}`.

### Task Endpoints

```
GET    /tasks
GET    /tasks?status=all|completed|incomplete
GET    /tasks?today=true
GET    /tasks?area=<name-or-uuid>
GET    /tasks?project=<name-or-uuid>
GET    /tasks?tag=<name>
GET    /tasks?include-future=true
GET    /tasks/search?q=<query>&in-notes=true
GET    /tasks/{uuid}
POST   /tasks
PATCH  /tasks/{uuid}
DELETE /tasks/{uuid}
POST   /tasks/{uuid}/complete
POST   /tasks/{uuid}/cancel
POST   /tasks/{uuid}/move-to-today
POST   /tasks/{uuid}/move-to-someday
```

**POST /tasks** (create):
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

**PATCH /tasks/{uuid}** (update):
```json
{
  "title": "New title",
  "notes": "New notes",
  "when": "tomorrow",
  "deadline": "2026-02-15",
  "tags": "new,tags"
}
```

### Project Endpoints

```
GET  /projects
GET  /projects?include-completed=true
GET  /projects/{uuid}
GET  /projects/{uuid}/tasks
GET  /projects/{uuid}/tasks?include-completed=true
GET  /projects/{uuid}/headings
POST /projects
```

**POST /projects**:
```json
{
  "title": "Project title",
  "notes": "Optional notes",
  "when": "today|tomorrow|evening|anytime|someday|YYYY-MM-DD",
  "deadline": "YYYY-MM-DD",
  "tags": "tag1,tag2",
  "area": "Area name",
  "todos": ["Task 1", "Task 2"]
}
```

### Area Endpoints

```
GET /areas
GET /areas/{uuid}
GET /areas/{uuid}/tasks
GET /areas/{uuid}/tasks?include_completed=true
GET /areas/{uuid}/projects
GET /areas/{uuid}/projects?include_completed=true
```

### Tag Endpoints

```
GET /tags
GET /tags/{name}/tasks
```

Tag names are URL-encoded: `/tags/my%20tag/tasks`

### Heading Endpoints

```
PATCH  /headings/{uuid}
DELETE /headings/{uuid}
```

### Response Types

**Success response:**
```json
{"success": true, "message": "task created"}
```

**Error response:**
```json
{"success": false, "message": "error description"}
```

Or for read errors:
```json
{"error": "error description"}
```

**TaskJSON:**
```json
{
  "uuid": "ABC123-DEF456-...",
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
  "project_uuid": "DEF456-...",
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

CLI:
```bash
thingies today --json
```

API:
```bash
curl http://localhost:8484/today
```

### Create task in project with deadline

CLI:
```bash
thingies tasks create "Review PR" --list "Work" --when today --deadline 2026-02-01
```

API:
```bash
curl -X POST http://localhost:8484/tasks \
  -H "Content-Type: application/json" \
  -d '{"title": "Review PR", "list": "Work", "when": "today", "deadline": "2026-02-01"}'
```

### Complete a task

CLI:
```bash
thingies tasks complete ABC123-DEF456-...
```

API:
```bash
curl -X POST http://localhost:8484/tasks/ABC123-DEF456-.../complete
```

### Search tasks

CLI:
```bash
thingies search "keyword" --in-notes
```

API:
```bash
curl "http://localhost:8484/tasks/search?q=keyword&in-notes=true"
```

### Get project with all tasks

CLI:
```bash
thingies projects show <uuid> --include-completed --json
```

API:
```bash
curl "http://localhost:8484/projects/<uuid>/tasks?include-completed=true"
```

### Get full snapshot

CLI:
```bash
thingies snapshot --json | jq
```

API:
```bash
curl http://localhost:8484/snapshot | jq -r '.snapshot'
```

## Gotchas

| Issue | Cause | Solution |
|-------|-------|----------|
| Write operations fail | Things 3 not running | Open Things 3 app |
| Repeating tasks not shown | Filtered by default | Use `--include-future` flag or `include-future=true` query param |
| Name ambiguous error | Multiple items match name | Use UUID instead |
| Create returns immediately | URL scheme is async | Task may take a moment to appear |
| Update activates Things | AppleScript switches focus | Normal behavior |

## Error Codes

| HTTP Status | Meaning |
|-------------|---------|
| 200 | Success |
| 400 | Bad request (missing required field, invalid UUID) |
| 404 | Resource not found |
| 500 | Server error (database error, Things integration failure) |
