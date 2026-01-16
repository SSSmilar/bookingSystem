package handlers

import (
	"bookingSystem/internal/models"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

var jwtKey = []byte("secret_key_for_bookingSystem")

type Handler struct {
	DB *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Handler {
	return &Handler{DB: db}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var creds models.Credentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "Error format decoding ", http.StatusBadRequest)
		return
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(creds.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Error hashing password", http.StatusInternalServerError)
		return
	}
	if _, err := h.DB.Exec(context.Background(),
		"INSERT INTO users (email, password_hash) VALUES ($1, $2)", creds.Email, string(hashedPassword)); err != nil {
		log.Println("Error Registering ", err)
		http.Error(w, "User already exist  ", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(map[string]string{"message": "User Registered Successfully"}); err != nil {
		log.Println("Error encoding response ", err)
	}
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var creds models.Credentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "Error format decoding", http.StatusBadRequest)
		return
	}
	var role string
	var storedHash string
	var userID int64
	if err := h.DB.QueryRow(context.Background(),
		"SELECT id , password_hash , role FROM users WHERE  email = $1", creds.Email).Scan(&userID, &storedHash, &role); err != nil {
		http.Error(w, "Invalid Email or Password please check this ", http.StatusUnauthorized)
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(creds.Password)); err != nil {
		http.Error(w, "Invalid Email or Password please check this ", http.StatusUnauthorized)
		return
	}
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &jwt.MapClaims{
		"user_id": userID,
		"email":   creds.Email,
		"role":    role,
		"exp":     expirationTime.Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		http.Error(w, "Error generation token", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{
		"token": tokenString,
		"role":  role,
	}); err != nil {
		log.Println("Error encoding response ", err)
	}
}
func (h *Handler) GetRooms(w http.ResponseWriter, r *http.Request) {
	rows, err := h.DB.Query(context.Background(), "SELECT id , name ,  capacity , description FROM rooms")
	if err != nil {
		http.Error(w, "Error DataBase", http.StatusInternalServerError)
		log.Println(err)
		return
	}
	defer rows.Close()

	var rooms []models.Room
	for rows.Next() {
		var r models.Room
		if err := rows.Scan(&r.ID, &r.Name, &r.Capacity, &r.Description); err != nil {
			continue
		}
		rooms = append(rooms, r)
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(rooms); err != nil {
		log.Println("Error encoding response ", err)
	}
}
func (h *Handler) AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("Authorization")
		if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
			tokenString = tokenString[7:]
		}
		claims := &jwt.MapClaims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})
		if err != nil || !token.Valid {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		userIdFloat, ok := (*claims)["user_id"].(float64)
		if !ok {
			http.Error(w, "Invalid Token claims ", http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), "user_id", int64(userIdFloat))
		next(w, r.WithContext(ctx))
	}
}
