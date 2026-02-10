# thingies

Go CLI and REST API for full CRUD access to Things 3 (macOS task manager).

- Reads via direct SQLite database access (read-only, no app launch)
- Creates via Things URL scheme (`things:///add`, `things:///add-project`)
- Updates/deletes/completes via AppleScript (`osascript`)
- REST API via built-in HTTP server (default port 8484)

Platform: macOS only. Requires Things 3 installed.

---

## Installation

```bash
cd ~/lrepos/thingies
make build                        # outputs bin/thingies
make install                      # copies to /usr/local/bin/thingies
```

Requires Go 1.21+ (project uses 1.25.5). No CGO needed (pure Go SQLite via `modernc.org/sqlite`).

---

## CLI Reference

Binary: `thingies`

### Global Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--db` | `-d` | auto-detect | Path to Things SQLite database |
| `--json` | `-j` | false | Output as JSON |
| `--no-color` | | false | Disable colored output |
| `--verbose` | `-v` | false | Verbose output |

### View Commands

These are read-only shortcuts to common task filters. All support `--json` output.

```bash
thingies today                    # Tasks in Today view
thingies inbox                    # Unprocessed tasks (start=0, no area/project)
thingies upcoming                 # Future scheduled tasks (start=2, startDate > today)
thingies someday                  # Deferred tasks (start=2, no startDate)
thingies anytime                  # Available tasks (start=1, or start=2 with startDate <= today)
thingies logbook -n 50            # Completed tasks, most recent first (default limit: 50)
thingies snapshot                 # Hierarchical text view: Today > Anytime > Upcoming > Someday > Inbox
thingies search <query>           # Search by title (incomplete tasks only by default)
thingies search <query> --in-notes  # Also search in task notes
```

### Tasks

**List tasks:**
```bash
thingies tasks list                           # All incomplete tasks
thingies tasks list --status all              # All tasks (incomplete + completed + canceled)
thingies tasks list --status completed        # Only completed
thingies tasks list --status canceled         # Only canceled
thingies tasks list --today                   # Only Today view tasks
thingies tasks list --area "Work"             # Filter by area (name or UUID)
thingies tasks list --project "Bills"         # Filter by project (name or UUID)
thingies tasks list --tag "urgent"            # Filter by tag name
thingies tasks list --include-future          # Include future repeating task instances
```

**Show task details:**
```bash
thingies tasks show <uuid>                    # Full task details including checklist items
```

**Create task:**
```bash
thingies tasks create "Title"
thingies tasks create "Title" --when today --list "Project Name" --heading "Section"
thingies tasks create "Title" --notes "Details" --deadline 2026-03-15 --tags "work,urgent"
```

Create flags:

| Flag | Description |
|------|-------------|
| `--notes` | Task notes (supports Markdown) |
| `--when` | Schedule: `today`, `tomorrow`, `evening`, `someday`, or `YYYY-MM-DD` |
| `--deadline` | Due date in `YYYY-MM-DD` format |
| `--tags` | Comma-separated tag names |
| `--list` | Project or area name to add task to |
| `--heading` | Heading within project |
| `--completed` | Create as already completed |
| `--canceled` | Create as already canceled |

**Update task:**
```bash
thingies tasks update <uuid> --title "New title"
thingies tasks update <uuid> --notes "Replacement notes"
thingies tasks update <uuid> --when today
thingies tasks update <uuid> --when tomorrow
thingies tasks update <uuid> --when anytime
thingies tasks update <uuid> --when someday
thingies tasks update <uuid> --when 2026-03-15    # Specific date (uses URL scheme + auth token)
thingies tasks update <uuid> --deadline 2026-03-15
thingies tasks update <uuid> --tags "work,urgent"  # Replaces all existing tags
```

Update flags:

| Flag | Description |
|------|-------------|
| `--title` | New title |
| `--notes` | Replace notes entirely |
| `--when` | `today`, `tomorrow`, `evening`, `anytime`, `someday`, or `YYYY-MM-DD` |
| `--deadline` | Due date `YYYY-MM-DD` |
| `--tags` | Comma-separated tags (replaces existing) |

**Complete/cancel/delete:**
```bash
thingies tasks complete <uuid>
thingies tasks cancel <uuid>
thingies tasks delete <uuid>                  # Moves to trash (prompts confirmation)
thingies tasks delete <uuid> -f               # Skip confirmation
```

