package models

import (
	"database/sql"
	"errors"
)

// User represents the data structure of our database row
type User struct {
	ID             int    `json:"id"`
	Username       string `json:"username"`
	HashedPassword string `json:"-"` // The "-" prevents the password from EVER leaking in JSON
	Email          string `json:"email"`
	SessionToken   string `json:"-"` // sql.NullString handles NULL values in the DB (like when logged out)
	CSRFToken      string `json:"-"`
}

// UserModel wraps the database connection pool
type UserModel struct {
	DB *sql.DB
}

func (m *UserModel) Insert(username, hashedPassword string, email string) (int, error) {
	query := `INSERT INTO users (username, hashed_password, email) VALUES (?, ?, ?) RETURNING id;`

	var id int
	err := m.DB.QueryRow(query, username, hashedPassword, email).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (m *UserModel) GetByUsername(username string) (*User, error) {
	query := `SELECT id, username, hashed_password FROM users WHERE username = ?`

	var user User
	err := m.DB.QueryRow(query, username).Scan(&user.ID, &user.Username, &user.HashedPassword)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("models: no matching record found")
		}
		return nil, err
	}

	return &user, nil
}

func (m *UserModel) UpdateSession(userID int, sessionToken *string, csrfToken *string) error {
	query := `UPDATE users SET session_token = ?, csrf_token = ? WHERE id = ?`

	_, err := m.DB.Exec(query, sessionToken, csrfToken, userID)
	return err
}

func (m *UserModel) GetBySession(sessionToken string, csrfToken string) (*User, error) {
	query := `SELECT id FROM users WHERE session_token = ? AND csrf_token = ?`

	var user User
	err := m.DB.QueryRow(query, sessionToken, csrfToken).Scan(&user.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("models: no matching record found")
		}
		return nil, err
	}

	return &user, nil
}
