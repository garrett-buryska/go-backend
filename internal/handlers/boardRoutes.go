package handlers

import (
	"encoding/json"
	"fmt"
	"go-backend/internal/models"
	"net/http"
	"strconv"
)

type BoardHandler struct {
	BoardModel *models.BoardModel
}

func (h *BoardHandler) Create(w http.ResponseWriter, r *http.Request) {
	// 1. get userID from context and title from form
	userID := r.Context().Value("user_id_key").(int)
	title := r.FormValue("title")

	// 2. validate title
	if title == "" {
		http.Error(w, "Invalid title", http.StatusBadRequest)
		return
	}

	// 3. insert into table
	boardID, err := h.BoardModel.Create(userID, title)
	if err != nil {
		http.Error(w, "Invalid board entry", http.StatusBadRequest)
		return
	}

	// share the NEWS!
	w.Header().Set("Location", fmt.Sprintf("/%d", boardID))
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Board created successfully!"))
}

func (h *BoardHandler) Update(w http.ResponseWriter, r *http.Request) {
	// 1. get userID from context and title from form
	userID := r.Context().Value("user_id_key").(int)
	title := r.FormValue("title")

	boardID, err := strconv.Atoi(r.PathValue("boardID"))
	if err != nil {
		http.Error(w, "Improper Board ID", http.StatusBadRequest)
		return
	}

	// 2. validate title
	if title == "" {
		http.Error(w, "Invalid title", http.StatusBadRequest)
		return
	}

	// 3. insert into table
	if err := h.BoardModel.Update(userID, boardID, title); err != nil {
		http.Error(w, "Invalid board entry", http.StatusBadRequest)
		return
	}

	// share the NEWS!
	w.Header().Set("Location", fmt.Sprintf("/%d", boardID))
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte("Board created successfully!"))
}

func (h *BoardHandler) Delete(w http.ResponseWriter, r *http.Request) {
	// 1. get required variables
	userID := r.Context().Value("user_id_key").(int)
	boardID, err := strconv.Atoi(r.PathValue("boardID"))
	if err != nil {
		http.Error(w, "Improper Board ID", http.StatusBadRequest)
		return
	}

	// 2. delete from table
	if err := h.BoardModel.Delete(userID, boardID); err != nil {
		http.Error(w, "Invalid board delete request", http.StatusBadRequest)
		return
	}

	// share the NEWS!
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte("Board deleted successfully!"))
}

func (h *BoardHandler) GetBoard(w http.ResponseWriter, r *http.Request) {
	// 1. get userID from context and boardID from path
	userID := r.Context().Value("user_id_key").(int)
	boardID, err := strconv.Atoi(r.PathValue("boardID"))
	if err != nil {
		http.Error(w, "Improper Board ID", http.StatusBadRequest)
		return
	}

	// 2. get the board
	board, err := h.BoardModel.GetBoard(boardID, userID)
	if err != nil {
		http.Error(w, "Board Not Found", http.StatusNotFound)
		return
	}

	// 3. share the board
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(board)
	if err != nil {
		// If encoding fails (rare, but good to handle)
		http.Error(w, "Error formatting response", http.StatusInternalServerError)
		return
	}
}
