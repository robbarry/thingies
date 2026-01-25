package db

import (
	"database/sql"
	"fmt"
	"time"

	"thingies/internal/models"
)

// scanTasks scans rows into a slice of Task
func scanTasks(rows *sql.Rows) ([]models.Task, error) {
	var tasks []models.Task

	for rows.Next() {
		var task models.Task
		var createdTS, modifiedTS, startTS, deadlineTS, completedTS sql.NullFloat64
		var isRepeating int

		err := rows.Scan(
			&task.UUID,
			&task.Title,
			&task.Notes,
			&task.Status,
			&task.Type,
			&createdTS,
			&modifiedTS,
			&startTS,
			&deadlineTS,
			&completedTS,
			&task.AreaName,
			&task.ProjectName,
			&task.Tags,
			&isRepeating,
			&task.TodayIndex,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}

		task.Created = timestampToNullTime(createdTS)
		task.Modified = timestampToNullTime(modifiedTS)
		task.Scheduled = thingsDateToNullTime(startTS)
		task.Deadline = thingsDateToNullTime(deadlineTS)
		task.Completed = timestampToNullTime(completedTS)
		task.IsRepeating = isRepeating == 1

		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return tasks, nil
}

// scanProjects scans rows into a slice of Project
func scanProjects(rows *sql.Rows) ([]models.Project, error) {
	var projects []models.Project

	for rows.Next() {
		var proj models.Project

		err := rows.Scan(
			&proj.UUID,
			&proj.Title,
			&proj.Notes,
			&proj.Status,
			&proj.AreaName,
			&proj.OpenTasks,
			&proj.TotalTasks,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan project: %w", err)
		}

		projects = append(projects, proj)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return projects, nil
}

// thingsDateEpoch is the epoch for Things 3 startDate/deadline fields.
// Things uses Nov 11, 2021 00:00:00 UTC as the reference date for these fields.
const thingsDateEpoch = 1636588800

// timestampToNullTime converts a Unix timestamp to sql.NullTime
// Used for creationDate, userModificationDate, stopDate which are Unix timestamps
func timestampToNullTime(ts sql.NullFloat64) sql.NullTime {
	if !ts.Valid || ts.Float64 == 0 {
		return sql.NullTime{}
	}
	return sql.NullTime{
		Time:  time.Unix(int64(ts.Float64), 0),
		Valid: true,
	}
}

// thingsDateToNullTime converts Things date fields (startDate, deadline) to sql.NullTime
// These fields use a different epoch than standard Unix time
func thingsDateToNullTime(ts sql.NullFloat64) sql.NullTime {
	if !ts.Valid || ts.Float64 == 0 {
		return sql.NullTime{}
	}
	// Convert from Things date epoch to Unix epoch
	unixTime := int64(ts.Float64) + thingsDateEpoch
	return sql.NullTime{
		Time:  time.Unix(unixTime, 0),
		Valid: true,
	}
}
