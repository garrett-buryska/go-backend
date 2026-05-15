package handlers

import (
	"context"
	"encoding/json"
	"go-backend/internal/models"
	"go-backend/internal/utils"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type UserHandler struct {
	UserModel *models.UserModel
}

// Create structs to handle the incoming JSON payloads
type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	// 1. Extract values from JSON body
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendJSONError(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	// 2. Validate password
	if len(req.Username) < 8 || len(req.Password) < 8 {
		utils.SendJSONError(w, "Username and password must be at least 8 characters", http.StatusNotAcceptable)
		return
	}

	// 3. Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		utils.SendJSONError(w, "Server error hashing password", http.StatusInternalServerError)
		return
	}

	// 4. Save to the database
	if _, err = h.UserModel.Insert(req.Username, string(hashedPassword)); err != nil {
		utils.SendJSONError(w, "User already exists", http.StatusConflict)
		return
	}

	// 5. Share the NEWS!
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "User registered successfully!"})
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	// 1. Extract values from JSON
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendJSONError(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	// 2. Get user by Username
	user, err := h.UserModel.GetByUsername(req.Username)
	if err != nil {
		utils.SendJSONError(w, "Invalid username/password", http.StatusUnauthorized)
		return
	}

	// 3. Validate Password
	err = bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(req.Password))
	if err != nil {
		utils.SendJSONError(w, "Invalid username/password", http.StatusUnauthorized)
		return
	}

	// 4. Generate SessionToken
	sessionToken, err := utils.GenerateToken(32)
	if err != nil {
		utils.SendJSONError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// 5. Generate CSRFToken
	csrfToken, err := utils.GenerateToken(32)
	if err != nil {
		utils.SendJSONError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// 6. Set tokens
	err = h.UserModel.UpdateSession(user.ID, &sessionToken, &csrfToken)
	if err != nil {
		utils.SendJSONError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// 7. Set cookies (HttpOnly keeps session safe from XSS, CSRF cookie is read by React)
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    sessionToken,
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		Path:     "/",
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "csrf_token",
		Value:    csrfToken,
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: false,
		Path:     "/",
	})

	// 8. Share the NEWS!
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Returning the username/ID is helpful for React to update its UI state
	response := map[string]any{
		"message":  "Logged in successfully!",
		"username": user.Username,
	}
	json.NewEncoder(w).Encode(response)
}

func (h *UserHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// 1. Get sessionToken
	sessionToken, err := r.Cookie("session_token")
	if err != nil || sessionToken.Value == "" {
		utils.SendJSONError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// 2. Get csrf
	csrfToken := r.Header.Get("X-CSRF-Token")

	// 3. Get user
	user, err := h.UserModel.GetBySession(sessionToken.Value, csrfToken)
	if err != nil {
		utils.SendJSONError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// 4. Purge tokens
	err = h.UserModel.UpdateSession(user.ID, nil, nil)
	if err != nil {
		utils.SendJSONError(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// 5. Purge Cookies
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		HttpOnly: true,
		Path:     "/",
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "csrf_token",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		HttpOnly: false,
		Path:     "/",
	})

	// 6. Share the NEWS!
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Logged out successfully!"})
}

func (h *UserHandler) RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. Get session token from cookie
		sessionCookie, err := r.Cookie("session_token")
		if err != nil {
			utils.SendJSONError(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// 2. get csrf token
		csrfToken := r.Header.Get("X-CSRF-Token")

		// 3. Verify tokens against the database
		user, err := h.UserModel.GetBySession(sessionCookie.Value, csrfToken)
		if err != nil {
			utils.SendJSONError(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// 4. proceed! with user_id context!
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), "user_id_key", user.ID)))
	}
}

func (h *UserHandler) Me(w http.ResponseWriter, r *http.Request) {
	// 1. Get sessionToken
	sessionToken, err := r.Cookie("session_token")
	if err != nil || sessionToken.Value == "" {
		utils.SendJSONError(w, "Not logged in1", http.StatusUnauthorized)
		return
	}

	// 2. Get csrf
	csrfToken := r.Header.Get("X-CSRF-Token")

	// 3. Get user
	user, err := h.UserModel.GetBySession(sessionToken.Value, csrfToken)
	if err != nil {
		utils.SendJSONError(w, "Not logged in2", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"username": user.Username})
}
