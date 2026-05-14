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

	h := &handlers.UserHandler{
		UserModel: userModel,
	}

	mux := http.NewServeMux()

	mux.HandleFunc("POST /register", h.Register)
	mux.HandleFunc("POST /login", h.Login)
	mux.HandleFunc("POST /logout", h.Logout) // FOR TESTING obv

	log.Println("Server starting on :8080")
	err = http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Fatal(err)
	}
}
