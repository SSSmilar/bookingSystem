package main

import (
	"bookingSystem/internal/handlers"
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
)

var db *pgxpool.Pool

func main() {
	var err error
	dbUrl := "postgres://postgres:password@localhost:5432/booking_db?sslmode=disable"
	db, err = pgxpool.New(context.Background(), dbUrl)
	if err != nil {
		fmt.Println("Error connecting to database:", err)
	}
	defer db.Close()

	err = db.Ping(context.Background())
	if err != nil {
		fmt.Println("DataBase unavailable :", err)
	}
	log.Println("Connected to database PostgreSQL")

	myHandler := handlers.New(db)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/rooms", myHandler.GetRooms)
	mux.HandleFunc("POST /api/register", myHandler.Register)
	mux.HandleFunc("POST /api/login", myHandler.Login)
	handler := CorsMiddleware(mux)
	log.Println("Server started on port 8080")
	if err := http.ListenAndServe(":8080", handler); err != nil {
		log.Fatal(err)
	}
}
func CorsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", " Content-Type,  Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}
