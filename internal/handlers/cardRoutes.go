package handlers

import (
	"encoding/json"
	"go-backend/internal/models"
	"go-backend/internal/utils"
	"net/http"
	"strconv"
)

type CardHandler struct {
	CardModel *models.CardModel
}

type CardRequest struct {
	Title    string  `json:"title"`
	Position float64 `json:"position"`
	Body     string  `json:"body,omitempty"`
}

func (h *CardHandler) Create(w http.ResponseWriter, r *http.Request) {
	// 1. get required resources
	userID := r.Context().Value("user_id_key").(int)

	columnID, err := strconv.Atoi(r.PathValue("columnID"))
	if err != nil {
		utils.SendJSONError(w, "Invalid Column ID", http.StatusBadRequest)
		return
	}

	var request CardRequest
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
	cardID, err := h.CardModel.Create(userID, columnID, request.Title, request.Position)
	if err != nil {
		utils.SendJSONError(w, "Failed to create card", http.StatusInternalServerError)
		return
	}

	// 4. share the NEWS!
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	response := map[string]any{
		"message": "Card created",
		"card": map[string]any{
			"id":       cardID,
			"title":    request.Title,
			"position": request.Position,
		},
	}
	json.NewEncoder(w).Encode(response)
}

func (h *CardHandler) Update(w http.ResponseWriter, r *http.Request) {
	// 1. get required resources
	userID := r.Context().Value("user_id_key").(int)

	columnID, err := strconv.Atoi(r.PathValue("columnID"))
	if err != nil {
		utils.SendJSONError(w, "Invalid Column ID", http.StatusBadRequest)
		return
	}

	cardID, err := strconv.Atoi(r.PathValue("cardID"))
	if err != nil {
		utils.SendJSONError(w, "Invalid Card ID", http.StatusBadRequest)
		return
	}

	var request CardRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		utils.SendJSONError(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	// 2. validate title
	if request.Title == "" {
		utils.SendJSONError(w, "Title is required", http.StatusBadRequest)
		return
	}

	// 3. put into table
	if err := h.CardModel.Update(userID, columnID, cardID, request.Title, request.Body, request.Position); err != nil {
		utils.SendJSONError(w, "Failed to update card", http.StatusInternalServerError)
		return
	}

	// 4. Respond with JSON so React can update its state
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]any{
		"message": "Card updated",
		"card": map[string]any{
			"id":       cardID,
			"title":    request.Title,
			"position": request.Position,
			"body":     request.Body,
		},
	}
	json.NewEncoder(w).Encode(response)
}

func (h *CardHandler) Delete(w http.ResponseWriter, r *http.Request) {
	// 1. get required resources
	userID := r.Context().Value("user_id_key").(int)

	columnID, err := strconv.Atoi(r.PathValue("columnID"))
	if err != nil {
		utils.SendJSONError(w, "Invalid Column ID", http.StatusBadRequest)
		return
	}

	cardID, err := strconv.Atoi(r.PathValue("cardID"))
	if err != nil {
		utils.SendJSONError(w, "Invalid Card ID", http.StatusBadRequest)
		return
	}

	// 2. delete from table
	if err := h.CardModel.Delete(userID, columnID, cardID); err != nil {
		utils.SendJSONError(w, "Failed to delete card", http.StatusInternalServerError)
		return
	}

	// 3. Share the NEWS!
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]any{
		"message":         "Card deleted successfully",
		"deleted_card_id": cardID,
	}
	json.NewEncoder(w).Encode(response)
}
