# thingies

Go CLI and REST API for full CRUD access to Things 3 (macOS task manager).

- Reads via direct SQLite database access (read-only, no app launch needed)
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

### Command Aliases

| Command | Aliases |
|---------|---------|
| `tasks` | `task`, `t` |
| `projects` | `project`, `p` |
| `areas` | `area`, `a` |
| `tags` | `tag` |
| `snapshot` | `all` |

### UUID Resolution

All commands that accept a UUID also accept:

- **Full UUID**: 22-character alphanumeric string (e.g., `6Cq1RzaLR7eFfjNL3Ymriw`)
- **Short UUID prefix**: Any unique prefix of a UUID (e.g., `6Cq1Rz`). If the prefix matches multiple items, the command returns an error listing the ambiguous matches.
- **Name** (areas and projects only): Exact name match. If multiple items share the same name, the command returns an error asking you to use the UUID instead.

Resolution order: full UUID check, then short prefix match, then name lookup (areas/projects).

### View Commands

Read-only shortcuts to common task filters. All support `--json` output.

```bash
thingies today                              # Tasks in Today view
thingies inbox                              # Unprocessed tasks (start=0)
thingies upcoming                           # Future scheduled tasks (start=2, startDate > today)
thingies someday                            # Deferred tasks (start=2, no startDate)
thingies anytime                            # Available tasks (start=1, or start=2 with startDate <= today)
thingies logbook                            # Completed tasks, most recent first (default limit: 50)
thingies logbook -n 100                     # Completed tasks with custom limit
thingies snapshot                           # Hierarchical view of all tasks
thingies search <query>                     # Search by title (incomplete tasks only by default)
thingies search <query> --in-notes          # Also search in task notes
thingies search <query> --include-future    # Include future instances of repeating tasks
```

The `snapshot` command outputs a styled hierarchical view (Today, Inbox, Upcoming, Someday, then Areas with Projects/Headings/Tasks). With `--json`, it returns a structured object:
```json
{
  "today": [TaskJSON, ...],
  "inbox": [TaskJSON, ...],
  "upcoming": [TaskJSON, ...],
  "someday": [TaskJSON, ...],
  "areas": [{"uuid": "...", "title": "...", "projects": [...], "tasks": [...]}, ...]
}
```

### Tasks

**List tasks:**
```bash
thingies tasks list                           # All incomplete tasks
thingies tasks list --status all              # All tasks (incomplete + completed + canceled)
thingies tasks list --status completed        # Only completed
thingies tasks list --status canceled         # Only canceled
thingies tasks list --today                   # Only Today view tasks
thingies tasks list --area "Work"             # Filter by area name (substring match, case-insensitive)
thingies tasks list --project "Bills"         # Filter by project name (substring match, case-insensitive)
thingies tasks list --tag "urgent"            # Filter by tag name (substring match, case-insensitive)
thingies tasks list --include-future          # Include future repeating task instances
```

**Show task details:**
```bash
thingies tasks show <uuid>                    # Full task details including checklist items
thingies tasks show 6Cq1Rz                    # Works with short UUID prefix
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
| `--list` | Project or area name to file task under |
| `--heading` | Heading within project (requires `--list`) |
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
| `--notes` | Replace notes entirely (not append) |
| `--when` | `today`, `tomorrow`, `evening`, `anytime`, `someday`, or `YYYY-MM-DD` |
| `--deadline` | Due date `YYYY-MM-DD` |
| `--tags` | Comma-separated tags (replaces all existing tags) |

**Complete/cancel/delete:**
```bash
thingies tasks complete <uuid>
thingies tasks cancel <uuid>
thingies tasks delete <uuid>                  # Moves to Things trash immediately (no confirmation prompt)
```

### Projects

```bash
thingies projects list                        # Active (incomplete) projects
thingies projects show <uuid>                 # Project details with task counts
thingies projects show "Project Name"         # By name (resolved to UUID)
thingies projects create "Title" --area "Work" --todos "Task 1\nTask 2"
thingies projects update <uuid> --title "New" --notes "Updated notes"
thingies projects update <uuid> --deadline 2026-03-01
thingies projects complete <uuid>
thingies projects delete <uuid>
```

Create flags: `--area`, `--notes`, `--when`, `--deadline`, `--tags`, `--todos` (newline-separated task titles in a single string).

Update flags: `--title`, `--notes`, `--deadline`, `--tags`.

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

### REST API Server

```bash
thingies serve                    # Start on 0.0.0.0:8484
thingies serve -p 3000            # Custom port
thingies serve --host 127.0.0.1   # Localhost only
```

The server handles graceful shutdown on SIGINT/SIGTERM with a 30-second timeout.

---

## REST API Reference

Default base URL: `http://localhost:8484`

