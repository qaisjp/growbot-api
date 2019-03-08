package models

import (
	"time"

	"github.com/jmoiron/sqlx/types"
)

type Event struct {
	ID         int       `json:"id" db:"id"`
	Summary    string    `json:"summary" db:"summary"`
	Start      time.Time `json:"start" db:"start"`
	Recurrence []string  `json:"recurrence" db:"recurrence"`
	UserID     int       `json:"user_id" db:"user_id"`
}

type EventAction struct {
	ID      int             `json:"id" db:"id"`
	Name    EventActionName `json:"name" db:"name"`
	Data    types.JSONText  `json:"data" db:"data"`
	PlantID int             `json:"plant_id" db:"plant_id"`
	EventID int             `json:"event_id" db:"event_id"`
}

type EventActionName int

const (
	EventActionPlantWater EventActionName = iota
	EventActionPlantCapturePhoto
)
