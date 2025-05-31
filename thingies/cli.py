"""Main CLI interface for thingies"""

import json
import sqlite3
from datetime import datetime
from pathlib import Path
from typing import Optional

import click
from rich import box
from rich.console import Console
from rich.table import Table

from thingies.config import DB_PATH

console = Console()


class ThingsDB:
    def __init__(self, db_path: Optional[str] = None):
        if db_path:
            self.db_path = Path(db_path)
        else:
            # Default Things3 database location
            self.db_path = DB_PATH

            if self.db_path.is_dir():
                # Find the ThingsData directory
                for item in self.db_path.iterdir():
                    if item.name.startswith("ThingsData-"):
                        self.db_path = (
                            item / "Things Database.thingsdatabase" / "main.sqlite"
                        )
                        break

        if not self.db_path.exists():
            raise click.ClickException(f"Database not found at {self.db_path}")

        # Connect read-only
        self.conn = sqlite3.connect(f"file:{self.db_path}?mode=ro", uri=True)
        self.conn.row_factory = sqlite3.Row

    def close(self):
        if self.conn:
            self.conn.close()

    def __enter__(self):
        return self

    def __exit__(self, exc_type, exc_val, exc_tb):
        self.close()


def format_date(timestamp: Optional[float]) -> str:
    """Convert unix timestamp to human readable date"""
    if not timestamp:
        return ""
    return datetime.fromtimestamp(timestamp).strftime("%Y-%m-%d %H:%M")


def format_date_short(timestamp: Optional[float]) -> str:
    """Convert unix timestamp to short date"""
    if not timestamp:
        return ""
    return datetime.fromtimestamp(timestamp).strftime("%Y-%m-%d")


@click.group()
@click.option("--db", "-d", help="Path to Things database (default: auto-detect)")
@click.option("--json", "output_json", is_flag=True, help="Output as JSON")
@click.pass_context
def cli(ctx, db, output_json):
    """Thingies - Terminal interface for Things3"""
    ctx.ensure_object(dict)
    ctx.obj["db_path"] = db
    ctx.obj["json"] = output_json


@cli.command()
@click.option(
    "--status",
    type=click.Choice(["all", "incomplete", "completed", "canceled"]),
    default="incomplete",
)
@click.option("--area", help="Filter by area name")
@click.option("--project", help="Filter by project name")
@click.option("--tag", help="Filter by tag name")
@click.option("--today", is_flag=True, help="Show only Today items")
@click.pass_context
def list(ctx, status, area, project, tag, today):
    """List tasks"""
    with ThingsDB(ctx.obj.get("db_path")) as db:
        # Build query
        query = """
            SELECT 
                t.uuid,
                t.title,
                t.status,
                t.type,
                datetime(t.creationDate, 'unixepoch') as created,
                datetime(t.userModificationDate, 'unixepoch') as modified,
                datetime(t.startDate, 'unixepoch') as scheduled,
                datetime(t.deadline, 'unixepoch') as due,
                datetime(t.stopDate, 'unixepoch') as completed,
                a.title as area_name,
                p.title as project_name,
                GROUP_CONCAT(tag.title, ', ') as tags
            FROM TMTask t
            LEFT JOIN TMArea a ON t.area = a.uuid
            LEFT JOIN TMTask p ON t.project = p.uuid AND p.type = 1 AND p.trashed = 0
            LEFT JOIN TMTaskTag tt ON t.uuid = tt.tasks
            LEFT JOIN TMTag tag ON tt.tags = tag.uuid
            WHERE t.type = 0 AND t.trashed = 0
        """

        conditions = []
        params = []

        # Status filter
        if status == "incomplete":
            conditions.append("t.status = 0")
        elif status == "completed":
            conditions.append("t.status = 3")
        elif status == "canceled":
            conditions.append("t.status = 2")

        # Today filter
        if today:
            conditions.append("t.todayIndex > 0")

        # Area filter
        if area:
            conditions.append("LOWER(a.title) LIKE LOWER(?)")
            params.append(f"%{area}%")

        # Project filter
        if project:
            conditions.append("LOWER(p.title) LIKE LOWER(?)")
            params.append(f"%{project}%")

        if conditions:
            query += " AND " + " AND ".join(conditions)

        query += " GROUP BY t.uuid"

        # Tag filter (having clause because of GROUP_CONCAT)
        if tag:
            query += " HAVING LOWER(tags) LIKE LOWER(?)"
            params.append(f"%{tag}%")

        query += ' ORDER BY COALESCE(t.todayIndex, 999999), t."index"'

        cursor = db.conn.execute(query, params)
        tasks = [dict(row) for row in cursor.fetchall()]

        if ctx.obj.get("json"):
            click.echo(json.dumps(tasks, indent=2))
        else:
            if not tasks:
                console.print("[yellow]No tasks found[/yellow]")
                return

            table = Table(title="Tasks", box=box.ROUNDED)
            table.add_column("Title", style="cyan", no_wrap=False)
            table.add_column("Status", style="green")
            table.add_column("Project", style="blue")
            table.add_column("Area", style="magenta")
            table.add_column("Due", style="red")
            table.add_column("Tags", style="yellow")

            for task in tasks:
                status_icon = (
                    "✓" if task["status"] == 3 else "○" if task["status"] == 0 else "✗"
                )
                due_date = format_date_short(
                    datetime.fromisoformat(task["due"]).timestamp()
                    if task["due"]
                    else None
                )

                table.add_row(
                    task["title"],
                    status_icon,
                    task["project_name"] or "",
                    task["area_name"] or "",
                    due_date,
                    task["tags"] or "",
                )

            console.print(table)
            console.print(f"\n[dim]Found {len(tasks)} task(s)[/dim]")


