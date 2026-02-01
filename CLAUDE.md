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
make test               # Run tests
make fmt                # Format code
make tidy               # go mod tidy
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

**Date epochs:**
- `creationDate`, `userModificationDate`, `stopDate`: Unix epoch (1970)
- `startDate`, `deadline`: Things epoch (2021-11-11 00:00:00 UTC) - constant `thingsDateEpoch = 1636588800` in `scanner.go`

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

### REST API Endpoints

The `serve` command starts an HTTP server (default port 8484). Key endpoints:
- `GET /health` - Health check
- `GET /today`, `/inbox`, `/upcoming`, `/someday`, `/anytime`, `/logbook`, `/deadlines` - Views
- `GET /tasks`, `POST /tasks`, `GET /tasks/{uuid}`, `PATCH /tasks/{uuid}`, `DELETE /tasks/{uuid}`
- `POST /tasks/{uuid}/complete`, `/cancel`, `/move-to-today`, `/move-to-someday`
- `GET /projects`, `GET /projects/{uuid}`, `GET /projects/{uuid}/tasks`, `GET /projects/{uuid}/headings`
- `GET /areas`, `GET /areas/{uuid}`, `GET /areas/{uuid}/tasks`, `GET /areas/{uuid}/projects`
- `GET /tags`, `GET /tags/{name}/tasks`
- `GET /snapshot` - Full hierarchical view as JSON

## Gotchas

- **Area visibility**: `visible = NULL` means visible (not `visible = 1`)
- **Today view**: Tasks in Today have `todayIndex > 0`
- **Repeating tasks**: `rt1_repeatingTemplate IS NOT NULL`; filtered by default, use `--include-future`
- **No CGO**: Uses `modernc.org/sqlite` pure Go driver
- **Shell completions**: `thingies completion bash/zsh/fish`