### Projects

```bash
thingies projects list                        # Active (incomplete) projects
thingies projects show <uuid>                 # Project details with task counts
thingies projects show "Project Name"         # By name (resolved to UUID)
thingies projects create "Title" --area "Work" --todos "Task 1\nTask 2"
thingies projects update <uuid> --title "New" --notes "Updated notes"
thingies projects complete <uuid>
thingies projects delete <uuid>
```

Create flags: `--area`, `--notes`, `--when`, `--deadline`, `--tags`, `--todos` (newline-separated task titles).

### Areas

```bash
thingies areas list                           # All visible areas
thingies areas show <uuid>                    # Area with project and task counts
thingies areas show "Work"                    # By name
thingies areas create "Name"
thingies areas update <uuid> --title "New Name"
thingies areas delete <uuid>
```

### Tags

```bash
thingies tags list                            # All tags with task usage counts
thingies tags create "Name"
thingies tags create "Name" --parent <uuid>   # Create nested tag
thingies tags update <uuid> --title "New Name"
thingies tags delete <uuid>
```

### Name Resolution

Most commands accept either a UUID or a name for areas and projects. Names are resolved to UUIDs via exact-match lookup. If multiple items share the same name, the command returns an error asking you to use the UUID instead.

### REST API Server

```bash
thingies serve                    # Start on 0.0.0.0:8484
thingies serve -p 3000            # Custom port
thingies serve --host 127.0.0.1   # Localhost only
```

---

## REST API Reference

Default base URL: `http://localhost:8484`

All responses are `Content-Type: application/json`. CORS is enabled (all origins).

### Health

```
GET /health
```

Response:
```json
{"status": "ok", "time": "2026-02-10T15:00:00Z"}
```

### Views

All view endpoints return `TaskJSON[]`.

```
GET /today
GET /inbox
GET /anytime
GET /upcoming
GET /someday
GET /logbook              ?limit=50          (default: 50)
GET /deadlines            ?days=7            (default: 7)
GET /snapshot                                (returns {"snapshot": "..."} hierarchical text)
```

### Tasks

**List tasks:**
```
GET /tasks                ?status=incomplete   (incomplete|completed|canceled|all)
                          &area=Work
                          &project=Bills
                          &tag=urgent
                          &today=true
                          &include-future=true
```

**Search tasks:**
```
GET /tasks/search         ?q=keyword           (required)
                          &in-notes=true
                          &include-future=true
```

**Get single task:**
```
GET /tasks/{uuid}
```

**Create task:**
```
POST /tasks
Content-Type: application/json

{
  "title": "Task title",          # required
  "notes": "Details",             # optional
  "when": "today",                # optional: today, tomorrow, evening, someday, YYYY-MM-DD
  "deadline": "2026-03-15",       # optional: YYYY-MM-DD
  "tags": "work,urgent",          # optional: comma-separated
  "list": "Project Name",         # optional: project or area name
  "heading": "Section"            # optional: heading within project
}
```

Success response:
```json
{"success": true, "message": "task created"}
```

**Update task:**
```
PATCH /tasks/{uuid}
Content-Type: application/json

{
  "title": "New title",           # optional
  "notes": "New notes",           # optional (replaces existing)
  "when": "tomorrow",             # optional
  "deadline": "2026-03-15",       # optional
  "tags": "work"                  # optional (replaces existing)
}
```

**Task actions:**
```
POST /tasks/{uuid}/complete
POST /tasks/{uuid}/cancel
POST /tasks/{uuid}/move-to-today
POST /tasks/{uuid}/move-to-someday
DELETE /tasks/{uuid}
```

All action endpoints return:
```json
{"success": true, "message": "task completed"}
```

### Projects

```
GET /projects              ?include-completed=true
GET /projects/{uuid}
GET /projects/{uuid}/tasks ?include-completed=true
GET /projects/{uuid}/headings
```

**Create project:**
```
POST /projects
Content-Type: application/json

{
  "title": "Project title",       # required
  "notes": "Description",         # optional
  "when": "today",                # optional
  "deadline": "2026-06-01",       # optional
  "tags": "work",                 # optional
  "area": "Work",                 # optional: area name
  "todos": ["Task 1", "Task 2"]   # optional: initial tasks
}
```

### Areas

