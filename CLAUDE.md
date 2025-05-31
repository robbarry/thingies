# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Thingies is a Python CLI tool that provides read-only access to the Things3 task management app database. It allows users to list tasks, projects, areas, and tags, as well as search through their tasks from the terminal.

## Development Commands

### Setup and Installation
```bash
# Install in development mode using uv
uv pip install -e .
```

### Running the CLI
```bash
# Run the CLI after installation
thingies --help

# Common development commands while testing
thingies list
thingies list --json
thingies projects
thingies areas
thingies tags
thingies search "keyword"
```

## Architecture and Structure

### Database Access
- The application provides **read-only** access to Things3 SQLite database
- Database location: `~/Library/Group Containers/JLMPQHK86H.com.culturedcode.ThingsMac/ThingsData-*/Things Database.thingsdatabase/main.sqlite`
- Uses SQLite URI mode with `?mode=ro` to ensure read-only connection
- The `ThingsDB` class in `thingies/cli.py` manages database connections

### Key Components

1. **CLI Structure** (`thingies/cli.py`):
   - Uses Click framework for command-line interface
   - Uses Rich library for formatted terminal output
   - Main commands: `list`, `projects`, `areas`, `tags`, `search`
   - Supports both human-readable tables and JSON output

2. **Database Schema Understanding**:
   - `TMTask` table: Contains tasks and projects (differentiated by `type` field)
   - `TMArea` table: Contains areas of responsibility
   - `TMTag` table: Contains tags
   - `TMTaskTag` table: Many-to-many relationship between tasks and tags

### Important Implementation Details

- Task types: `type = 0` for tasks, `type = 1` for projects
- Task status: `status = 0` for incomplete, `status = 3` for completed, `status = 2` for canceled
- The `todayIndex` field indicates if a task is in the Today view
- All queries filter out trashed items (`trashed = 0`)
- Dates are stored as Unix timestamps and need conversion

## Development Notes

- Python 3.13+ is required
- Uses `uv` as the package manager (per user's global CLAUDE.md)
- No test suite currently exists
- No linting or formatting tools are configured