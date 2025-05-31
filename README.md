# Thingies - Terminal Interface for Things3

A Python CLI tool for interacting with the Things3 task management app database.

## Installation

```bash
uv pip install -e .
```

## Usage

### Basic Commands

```bash
# List all incomplete tasks
thingies list

# List completed tasks
thingies list --status completed

# List tasks in Today view
thingies list --today

# Filter by project or area
thingies list --project "Work" --area "Home"

# List all projects
thingies projects

# List all areas
thingies areas

# List all tags
thingies tags

# Search tasks
thingies search "meeting"
thingies search "budget" --in-notes
```

### Output Formats

```bash
# Human-readable output (default)
thingies list

# JSON output
thingies --json list
```

### Custom Database Location

```bash
# Use a specific database file
thingies --db /path/to/database.sqlite list
```

## Features

- **Read-only access** to Things3 database
- Human-readable tables with rich formatting
- JSON output for scripting
- Search by title and notes
- Filter by status, area, project, and tags
- View task counts for projects and areas

## Database Location

By default, thingies looks for the Things3 database at:
```
~/Library/Group Containers/JLMPQHK86H.com.culturedcode.ThingsMac/ThingsData-*/Things Database.thingsdatabase/main.sqlite
```

## Safety

- Always uses read-only database connections
- Things3 must be closed before accessing the database
- Recommend backing up your database before use