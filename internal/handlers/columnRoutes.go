package handlers

import (
	"encoding/json"
	"go-backend/internal/models"
	"go-backend/internal/utils"
	"net/http"
	"strconv"
)

type ColumnHandler struct {
	ColumnModel *models.ColumnModel
}

type ColumnRequest struct {
	Title    string  `json:"title"`
	Position float64 `json:"position"`
}

func (h *ColumnHandler) Create(w http.ResponseWriter, r *http.Request) {
	// 1. get userID from context and title from form
	userID := r.Context().Value("user_id_key").(int)

	boardID, err := strconv.Atoi(r.PathValue("boardID"))
	if err != nil {
		utils.SendJSONError(w, "Invalid Board ID", http.StatusBadRequest)
		return
	}

	var request ColumnRequest
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
	columnID, err := h.ColumnModel.Create(userID, boardID, request.Title, request.Position)
	if err != nil {
		utils.SendJSONError(w, "Failed to create column", http.StatusInternalServerError)
		return
	}

	// 4. share the NEWS!
	// 4. share the NEWS!
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	response := map[string]any{
		"message": "Column created",
		"column": map[string]any{
			"id":       columnID,
			"title":    request.Title,
			"position": request.Position,
		},
	}
	json.NewEncoder(w).Encode(response)
}

func (h *ColumnHandler) Update(w http.ResponseWriter, r *http.Request) {
	// 1. get userID from context and title from form
	userID := r.Context().Value("user_id_key").(int)

	boardID, err := strconv.Atoi(r.PathValue("boardID"))
	if err != nil {
		utils.SendJSONError(w, "Invalid Board ID", http.StatusBadRequest)
		return
	}

	columnID, err := strconv.Atoi(r.PathValue("columnID"))
	if err != nil {
		utils.SendJSONError(w, "Invalid Column ID", http.StatusBadRequest)
		return
	}

	var request ColumnRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		utils.SendJSONError(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	// 2. validate title
	if request.Title == "" {
		utils.SendJSONError(w, "Title is required", http.StatusBadRequest)
		return
	}

	// 3. update table
	if err := h.ColumnModel.Update(userID, boardID, columnID, request.Title, request.Position); err != nil {
		utils.SendJSONError(w, "Failed to update column", http.StatusInternalServerError)
		return
	}

	// 4. share the NEWS!
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]any{
		"message": "Column updated",
		"column": map[string]any{
			"id":       columnID,
			"title":    request.Title,
			"position": request.Position,
		},
	}
	json.NewEncoder(w).Encode(response)
}

func (h *ColumnHandler) Delete(w http.ResponseWriter, r *http.Request) {
	// 1. get required variables
	userID := r.Context().Value("user_id_key").(int)

	boardID, err := strconv.Atoi(r.PathValue("boardID"))
	if err != nil {
		utils.SendJSONError(w, "Invalid Board ID", http.StatusBadRequest)
		return
	}

	columnID, err := strconv.Atoi(r.PathValue("columnID"))
	if err != nil {
		utils.SendJSONError(w, "Invalid Column ID", http.StatusBadRequest)
		return
	}

	// 2. delete from table
	if err = h.ColumnModel.Delete(userID, boardID, columnID); err != nil {
		utils.SendJSONError(w, "Failed to delete column", http.StatusInternalServerError)
		return
	}

	// 3. Share the NEWS!
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]any{
		"message":           "Column deleted successfully",
		"deleted_column_id": columnID,
	}
	json.NewEncoder(w).Encode(response)
}
