# Thingies

A Go CLI and REST API for [Things 3](https://culturedcode.com/things/) with full CRUD access.

Reads are fast -- they go straight to the Things 3 SQLite database without launching the app. Creates use the Things URL scheme. Updates, deletes, and completions go through AppleScript. The REST API wraps all of this behind a JSON HTTP server.

## Requirements

- macOS with Things 3 installed
- Go 1.21+ (for building from source)

## Install

```bash
make install    # builds and copies to /usr/local/bin
```

Or build without installing:

```bash
make build      # outputs to bin/thingies
```

## CLI Usage

### Views

```bash
thingies today              # Today's tasks
thingies inbox              # Inbox (no area/project, not scheduled)
thingies upcoming           # Future scheduled tasks
thingies someday            # Tasks deferred to someday
thingies anytime            # Available but not scheduled
thingies logbook -n 50      # Completed tasks (default 50)
thingies snapshot           # Hierarchical view (areas -> projects -> tasks)
```

### Search

```bash
thingies search "keyword"
thingies search "keyword" --in-notes        # Also search notes
thingies search "keyword" --include-future   # Include future repeating task instances
```

### Tasks

```bash
thingies tasks list                                    # All incomplete tasks
thingies tasks list --status all                       # Include completed/canceled
thingies tasks list --today                            # Only Today view
thingies tasks list --area "Work"                      # Filter by area
thingies tasks list --project "Bills" --tag "urgent"   # Filter by project and tag
thingies tasks list --include-future                   # Include future repeating instances

thingies tasks show <uuid>                             # Full task details
thingies tasks create "New task"                       # Create task
thingies tasks create "New task" --when today --list "Project" --heading "Section"
thingies tasks create "New task" --deadline 2026-02-15 # With due date

thingies tasks update <uuid> --title "New" --notes "Updated"
thingies tasks update <uuid> --when tomorrow           # Schedule for tomorrow
thingies tasks update <uuid> --when 2026-03-15         # Schedule to specific date
thingies tasks update <uuid> --deadline 2026-02-15     # Set due date
thingies tasks complete <uuid>
thingies tasks cancel <uuid>
thingies tasks delete <uuid>
```

The `--when` flag accepts: `today`, `tomorrow`, `evening`, `anytime`, `someday`, or a date as `YYYY-MM-DD`.

### Projects

```bash
thingies projects list
thingies projects show <uuid>
thingies projects create "New project" --area "Work" --todos "Task 1\nTask 2"
thingies projects create "New project" --deadline 2026-03-01
thingies projects update <uuid> --notes "# Markdown supported"
thingies projects update <uuid> --deadline 2026-03-01
thingies projects complete <uuid>
thingies projects delete <uuid>
```

### Areas

```bash
thingies areas list
thingies areas show <uuid>
thingies areas create "Name"
thingies areas update <uuid> --title "New name"
thingies areas delete <uuid>
```

### Tags

```bash
thingies tags list
thingies tags create "Name"
thingies tags create "Nested" --parent <uuid>
thingies tags update <uuid> --title "New name"
thingies tags delete <uuid>
```

### Global Flags

```
--db, -d       Path to Things database (default: auto-detect)
--json, -j     Output as JSON
--no-color     Disable colors
--verbose, -v  Verbose output
```

### Command Aliases

```
tasks    -> task, t
projects -> project, p
areas    -> area, a
tags     -> tag
snapshot -> all
```

### Name and UUID Resolution

Most commands accept either a full UUID or a name for areas and projects. Names are resolved to UUIDs automatically. If multiple items share the same name, you'll be prompted to use the UUID instead.

Short UUID prefixes also work -- you can pass the first few characters of a UUID and thingies will resolve it as long as the prefix is unambiguous.

## REST API

```bash
thingies serve                   # Start on 0.0.0.0:8484
thingies serve -p 3000           # Custom port
thingies serve --host 127.0.0.1  # Localhost only
```

All responses are JSON. CORS is enabled for all origins.

### Endpoints

**Views:**
- `GET /today` - Today's tasks
- `GET /inbox` - Inbox tasks
- `GET /upcoming` - Upcoming scheduled tasks
- `GET /someday` - Someday tasks
- `GET /anytime` - Anytime tasks
- `GET /logbook` - Completed tasks (query: `limit`, default 50)
- `GET /deadlines` - Tasks with upcoming deadlines (query: `days`, default 7)
- `GET /snapshot` - Full hierarchical view

**Tasks:**
- `GET /tasks` - List tasks (query: `status`, `area`, `project`, `tag`, `today`, `include-future`)
- `GET /tasks/search?q=query` - Search tasks (query: `in-notes`, `include-future`)
- `GET /tasks/{uuid}` - Get task
- `POST /tasks` - Create task (body: `title`, `notes`, `when`, `deadline`, `tags`, `list`, `heading`)
- `PATCH /tasks/{uuid}` - Update task (body: `title`, `notes`, `when`, `deadline`, `tags`)
- `DELETE /tasks/{uuid}` - Delete task
- `POST /tasks/{uuid}/complete` - Mark complete
- `POST /tasks/{uuid}/cancel` - Mark canceled
- `POST /tasks/{uuid}/move-to-today` - Move to Today
- `POST /tasks/{uuid}/move-to-someday` - Move to Someday

**Projects:**
- `GET /projects` - List projects (query: `include-completed`)
- `GET /projects/{uuid}` - Get project
- `GET /projects/{uuid}/tasks` - Get project tasks (query: `include-completed`)
- `GET /projects/{uuid}/headings` - Get project headings
- `POST /projects` - Create project (body: `title`, `notes`, `when`, `deadline`, `tags`, `area`, `todos`)

**Areas:**
- `GET /areas` - List areas
- `GET /areas/{uuid}` - Get area
- `GET /areas/{uuid}/tasks` - Get area tasks (query: `include_completed`)
- `GET /areas/{uuid}/projects` - Get area projects (query: `include_completed`)

**Tags:**
- `GET /tags` - List tags
- `GET /tags/{name}/tasks` - Get tasks by tag

**Headings:**
- `PATCH /headings/{uuid}` - Update heading (body: `title`)
- `DELETE /headings/{uuid}` - Delete heading

**Health:**
- `GET /health` - Health check

## How It Works

- **Reads** go directly to the Things 3 SQLite database (read-only, no app launch needed)
- **Creates** use the Things URL scheme (`things:///add`, `things:///add-project`)
- **Updates/Deletes/Completes** use AppleScript via `osascript`
- **Specific date scheduling** uses the Things URL scheme with an auth token (AppleScript cannot set activation dates)

The database is accessed read-only using a pure Go SQLite driver (`modernc.org/sqlite` -- no CGO required). The database path is auto-detected from the standard Things 3 location.

## Shell Completions

```bash
thingies completion bash > /etc/bash_completion.d/thingies
thingies completion zsh > "${fpath[1]}/_thingies"
thingies completion fish > ~/.config/fish/completions/thingies.fish
```

## Development

```bash
make build    # Build to bin/thingies
make install  # Install to /usr/local/bin
make test     # Run tests
make fmt      # Format code
make tidy     # go mod tidy
```
