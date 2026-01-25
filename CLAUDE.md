# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Thingies is a Go CLI tool that provides full CRUD access to the Things 3 task management app. It offers:

- **Reads**: Direct SQLite database access (fast, no app launch needed)
- **Creates**: Things URL scheme (`things:///add`, `things:///add-project`)
- **Updates/Deletes/Completes**: AppleScript via `osascript`

## Development Commands

### Build and Install
```bash
# Build the binary
make build              # or: go build -o bin/thingies ./cmd/thingies

# Install to /usr/local/bin
make install

# Clean build artifacts
make clean
```

### Running the CLI
```bash
./bin/thingies --help
```

### Commands Reference

**Shortcuts:**
```bash
thingies today                    # Today's tasks
thingies inbox                    # Inbox tasks (no area/project, not scheduled)
thingies search <query>           # Search by title
thingies search <query> --in-notes           # Also search notes
thingies search <query> --include-future     # Include future repeating instances
thingies snapshot                 # Hierarchical view (areas → projects → tasks)
thingies snapshot --json          # Full hierarchy as JSON
```

**Tasks:**
```bash
thingies tasks list                           # Incomplete tasks
thingies tasks list --status all              # All tasks (incomplete, completed, canceled)
thingies tasks list --status completed        # Only completed
thingies tasks list --today                   # Only Today view
thingies tasks list --area "Work"             # Filter by area
thingies tasks list --project "Bills"         # Filter by project
thingies tasks list --tag "urgent"            # Filter by tag
thingies tasks list --include-future          # Include future repeating instances
thingies tasks show <uuid>                    # Full task details
thingies tasks create "Title"                 # Create task (URL scheme)
thingies tasks create "Title" --when today    # Schedule for today
thingies tasks create "Title" --deadline 2026-02-01 --tags "work,urgent"
thingies tasks create "Title" --list "Project Name" --heading "Section"
thingies tasks update <uuid> --title "New"    # Update title (AppleScript)
thingies tasks update <uuid> --notes "..."    # Update notes
thingies tasks update <uuid> --when today     # Reschedule (today/tomorrow/evening/anytime/someday/YYYY-MM-DD)
thingies tasks update <uuid> --deadline 2026-02-01 --tags "a,b"
thingies tasks complete <uuid>                # Mark complete (AppleScript)
thingies tasks cancel <uuid>                  # Mark canceled (AppleScript)
thingies tasks delete <uuid>                  # Move to trash (AppleScript)
thingies tasks delete <uuid> -f               # Skip confirmation
```

**Projects:**
```bash
thingies projects list                        # Active projects
thingies projects list --include-completed    # Include completed projects
thingies projects show <uuid>                 # Project details + tasks
thingies projects show <uuid> --include-completed  # Include completed tasks
thingies projects create "Title"              # Create project (URL scheme)
thingies projects create "Title" --area "Work" --todos "Task 1\nTask 2"
thingies projects update <uuid> --title "New" # Update title (AppleScript)
thingies projects update <uuid> --notes "..." # Update notes (supports Markdown)
thingies projects update <uuid> --deadline 2026-02-01 --tags "a,b"
thingies projects complete <uuid>             # Mark complete (AppleScript)
thingies projects delete <uuid>               # Move to trash (AppleScript)
```

**Areas:**
```bash
thingies areas list                           # All visible areas
thingies areas show <uuid>                    # Area with projects and loose tasks
thingies areas show <uuid> --include-completed
```

**Tags:**
```bash
thingies tags list                            # All tags with usage counts
```

### Global Flags
```
--db, -d     Path to Things database (default: auto-detect)
--json, -j   Output as JSON
--no-color   Disable colors
--verbose    Verbose output
```

## Architecture and Structure

