package db

import (
	"database/sql"
	"fmt"
	"strings"

	"thingies/internal/models"
)

// TaskFilter contains filters for listing tasks
type TaskFilter struct {
	Status        string // "all", "incomplete", "completed", "canceled"
	Area          string
	Project       string
	Tag           string
	Today         bool
	IncludeFuture bool
}

// ListTasks returns tasks matching the filter
func (db *ThingsDB) ListTasks(filter TaskFilter) ([]models.Task, error) {
	query := `
		SELECT
			t.uuid,
			t.title,
			t.notes,
			t.status,
			t.type,
			t.creationDate,
			t.userModificationDate,
			t.startDate,
			t.deadline,
			t.stopDate,
			COALESCE(a.title, pa.title, hpa.title) as area_name,
			COALESCE(p.title, hp.title) as project_name,
			h.title as heading_name,
			GROUP_CONCAT(tag.title, ', ') as tags,
			CASE WHEN t.rt1_repeatingTemplate IS NOT NULL THEN 1 ELSE 0 END as is_repeating,
			t.todayIndex
		FROM TMTask t
		LEFT JOIN TMArea a ON t.area = a.uuid
		LEFT JOIN TMTask p ON t.project = p.uuid AND p.type = 1 AND p.trashed = 0
		LEFT JOIN TMArea pa ON p.area = pa.uuid
		LEFT JOIN TMTask p2 ON t.project = p2.uuid AND p2.type = 1
		LEFT JOIN TMTask h ON t.heading = h.uuid
		LEFT JOIN TMTask hp ON h.project = hp.uuid AND hp.type = 1
		LEFT JOIN TMArea hpa ON hp.area = hpa.uuid
		LEFT JOIN TMTaskTag tt ON t.uuid = tt.tasks
		LEFT JOIN TMTag tag ON tt.tags = tag.uuid
		WHERE t.type = 0 AND t.trashed = 0
			AND (t.project IS NULL OR p2.trashed = 0)
			AND (h.uuid IS NULL OR hp.uuid IS NULL OR hp.trashed = 0)
	`

	var conditions []string
	var params []interface{}

	// Status filter
	switch filter.Status {
	case "incomplete", "":
		conditions = append(conditions, "t.status = 0")
	case "completed":
		conditions = append(conditions, "t.status = 3")
	case "canceled":
		conditions = append(conditions, "t.status = 2")
	// "all" - no filter
	}

	// Today filter based on things.py logic:
	// 1. Anytime tasks with start dates (start=1, startDate set)
	// 2. Someday tasks with past start dates (start=2, startDate <= today)
	// 3. Overdue tasks by deadline (no startDate, deadline <= today, not suppressed)
	if filter.Today {
		todayPacked := TodayPackedDate()
		conditions = append(conditions, fmt.Sprintf(`(
			(t.start = 1 AND t.startDate IS NOT NULL)
			OR (t.start = 2 AND t.startDate IS NOT NULL AND t.startDate <= %d)
			OR (t.startDate IS NULL AND t.deadline IS NOT NULL AND t.deadline <= %d AND t.deadlineSuppressionDate IS NULL)
		)`, todayPacked, todayPacked))
	}

	// Area filter
	if filter.Area != "" {
		conditions = append(conditions, "LOWER(a.title) LIKE LOWER(?)")
		params = append(params, "%"+filter.Area+"%")
	}

	// Project filter
	if filter.Project != "" {
		conditions = append(conditions, "LOWER(p.title) LIKE LOWER(?)")
		params = append(params, "%"+filter.Project+"%")
	}

	// Future repeating tasks filter (startDate is packed date format, not Unix timestamp)
	if !filter.IncludeFuture {
		todayPacked := TodayPackedDate()
		conditions = append(conditions, fmt.Sprintf("(t.rt1_repeatingTemplate IS NULL OR t.startDate IS NULL OR t.startDate <= %d)", todayPacked))
	}

	if len(conditions) > 0 {
		query += " AND " + strings.Join(conditions, " AND ")
	}

	query += " GROUP BY t.uuid"

	// Tag filter (HAVING because of GROUP_CONCAT)
	if filter.Tag != "" {
		query += " HAVING LOWER(tags) LIKE LOWER(?)"
		params = append(params, "%"+filter.Tag+"%")
	}

	query += ` ORDER BY COALESCE(t.todayIndex, 999999), t."index"`

	rows, err := db.conn.Query(query, params...)
	if err != nil {
		return nil, fmt.Errorf("failed to query tasks: %w", err)
	}
	defer rows.Close()

	return scanTasks(rows)
}