All responses are `Content-Type: application/json`. CORS is enabled (all origins). The server accepts OPTIONS preflight requests.

### Health

```
GET /health
```

Response:
```json
{"status": "ok", "time": "2026-02-10T15:00:00Z"}
```

### Views

All view endpoints return `TaskJSON[]` except `/snapshot`.

```
GET /today
GET /inbox
GET /anytime
GET /upcoming
GET /someday
GET /logbook              ?limit=50          (default: 50)
GET /deadlines            ?days=7            (default: 7, API-only, no CLI equivalent)
GET /snapshot                                (returns {"snapshot": "..."} hierarchical text)
```

### Tasks

**List tasks:**
```
GET /tasks                ?status=incomplete   (incomplete|completed|canceled|all; default: incomplete)
                          &area=Work           (substring match, case-insensitive)
                          &project=Bills       (substring match, case-insensitive)
                          &tag=urgent          (substring match, case-insensitive)
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

The `{uuid}` path parameter also accepts short UUID prefixes.

**Create task:**
```
POST /tasks
Content-Type: application/json

{
  "title": "Task title",          // required
  "notes": "Details",             // optional
  "when": "today",                // optional: today, tomorrow, evening, someday, YYYY-MM-DD
  "deadline": "2026-03-15",       // optional: YYYY-MM-DD
  "tags": "work,urgent",          // optional: comma-separated
  "list": "Project Name",         // optional: project or area name
  "heading": "Section"            // optional: heading within project
}
```

Note: The request body parser uses `DisallowUnknownFields()`. Sending unrecognized fields returns a 400 error.

Success response:
```json
{"success": true, "message": "task created"}
```

**Update task:**
```
PATCH /tasks/{uuid}
Content-Type: application/json

{
  "title": "New title",           // optional
  "notes": "New notes",           // optional (replaces existing entirely)
  "when": "tomorrow",             // optional
  "deadline": "2026-03-15",       // optional
  "tags": "work"                  // optional (replaces all existing tags)
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
  "title": "Project title",       // required
  "notes": "Description",         // optional
  "when": "today",                // optional
  "deadline": "2026-06-01",       // optional
  "tags": "work",                 // optional
  "area": "Work",                 // optional: area name
  "todos": ["Task 1", "Task 2"]  // optional: initial tasks (string array)
}
```

### Areas

```
GET /areas
GET /areas/{uuid}
GET /areas/{uuid}/tasks     ?include_completed=true
GET /areas/{uuid}/projects  ?include_completed=true
```

Note: Area sub-resource endpoints use `include_completed` (underscore), not `include-completed` (hyphen).

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

PATCH response:
```json
{"status": "updated", "uuid": "...", "title": "New Name"}
```

DELETE response:
```json
{"status": "deleted", "uuid": "..."}
```

### Error Responses

Errors come in two formats depending on the endpoint:

From task write handlers (`POST /tasks`, `PATCH /tasks/{uuid}`, action endpoints):
```json
{"success": false, "message": "error description"}
```

From task read handlers (`GET /tasks`, `GET /tasks/{uuid}`, `GET /tasks/search`) and heading handlers:
```json
{"error": "error description"}
```

| HTTP Status | Meaning |
|-------------|---------|
| 400 | Missing required parameter, invalid request body, or unknown field in JSON |
| 404 | Task/project/area/heading not found, or ambiguous UUID prefix |
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
  "uuid": "6Cq1RzaLR7eFfjNL3Ymriw",
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
  "project_uuid": "7Xm2TpbMQ4gHhjOK4Znsjx",
  "project_name": "Project Title",
  "heading_uuid": "8Yn3UqcNR5hIikPL5Aotky",
  "heading_name": "Section Name",
  "tags": "tag1, tag2",
  "is_repeating": false,
  "checklist_items": [
    {"uuid": "9Zo4VrdOS6iJjlQM6Bpulz", "title": "Step 1", "completed": false, "index": 0}
  ]
}
```

Things UUIDs are 22-character base62 alphanumeric strings (not the standard 8-4-4-4-12 UUID format).

All date/time fields are RFC 3339 strings. Fields with empty/null values are omitted from JSON output (`omitempty`).

### Project JSON Schema

```json
{
  "uuid": "7Xm2TpbMQ4gHhjOK4Znsjx",
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
  "uuid": "8Yn3UqcNR5hIikPL5Aotky",
  "title": "string",
  "open_tasks": 10,
  "active_projects": 3
}
```

