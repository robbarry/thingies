# Thingies

A Go CLI and REST API for Things 3 with full CRUD access.

## Install

```bash
make install  # builds and copies to /usr/local/bin
```

## CLI Usage

```bash
# Views
thingies today
thingies inbox
thingies upcoming
thingies someday
thingies anytime
thingies logbook -n 50
thingies snapshot

# Search
thingies search "keyword"
thingies search "keyword" --in-notes

# Tasks
thingies tasks list
thingies tasks list --today --area "Work" --project "Bills" --tag "urgent"
thingies tasks show <uuid>
thingies tasks create "New task" --when today --list "Project" --heading "Section"
thingies tasks update <uuid> --title "New" --notes "Updated" --when tomorrow
thingies tasks complete <uuid>
thingies tasks cancel <uuid>
thingies tasks delete <uuid>

# Projects
thingies projects list
thingies projects show <uuid>
thingies projects create "New project" --area "Work" --todos "Task 1\nTask 2"
thingies projects update <uuid> --notes "# Markdown supported"
thingies projects complete <uuid>
thingies projects delete <uuid>

# Areas
thingies areas list
thingies areas show <uuid>
thingies areas create "Name"
thingies areas update <uuid> --title "New name"
thingies areas delete <uuid>

# Tags
thingies tags list
thingies tags create "Name"
thingies tags create "Nested" --parent <uuid>
thingies tags update <uuid> --title "New name"
thingies tags delete <uuid>

# JSON output
thingies --json today
thingies --json tasks list
```

## REST API

```bash
thingies serve              # 0.0.0.0:8484
thingies serve -p 3000      # custom port
```

Key endpoints:
- `GET /today`, `/inbox`, `/upcoming`, `/someday`, `/anytime`, `/logbook`
- `GET /tasks`, `POST /tasks`, `GET /tasks/{uuid}`, `PATCH /tasks/{uuid}`, `DELETE /tasks/{uuid}`
- `POST /tasks/{uuid}/complete`, `/cancel`, `/move-to-today`
- `GET /projects`, `/areas`, `/tags`
- `GET /snapshot`

## How It Works

- **Reads**: Direct SQLite access (fast, no app launch)
- **Creates**: Things URL scheme (`things:///add`)
- **Updates/Deletes**: AppleScript via `osascript`

## Requirements

- macOS with Things 3 installed
- Go 1.21+ (for building)