// GetTask returns a single task by UUID
func (db *ThingsDB) GetTask(uuid string) (*models.Task, error) {
	query := `
		SELECT
			t.uuid,
			t.title,
			t.notes,
			t.status,
			t.type,
			t.creationDate,
			t.userModificationDate,
			t.startDate,
			t.deadline,
			t.stopDate,
			COALESCE(a.title, pa.title, hpa.title) as area_name,
			COALESCE(p.title, hp.title) as project_name,
			h.title as heading_name,
			GROUP_CONCAT(tag.title, ', ') as tags,
			CASE WHEN t.rt1_repeatingTemplate IS NOT NULL THEN 1 ELSE 0 END as is_repeating,
			t.todayIndex
		FROM TMTask t
		LEFT JOIN TMArea a ON t.area = a.uuid
		LEFT JOIN TMTask p ON t.project = p.uuid AND p.type = 1
		LEFT JOIN TMArea pa ON p.area = pa.uuid
		LEFT JOIN TMTask h ON t.heading = h.uuid
		LEFT JOIN TMTask hp ON h.project = hp.uuid AND hp.type = 1
		LEFT JOIN TMArea hpa ON hp.area = hpa.uuid
		LEFT JOIN TMTaskTag tt ON t.uuid = tt.tasks
		LEFT JOIN TMTag tag ON tt.tags = tag.uuid
		WHERE t.uuid = ?
		GROUP BY t.uuid
	`

	rows, err := db.conn.Query(query, uuid)
	if err != nil {
		return nil, fmt.Errorf("failed to query task: %w", err)
	}
	defer rows.Close()

	tasks, err := scanTasks(rows)
	if err != nil {
		return nil, err
	}

	if len(tasks) == 0 {
		return nil, fmt.Errorf("task not found: %s", uuid)
	}

	return &tasks[0], nil
}

// ListProjects returns all projects
func (db *ThingsDB) ListProjects(includeCompleted bool) ([]models.Project, error) {
	query := `
		SELECT
			p.uuid,
			p.title,
			p.notes,
			p.status,
			a.title as area_name,
			COUNT(DISTINCT CASE WHEN t.type = 0 AND t.status = 0 THEN t.uuid END) as open_tasks,
			COUNT(DISTINCT CASE WHEN t.type = 0 THEN t.uuid END) as total_tasks
		FROM TMTask p
		LEFT JOIN TMArea a ON p.area = a.uuid
		LEFT JOIN TMTask t ON t.project = p.uuid AND t.trashed = 0
		WHERE p.type = 1 AND p.trashed = 0
	`

	if !includeCompleted {
		query += " AND p.status = 0"
	}

	query += ` GROUP BY p.uuid ORDER BY p."index"`

	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query projects: %w", err)
	}
	defer rows.Close()

	return scanProjects(rows)
}

// GetProject returns a single project by UUID
func (db *ThingsDB) GetProject(uuid string) (*models.Project, error) {
	query := `
		SELECT
			p.uuid,
			p.title,
			p.notes,
			p.status,
			a.title as area_name,
			COUNT(DISTINCT CASE WHEN t.type = 0 AND t.status = 0 THEN t.uuid END) as open_tasks,
			COUNT(DISTINCT CASE WHEN t.type = 0 THEN t.uuid END) as total_tasks
		FROM TMTask p
		LEFT JOIN TMArea a ON p.area = a.uuid
		LEFT JOIN TMTask t ON t.project = p.uuid AND t.trashed = 0
		WHERE p.uuid = ? AND p.type = 1
		GROUP BY p.uuid
	`

	rows, err := db.conn.Query(query, uuid)
	if err != nil {
		return nil, fmt.Errorf("failed to query project: %w", err)
	}
	defer rows.Close()

	projects, err := scanProjects(rows)
	if err != nil {
		return nil, err
	}

	if len(projects) == 0 {
		return nil, fmt.Errorf("project not found: %s", uuid)
	}

	return &projects[0], nil
}

