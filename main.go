package main

import (
	"context"
	"encoding/json"
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

	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/rooms", GetRooms)
}
func GetRooms(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query(context.Background(), "SELECT id , name , capacity ,  description FROM rooms")
	if err != nil {
		http.Error(w, "Error DataBase", http.StatusInternalServerError)
		log.Println(err)
		return
	}
	defer rows.Close()
	var roms []Room
	for rows.Next() {
		var r Room
		if err := rows.Scan(&r.ID, &r.Name, &r.Capacity, &r.Description); err != nil {
			log.Println(err)
			continue
		}
		roms = append(roms, r)
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(roms); err != nil {
		log.Println(err)
	}
}
