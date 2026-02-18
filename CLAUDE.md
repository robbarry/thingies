# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Thingies is a Go CLI tool and REST API server that provides full CRUD access to the Things 3 task management app. It offers:

- **Reads**: Direct SQLite database access (fast, no app launch needed)
- **Creates**: Things URL scheme (`things:///add`, `things:///add-project`)
- **Updates/Deletes/Completes**: AppleScript via `osascript`
- **REST API**: HTTP server for programmatic access

## Development Commands

```bash
make build              # Build to bin/thingies
make install            # Install to /usr/local/bin
make test               # Run tests (go test ./...)
make fmt                # Format code (go fmt ./...)
make tidy               # go mod tidy
make run ARGS="today"   # Build and run with args
```

Run a single test:
```bash
go test ./internal/db -run TestFunctionName -v
```

### Commands Reference

**Views (shortcuts to common filters):**
```bash
thingies today                    # Today's tasks
thingies inbox                    # Inbox tasks (no area/project, not scheduled)
thingies upcoming                 # Future scheduled tasks
thingies someday                  # Tasks deferred to someday
thingies anytime                  # Available but not scheduled
thingies logbook -n 50            # Completed tasks (default 50)
thingies search <query>                       # Search by title
thingies search <query> --in-notes            # Also search notes
thingies search <query> --include-future      # Include future repeating tasks
thingies snapshot                 # Hierarchical view (areas → projects → tasks)
```

**Tasks:**
```bash
thingies tasks list                           # Incomplete tasks
thingies tasks list --status all              # All (incomplete, completed, canceled)
thingies tasks list --today                   # Only Today view
thingies tasks list --area "Work"             # Filter by area (name or UUID)
thingies tasks list --project "Bills"         # Filter by project (name or UUID)
thingies tasks list --tag "urgent"            # Filter by tag
thingies tasks list --include-future          # Include future repeating task instances
thingies tasks show <uuid>                    # Full task details
thingies tasks create "Title"                 # Create task
thingies tasks create "Title" --when today --list "Project" --heading "Section"
thingies tasks create "Title" --deadline 2026-02-15  # With due date
thingies tasks update <uuid> --title "New" --notes "..." --when tomorrow
thingies tasks update <uuid> --when 2026-03-15  # Schedule to specific date (uses URL scheme)
thingies tasks update <uuid> --deadline 2026-03-01  # Set due date
thingies tasks complete <uuid>                # Mark complete
thingies tasks cancel <uuid>                  # Mark canceled
thingies tasks delete <uuid>                  # Move to trash
```

**Projects:**
```bash
thingies projects list                        # Active projects
thingies projects show <uuid>                 # Project details + tasks
thingies projects create "Title" --area "Work" --todos "Task 1\nTask 2"
thingies projects create "Title" --deadline 2026-03-01  # With due date
thingies projects update <uuid> --title "New" --notes "..."
thingies projects update <uuid> --deadline 2026-03-01   # Set due date
thingies projects complete <uuid>
thingies projects delete <uuid>
```

**Areas:**
```bash
thingies areas list                           # All visible areas
thingies areas show <uuid>                    # Area with projects and loose tasks
thingies areas create "Name"                  # Create area
thingies areas update <uuid> --title "New"
thingies areas delete <uuid>
```

**Tags:**
```bash
thingies tags list                            # All tags with usage counts
thingies tags create "Name"                   # Create tag
thingies tags create "Name" --parent <uuid>   # Create nested tag
thingies tags update <uuid> --title "New"
thingies tags delete <uuid>
```

**REST API Server:**
```bash
thingies serve                    # Start on 0.0.0.0:8484
thingies serve -p 3000            # Custom port
thingies serve --host 127.0.0.1   # Localhost only
```

### Global Flags
```
--db, -d     Path to Things database (default: auto-detect)
--json, -j   Output as JSON
--no-color   Disable colors
--verbose    Verbose output
```

### `--when` Values
The `--when` flag accepts: `today`, `tomorrow`, `evening`, `anytime`, `someday`, or a specific date as `YYYY-MM-DD`. Specific dates require the URL scheme with an auth token (handled automatically).

