
# Things 3 Task Management

## Behavioral Rules

- **Present ONE task at a time** (ADHD consideration)
- Focus on completion over optimization
- Create minimal tasks - only what's requested
- Skip unnecessary detail unless asked
- Provide concise answers

When Rob asks "What's next?" or about tasks:
1. Get snapshot for overview: `thingies snapshot`
2. Check today's tasks: `thingies today`
3. Check inbox for items needing processing: `thingies inbox`

## Overview

Things 3 is Rob's primary task management app (macOS/iOS). Two interfaces are available:

1. **`thingies` CLI** (preferred) - Go binary with direct database access + AppleScript
2. **REST API** (backup) - HTTP API running on VEGA over Tailscale

**Always prefer the CLI** - it's faster, works offline, and doesn't require network access.

## Data Structure

```
Area (top-level container)
└── Project
    └── Heading (section within project)
        └── Task
```

### Current Areas
| Area | Purpose |
|------|---------|
| **Work** | Work-related tasks and projects |
| **Home** | Personal/household tasks and projects |
| **Auto** | Container for AI-generated tasks |

### Special Projects
| Project | Area | Description |
|---------|------|-------------|
| **Want To** | Home | Uses headings: Read, Buy, Do, Travel, Watch, Write |
| **Bills** | Home | Recurring payments and financial tasks |

### Views
| View | Command | Description |
|------|---------|-------------|
| Today | `thingies today` | Tasks scheduled for today |
| Inbox | `thingies inbox` | Unprocessed tasks (triage queue) |
| Upcoming | `thingies upcoming` | Future scheduled tasks (shows dates) |
| Anytime | `thingies anytime` | Available tasks with no specific date |
| Someday | `thingies someday` | Deferred/backlog tasks |
| Logbook | `thingies logbook` | Completed tasks |

---

## CLI: `thingies`

### Installation

The source code lives at `~/lrepos/thingies` (same path on all machines).

```bash
cd ~/lrepos/thingies
make build
ln -sf ~/lrepos/thingies/bin/thingies ~/bin/thingies
```

If the binary is missing or outdated, rebuild it with `make build`.

### Quick Start

```bash
# Orientation - see everything
thingies snapshot

# Today's tasks
thingies today

# Inbox
thingies inbox

# Search
thingies search "keyword"
```

### Reading Tasks

```bash
# Views (shortcuts)
thingies today                              # Today's tasks
thingies inbox                              # Inbox (unprocessed)
thingies upcoming                           # Future scheduled tasks (shows dates)
thingies someday                            # Deferred tasks
thingies anytime                            # Available but not scheduled
thingies logbook -n 50                      # Completed tasks (default 50)
thingies snapshot                           # Hierarchical view of everything

# List with filters
thingies tasks list                         # All incomplete tasks
thingies tasks list --today                 # Only Today view
thingies tasks list --project "Bills"       # Filter by project (name or UUID)
thingies tasks list --area "Work"           # Filter by area (name or UUID)
thingies tasks list --tag "urgent"          # Filter by tag
thingies tasks list --status completed      # Completed tasks
thingies tasks list --status all            # All statuses

# Show details (accepts name or UUID)
thingies tasks show <uuid>
thingies projects show "Want To"            # By name
thingies projects show <uuid>               # By UUID
thingies areas show "Work"                  # By name
thingies areas show <uuid>                  # By UUID

# Search
thingies search "keyword"
thingies search "keyword" --in-notes

# List entities
thingies projects list
thingies areas list
thingies tags list

# JSON output (for scripting)
thingies --json today
thingies --json tasks list
```

### Creating Tasks

```bash
# Basic task
thingies tasks create "Call dentist"

# Task for today
thingies tasks create "Review PR" --when today

# Task with all options
thingies tasks create "Quarterly review" \
  --when today \
  --deadline 2026-02-01 \
  --tags "work,finance" \
  --notes "Focus on Q4 numbers" \
  --list "Project Name" \
  --heading "Section"

# Create project
thingies projects create "New Project" --area "Work"
```

### Updating Tasks

```bash
# Update title
thingies tasks update <uuid> --title "New title"

# Update notes (supports Markdown)
thingies tasks update <uuid> --notes "# Header\n\nNotes here"

# Reschedule
thingies tasks update <uuid> --when today
thingies tasks update <uuid> --when tomorrow
thingies tasks update <uuid> --when 2026-02-15

# Set deadline
thingies tasks update <uuid> --deadline 2026-02-01

# Update tags
thingies tasks update <uuid> --tags "work,urgent"

# Update project
thingies projects update <uuid> --notes "Project notes with **Markdown**"
```

### Completing/Deleting Tasks

```bash
# Complete
thingies tasks complete <uuid>
thingies projects complete <uuid>

# Cancel
thingies tasks cancel <uuid>

# Delete (moves to trash)
thingies tasks delete <uuid>
thingies tasks delete <uuid> -f              # Skip confirmation
thingies projects delete <uuid>
```

### Common Workflows

**Morning review:**
```bash
thingies snapshot                            # See everything
thingies inbox                               # Check for unprocessed items
thingies today                               # Focus on today
```

**Quick add to today:**
```bash
thingies tasks create "Quick task" --when today
```

**Find and complete:**
```bash
# Search, get UUID from output, then complete
thingies search "dentist"
thingies tasks complete <uuid-from-search>
```

**Add to Want To project:**
```bash
# First find the project UUID
thingies projects list
thingies tasks create "Book title" --list "Want To" --heading "Read"
```

---

## Local REST API Server

The CLI includes a built-in HTTP server for programmatic access:

```bash
thingies serve                    # Start on 0.0.0.0:8484 (default)
thingies serve -p 3000            # Custom port
thingies serve --host 127.0.0.1   # Localhost only
```

Endpoints mirror the CLI: `/today`, `/inbox`, `/upcoming`, `/tasks`, `/projects`, `/areas`, `/snapshot`, etc.

---

## Remote REST API (Backup)

Use when the CLI is unavailable (e.g., remote access, different machine without the code).

**Base URL:** `${THINGS_API_URL}` (defaults to `https://vega.taildef9.ts.net`)

```bash
# Snapshot
curl -s "${THINGS_API_URL:-https://vega.taildef9.ts.net}/snapshot" | jq -r .snapshot

# Today
curl -s "${THINGS_API_URL:-https://vega.taildef9.ts.net}/today" | jq

# Add task
curl -s -X POST "${THINGS_API_URL:-https://vega.taildef9.ts.net}/tasks" \
  -H "Content-Type: application/json" \
  -d '{"title": "New task", "when": "today"}' | jq

# Complete task
curl -s -X POST "${THINGS_API_URL:-https://vega.taildef9.ts.net}/tasks/{uuid}/complete" | jq
```

See `references/rest-api.md` for complete API documentation.

### API Troubleshooting

```bash
# Check if API is responding
curl -s "${THINGS_API_URL:-https://vega.taildef9.ts.net}/health" | jq

# Restart if needed
ssh vega 'PATH=/Users/rob/.nvm/versions/node/v22.19.0/bin:$PATH pm2 restart things-api'
```

---

## Resources

- `references/rest-api.md` - Complete REST API endpoint documentation
