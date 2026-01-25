# Thingies

A Go CLI for Things 3 with full CRUD access.

## Install

```bash
make build
ln -s $(pwd)/bin/thingies ~/bin/thingies  # or copy to /usr/local/bin
```

## Usage

```bash
# Today's tasks
thingies today

# Inbox
thingies inbox

# Hierarchical view
thingies snapshot

# Search
thingies search "keyword"
thingies search "keyword" --in-notes

# List/filter tasks
thingies tasks list
thingies tasks list --today
thingies tasks list --project "Bills"
thingies tasks list --area "Work"
thingies tasks list --status completed

# Task CRUD
thingies tasks show <uuid>
thingies tasks create "New task" --when today
thingies tasks update <uuid> --notes "Updated notes"
thingies tasks complete <uuid>
thingies tasks delete <uuid>

# Projects
thingies projects list
thingies projects show <uuid>
thingies projects create "New project" --area "Work"
thingies projects update <uuid> --notes "# Markdown supported"
thingies projects complete <uuid>
thingies projects delete <uuid>

# Areas & Tags
thingies areas list
thingies areas show <uuid>
thingies tags list

# JSON output
thingies --json today
thingies --json tasks list
```

## How It Works

- **Reads**: Direct SQLite access (fast, no app launch)
- **Creates**: Things URL scheme (`things:///add`)
- **Updates/Deletes**: AppleScript via `osascript`

## Requirements

- macOS with Things 3 installed
- Go 1.21+ (for building)