### Command Aliases
```
tasks → task, t          projects → project, p
areas → area, a          tags → tag
snapshot → all
```

### Name and UUID Resolution
Most commands accept either a UUID, a short UUID prefix, or a name for areas/projects. Resolution order:
1. Full UUID (22 alphanumeric chars) -- verified directly
2. Short UUID prefix (any alphanumeric prefix) -- matched via `LIKE prefix%`; errors on ambiguous matches
3. Name lookup (for projects and areas only)

Short prefixes work in both CLI commands and REST API endpoints. The API resolves short UUIDs in path parameters (e.g., `GET /tasks/abc123` resolves to the full UUID).

Delete commands execute immediately without confirmation prompts.

## Architecture

### Key Packages
- `internal/cmd/` - Cobra commands organized by resource (`tasks/`, `projects/`, `areas/`, `tags/`) plus view commands (`today.go`, `inbox.go`, `upcoming.go`, `someday.go`, `anytime.go`, `logbook.go`). The `serve` command registers itself via `init()` in `serve.go`.
- `internal/db/` - SQLite database layer: `db.go` (connection), `queries.go` (SQL + UUID prefix resolution), `scanner.go` (row scanning), `resolve.go` (name/UUID resolution with fallback chain)
- `internal/server/` - HTTP REST API server using Go 1.22+ `ServeMux` routing patterns (`GET /tasks/{uuid}`). Handlers split across: `handlers_tasks.go` (list/get/search), `tasks.go` (create/update/delete + request types), `handlers_views.go` (today/inbox/etc.), `headings.go`
- `internal/things/` - Things 3 integration: `urlscheme.go` (URL builder), `applescript.go` (osascript), `opener.go` (macOS open)
- `internal/models/` - Data models: Task, Project, Area, Tag, Heading, ChecklistItem. Each model has a `ToJSON()` method producing a clean serializable struct.
- `internal/output/` - Formatters: table (lipgloss) and JSON
- `internal/cmd/shared/` - Flag propagation helpers (`GetDBPath`, `IsJSON`, `GetFormatter`) that walk the command tree to find persistent flags

### Database
- Path: `~/Library/Group Containers/JLMPQHK86H.com.culturedcode.ThingsMac/ThingsData-*/Things Database.thingsdatabase/main.sqlite`
- Read-only connection: `file:{path}?mode=ro`
- Pure Go SQLite: `modernc.org/sqlite` (no CGO)

**Key tables:**
- `TMTask` (type: 0=task, 1=project, 2=heading; status: 0=incomplete, 2=canceled, 3=completed)
- `TMArea` (visible: NULL means visible)
- `TMTag`, `TMTaskTag` (many-to-many)

**Date formats:**
- `creationDate`, `userModificationDate`, `stopDate`: Unix timestamps (seconds since 1970)
- `startDate`, `deadline`: Binary-packed dates (NOT timestamps)
  - Format: `year << 16 | month << 12 | day << 7`
  - Example: 2026-01-31 = `(2026 << 16) | (1 << 12) | (31 << 7)` = `132775936`
  - See `DateToPackedInt()` in `db.go` and `thingsDateToNullTime()` in `scanner.go`

### Things Integration

**URL Scheme (creates):**
```
things:///add?title=X&notes=Y&when=today&list=ProjectName
things:///add-project?title=X&area=AreaName&to-dos=task1%0Atask2
```
Note: Spaces must be `%20`, not `+` (see `urlscheme.go`).

**AppleScript (updates/deletes/completes):**
```applescript
tell application "Things3"
    set status of to do id "UUID" to completed
    set notes of to do id "UUID" to "New notes"
end tell
```
AppleScript can update notes without an auth token (unlike URL scheme).
Note: `activation date` is read-only in AppleScript, so specific date scheduling uses the URL scheme with an auth token from `TMSettings.uriSchemeAuthenticationToken`.

**URL Scheme (updates with specific dates):**
```
things:///update?id=UUID&auth-token=TOKEN&when=2026-03-15
```

### REST API Endpoints

The `serve` command starts an HTTP server (default port 8484). All responses are JSON. CORS is enabled (`*` origin) for local development.