// GetProjectTasks returns tasks belonging to a project
func (db *ThingsDB) GetProjectTasks(projectUUID string, includeCompleted bool) ([]models.Task, error) {
	query := `
		SELECT
			t.uuid,
			t.title,
			t.notes,
			t.status,
			t.type,
			t.creationDate,
			t.userModificationDate,
			t.startDate,
			t.deadline,
			t.stopDate,
			COALESCE(a.title, pa.title, hpa.title) as area_name,
			COALESCE(p.title, hp.title) as project_name,
			h.title as heading_name,
			GROUP_CONCAT(tag.title, ', ') as tags,
			CASE WHEN t.rt1_repeatingTemplate IS NOT NULL THEN 1 ELSE 0 END as is_repeating,
			t.todayIndex
		FROM TMTask t
		LEFT JOIN TMArea a ON t.area = a.uuid
		LEFT JOIN TMTask p ON t.project = p.uuid AND p.type = 1
		LEFT JOIN TMArea pa ON p.area = pa.uuid
		LEFT JOIN TMTask h ON t.heading = h.uuid
		LEFT JOIN TMTask hp ON h.project = hp.uuid AND hp.type = 1
		LEFT JOIN TMArea hpa ON hp.area = hpa.uuid
		LEFT JOIN TMTaskTag tt ON t.uuid = tt.tasks
		LEFT JOIN TMTag tag ON tt.tags = tag.uuid
		WHERE (t.project = ? OR hp.uuid = ?) AND t.type = 0 AND t.trashed = 0
	`

	if !includeCompleted {
		query += " AND t.status = 0"
	}

	query += ` GROUP BY t.uuid ORDER BY t."index"`

	rows, err := db.conn.Query(query, projectUUID, projectUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to query project tasks: %w", err)
	}
	defer rows.Close()

	return scanTasks(rows)
}

// ListAreas returns all visible areas
func (db *ThingsDB) ListAreas() ([]models.Area, error) {
	query := `
		SELECT
			a.uuid,
			a.title,
			COUNT(DISTINCT CASE WHEN t.type = 0 AND t.status = 0 THEN t.uuid END) as open_tasks,
			COUNT(DISTINCT CASE WHEN t.type = 1 AND t.status = 0 THEN t.uuid END) as active_projects
		FROM TMArea a
		LEFT JOIN TMTask t ON t.area = a.uuid AND t.trashed = 0
		WHERE a.visible IS NULL OR a.visible != 0
		GROUP BY a.uuid
		ORDER BY a."index"
	`

	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query areas: %w", err)
	}
	defer rows.Close()

	var areas []models.Area
	for rows.Next() {
		var area models.Area
		if err := rows.Scan(&area.UUID, &area.Title, &area.OpenTasks, &area.ActiveProjects); err != nil {
			return nil, fmt.Errorf("failed to scan area: %w", err)
		}
		areas = append(areas, area)
	}

	return areas, rows.Err()
}

// GetArea returns a single area by UUID
func (db *ThingsDB) GetArea(uuid string) (*models.Area, error) {
	query := `
		SELECT
			a.uuid,
			a.title,
			COUNT(DISTINCT CASE WHEN t.type = 0 AND t.status = 0 THEN t.uuid END) as open_tasks,
			COUNT(DISTINCT CASE WHEN t.type = 1 AND t.status = 0 THEN t.uuid END) as active_projects
		FROM TMArea a
		LEFT JOIN TMTask t ON t.area = a.uuid AND t.trashed = 0
		WHERE a.uuid = ?
		GROUP BY a.uuid
	`

	var area models.Area
	err := db.conn.QueryRow(query, uuid).Scan(&area.UUID, &area.Title, &area.OpenTasks, &area.ActiveProjects)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("area not found: %s", uuid)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query area: %w", err)
	}

	return &area, nil
}

// GetAreaProjects returns projects belonging to an area
func (db *ThingsDB) GetAreaProjects(areaUUID string, includeCompleted bool) ([]models.Project, error) {
	query := `
		SELECT
			p.uuid,
			p.title,
			p.notes,
			p.status,
			a.title as area_name,
			COUNT(DISTINCT CASE WHEN t.type = 0 AND t.status = 0 THEN t.uuid END) as open_tasks,
			COUNT(DISTINCT CASE WHEN t.type = 0 THEN t.uuid END) as total_tasks
		FROM TMTask p
		LEFT JOIN TMArea a ON p.area = a.uuid
		LEFT JOIN TMTask t ON t.project = p.uuid AND t.trashed = 0
		WHERE p.area = ? AND p.type = 1 AND p.trashed = 0
	`

	if !includeCompleted {
		query += " AND p.status = 0"
	}

	query += ` GROUP BY p.uuid ORDER BY p."index"`

	rows, err := db.conn.Query(query, areaUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to query area projects: %w", err)
	}
	defer rows.Close()

	return scanProjects(rows)
}

