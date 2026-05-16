package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "modernc.org/sqlite"
)

// InitDB opens a database connection, applies performance settings,
// and ensures required tables exist.
func InitDB(dbPath string) (*sql.DB, error) {
	// 1. Open the database connection
	// Go doesn't actually connect here, it just validates the arguments.
	// Format the DSN with query parameters for the go-sqlite3 driver
	dsn := fmt.Sprintf("%s?_journal=WAL&_fk=1&_busy_timeout=5000&_sync=NORMAL", dbPath)

	db, err := sql.Open("sqlite", dsn)
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
	db.SetMaxOpenConns(25) // Allow concurrent reads
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Hour)

	// 4. Run migrations (create tables if they don't exist)
	err = createTables(db)
	if err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	log.Println("Database connection established and configured.")
	return db, nil
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
		session_token TEXT UNIQUE,
		csrf_token TEXT
	);
	
	CREATE TABLE IF NOT EXISTS boards (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		title TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
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
		body TEXT NOT NULL,
		position REAL NOT NULL,
		FOREIGN KEY (column_id) REFERENCES columns(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_boards_user_id ON boards(user_id);
	CREATE INDEX IF NOT EXISTS idx_columns_board_id ON columns(board_id);
	CREATE INDEX IF NOT EXISTS idx_cards_column_id ON cards(column_id);

	CREATE TRIGGER IF NOT EXISTS update_board_timestamp
	AFTER UPDATE ON boards
	FOR EACH ROW
	BEGIN
		UPDATE boards SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
	END;

	CREATE TRIGGER IF NOT EXISTS update_board_on_column_insert
	AFTER INSERT ON columns
	FOR EACH ROW
	BEGIN
		UPDATE boards SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.board_id;
	END;

	CREATE TRIGGER IF NOT EXISTS update_board_on_column_update
	AFTER UPDATE ON columns
	FOR EACH ROW
	BEGIN
		UPDATE boards SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.board_id;
	END;

	CREATE TRIGGER IF NOT EXISTS update_board_on_column_delete
	AFTER DELETE ON columns
	FOR EACH ROW
	BEGIN
		UPDATE boards SET updated_at = CURRENT_TIMESTAMP WHERE id = OLD.board_id;
	END;

	CREATE TRIGGER IF NOT EXISTS update_board_on_card_insert
	AFTER INSERT ON cards
	FOR EACH ROW
	BEGIN
		UPDATE boards 
		SET updated_at = CURRENT_TIMESTAMP 
		WHERE id = (SELECT board_id FROM columns WHERE id = NEW.column_id);
	END;

	CREATE TRIGGER IF NOT EXISTS update_board_on_card_update
	AFTER UPDATE ON cards
	FOR EACH ROW
	BEGIN
		UPDATE boards 
		SET updated_at = CURRENT_TIMESTAMP 
		WHERE id = (SELECT board_id FROM columns WHERE id = NEW.column_id);
	END;

	CREATE TRIGGER IF NOT EXISTS update_board_on_card_delete
	AFTER DELETE ON cards
	FOR EACH ROW
	BEGIN
		UPDATE boards 
		SET updated_at = CURRENT_TIMESTAMP 
		WHERE id = (SELECT board_id FROM columns WHERE id = OLD.column_id);
	END;
	`

	_, err := db.Exec(query)
	return err
}
