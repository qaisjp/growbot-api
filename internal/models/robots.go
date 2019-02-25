package models

import "time"

type Robot struct {
	ID         string `json:"id" db:"id"`
	AdminToken string `json:"-" db:"admin_token"`
	UserID     string `json:"user_id,omitempty" db:"user_id"`
	RoomID     string `json:"room_id,omitempty" db:"room_id"`

	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type RobotState struct {
	ID           string `json:"id" db:"id"`
	BatteryLevel int    `json:"battery_level" db:"battery_level"`
	WaterLevel   int    `json:"water_level" db:"water_level"`
	Distress     bool   `json:"distress" db:"distress"`

	SeenAt time.Time `json:"seen_at" db:"seen_at"`
}
