package models

import (
	"database/sql"
	"errors"
)

type Column struct {
	ID       int     `json:"id"`
	BoardID  int     `json:"board_id"`
	Title    string  `json:"title"`
	Position float64 `json:"position"`
	Cards    []Card  `json:"cards"`
}

type ColumnModel struct {
	DB *sql.DB
}

func (m *ColumnModel) Create(userID, boardID int, title string, position float64) (int, error) {
	authQuery := `SELECT 1 FROM boards WHERE user_id = ? AND id = ?;`

	var exists int
	err := m.DB.QueryRow(authQuery, userID, boardID).Scan(&exists)
	if err != nil {
		return 0, err
	}

	query := `
		INSERT INTO columns (board_id, title, position) 
		SELECT ?, ?, ? 
		WHERE EXISTS (
            SELECT 1 
            FROM boards 
            WHERE boards.user_id = ? 
              	AND boards.id = ?
        ) 
		RETURNING id;
	`

	var id int
	err = m.DB.QueryRow(query, boardID, title, position, userID, boardID).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (m *ColumnModel) Update(userID, boardID, columnID int, title string, position float64) error {
	query := `
		UPDATE columns
		SET title = ?, position = ?
		WHERE id = ? 
			AND board_id IN (
				SELECT id 
				FROM boards 
				WHERE id = ? AND user_id = ?
			);
	`

	result, err := m.DB.Exec(query, title, position, columnID, boardID, userID)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return errors.New("unauthorized or record not found")
	}

	return nil
}

func (m *ColumnModel) Delete(userID, boardID, columnID int) error {
	query := `
		DELETE FROM columns
		WHERE id = ? 
			AND board_id IN (
				SELECT id 
				FROM boards 
				WHERE id = ? AND user_id = ?
			);
	`

	result, err := m.DB.Exec(query, columnID, boardID, userID)
	if err != nil {
		return errors.New("Invalid delete request")
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return errors.New("unauthorized or record not found")
	}

	return nil
}
