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
thingies search <query>           # Search by title
thingies search <query> --in-notes           # Also search notes
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
thingies tasks show <uuid>                    # Full task details
thingies tasks create "Title"                 # Create task
thingies tasks create "Title" --when today --list "Project" --heading "Section"
thingies tasks update <uuid> --title "New" --notes "..." --when tomorrow
thingies tasks update <uuid> --when 2026-03-15  # Schedule to specific date (uses URL scheme)
thingies tasks complete <uuid>                # Mark complete
thingies tasks cancel <uuid>                  # Mark canceled
thingies tasks delete <uuid>                  # Move to trash
```

**Projects:**
```bash
thingies projects list                        # Active projects
thingies projects show <uuid>                 # Project details + tasks
thingies projects create "Title" --area "Work" --todos "Task 1\nTask 2"
thingies projects update <uuid> --title "New" --notes "..."
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

### Name Resolution
Most commands accept either a UUID or a name for areas/projects. Names are resolved to UUIDs automatically; if multiple items match a name, you'll be prompted to use the UUID instead.

## Architecture

### Key Packages
- `internal/cmd/` - Cobra commands organized by resource (`tasks/`, `projects/`, `areas/`, `tags/`) plus view commands (`today.go`, `inbox.go`, `upcoming.go`, `someday.go`, `anytime.go`, `logbook.go`)
- `internal/db/` - SQLite database layer: `db.go` (connection), `queries.go` (SQL), `scanner.go` (row scanning), `resolve.go` (name→UUID resolution)
- `internal/server/` - HTTP REST API server with handlers for tasks, projects, areas, tags, and views
- `internal/things/` - Things 3 integration: `urlscheme.go` (URL builder), `applescript.go` (osascript), `opener.go` (macOS open)
- `internal/models/` - Data models: Task, Project, Area, Tag, Heading, ChecklistItem
- `internal/output/` - Formatters: table (lipgloss) and JSON

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

The `serve` command starts an HTTP server (default port 8484). All responses are JSON.

**Views:**
- `GET /today`, `/inbox`, `/upcoming`, `/someday`, `/anytime`, `/logbook`, `/deadlines`
- `GET /snapshot` - Full hierarchical view as JSON

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

## Gotchas

- **Area visibility**: `visible = NULL` means visible (not `visible = 1`)
- **Today view logic**: Tasks appear in Today if:
  - `start=1` AND `startDate` is set (Anytime tasks moved to Today), OR
  - `start=2` AND `startDate <= today` (Someday tasks with past/current start date), OR
  - `deadline <= today` AND `deadlineSuppressionDate IS NULL` (overdue by deadline)
- **Repeating tasks**: `rt1_repeatingTemplate IS NOT NULL`; future instances filtered by default, use `--include-future`
- **Start field values**: `0` = Inbox, `1` = Anytime, `2` = Someday (scheduled or deferred)
- **No CGO**: Uses `modernc.org/sqlite` pure Go driver (no C compiler needed)
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
