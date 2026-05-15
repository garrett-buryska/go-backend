package models

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
)

type Board struct {
	ID      int      `json:"id"`
	UserID  int      `json:"user_id"`
	Title   string   `json:"title"`
	Columns []Column `json:"columns"`
}

type BoardModel struct {
	DB *sql.DB
}

func (m *BoardModel) Create(userID int, title string) (int, error) {
	query := `INSERT INTO boards (user_id, title) VALUES (?, ?) RETURNING id;`

	var id int
	err := m.DB.QueryRow(query, userID, title).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (m *BoardModel) Update(userID, boardID int, title string) error {
	query := `UPDATE boards SET title = ? WHERE user_id = ? AND id = ?;`

	result, err := m.DB.Exec(query, title, userID, boardID)
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

func (m *BoardModel) Delete(userID, boardID int) error {
	query := `DELETE from boards WHERE user_id = ? AND id = ?`

	result, err := m.DB.Exec(query, userID, boardID)
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

func (m *BoardModel) GetBoard(boardID, userID int) (*Board, error) {
	query := `
		SELECT boards.id, boards.title,
		(
			SELECT json_group_array(
				json_object(
					'id', columns.id,
					'title', columns.title,
					'position', columns.position,
					'cards', (
						SELECT json_group_array(
							json_object(
								'id', cards.id,
								'title', cards.title,
								'position', cards.position,
								'body', cards.body
							)
						)
						FROM cards
						WHERE cards.column_id = columns.id
					)
				)
			)
			FROM columns
			WHERE columns.board_id = boards.id
		) AS columns
		FROM boards
		WHERE boards.id = ? AND boards.user_id = ?;
	`

	var board Board
	var columnsRaw []byte

	err := m.DB.QueryRow(query, boardID, userID).Scan(&board.ID, &board.Title, &columnsRaw)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("models: no matching record found")
		}
		return nil, err
	}

	if err := json.Unmarshal(columnsRaw, &board.Columns); err != nil {
		return nil, fmt.Errorf("failed to parse columns: %w", err)
	}

	return &board, nil
}
