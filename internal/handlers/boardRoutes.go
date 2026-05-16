package handlers

import (
	"encoding/json"
	"go-backend/internal/models"
	"go-backend/internal/utils"
	"net/http"
	"strconv"
)

type BoardHandler struct {
	BoardModel *models.BoardModel
}

type BoardRequest struct {
	Title string `json:"title"`
}

func (h *BoardHandler) Create(w http.ResponseWriter, r *http.Request) {
	// 1. get userID from context and title from form
	userID := r.Context().Value("user_id_key").(int)

	var request BoardRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		utils.SendJSONError(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	// 2. validate title
	if request.Title == "" {
		utils.SendJSONError(w, "Title is required", http.StatusBadRequest)
		return
	}

	// 3. insert into table
	boardID, err := h.BoardModel.Create(userID, request.Title)
	if err != nil {
		utils.SendJSONError(w, "Failed to create board", http.StatusInternalServerError)
		return
	}

	// share the NEWS!
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	response := map[string]any{
		"message": "Board created",
		"board": map[string]any{
			"id":    boardID,
			"title": request.Title,
		},
	}
	json.NewEncoder(w).Encode(response)
}

func (h *BoardHandler) Update(w http.ResponseWriter, r *http.Request) {
	// 1. get userID from context and title from form
	userID := r.Context().Value("user_id_key").(int)

	boardID, err := strconv.Atoi(r.PathValue("boardID"))
	if err != nil {
		utils.SendJSONError(w, "Invalid Board ID", http.StatusBadRequest)
		return
	}

	var request BoardRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		utils.SendJSONError(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	// 2. validate title
	if request.Title == "" {
		utils.SendJSONError(w, "Title is required", http.StatusBadRequest)
		return
	}

	// 3. insert into table
	if err := h.BoardModel.Update(userID, boardID, request.Title); err != nil {
		utils.SendJSONError(w, "Failed to update board", http.StatusInternalServerError)
		return
	}

	// 4. Respond with JSON so React can update its state
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]any{
		"message": "Board updated",
		"board": map[string]any{
			"id":    boardID,
			"title": request.Title,
		},
	}
	json.NewEncoder(w).Encode(response)
}

func (h *BoardHandler) Delete(w http.ResponseWriter, r *http.Request) {
	// 1. get required variables
	userID := r.Context().Value("user_id_key").(int)

	boardID, err := strconv.Atoi(r.PathValue("boardID"))
	if err != nil {
		utils.SendJSONError(w, "Invalid Board ID", http.StatusBadRequest)
		return
	}

	// 2. delete from table
	if err := h.BoardModel.Delete(userID, boardID); err != nil {
		utils.SendJSONError(w, "Failed to delete board", http.StatusInternalServerError)
		return
	}

	// 3. Share the NEWS!
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]any{
		"message":          "Board deleted successfully",
		"deleted_board_id": boardID,
	}
	json.NewEncoder(w).Encode(response)
}

func (h *BoardHandler) GetBoard(w http.ResponseWriter, r *http.Request) {
	// 1. get userID from context and boardID from path
	userID := r.Context().Value("user_id_key").(int)
	boardID, err := strconv.Atoi(r.PathValue("boardID"))
	if err != nil {
		utils.SendJSONError(w, "Improper Board ID", http.StatusBadRequest)
		return
	}

	// 2. get the board
	board, err := h.BoardModel.GetBoard(boardID, userID)
	if err != nil {
		utils.SendJSONError(w, "Board Not Found", http.StatusNotFound)
		return
	}

	// 3. share the board
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(board)
	if err != nil {
		utils.SendJSONError(w, "Error formatting response", http.StatusInternalServerError)
		return
	}
}

// Add this anywhere in the file
func (h *BoardHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id_key").(int)

	boards, err := h.BoardModel.GetAllForUser(userID)
	if err != nil {
		utils.SendJSONError(w, "Failed to fetch boards", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(boards)
}
