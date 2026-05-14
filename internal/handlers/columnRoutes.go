package handlers

import (
	"fmt"
	"go-backend/internal/models"
	"net/http"
	"strconv"
)

type ColumnHandler struct {
	ColumnModel *models.ColumnModel
}

func (h *ColumnHandler) Create(w http.ResponseWriter, r *http.Request) {
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

	// 2. validate title
	if title == "" {
		http.Error(w, "Invalid title", http.StatusBadRequest)
		return
	}

	// 3. insert into table
	columnID, err := h.ColumnModel.Create(userID, boardID, title, position)
	if err != nil {
		http.Error(w, "Invalid column entry", http.StatusBadRequest)
		return
	}

	// 4. share the NEWS!
	w.Header().Set("Location", fmt.Sprintf("/%d/%d", boardID, columnID))
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Column created successfully!"))
}

func (h *ColumnHandler) Update(w http.ResponseWriter, r *http.Request) {
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
		http.Error(w, "Improper Board ID", http.StatusBadRequest)
		return
	}

	// 2. validate title
	if title == "" {
		http.Error(w, "Invalid title", http.StatusBadRequest)
		return
	}

	// 3. update table
	if err := h.ColumnModel.Update(userID, boardID, columnID, title, position); err != nil {
		http.Error(w, "Invalid column entry", http.StatusBadRequest)
		return
	}

	// 4. share the NEWS!
	w.Header().Set("Location", fmt.Sprintf("/%d/%d", boardID, columnID))
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte("Column updated successfully!"))
}

func (h *ColumnHandler) Delete(w http.ResponseWriter, r *http.Request) {
	// 1. get required variables
	userID := r.Context().Value("user_id_key").(int)
	boardID, err := strconv.Atoi(r.PathValue("boardID"))
	if err != nil {
		http.Error(w, "Improper Board ID", http.StatusBadRequest)
		return
	}
	columnID, err := strconv.Atoi(r.PathValue("columnID"))
	if err != nil {
		http.Error(w, "Improper Board ID", http.StatusBadRequest)
		return
	}

	// 2. delete from table
	if err = h.ColumnModel.Delete(userID, boardID, columnID); err != nil {
		http.Error(w, "Improper delete request", http.StatusBadRequest)
		return
	}

	// 3. Share the NEWS!
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte("Column deleted successfully!"))
}
