package main

import (
	"context"
	"go-backend/internal/database"
	"go-backend/internal/handlers"
	"go-backend/internal/models"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. Identify the origin of the request
		origin := r.Header.Get("Origin")

		// 2. Allow local React/Vite dev servers (Add your production domain here later!)
		allowedOrigins := map[string]bool{
			"http://localhost:3000":                 true, // Standard React app
			"http://localhost:5173":                 true, // Vite app
			"https://cardboard.garrettburyska.work": true,
		}

		if allowedOrigins[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}

		// 3. Set the required CORS headers
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		// Notice X-CSRF-Token is explicitly allowed here
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		// CRITICAL: This allows the browser to send your HttpOnly session cookie
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		// 4. Handle preflight OPTIONS requests immediately
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		// 5. Pass the request down the chain to your actual handlers
		next.ServeHTTP(w, r)
	})
}

func main() {
	db, err := database.InitDB("./data/app.db")
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	userModel := &models.UserModel{DB: db}

	uh := &handlers.UserHandler{
		UserModel: userModel,
	}

	boardModel := &models.BoardModel{DB: db}

	bh := &handlers.BoardHandler{
		BoardModel: boardModel,
	}

	columnModel := &models.ColumnModel{DB: db}

	ch := &handlers.ColumnHandler{
		ColumnModel: columnModel,
	}

	cardModel := &models.CardModel{DB: db}

	cah := &handlers.CardHandler{
		CardModel: cardModel,
	}

	mux := http.NewServeMux()

	// AUTHS
	mux.HandleFunc("POST /api/auth/register", uh.Register)
	mux.HandleFunc("POST /api/auth/login", uh.Login)
	mux.HandleFunc("POST /api/auth/logout", uh.Logout)
	mux.HandleFunc("GET /api/auth/me", uh.Me)

	// BOARDS
	mux.HandleFunc("GET /api/boards", uh.RequireAuth(bh.GetAll))
	mux.HandleFunc("GET /api/boards/{boardID}", uh.RequireAuth(bh.GetBoard))
	mux.HandleFunc("POST /api/boards", uh.RequireAuth(bh.Create))
	mux.HandleFunc("PUT /api/boards/{boardID}", uh.RequireAuth(bh.Update))
	mux.HandleFunc("DELETE /api/boards/{boardID}", uh.RequireAuth(bh.Delete))

	// COLUMNS (Nested under the board they belong to)
	mux.HandleFunc("POST /api/boards/{boardID}/columns", uh.RequireAuth(ch.Create))
	mux.HandleFunc("PUT /api/boards/{boardID}/columns/{columnID}", uh.RequireAuth(ch.Update))
	mux.HandleFunc("DELETE /api/boards/{boardID}/columns/{columnID}", uh.RequireAuth(ch.Delete))

	// CARDS (Nested under columns)
	mux.HandleFunc("POST /api/columns/{columnID}/cards", uh.RequireAuth(cah.Create))
	mux.HandleFunc("PUT /api/columns/{columnID}/cards/{cardID}", uh.RequireAuth(cah.Update))
	mux.HandleFunc("DELETE /api/columns/{columnID}/cards/{cardID}", uh.RequireAuth(cah.Delete))

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      corsMiddleware(mux),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	// 1. Run the server in a goroutine so it doesn't block the rest of the code
	go func() {
		log.Println("Server starting on :8080")
		if err = srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	// 2. Wait for an interrupt signal (Ctrl+C)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	log.Println("Shutting down server...")

	// 3. Give the server 5 seconds to finish active requests
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited properly")
}