### Project Layout
```
thingies/
├── cmd/thingies/main.go          # Entry point
├── internal/
│   ├── cmd/                      # Cobra commands
│   │   ├── root.go               # Root + global flags
│   │   ├── shared/               # Shared helpers (flag access, formatter selection)
│   │   ├── tasks/                # Task CRUD commands
│   │   ├── projects/             # Project commands
│   │   ├── areas/                # Area commands
│   │   ├── tags/                 # Tag commands
│   │   ├── today.go              # thingies today
│   │   ├── inbox.go              # thingies inbox
│   │   ├── search.go             # thingies search
│   │   └── snapshot.go           # thingies snapshot
│   ├── db/
│   │   ├── db.go                 # Database connection
│   │   ├── queries.go            # SQL queries
│   │   └── scanner.go            # Row scanning utilities
│   ├── models/
│   │   ├── task.go               # Task model
│   │   ├── project.go            # Project model
│   │   ├── area.go               # Area model
│   │   ├── tag.go                # Tag model
│   │   └── common.go             # Status/Type enums
│   ├── things/
│   │   ├── urlscheme.go          # URL scheme builder
│   │   ├── applescript.go        # AppleScript executor
│   │   └── opener.go             # macOS open command
│   └── output/
│       ├── formatter.go          # Interface
│       ├── table.go              # lipgloss table output
│       └── json.go               # JSON output
├── go.mod
├── Makefile
└── CLAUDE.md
```

### Database Access
- Path: `~/Library/Group Containers/JLMPQHK86H.com.culturedcode.ThingsMac/ThingsData-*/Things Database.thingsdatabase/main.sqlite`
- Read-only connection: `file:{path}?mode=ro`
- Pure Go SQLite: `modernc.org/sqlite` (no CGO)

### Key Database Tables
- `TMTask` (type: 0=task, 1=project, 2=heading; status: 0=incomplete, 2=canceled, 3=completed)
- `TMArea` (visible: NULL means visible)
- `TMTag`
- `TMTaskTag` (many-to-many)

### Date Handling
Things 3 uses different epochs for different date fields:
- `creationDate`, `userModificationDate`, `stopDate`: Unix timestamps (1970-01-01 epoch)
- `startDate`, `deadline`: Things epoch (2021-11-11 00:00:00 UTC)

### URL Scheme (Creates/Updates)
```
things:///add?title=X&notes=Y&when=today&deadline=2026-01-26&tags=a,b&list=ProjectName
things:///add-project?title=X&area=AreaName&to-dos=task1%0Atask2
things:///update?id=UUID&auth-token=TOKEN&title=NewTitle
```
Note: Spaces must be encoded as `%20`, not `+`.

### AppleScript (Deletes/Completes/Updates)
```applescript
tell application "Things3"
    delete to do id "UUID"
    set status of to do id "UUID" to completed
    set notes of project id "UUID" to "New notes content"
    set notes of to do id "UUID" to "New notes content"
end tell
```
Note: AppleScript can update notes without needing an auth token (unlike URL scheme updates).

## Dependencies

- `github.com/spf13/cobra` - CLI framework
- `github.com/charmbracelet/lipgloss` - Terminal styling
- `modernc.org/sqlite` - Pure Go SQLite driver

## Known Issues / Gotchas

### Things date epoch
The `startDate` and `deadline` fields use a custom epoch (Nov 11, 2021 00:00:00 UTC), not Unix epoch. The constant `thingsDateEpoch = 1636588800` in `scanner.go` handles this. If dates look wrong (e.g., showing 1974 or 2005), the epoch calculation may need adjustment.

### Area visibility
Things stores `visible = NULL` for visible areas (not `visible = 1`). The query uses `WHERE a.visible IS NULL OR a.visible != 0`.

### Today view detection
Tasks appear in Today when `todayIndex > 0`. Tasks not in Today have `todayIndex = 0` or negative values.

### URL encoding
Things URL scheme requires `%20` for spaces, not `+`. Go's default `url.Values.Encode()` uses `+`, so we replace it in `urlscheme.go`.

### Repeating tasks
Repeating tasks have `rt1_repeatingTemplate IS NOT NULL`. By default, future instances are filtered out. Use `--include-future` to show them.

## Development Notes

- No CGO required (pure Go SQLite driver)
- Uses `uv` for any Python scripts (legacy code)
- The Python version in `thingies/` directory can be removed once Go version is complete
- Shell completions available via `thingies completion bash/zsh/fish`
