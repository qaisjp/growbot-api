package models

import "time"

// Room is a room in a user's world
type Room struct {
	ID     int    `json:"id" db:"id"`
	UserID int    `json:"user_id" db:"user_id"`
	Name   string `json:"name" db:"name"`

	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}