```
GET /areas
GET /areas/{uuid}
GET /areas/{uuid}/tasks     ?include_completed=true
GET /areas/{uuid}/projects  ?include_completed=true
```

### Tags

```
GET /tags
GET /tags/{name}/tasks
```

Tag names in URL paths are URL-decoded, so spaces and special characters work (e.g., `/tags/my%20tag/tasks`).

### Headings

```
PATCH /headings/{uuid}     {"title": "New Name"}
DELETE /headings/{uuid}
```

### Error Responses

All errors return JSON:
```json
{"success": false, "message": "error description"}
```

or (for some endpoints):
```json
{"error": "error description"}
```

| HTTP Status | Meaning |
|-------------|---------|
| 400 | Missing required parameter or invalid request body |
| 404 | Task/project/area not found |
| 500 | Database error or AppleScript failure |

---

## Data Model

### Hierarchy

```
Area
  Project
    Heading (section within project)
      Task
        ChecklistItem
```

Tasks can also exist directly under an Area (no project) or in the Inbox (no area or project).

### Task JSON Schema

```json
{
  "uuid": "XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX",
  "title": "string",
  "notes": "string",
  "status": "incomplete | completed | canceled",
  "type": "Task | Project | Heading",
  "created": "2026-01-15T10:30:00Z",
  "modified": "2026-01-15T10:30:00Z",
  "scheduled": "2026-02-01T00:00:00Z",
  "due": "2026-02-15T00:00:00Z",
  "completed": "2026-02-10T14:00:00Z",
  "area_name": "Work",
  "project_uuid": "XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX",
  "project_name": "Project Title",
  "heading_uuid": "XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX",
  "heading_name": "Section Name",
  "tags": "tag1, tag2",
  "is_repeating": false,
  "checklist_items": [
    {"uuid": "...", "title": "Step 1", "completed": false, "index": 0}
  ]
}
```

All date/time fields are RFC 3339 strings. Empty/null fields are omitted from JSON output.

### Project JSON Schema

```json
{
  "uuid": "XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX",
  "title": "string",
  "notes": "string",
  "status": "incomplete | completed | canceled",
  "area_name": "Work",
  "open_tasks": 5,
  "total_tasks": 12
}
```

### Area JSON Schema

```json
{
  "uuid": "XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX",
  "title": "string",
  "open_tasks": 10,
  "active_projects": 3
}
```

### Tag JSON Schema

```json
{
  "uuid": "XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX",
  "title": "string",
  "shortcut": "s",
  "task_count": 7
}
```

### Heading JSON Schema

```json
{
  "uuid": "XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX",
  "title": "string",
  "index": 0
}
```

### Status Values

| Value | Integer | Meaning |
|-------|---------|---------|
| incomplete | 0 | Open/active |
| canceled | 2 | Canceled |
| completed | 3 | Done |

### Start Field Values (internal)

| Value | Meaning |
|-------|---------|
| 0 | Inbox |
| 1 | Anytime |
| 2 | Someday (scheduled or deferred) |

---

## Common Workflows

### Morning review

```bash
thingies snapshot        # see everything at a glance
thingies inbox           # triage unprocessed items
thingies today           # focus on today
```

### Quick add to today

```bash
thingies tasks create "Fix bug in parser" --when today --tags "work"
```

### Find and complete a task

```bash
thingies search "dentist"
# note the UUID from output
thingies tasks complete <uuid>
```

### Add task to a specific project and heading

```bash
thingies tasks create "Read Dune" --list "Want To" --heading "Read"
```

### Reschedule a task to a specific date

```bash
thingies tasks update <uuid> --when 2026-03-15
```

### Create a project with initial tasks

```bash
thingies projects create "Q1 Review" --area "Work" --todos "Gather metrics\nDraft report\nSchedule meeting"
```

### REST API: create a task

```bash
curl -s -X POST http://localhost:8484/tasks \
  -H "Content-Type: application/json" \
  -d '{"title": "New task", "when": "today"}' | jq
```

### REST API: complete a task

```bash
curl -s -X POST http://localhost:8484/tasks/<uuid>/complete | jq
```

### REST API: get today's tasks

```bash
curl -s http://localhost:8484/today | jq
```

### REST API: full snapshot

```bash
curl -s http://localhost:8484/snapshot | jq -r .snapshot
```

---

## Gotchas