// GetAreaTasks returns tasks directly under an area (not in projects)
func (db *ThingsDB) GetAreaTasks(areaUUID string, includeCompleted bool) ([]models.Task, error) {
	query := `
		SELECT
			t.uuid,
			t.title,
			t.notes,
			t.status,
			t.type,
			t.creationDate,
			t.userModificationDate,
			t.startDate,
			t.deadline,
			t.stopDate,
			a.title as area_name,
			NULL as project_name,
			NULL as heading_name,
			GROUP_CONCAT(tag.title, ', ') as tags,
			CASE WHEN t.rt1_repeatingTemplate IS NOT NULL THEN 1 ELSE 0 END as is_repeating,
			t.todayIndex
		FROM TMTask t
		LEFT JOIN TMArea a ON t.area = a.uuid
		LEFT JOIN TMTaskTag tt ON t.uuid = tt.tasks
		LEFT JOIN TMTag tag ON tt.tags = tag.uuid
		WHERE t.area = ? AND t.project IS NULL AND t.type = 0 AND t.trashed = 0
	`

	if !includeCompleted {
		query += " AND t.status = 0"
	}

	query += ` GROUP BY t.uuid ORDER BY t."index"`

	rows, err := db.conn.Query(query, areaUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to query area tasks: %w", err)
	}
	defer rows.Close()

	return scanTasks(rows)
}

// ListTags returns all tags
func (db *ThingsDB) ListTags() ([]models.Tag, error) {
	query := `
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
	`

	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query tags: %w", err)
	}
	defer rows.Close()

	var tags []models.Tag
	for rows.Next() {
		var tag models.Tag
		if err := rows.Scan(&tag.UUID, &tag.Title, &tag.Shortcut, &tag.TaskCount); err != nil {
			return nil, fmt.Errorf("failed to scan tag: %w", err)
		}
		tags = append(tags, tag)
	}

	return tags, rows.Err()
}

// Search searches for tasks by title and optionally notes
func (db *ThingsDB) Search(term string, includeNotes, includeFuture bool) ([]models.Task, error) {
	query := `
		SELECT
			t.uuid,
			t.title,
			t.notes,
			t.status,
			t.type,
			t.creationDate,
			t.userModificationDate,
			t.startDate,
			t.deadline,
			t.stopDate,
			COALESCE(a.title, pa.title, hpa.title) as area_name,
			COALESCE(p.title, hp.title) as project_name,
			h.title as heading_name,
			GROUP_CONCAT(tag.title, ', ') as tags,
			CASE WHEN t.rt1_repeatingTemplate IS NOT NULL THEN 1 ELSE 0 END as is_repeating,
			t.todayIndex
		FROM TMTask t
		LEFT JOIN TMArea a ON t.area = a.uuid
		LEFT JOIN TMTask p ON t.project = p.uuid AND p.type = 1 AND p.trashed = 0
		LEFT JOIN TMArea pa ON p.area = pa.uuid
		LEFT JOIN TMTask p2 ON t.project = p2.uuid AND p2.type = 1
		LEFT JOIN TMTask h ON t.heading = h.uuid
		LEFT JOIN TMTask hp ON h.project = hp.uuid AND hp.type = 1
		LEFT JOIN TMArea hpa ON hp.area = hpa.uuid
		LEFT JOIN TMTaskTag tt ON t.uuid = tt.tasks
		LEFT JOIN TMTag tag ON tt.tags = tag.uuid
		WHERE t.trashed = 0
			AND (t.project IS NULL OR p2.trashed = 0)
			AND (h.uuid IS NULL OR hp.uuid IS NULL OR hp.trashed = 0)
			AND (LOWER(t.title) LIKE LOWER(?)
	`

	params := []interface{}{"%" + term + "%"}

	if includeNotes {
		query += " OR LOWER(t.notes) LIKE LOWER(?)"
		params = append(params, "%"+term+"%")
	}

	query += ")"

	if !includeFuture {
		todayPacked := TodayPackedDate()
		query += fmt.Sprintf(" AND (t.rt1_repeatingTemplate IS NULL OR t.startDate IS NULL OR t.startDate <= %d)", todayPacked)
	}

	query += ` GROUP BY t.uuid ORDER BY t.type, t."index"`

	rows, err := db.conn.Query(query, params...)
	if err != nil {
		return nil, fmt.Errorf("failed to search: %w", err)
	}
	defer rows.Close()

	return scanTasks(rows)
}

