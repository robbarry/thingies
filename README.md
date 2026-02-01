# Thingies

A Go CLI and REST API for Things 3 with full CRUD access.

## Requirements

- macOS with Things 3 installed
- Go 1.21+ (for building)

## Install

```bash
make install  # builds and copies to /usr/local/bin
```

Or build without installing:

```bash
make build    # outputs to bin/thingies
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
thingies search "keyword" --in-notes  # Also search notes
```

### Tasks

```bash
thingies tasks list                                    # All incomplete tasks
thingies tasks list --status all                       # Include completed/canceled
thingies tasks list --today                            # Only Today view
thingies tasks list --area "Work"                      # Filter by area
thingies tasks list --project "Bills" --tag "urgent"   # Filter by project and tag

thingies tasks show <uuid>                             # Full task details
thingies tasks create "New task"                       # Create task
thingies tasks create "New task" --when today --list "Project" --heading "Section"
thingies tasks create "New task" --deadline 2026-02-15 # With due date

thingies tasks update <uuid> --title "New" --notes "Updated" --when tomorrow
thingies tasks update <uuid> --deadline 2026-02-15     # Set due date
thingies tasks complete <uuid>
thingies tasks cancel <uuid>
thingies tasks delete <uuid>
```

### Projects

```bash
thingies projects list
thingies projects show <uuid>
thingies projects create "New project" --area "Work" --todos "Task 1\nTask 2"
thingies projects create "New project" --deadline 2026-03-01  # With due date
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
--db, -d     Path to Things database (default: auto-detect)
--json, -j   Output as JSON
--no-color   Disable colors
--verbose    Verbose output
```

### Name Resolution

Most commands accept either a UUID or a name for areas/projects. Names are resolved to UUIDs automatically; if multiple items share a name, you'll be prompted to use the UUID.

## REST API

```bash
thingies serve              # Start on 0.0.0.0:8484
thingies serve -p 3000      # Custom port
thingies serve --host 127.0.0.1  # Localhost only
```

### Endpoints

**Views:**
- `GET /today`, `/inbox`, `/upcoming`, `/someday`, `/anytime`, `/logbook`, `/deadlines`
- `GET /snapshot` - Full hierarchical view as JSON

**Tasks:**
- `GET /tasks` - List tasks
- `GET /tasks/search?q=query` - Search tasks
- `POST /tasks` - Create task
- `GET /tasks/{uuid}` - Get task
- `PATCH /tasks/{uuid}` - Update task
- `DELETE /tasks/{uuid}` - Delete task
- `POST /tasks/{uuid}/complete`, `/cancel`, `/move-to-today`, `/move-to-someday`

**Projects:**
- `GET /projects` - List projects
- `POST /projects` - Create project
- `GET /projects/{uuid}` - Get project
- `GET /projects/{uuid}/tasks` - Get project tasks
- `GET /projects/{uuid}/headings` - Get project headings

**Areas:**
- `GET /areas` - List areas
- `GET /areas/{uuid}` - Get area
- `GET /areas/{uuid}/tasks` - Get area tasks
- `GET /areas/{uuid}/projects` - Get area projects

**Tags:**
- `GET /tags` - List tags
- `GET /tags/{name}/tasks` - Get tasks by tag

**Headings:**
- `PATCH /headings/{uuid}` - Update heading
- `DELETE /headings/{uuid}` - Delete heading

**Health:**
- `GET /health` - Health check

## How It Works

- **Reads**: Direct SQLite access (fast, no app launch needed)
- **Creates**: Things URL scheme (`things:///add`, `things:///add-project`)
- **Updates/Deletes/Completes**: AppleScript via `osascript`

The database is accessed read-only using a pure Go SQLite driver (no CGO required).

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