**Area visibility in the database:** `visible = NULL` means the area is visible. Do not filter on `visible = 1`.

**Today view logic:** A task appears in Today if any of these are true:
- `start = 1` AND `startDate IS NOT NULL` (Anytime task moved to Today)
- `start = 2` AND `startDate <= todayPackedDate` (Someday task whose scheduled date has arrived)
- `deadline <= todayPackedDate` AND `deadlineSuppressionDate IS NULL` (overdue by deadline)

**Binary-packed dates:** The `startDate` and `deadline` columns in TMTask are NOT Unix timestamps. They use a packed format: `year << 16 | month << 12 | day << 7`. Example: 2026-01-31 = 132775936. Use `DateToPackedInt()` in `db.go` to compute these.

**Unix timestamp dates:** The `creationDate`, `userModificationDate`, and `stopDate` columns ARE standard Unix timestamps (seconds since 1970).

**Repeating tasks:** Tasks with `rt1_repeatingTemplate IS NOT NULL` are instances of repeating tasks. Future instances are filtered out by default in list queries. Use `--include-future` to see them.

**Specific date scheduling:** AppleScript cannot set `activation date` (it is read-only). Scheduling a task to a specific date (YYYY-MM-DD) requires the Things URL scheme (`things:///update`) with an auth token from `TMSettings.uriSchemeAuthenticationToken`.

**URL scheme encoding:** Things does not decode `+` as space. The URL builder replaces `+` with `%20` in all query parameters.

**Notes updates replace entirely:** The `--notes` flag and the API `notes` field replace the full notes content. There is no append/prepend in the CLI (though the URL scheme supports `prepend-notes` and `append-notes`).

**Tags replace on update:** Setting tags via update replaces all existing tags on the task. There is no additive tag operation in the CLI.

**Database is read-only:** The SQLite connection opens in `mode=ro`. All writes go through AppleScript or the URL scheme, never direct SQL.

**No CGO required:** Uses `modernc.org/sqlite` pure Go driver. No C compiler needed to build.

---

## Architecture

```
cmd/thingies/main.go              # entry point
internal/cmd/                     # CLI commands (cobra)
  root.go                         # root command, global flags
  serve.go                        # HTTP server command
  today.go, inbox.go, ...         # view commands
  tasks/                          # tasks subcommands (list, show, create, update, complete, cancel, delete)
  projects/                       # projects subcommands
  areas/                          # areas subcommands
  tags/                           # tags subcommands
  shared/shared.go                # shared utilities (GetDBPath)
internal/db/                      # SQLite database layer
  db.go                           # connection, DateToPackedInt()
  queries.go                      # all SQL queries, TaskFilter
  scanner.go                      # row scanning, thingsDateToNullTime()
  resolve.go                      # name-to-UUID resolution
internal/server/                  # HTTP REST API
  server.go                       # routes, middleware, CORS, snapshot
  handlers_tasks.go               # GET /tasks, GET /tasks/{uuid}, GET /tasks/search
  handlers_views.go               # GET /today, /inbox, /upcoming, etc.
  tasks.go                        # POST/PATCH/DELETE task handlers, request/response types
  headings.go                     # PATCH/DELETE heading handlers
internal/things/                  # Things 3 integration
  urlscheme.go                    # URL builders (AddParams, AddProjectParams, UpdateParams)
  applescript.go                  # AppleScript operations (update, complete, cancel, delete, move)
  opener.go                       # macOS `open` command wrapper
internal/models/                  # data models
  task.go                         # Task, TaskJSON
  project.go                      # Project, ProjectJSON
  area.go                         # Area
  tag.go                          # Tag, TagJSON
  heading.go                      # Heading
  checklist.go                    # ChecklistItem
  common.go                       # TaskStatus, TaskType enums
internal/output/                  # formatters
  table.go                        # lipgloss table output
  json.go                         # JSON output
  formatter.go                    # output interface
```

### Database Location

Auto-detected at: `~/Library/Group Containers/JLMPQHK86H.com.culturedcode.ThingsMac/ThingsData-*/Things Database.thingsdatabase/main.sqlite`

### Key Dependencies

| Package | Purpose |
|---------|---------|
| `modernc.org/sqlite` | Pure Go SQLite driver (no CGO) |
| `github.com/spf13/cobra` | CLI framework |
| `github.com/charmbracelet/lipgloss` | Terminal table styling |
