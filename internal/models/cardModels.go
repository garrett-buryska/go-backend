package models

import (
	"database/sql"
	"errors"
)

type Card struct {
	ID       int     `json:"id"`
	ColumnID int     `json:"column_id"`
	Title    string  `json:"title"`
	Body     string  `json:"body"`
	Position float64 `json:"position"`
}

type CardModel struct {
	DB *sql.DB
}

func (m *CardModel) Create(userID, boardID, columnID int, title string, position float64) (int, error) {
	query := `
		INSERT INTO cards (column_id, title, position) 
		SELECT ?, ?, ?
		WHERE EXISTS (
            SELECT 1 
            FROM columns 
            INNER JOIN boards ON columns.board_id = boards.id 
            WHERE boards.user_id = ? 
              	AND boards.id = ? 
              	AND columns.id = ?
        ) 
		RETURNING id;
	`

	var id int
	err := m.DB.QueryRow(query, columnID, title, position, userID, boardID, columnID).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (m *CardModel) Update(userID, boardID, columnID, cardID int, title, body string, position float64) error {
	query := `
		UPDATE cards
		SET title = ?, position = ?, body = ?
		WHERE id = ? 
			AND column_id IN (
				SELECT columns.id 
				FROM columns
				INNER JOIN boards ON columns.board_id = boards.id
				WHERE boards.user_id = ? AND boards.id = ? AND columns.id = ?
			);
	`

	result, err := m.DB.Exec(query, title, position, body, cardID, userID, boardID, columnID)
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

func (m *CardModel) Delete(userID, boardID, columnID, cardID int) error {
	query := `
		DELETE FROM cards
		WHERE id = ? 
			AND column_id IN (
				SELECT columns.id 
				FROM columns 
				INNER JOIN boards ON columns.board_id = boards.id
				WHERE boards.user_id = ? AND boards.id = ? AND columns.id = ?
			);
	`

	result, err := m.DB.Exec(query, cardID, userID, boardID, columnID)
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
