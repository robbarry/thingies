package db

import (
	"database/sql"
	"fmt"
	"regexp"
	"strings"
)

// uuidPattern matches Things UUIDs: 22-character base62 alphanumeric strings
var uuidPattern = regexp.MustCompile(`^[0-9A-Za-z]{22}$`)

// LooksLikeUUID checks if string appears to be a full Things UUID (22 alphanumeric chars)
func LooksLikeUUID(s string) bool {
	return uuidPattern.MatchString(s)
}

// GetProjectUUIDByName looks up a project by name, returns UUID
// Returns error if no match or multiple matches found
func (db *ThingsDB) GetProjectUUIDByName(name string) (string, error) {
	query := `SELECT uuid FROM TMTask WHERE type = 1 AND trashed = 0 AND title = ?`
	rows, err := db.conn.Query(query, name)
	if err != nil {
		return "", fmt.Errorf("failed to query project: %w", err)
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
	if err := rows.Err(); err != nil {
		return "", fmt.Errorf("error iterating projects: %w", err)
	}

	if len(uuids) == 0 {
		return "", fmt.Errorf("project not found: %s", name)
	}
	if len(uuids) > 1 {
		return "", fmt.Errorf("multiple projects match '%s', use UUID", name)
	}
	return uuids[0], nil
}

// GetAreaUUIDByName looks up an area by name, returns UUID
// Returns error if no match or multiple matches found
func (db *ThingsDB) GetAreaUUIDByName(name string) (string, error) {
	query := `SELECT uuid FROM TMArea WHERE title = ?`
	rows, err := db.conn.Query(query, name)
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
	if err := rows.Err(); err != nil {
		return "", fmt.Errorf("error iterating areas: %w", err)
	}

	if len(uuids) == 0 {
		return "", fmt.Errorf("area not found: %s", name)
	}
	if len(uuids) > 1 {
		return "", fmt.Errorf("multiple areas match '%s', use UUID", name)
	}
	return uuids[0], nil
}

// ResolveProjectID returns UUID for a project given name, full UUID, or short UUID prefix.
// Tries in order: full UUID match, short UUID prefix, name lookup.
func (db *ThingsDB) ResolveProjectID(nameOrUUID string) (string, error) {
	if LooksLikeUUID(nameOrUUID) {
		// Verify the UUID exists
		var exists int
		err := db.conn.QueryRow(`SELECT 1 FROM TMTask WHERE uuid = ? AND type = 1 AND trashed = 0`, nameOrUUID).Scan(&exists)
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("project not found: %s", nameOrUUID)
		}
		if err != nil {
			return "", fmt.Errorf("failed to query project: %w", err)
		}
		return nameOrUUID, nil
	}

	// Try short UUID prefix resolution first
	resolved, err := db.ResolveProjectUUID(nameOrUUID)
	if err == nil {
		return resolved, nil
	}
	// Only fall through to name lookup for "not found" errors;
	// surface ambiguous prefix and DB errors immediately
	if !strings.Contains(err.Error(), "not found") {
		return "", err
	}

	// Fall back to name lookup
	return db.GetProjectUUIDByName(nameOrUUID)
}

// ResolveAreaID returns UUID for an area given name, full UUID, or short UUID prefix.
// Tries in order: full UUID match, short UUID prefix, name lookup.
func (db *ThingsDB) ResolveAreaID(nameOrUUID string) (string, error) {
	if LooksLikeUUID(nameOrUUID) {
		// Verify the UUID exists
		var exists int
		err := db.conn.QueryRow(`SELECT 1 FROM TMArea WHERE uuid = ?`, nameOrUUID).Scan(&exists)
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("area not found: %s", nameOrUUID)
		}
		if err != nil {
			return "", fmt.Errorf("failed to query area: %w", err)
		}
		return nameOrUUID, nil
	}

	// Try short UUID prefix resolution first
	resolved, err := db.ResolveAreaUUID(nameOrUUID)
	if err == nil {
		return resolved, nil
	}
	// Only fall through to name lookup for "not found" errors;
	// surface ambiguous prefix and DB errors immediately
	if !strings.Contains(err.Error(), "not found") {
		return "", err
	}

	// Fall back to name lookup
	return db.GetAreaUUIDByName(nameOrUUID)
}
