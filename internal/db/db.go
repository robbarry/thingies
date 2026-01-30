package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

// ThingsDB provides read-only access to the Things 3 database
type ThingsDB struct {
	conn *sql.DB
	path string
}

// DefaultDBPath returns the default Things 3 database path
func DefaultDBPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	basePath := filepath.Join(home, "Library", "Group Containers", "JLMPQHK86H.com.culturedcode.ThingsMac")

	// Find ThingsData-* directory
	entries, err := os.ReadDir(basePath)
	if err != nil {
		return "", fmt.Errorf("failed to read Things container: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), "ThingsData-") {
			dbPath := filepath.Join(basePath, entry.Name(), "Things Database.thingsdatabase", "main.sqlite")
			if _, err := os.Stat(dbPath); err == nil {
				return dbPath, nil
			}
		}
	}

	return "", fmt.Errorf("Things 3 database not found in %s", basePath)
}

// Open opens a read-only connection to the Things database
func Open(dbPath string) (*ThingsDB, error) {
	if dbPath == "" {
		var err error
		dbPath, err = DefaultDBPath()
		if err != nil {
			return nil, err
		}
	}

	// Open in read-only mode
	connStr := fmt.Sprintf("file:%s?mode=ro", dbPath)
	conn, err := sql.Open("sqlite", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	if err := conn.Ping(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return &ThingsDB{
		conn: conn,
		path: dbPath,
	}, nil
}

// Close closes the database connection
func (db *ThingsDB) Close() error {
	if db.conn != nil {
		return db.conn.Close()
	}
	return nil
}

// Conn returns the underlying database connection
func (db *ThingsDB) Conn() *sql.DB {
	return db.conn
}

// Path returns the database path
func (db *ThingsDB) Path() string {
	return db.path
}

// TodayPackedDate returns today's date in Things packed format
// Things packs dates as: year << 16 | month << 12 | day << 7
func TodayPackedDate() int {
	return DateToPackedInt(time.Now())
}

// DateToPackedInt converts a time.Time to Things packed date format
// Things packs dates as: year << 16 | month << 12 | day << 7
func DateToPackedInt(t time.Time) int {
	return (t.Year() << 16) | (int(t.Month()) << 12) | (t.Day() << 7)
}
