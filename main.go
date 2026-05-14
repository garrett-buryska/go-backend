package main

import (
	"go-backend/internal/database"
	"go-backend/internal/handlers"
	"go-backend/internal/models"
	"log"
	"net/http"
)

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
	mux.HandleFunc("POST /register", uh.Register) // REGISTER
	mux.HandleFunc("POST /login", uh.Login)       // LOGIN
	mux.HandleFunc("POST /logout", uh.Logout)     // LOGOUT

	// PROTECTED
	// board
	mux.HandleFunc("GET /{boardID}", uh.RequireAuth(bh.GetBoard))  // FETCH board (and cards and columns)
	mux.HandleFunc("POST /", uh.RequireAuth(bh.Create))            // CREATE board
	mux.HandleFunc("PUT /{boardID}", uh.RequireAuth(bh.Update))    // UPDATE board
	mux.HandleFunc("DELETE /{boardID}", uh.RequireAuth(bh.Delete)) // DELETE board
	// column
	mux.HandleFunc("POST /{boardID}/", uh.RequireAuth(ch.Create))             // CREATE column
	mux.HandleFunc("PUT /{boardID}/{columnID}", uh.RequireAuth(ch.Update))    // UPDATE column
	mux.HandleFunc("DELETE /{boardID}/{columnID}", uh.RequireAuth(ch.Delete)) // DELETE column
	// card
	mux.HandleFunc("POST /{boardID}/{columnID}/", uh.RequireAuth(cah.Create))           // CREATE card
	mux.HandleFunc("PUT /{boardID}/{columnID}/{cardID}", uh.RequireAuth(cah.Update))    // UPDATE card
	mux.HandleFunc("DELETE /{boardID}/{columnID}/{cardID}", uh.RequireAuth(cah.Delete)) // DELETE card

	log.Println("Server starting on :8080")
	err = http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Fatal(err)
	}
}