@cli.command()
@click.option("--include-completed", is_flag=True, help="Include completed tasks")
@click.pass_context
def projects(ctx, include_completed):
    """List projects"""
    with ThingsDB(ctx.obj.get("db_path")) as db:
        query = """
            SELECT 
                p.uuid,
                p.title,
                p.status,
                a.title as area_name,
                COUNT(DISTINCT CASE WHEN t.type = 0 AND t.status = 0 THEN t.uuid END) as open_tasks,
                COUNT(DISTINCT CASE WHEN t.type = 0 THEN t.uuid END) as total_tasks
            FROM TMTask p
            LEFT JOIN TMArea a ON p.area = a.uuid
            LEFT JOIN TMTask t ON t.project = p.uuid AND t.trashed = 0
            WHERE p.type = 1 AND p.trashed = 0
        """

        if not include_completed:
            query += " AND p.status = 0"

        query += ' GROUP BY p.uuid ORDER BY p."index"'

        cursor = db.conn.execute(query)
        projects = [dict(row) for row in cursor.fetchall()]

        if ctx.obj.get("json"):
            click.echo(json.dumps(projects, indent=2))
        else:
            if not projects:
                console.print("[yellow]No projects found[/yellow]")
                return

            table = Table(title="Projects", box=box.ROUNDED)
            table.add_column("Project", style="cyan", no_wrap=False)
            table.add_column("Area", style="magenta")
            table.add_column("Tasks", style="green", justify="right")
            table.add_column("Status", style="yellow")

            for proj in projects:
                status = "Active" if proj["status"] == 0 else "Completed"
                task_count = f"{proj['open_tasks']}/{proj['total_tasks']}"

                table.add_row(
                    proj["title"], proj["area_name"] or "", task_count, status
                )

            console.print(table)
            console.print(f"\n[dim]Found {len(projects)} project(s)[/dim]")


@cli.command()
@click.pass_context
def areas(ctx):
    """List areas"""
    with ThingsDB(ctx.obj.get("db_path")) as db:
        query = """
            SELECT 
                a.uuid,
                a.title,
                COUNT(DISTINCT CASE WHEN t.type = 0 AND t.status = 0 THEN t.uuid END) as open_tasks,
                COUNT(DISTINCT CASE WHEN t.type = 1 AND t.status = 0 THEN t.uuid END) as active_projects
            FROM TMArea a
            LEFT JOIN TMTask t ON t.area = a.uuid AND t.trashed = 0
            WHERE a.visible = 1
            GROUP BY a.uuid
            ORDER BY a.\"index\"
        """

        cursor = db.conn.execute(query)
        areas = [dict(row) for row in cursor.fetchall()]

        if ctx.obj.get("json"):
            click.echo(json.dumps(areas, indent=2))
        else:
            if not areas:
                console.print("[yellow]No areas found[/yellow]")
                return

            table = Table(title="Areas", box=box.ROUNDED)
            table.add_column("Area", style="cyan", no_wrap=False)
            table.add_column("Open Tasks", style="green", justify="right")
            table.add_column("Active Projects", style="blue", justify="right")

            for area in areas:
                table.add_row(
                    area["title"], str(area["open_tasks"]), str(area["active_projects"])
                )

            console.print(table)
            console.print(f"\n[dim]Found {len(areas)} area(s)[/dim]")