// GetInboxTasks returns tasks in the inbox (start = 0, meaning unprocessed)
func (db *ThingsDB) GetInboxTasks() ([]models.Task, error) {
	query := `
		SELECT
			t.uuid,
			t.title,
			t.notes,
			t.status,
			t.type,
			t.creationDate,
			t.userModificationDate,
			t.startDate,
			t.deadline,
			t.stopDate,
			COALESCE(a.title, pa.title, hpa.title) as area_name,
			COALESCE(p.title, hp.title) as project_name,
			h.title as heading_name,
			GROUP_CONCAT(tag.title, ', ') as tags,
			CASE WHEN t.rt1_repeatingTemplate IS NOT NULL THEN 1 ELSE 0 END as is_repeating,
			t.todayIndex
		FROM TMTask t
		LEFT JOIN TMArea a ON t.area = a.uuid
		LEFT JOIN TMTask p ON t.project = p.uuid AND p.type = 1
		LEFT JOIN TMArea pa ON p.area = pa.uuid
		LEFT JOIN TMTask h ON t.heading = h.uuid
		LEFT JOIN TMTask hp ON h.project = hp.uuid AND hp.type = 1
		LEFT JOIN TMArea hpa ON hp.area = hpa.uuid
		LEFT JOIN TMTaskTag tt ON t.uuid = tt.tasks
		LEFT JOIN TMTag tag ON tt.tags = tag.uuid
		WHERE t.type = 0 AND t.trashed = 0 AND t.status = 0
			AND t.start = 0
		GROUP BY t.uuid
		ORDER BY t."index"
	`

	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query inbox: %w", err)
	}
	defer rows.Close()

	return scanTasks(rows)
}

// GetUpcomingTasks returns tasks scheduled for the future (start=2/Someday with future startDate)
func (db *ThingsDB) GetUpcomingTasks() ([]models.Task, error) {
	todayPacked := TodayPackedDate()
	query := fmt.Sprintf(`
		SELECT
			t.uuid,
			t.title,
			t.notes,
			t.status,
			t.type,
			t.creationDate,
			t.userModificationDate,
			t.startDate,
			t.deadline,
			t.stopDate,
			COALESCE(a.title, pa.title, hpa.title) as area_name,
			COALESCE(p.title, hp.title) as project_name,
			h.title as heading_name,
			GROUP_CONCAT(tag.title, ', ') as tags,
			CASE WHEN t.rt1_repeatingTemplate IS NOT NULL THEN 1 ELSE 0 END as is_repeating,
			t.todayIndex
		FROM TMTask t
		LEFT JOIN TMArea a ON t.area = a.uuid
		LEFT JOIN TMTask p ON t.project = p.uuid AND p.type = 1
		LEFT JOIN TMArea pa ON p.area = pa.uuid
		LEFT JOIN TMTask h ON t.heading = h.uuid
		LEFT JOIN TMTask hp ON h.project = hp.uuid AND hp.type = 1
		LEFT JOIN TMArea hpa ON hp.area = hpa.uuid
		LEFT JOIN TMTaskTag tt ON t.uuid = tt.tasks
		LEFT JOIN TMTag tag ON tt.tags = tag.uuid
		WHERE t.type = 0 AND t.trashed = 0 AND t.status = 0
			AND t.start = 2
			AND t.startDate IS NOT NULL AND t.startDate > %d
		GROUP BY t.uuid
		ORDER BY t.startDate, t."index"
	`, todayPacked)

	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query upcoming tasks: %w", err)
	}
	defer rows.Close()

	return scanTasks(rows)
}

