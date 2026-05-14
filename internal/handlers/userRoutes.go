package handlers

import (
	"context"
	"go-backend/internal/models"
	"go-backend/internal/utils"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type UserHandler struct {
	UserModel *models.UserModel
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	// 1. Extract values
	username := r.FormValue("username")
	password := r.FormValue("password")
	email := r.FormValue("email")

	// 2. Validate password
	if len(username) < 8 || len(password) < 8 {
		http.Error(w, "Username and password must be at least 8 characters", http.StatusNotAcceptable)
		return
	}

	// 3. Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		http.Error(w, "Server error hashing password", http.StatusInternalServerError)
		return
	}

	// 4. Save to the database
	if _, err = h.UserModel.Insert(username, string(hashedPassword), email); err != nil {
		http.Error(w, "User already exists", http.StatusConflict)
		return
	}

	// 5. Share the NEWS!
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("User registered successfully!"))
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	// 1. Extract values
	username := r.FormValue("username")
	password := r.FormValue("password")

	// 2. Get by user by Username
	user, err := h.UserModel.GetByUsername(username)
	if err != nil {
		http.Error(w, "Invalid username/password", http.StatusUnauthorized)
		return
	}

	// 3. Validate Password
	err = bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(password))
	if err != nil {
		http.Error(w, "Invalid username/password", http.StatusUnauthorized)
		return
	}

	// 4. Generate SessionToken
	sessionToken, err := utils.GenerateToken(32)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// 5. Generate CSRFToken
	csrfToken, err := utils.GenerateToken(32)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// 6. Set tokens
	err = h.UserModel.UpdateSession(user.ID, &sessionToken, &csrfToken)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// 7. Set cookies
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    sessionToken,
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "csrf_token",
		Value:    csrfToken,
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: false,
	})

	// 8. Share the NEWS!
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Logged in successfully!"))
}

func (h *UserHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// 1. Get sessionToken
	sessionToken, err := r.Cookie("session_token")
	if err != nil || sessionToken.Value == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// 2. Get csrf
	csrfToken := r.Header.Get("X-CSRF-Token")

	// 3. Get user
	user, err := h.UserModel.GetBySession(sessionToken.Value, csrfToken)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// 4. Purge tokens
	err = h.UserModel.UpdateSession(user.ID, nil, nil)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// 5. Purge Cookies
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		HttpOnly: true,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "csrf_token",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		HttpOnly: false,
	})

	// 6. Share the NEWS!
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Logged out successfully!"))
}

func (h *UserHandler) RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. Get session token from cookie
		sessionCookie, err := r.Cookie("session_token")
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// 2. get csrf token
		csrfToken := r.Header.Get("X-CSRF-Token")

		// 3. Verify tokens against the database
		user, err := h.UserModel.GetBySession(sessionCookie.Value, csrfToken)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// 4. proceed! with user_id context!
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), "user_id_key", user.ID)))
	}
}
