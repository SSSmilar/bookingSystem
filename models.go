package main

import "time"

type Room struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	ID          int64  `json:"id"`
	Capacity    int64  `json:"capacity"`
}

type User struct {
	ID       int64  `json:"id"`
	Email    string `json:"email"`
	Password string `json:"password,omitempty"`
	Role     string `json:"role"`
}

type Booking struct {
	ID        int64     `json:"id"`
	RoomID    int64     `json:"roomId"`
	UserID    int64     `json:"userId"`
	Titel     string    `json:"title"`
	StartTime time.Time `json:"startTime"`
	EndTime   time.Time `json:"endTime"`
}
type Credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
