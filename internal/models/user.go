package models

import "time"

// User is a user account in the system
type User struct {
	ID        int    `json:"id" db:"id"`
	Forename  string `json:"forename" db:"forename"`
	Surname   string `json:"surname" db:"surname"`
	Password  string `json:"password" db:"password"`
	Email     string `json:"email" db:"email"`
	Activated bool   `json:"is_activated" db:"is_activated"`

	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}