// GetSomedayTasks returns someday tasks (start=2/Someday with no startDate)
func (db *ThingsDB) GetSomedayTasks() ([]models.Task, error) {
	query := `
		SELECT
			t.uuid,
			t.title,
			t.notes,
			t.status,
			t.type,
			t.creationDate,
			t.userModificationDate,
			t.startDate,
			t.deadline,
			t.stopDate,
			COALESCE(a.title, pa.title, hpa.title) as area_name,
			COALESCE(p.title, hp.title) as project_name,
			h.title as heading_name,
			GROUP_CONCAT(tag.title, ', ') as tags,
			CASE WHEN t.rt1_repeatingTemplate IS NOT NULL THEN 1 ELSE 0 END as is_repeating,
			t.todayIndex
		FROM TMTask t
		LEFT JOIN TMArea a ON t.area = a.uuid
		LEFT JOIN TMTask p ON t.project = p.uuid AND p.type = 1
		LEFT JOIN TMArea pa ON p.area = pa.uuid
		LEFT JOIN TMTask h ON t.heading = h.uuid
		LEFT JOIN TMTask hp ON h.project = hp.uuid AND hp.type = 1
		LEFT JOIN TMArea hpa ON hp.area = hpa.uuid
		LEFT JOIN TMTaskTag tt ON t.uuid = tt.tasks
		LEFT JOIN TMTag tag ON tt.tags = tag.uuid
		WHERE t.type = 0 AND t.trashed = 0 AND t.status = 0
			AND t.start = 2
			AND t.startDate IS NULL
		GROUP BY t.uuid
		ORDER BY t."index"
	`

	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query someday tasks: %w", err)
	}
	defer rows.Close()

	return scanTasks(rows)
}

// GetAuthToken retrieves the URL scheme authentication token
func (db *ThingsDB) GetAuthToken() (string, error) {
	var settings sql.NullString
	err := db.conn.QueryRow("SELECT plistData FROM TMSettings LIMIT 1").Scan(&settings)
	if err != nil {
		return "", fmt.Errorf("failed to query settings: %w", err)
	}

	// The auth token is stored in the plist data - for now return empty
	// TODO: Parse plist to extract uriSchemeAuthenticationToken
	return "", nil
}

// ResolveAreaUUID resolves a short UUID prefix to a full area UUID
func (db *ThingsDB) ResolveAreaUUID(prefix string) (string, error) {
	query := `SELECT uuid FROM TMArea WHERE uuid LIKE ? || '%'`
	rows, err := db.conn.Query(query, prefix)
	if err != nil {
		return "", fmt.Errorf("failed to query area: %w", err)
	}
	defer rows.Close()

	var uuids []string
	for rows.Next() {
		var uuid string
		if err := rows.Scan(&uuid); err != nil {
			return "", err
		}
		uuids = append(uuids, uuid)
	}

	if len(uuids) == 0 {
		return "", fmt.Errorf("area not found: %s", prefix)
	}
	if len(uuids) > 1 {
		return "", fmt.Errorf("ambiguous area prefix '%s' matches %d areas", prefix, len(uuids))
	}
	return uuids[0], nil
}

// ResolveTagUUID resolves a short UUID prefix to a full tag UUID
func (db *ThingsDB) ResolveTagUUID(prefix string) (string, error) {
	query := `SELECT uuid FROM TMTag WHERE uuid LIKE ? || '%'`
	rows, err := db.conn.Query(query, prefix)
	if err != nil {
		return "", fmt.Errorf("failed to query tag: %w", err)
	}
	defer rows.Close()

	var uuids []string
	for rows.Next() {
		var uuid string
		if err := rows.Scan(&uuid); err != nil {
			return "", err
		}
		uuids = append(uuids, uuid)
	}

	if len(uuids) == 0 {
		return "", fmt.Errorf("tag not found: %s", prefix)
	}
	if len(uuids) > 1 {
		return "", fmt.Errorf("ambiguous tag prefix '%s' matches %d tags", prefix, len(uuids))
	}
	return uuids[0], nil
}

// GetTag returns a single tag by UUID
func (db *ThingsDB) GetTag(uuid string) (*models.Tag, error) {
	query := `
		SELECT
			t.uuid,
			t.title,
			t.shortcut,
			COUNT(DISTINCT tt.tasks) as task_count
		FROM TMTag t
		LEFT JOIN TMTaskTag tt ON t.uuid = tt.tags
		LEFT JOIN TMTask task ON tt.tasks = task.uuid AND task.trashed = 0
		WHERE t.uuid = ?
		GROUP BY t.uuid
	`

	var tag models.Tag
	err := db.conn.QueryRow(query, uuid).Scan(&tag.UUID, &tag.Title, &tag.Shortcut, &tag.TaskCount)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("tag not found: %s", uuid)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query tag: %w", err)
	}

	return &tag, nil
}
