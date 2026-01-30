package db

import (
	"fmt"
	"regexp"
)

// uuidPattern matches Things UUIDs: XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX (36 chars with dashes)
var uuidPattern = regexp.MustCompile(`^[0-9A-Fa-f]{8}-[0-9A-Fa-f]{4}-[0-9A-Fa-f]{4}-[0-9A-Fa-f]{4}-[0-9A-Fa-f]{12}$`)

// LooksLikeUUID checks if string appears to be a UUID
// Things UUIDs are 36 chars with dashes: XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX
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

	if len(uuids) == 0 {
		return "", fmt.Errorf("area not found: %s", name)
	}
	if len(uuids) > 1 {
		return "", fmt.Errorf("multiple areas match '%s', use UUID", name)
	}
	return uuids[0], nil
}

// ResolveProjectID returns UUID for a project given name or UUID
// If the input looks like a UUID, it's returned as-is (after validation)
// Otherwise, it's looked up by name
func (db *ThingsDB) ResolveProjectID(nameOrUUID string) (string, error) {
	if LooksLikeUUID(nameOrUUID) {
		// Verify the UUID exists
		var exists int
		err := db.conn.QueryRow(`SELECT 1 FROM TMTask WHERE uuid = ? AND type = 1 AND trashed = 0`, nameOrUUID).Scan(&exists)
		if err != nil {
			return "", fmt.Errorf("project not found: %s", nameOrUUID)
		}
		return nameOrUUID, nil
	}
	return db.GetProjectUUIDByName(nameOrUUID)
}

// ResolveAreaID returns UUID for an area given name or UUID
// If the input looks like a UUID, it's returned as-is (after validation)
// Otherwise, it's looked up by name
func (db *ThingsDB) ResolveAreaID(nameOrUUID string) (string, error) {
	if LooksLikeUUID(nameOrUUID) {
		// Verify the UUID exists
		var exists int
		err := db.conn.QueryRow(`SELECT 1 FROM TMArea WHERE uuid = ?`, nameOrUUID).Scan(&exists)
		if err != nil {
			return "", fmt.Errorf("area not found: %s", nameOrUUID)
		}
		return nameOrUUID, nil
	}
	return db.GetAreaUUIDByName(nameOrUUID)
}
