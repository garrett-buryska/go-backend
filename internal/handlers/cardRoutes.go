package handlers

import (
	"fmt"
	"go-backend/internal/models"
	"net/http"
	"strconv"
)

type CardHandler struct {
	CardModel *models.CardModel
}

func (h *CardHandler) Create(w http.ResponseWriter, r *http.Request) {
	// 1. get userID from context and title from form
	userID := r.Context().Value("user_id_key").(int)
	title := r.FormValue("title")
	position, err := strconv.ParseFloat(r.FormValue("position"), 64)
	if err != nil {
		http.Error(w, "Improper position", http.StatusBadRequest)
		return
	}
	boardID, err := strconv.Atoi(r.PathValue("boardID"))
	if err != nil {
		http.Error(w, "Improper Board ID", http.StatusBadRequest)
		return
	}
	columnID, err := strconv.Atoi(r.PathValue("columnID"))
	if err != nil {
		http.Error(w, "Improper Column ID", http.StatusBadRequest)
		return
	}

	// 2. validate title
	if title == "" {
		http.Error(w, "Invalid title", http.StatusBadRequest)
		return
	}

	// 3. insert into table
	cardID, err := h.CardModel.Create(userID, boardID, columnID, title, position)
	if err != nil {
		http.Error(w, "Invalid card entry", http.StatusBadRequest)
		return
	}

	// 4. share the NEWS!
	w.Header().Set("Location", fmt.Sprintf("/%d/%d/%d", boardID, columnID, cardID))
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Card created successfully!"))
}

func (h *CardHandler) Update(w http.ResponseWriter, r *http.Request) {
	// 1. get userID from context and title from form
	userID := r.Context().Value("user_id_key").(int)
	title := r.FormValue("title")
	body := r.FormValue("body")

	position, err := strconv.ParseFloat(r.FormValue("position"), 64)
	if err != nil {
		http.Error(w, "Improper position", http.StatusBadRequest)
		return
	}

	boardID, err := strconv.Atoi(r.PathValue("boardID"))
	if err != nil {
		http.Error(w, "Improper Board ID", http.StatusBadRequest)
		return
	}

	columnID, err := strconv.Atoi(r.PathValue("columnID"))
	if err != nil {
		http.Error(w, "Improper Column ID", http.StatusBadRequest)
		return
	}

	cardID, err := strconv.Atoi(r.PathValue("cardID"))
	if err != nil {
		http.Error(w, "Improper Card ID", http.StatusBadRequest)
	}

	// 2. validate title
	if title == "" {
		http.Error(w, "Invalid title", http.StatusBadRequest)
		return
	}

	// 3. put into table
	if err := h.CardModel.Update(userID, boardID, columnID, cardID, title, body, position); err != nil {
		http.Error(w, "Invalid card entry", http.StatusBadRequest)
		return
	}

	// 4. share the NEWS!
	w.Header().Set("Location", fmt.Sprintf("/%d/%d/%d", boardID, columnID, cardID))
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte("Card updated successfully!"))
}

func (h *CardHandler) Delete(w http.ResponseWriter, r *http.Request) {
	// 1. get required variables
	userID := r.Context().Value("user_id_key").(int)
	boardID, err := strconv.Atoi(r.PathValue("boardID"))
	if err != nil {
		http.Error(w, "Improper Board ID", http.StatusBadRequest)
		return
	}

	columnID, err := strconv.Atoi(r.PathValue("columnID"))
	if err != nil {
		http.Error(w, "Improper Column ID", http.StatusBadRequest)
		return
	}

	cardID, err := strconv.Atoi(r.PathValue("cardID"))
	if err != nil {
		http.Error(w, "Improper Card ID", http.StatusBadRequest)
	}

	// 2. delete from table
	if err := h.CardModel.Delete(userID, boardID, columnID, cardID); err != nil {
		http.Error(w, "Improper delete request", http.StatusBadRequest)
		return
	}

	// 3. Share the NEWS!
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte("Card deleted successfully!"))
}