**Views:**
- `GET /today`, `/inbox`, `/upcoming`, `/someday`, `/anytime`
- `GET /logbook` - Completed tasks (query param: `limit`, default 50)
- `GET /deadlines` - API-only (no CLI equivalent), returns tasks with upcoming deadlines (query param: `days`, default 7)
- `GET /snapshot` - Full hierarchical view as JSON text (sections: TODAY, ANYTIME, UPCOMING, SOMEDAY, INBOX)

**Tasks:**
- `GET /tasks` - List tasks (query params: `status`, `area`, `project`, `tag`, `today`, `include-future`)
- `GET /tasks/search?q=term` - Search (query params: `in-notes`, `include-future`)
- `GET /tasks/{uuid}`, `POST /tasks`, `PATCH /tasks/{uuid}`, `DELETE /tasks/{uuid}`
- `POST /tasks/{uuid}/complete`, `/cancel`, `/move-to-today`, `/move-to-someday`

**Projects:**
- `GET /projects` - List (query param: `include-completed`)
- `POST /projects`, `GET /projects/{uuid}`
- `GET /projects/{uuid}/tasks`, `GET /projects/{uuid}/headings`

**Areas:**
- `GET /areas`, `GET /areas/{uuid}`
- `GET /areas/{uuid}/tasks`, `GET /areas/{uuid}/projects` (query param: `include_completed`)

**Tags:**
- `GET /tags`, `GET /tags/{name}/tasks`

**Headings:**
- `PATCH /headings/{uuid}`, `DELETE /headings/{uuid}`

**Health:**
- `GET /health`

## Testing Patterns

Tests avoid side effects from Things 3 integration (AppleScript, URL scheme). Common approaches:
- **Server handler tests**: Use `httptest.NewRequest` and `httptest.NewRecorder` to test handlers directly without a running server or database
- **Convention enforcement**: AST-based tests parse source files to verify constraints (e.g., `delete.go` must not import `bufio` or reference `os.Stdin`, ensuring no interactive prompts)
- **Decode-only tests**: Test JSON decoding paths separately from handler side effects to avoid launching external apps

Tests are sparse -- mainly covering recent regressions and conventions. No mocking framework; tests either parse the source AST or test isolated decode logic.

## Gotchas

- **Area visibility**: `visible = NULL` means visible (not `visible = 1`)
- **Today view logic**: Tasks appear in Today if:
  - `start=1` AND `startDate` is set (Anytime tasks moved to Today), OR
  - `start=2` AND `startDate <= today` (Someday tasks with past/current start date), OR
  - `deadline <= today` AND `deadlineSuppressionDate IS NULL` (overdue by deadline)
- **Repeating tasks**: `rt1_repeatingTemplate IS NOT NULL`; future instances filtered by default, use `--include-future`
- **Start field values**: `0` = Inbox, `1` = Anytime, `2` = Someday (scheduled or deferred)
- **No CGO**: Uses `modernc.org/sqlite` pure Go driver (no C compiler needed)
- **Duplicate helper names**: The server package has two sets of `writeJSON`/`writeError` -- one set as package-level functions in `tasks.go` (used by create/update/delete handlers, returns `APIResponse` struct) and one set as methods on `Server` in `handlers_tasks.go` (used by list/get/search handlers, returns raw JSON). New handlers should use the `Server` method variants (`s.writeJSON`, `s.writeError`).
- **Shell completions**: `thingies completion bash/zsh/fish`
- **Go version**: Requires Go 1.21+ (current: 1.25.5 per go.mod)

<!-- SKILL-MAINTENANCE-START -->
## Skill Maintenance

This tool's skill has moved to the **wsj-ai-plugin-marketplace**:
`~/lrepos/wsj-ai-plugin-marketplace/plugins/things/skills/things/`

The old copy in wsj-ai-skills (`install/skills/things/`) is deprecated.

When making significant changes to this tool:
1. Update the skill in wsj-ai-plugin-marketplace to reflect the changes
2. Key files: `SKILL.md` (main skill) and `references/rest-api.md` (API docs)

The skill should document:
- Core commands and workflows
- Common usage patterns
- Non-obvious behaviors or gotchas
<!-- SKILL-MAINTENANCE-END -->