### Tag JSON Schema

```json
{
  "uuid": "9Zo4VrdOS6iJjlQM6Bpulz",
  "title": "string",
  "shortcut": "s",
  "task_count": 7
}
```

### Heading JSON Schema

```json
{
  "uuid": "1Ab5WseOT7jKkmRN7Cqvma",
  "title": "string",
  "index": 0
}
```

### ChecklistItem JSON Schema

```json
{
  "uuid": "2Bc6XtfPU8kLlnSO8Drwnb",
  "title": "string",
  "completed": false,
  "index": 0
}
```

### Status Values

| Value | Integer (DB) | Meaning |
|-------|-------------|---------|
| `incomplete` | 0 | Open/active |
| `canceled` | 2 | Canceled |
| `completed` | 3 | Done |

Note: Status integer 1 is not used.

### Start Field Values (internal DB column)

| Value | Meaning |
|-------|---------|
| 0 | Inbox |
| 1 | Anytime |
| 2 | Someday (may have a scheduled startDate, or be purely deferred) |

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
# note the UUID (or short prefix) from output
thingies tasks complete 6Cq1Rz
```

### Add task to a specific project and heading

```bash
thingies tasks create "Read Dune" --list "Want To" --heading "Read"
```

### Reschedule a task to a specific date

```bash
thingies tasks update <uuid> --when 2026-03-15
```

### Defer a task to someday

```bash
thingies tasks update <uuid> --when someday
```

### Create a project with initial tasks

```bash
thingies projects create "Q1 Review" --area "Work" --todos "Gather metrics\nDraft report\nSchedule meeting"
```

### Check upcoming deadlines via API

```bash
curl -s http://localhost:8484/deadlines?days=14 | jq '.[].title'
```

### REST API: create a task

```bash
curl -s -X POST http://localhost:8484/tasks \
  -H "Content-Type: application/json" \
  -d '{"title": "New task", "when": "today"}' | jq
```

### REST API: complete a task

```bash
curl -s -X POST http://localhost:8484/tasks/6Cq1Rz/complete | jq
```

### REST API: get today's tasks

```bash
curl -s http://localhost:8484/today | jq
```

### REST API: full snapshot

```bash
curl -s http://localhost:8484/snapshot | jq -r .snapshot
```

### REST API: search with notes

```bash
curl -s 'http://localhost:8484/tasks/search?q=meeting&in-notes=true' | jq
```

---

## Gotchas

**Area visibility in the database:** The query uses `WHERE a.visible IS NULL OR a.visible != 0`. In practice, `visible = NULL` means visible.

**Today view logic:** A task appears in Today if any of these are true:
- `start = 1` AND `startDate IS NOT NULL` (Anytime task moved to Today)
- `start = 2` AND `startDate IS NOT NULL` AND `startDate <= todayPackedDate` (Someday task whose scheduled date has arrived)
- `startDate IS NULL` AND `deadline <= todayPackedDate` AND `deadlineSuppressionDate IS NULL` (overdue by deadline, no start date set)

**Binary-packed dates:** The `startDate` and `deadline` columns in TMTask are NOT Unix timestamps. They use a packed format: `year << 16 | month << 12 | day << 7`. Example: 2026-01-31 = `(2026 << 16) | (1 << 12) | (31 << 7)` = 132775936. See `DateToPackedInt()` in `db.go`.

**Unix timestamp dates:** The `creationDate`, `userModificationDate`, and `stopDate` columns ARE standard Unix timestamps (seconds since epoch 1970).

**Repeating tasks:** Tasks with `rt1_repeatingTemplate IS NOT NULL` are instances of repeating tasks. Future instances are filtered out by default in list queries. Use `--include-future` (CLI) or `include-future=true` (API) to see them.

**Specific date scheduling:** AppleScript cannot set `activation date` (it is read-only). Scheduling a task to a specific YYYY-MM-DD date requires the Things URL scheme (`things:///update`) with an auth token automatically retrieved from `TMSettings.uriSchemeAuthenticationToken`. Named values (`today`, `tomorrow`, `anytime`, `someday`) use AppleScript's `move to list` instead.

**URL scheme encoding:** Things does not decode `+` as space. The URL builder replaces `+` with `%20` in all query parameters (see `urlscheme.go`).

**Notes updates replace entirely:** The `--notes` flag and the API `notes` field replace the full notes content. There is no append/prepend operation in the CLI or API (though the underlying URL scheme supports `prepend-notes` and `append-notes` parameters in `UpdateParams`).

**Tags replace on update:** Setting tags via update replaces all existing tags on the task. There is no additive tag operation in the CLI or API.

