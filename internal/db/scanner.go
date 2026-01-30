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
			&task.ProjectUUID,
			&task.ProjectName,
			&task.HeadingUUID,
			&task.HeadingName,
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
// These fields are binary-packed dates, not timestamps:
//   - Bits 16-26: Year (11 bits)
//   - Bits 12-15: Month (4 bits)
//   - Bits 7-11: Day (5 bits)
func thingsDateToNullTime(ts sql.NullFloat64) sql.NullTime {
	if !ts.Valid || ts.Float64 == 0 {
		return sql.NullTime{}
	}

	packed := int(ts.Float64)

	// Extract year, month, day from packed bits
	year := (packed & 0x7FF0000) >> 16  // bits 16-26
	month := (packed & 0xF000) >> 12    // bits 12-15
	day := (packed & 0xF80) >> 7        // bits 7-11

	return sql.NullTime{
		Time:  time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC),
		Valid: true,
	}
}
