package models

import (
	"github.com/lib/pq"

	"github.com/google/uuid"

	"github.com/jmoiron/sqlx/types"
)

type Event struct {
	ID          int            `json:"id" db:"id"`
	Summary     string         `json:"summary" db:"summary"`
	Recurrences pq.StringArray `json:"recurrences" db:"recurrence"`
	UserID      int            `json:"user_id" db:"user_id"`
	Ephemeral   bool           `json:"ephemeral,omitempty" db:"ephemeral"`
}

type EventAction struct {
	ID      int            `json:"id" db:"id"`
	Name    string         `json:"name" db:"name"`
	Data    types.JSONText `json:"data" db:"data"`
	PlantID *int           `json:"plant_id,omitempty" db:"plant_id"`
	RobotID uuid.UUID      `json:"robot_id" db:"robot_id"`
	EventID int            `json:"event_id" db:"event_id"`
}

const (
	EventActionPlantWater        = "PLANT_WATER"
	EventActionPlantCapturePhoto = "PLANT_CAPTURE_PHOTO"
	EventActionRobotRandomMove   = "ROBOT_RANDOM_MOVE"
)