**Database is read-only:** The SQLite connection opens in `mode=ro`. All writes go through AppleScript or the Things URL scheme, never direct SQL.

**Delete has no confirmation:** `thingies tasks delete` (and project/area/tag delete) executes immediately via AppleScript with no confirmation prompt. The item is moved to Things' trash.

**Task create does not return UUID:** Creating tasks via the Things URL scheme (`things:///add`) does not return the UUID of the created task. The CLI prints the title, but the UUID must be found via search afterward.

**API error format inconsistency:** Task read endpoints (`GET /tasks`, `GET /tasks/{uuid}`, `GET /tasks/search`) return errors as `{"error": "..."}`. Task write endpoints and other handlers return `{"success": false, "message": "..."}`.

**No CGO required:** Uses `modernc.org/sqlite` pure Go driver. No C compiler needed to build.

**Snapshot API vs CLI:** The REST API's `GET /snapshot` returns `{"snapshot": "text..."}` (a flat text representation). The CLI's `thingies snapshot --json` returns a structured JSON object with `today`, `inbox`, `upcoming`, `someday`, and `areas` arrays.

---

## Architecture

```
cmd/thingies/main.go              # entry point
internal/cmd/                     # CLI commands (cobra)
  root.go                         # root command, global flags
  serve.go                        # HTTP server command
  today.go, inbox.go, ...         # view commands
  search.go                       # search command
  snapshot.go                     # snapshot command (alias: all)
  logbook.go                      # logbook command
  tasks/                          # tasks subcommands (list, show, create, update, complete, cancel, delete)
  projects/                       # projects subcommands (list, show, create, update, complete, delete)
  areas/                          # areas subcommands (list, show, create, update, delete)
  tags/                           # tags subcommands (list, create, update, delete)
  shared/shared.go                # shared utilities (GetDBPath, GetFormatter, IsJSON, IsNoColor)
internal/db/                      # SQLite database layer
  db.go                           # connection, DateToPackedInt(), TodayPackedDate()
  queries.go                      # all SQL queries, TaskFilter, Resolve*UUID functions
  scanner.go                      # row scanning, thingsDateToNullTime()
  resolve.go                      # name-to-UUID resolution (ResolveProjectID, ResolveAreaID)
internal/server/                  # HTTP REST API
  server.go                       # routes, middleware, CORS, snapshot builder, area/project/tag handlers
  handlers_tasks.go               # GET /tasks, GET /tasks/{uuid}, GET /tasks/search
  handlers_views.go               # GET /today, /inbox, /upcoming, /someday, /anytime, /logbook, /deadlines
  tasks.go                        # POST/PATCH/DELETE task handlers, POST /projects, request/response types
  headings.go                     # PATCH/DELETE heading handlers
internal/things/                  # Things 3 integration
  urlscheme.go                    # URL builders (AddParams, AddProjectParams, UpdateParams)
  applescript.go                  # AppleScript operations (update, complete, cancel, delete, move, create area/tag)
  opener.go                       # macOS `open` command wrapper
internal/models/                  # data models
  task.go                         # Task, TaskJSON, ToJSON()
  project.go                      # Project, ProjectJSON, ToJSON()
  area.go                         # Area
  tag.go                          # Tag, TagJSON, ToJSON()
  heading.go                      # Heading
  checklist.go                    # ChecklistItem
  common.go                       # TaskStatus, TaskType enums with String() and Icon()
internal/output/                  # formatters
  table.go                        # lipgloss table output
  json.go                         # JSON output
  formatter.go                    # Formatter interface
```

### Database Location

Auto-detected at: `~/Library/Group Containers/JLMPQHK86H.com.culturedcode.ThingsMac/ThingsData-*/Things Database.thingsdatabase/main.sqlite`

Override with `--db` flag or `-d` shorthand.

### Key Dependencies

| Package | Purpose |
|---------|---------|
| `modernc.org/sqlite` | Pure Go SQLite driver (no CGO) |
| `github.com/spf13/cobra` | CLI framework |
| `github.com/charmbracelet/lipgloss` | Terminal table styling |

### Key Database Tables

| Table | Purpose |
|-------|---------|
| `TMTask` | Tasks, projects, and headings (distinguished by `type`: 0=task, 1=project, 2=heading) |
| `TMArea` | Areas |
| `TMTag` | Tags |
| `TMTaskTag` | Many-to-many join table (tasks to tags) |
| `TMChecklistItem` | Checklist items within tasks |
| `TMSettings` | App settings including `uriSchemeAuthenticationToken` |