@cli.command()
@click.pass_context
def tags(ctx):
    """List all tags"""
    with ThingsDB(ctx.obj.get("db_path")) as db:
        query = """
            SELECT 
                t.uuid,
                t.title,
                t.shortcut,
                COUNT(DISTINCT tt.tasks) as task_count
            FROM TMTag t
            LEFT JOIN TMTaskTag tt ON t.uuid = tt.tags
            LEFT JOIN TMTask task ON tt.tasks = task.uuid AND task.trashed = 0
            GROUP BY t.uuid
            ORDER BY t.title
        """

        cursor = db.conn.execute(query)
        tags = [dict(row) for row in cursor.fetchall()]

        if ctx.obj.get("json"):
            click.echo(json.dumps(tags, indent=2))
        else:
            if not tags:
                console.print("[yellow]No tags found[/yellow]")
                return

            table = Table(title="Tags", box=box.ROUNDED)
            table.add_column("Tag", style="cyan")
            table.add_column("Shortcut", style="yellow")
            table.add_column("Tasks", style="green", justify="right")

            for tag in tags:
                table.add_row(
                    tag["title"], tag["shortcut"] or "", str(tag["task_count"])
                )

            console.print(table)
            console.print(f"\n[dim]Found {len(tags)} tag(s)[/dim]")


@cli.command()
@click.argument("search_term")
@click.option("--in-notes", is_flag=True, help="Search in notes as well")
@click.pass_context
def search(ctx, search_term, in_notes):
    """Search for tasks by title (and optionally notes)"""
    with ThingsDB(ctx.obj.get("db_path")) as db:
        query = """
            SELECT 
                t.uuid,
                t.title,
                t.notes,
                t.status,
                t.type,
                p.title as project_name,
                a.title as area_name
            FROM TMTask t
            LEFT JOIN TMArea a ON t.area = a.uuid
            LEFT JOIN TMTask p ON t.project = p.uuid AND p.type = 1 AND p.trashed = 0
            WHERE t.trashed = 0 AND (
                LOWER(t.title) LIKE LOWER(?)
        """

        params = [f"%{search_term}%"]

        if in_notes:
            query += " OR LOWER(t.notes) LIKE LOWER(?)"
            params.append(f"%{search_term}%")

        query += ') ORDER BY t.type, t."index"'

        cursor = db.conn.execute(query, params)
        results = [dict(row) for row in cursor.fetchall()]

        if ctx.obj.get("json"):
            click.echo(json.dumps(results, indent=2))
        else:
            if not results:
                console.print(f"[yellow]No results found for '{search_term}'[/yellow]")
                return

            table = Table(title=f"Search Results for '{search_term}'", box=box.ROUNDED)
            table.add_column("Type", style="yellow")
            table.add_column("Title", style="cyan", no_wrap=False)
            table.add_column("Project", style="blue")
            table.add_column("Area", style="magenta")
            table.add_column("Status", style="green")

            for result in results:
                type_name = (
                    "Task"
                    if result["type"] == 0
                    else "Project" if result["type"] == 1 else "Heading"
                )
                status = (
                    "Open"
                    if result["status"] == 0
                    else "Done" if result["status"] == 3 else "Canceled"
                )

                table.add_row(
                    type_name,
                    result["title"],
                    result["project_name"] or "",
                    result["area_name"] or "",
                    status,
                )

            console.print(table)
            console.print(f"\n[dim]Found {len(results)} result(s)[/dim]")


if __name__ == "__main__":
    cli()
