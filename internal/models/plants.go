package models

import (
	"time"

	"github.com/google/uuid"
)

type Plant struct {
	ID           uuid.UUID `json:"id" db:"id"`
	Name         string    `json:"name" db:"name"`
	UserID       int       `json:"user_id" db:"user_id"`
	SoilMoisture *int      `json:"soil_moisture" db:"soil_moisture"`
}

type PlantPhoto struct {
	ID        int       `json:"id" db:"id"`
	Filename  uuid.UUID `json:"filename" db:"filename"`
	PlantID   uuid.UUID `json:"plant_id" db:"plant_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}
