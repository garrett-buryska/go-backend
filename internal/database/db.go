package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// InitDB opens a database connection, applies performance settings,
// and ensures required tables exist.
func InitDB(dbPath string) (*sql.DB, error) {
	// 1. Open the database connection
	// Go doesn't actually connect here, it just validates the arguments.
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// 2. Ping to verify the file is accessible and the connection works
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// 3. Configure Connection Pooling (CRITICAL FOR SQLITE)
	// SQLite handles concurrent reads well (with WAL), but concurrent writes can lock the database.
	// Limiting open connections prevents "database is locked" errors during high traffic.
	db.SetMaxOpenConns(1) // Only allow one operation at a time to write safely
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(time.Hour)

	// 4. Apply SQLite-specific performance PRAGMAs
	err = setupPragmas(db)
	if err != nil {
		return nil, fmt.Errorf("failed to set pragmas: %w", err)
	}

	// 5. Run migrations (create tables if they don't exist)
	err = createTables(db)
	if err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	log.Println("Database connection established and configured.")
	return db, nil
}

func setupPragmas(db *sql.DB) error {
	pragmas := []string{
		"PRAGMA journal_mode = WAL;",   // Write-Ahead Logging: drastically improves concurrency
		"PRAGMA synchronous = NORMAL;", // Safe enough for WAL, much faster than FULL
		"PRAGMA foreign_keys = ON;",    // Enforce foreign key relationships
		"PRAGMA busy_timeout = 5000;",  // Wait up to 5 seconds if the DB is locked before erroring
	}

	for _, pragma := range pragmas {
		_, err := db.Exec(pragma)
		if err != nil {
			return err
		}
	}
	return nil
}

// createTables ensures our database schema exists
func createTables(db *sql.DB) error {
	// We use "IF NOT EXISTS" so this is safe to run every time the app starts.
	// This matches the User struct in internal/models/user.go
	query := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL UNIQUE,
		hashed_password TEXT NOT NULL,
		email TEXT NOT NULL,
		session_token TEXT UNIQUE,
		csrf_token TEXT
	);
	
	CREATE TABLE IF NOT EXISTS boards (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		title TEXT NOT NULL,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);
	
	CREATE TABLE IF NOT EXISTS columns (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		board_id INTEGER NOT NULL,
		title TEXT NOT NULL,
		position REAL NOT NULL,
		FOREIGN KEY (board_id) REFERENCES boards(id) ON DELETE CASCADE
	);
	
	CREATE TABLE IF NOT EXISTS cards (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		column_id INTEGER NOT NULL,
		title TEXT NOT NULL,
		body TEXT,
		position REAL NOT NULL,
		FOREIGN KEY (column_id) REFERENCES columns(id) ON DELETE CASCADE
	);
	`

	_, err := db.Exec(query)
	return err
}
