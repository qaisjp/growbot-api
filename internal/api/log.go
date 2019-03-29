package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	LogSeverityInfo int = iota
	LogSeveritySuccess
	LogSeverityWarning
	LogSeverityDanger
)

type LogEntry struct {
	ID        int        `json:"id" db:"id"`
	UserID    int        `json:"user_id" db:"user_id"`
	Type      string     `json:"type" db:"type"`
	Message   string     `json:"message" db:"message"`
	Severity  int        `json:"severity" db:"severity"`
	RobotID   *uuid.UUID `json:"robot_id,omitempty" db:"robot_id"`
	PlantID   *uuid.UUID `json:"plant_id,omitempty" db:"plant_id"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
}

// LogListGet returns a list of log entries
func (a *API) LogListGet(c *gin.Context) {
	userID := c.GetInt("user_id")

	entries := []LogEntry{}

	err := a.DB.Select(&entries, "select * from log where user_id=$1", userID)
	if err != nil {
		a.error(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"entries": entries,
	})
}

// LogCheck is a middleware to check whether the passed entry id exists,
// and (if logged in) confirms whether the currently logged in user owns that entry
func (a *API) LogCheck(c *gin.Context) {
	// idStr := c.Param("id")
	// id, err := strconv.Atoi(idStr)
	// if err != nil {
	// 	a.error(c, http.StatusBadRequest, err.Error())
	// 	c.Abort()
	// 	return
	// }

	// plant := models.Plant{}
	// err = a.DB.Get(&plant, "select * from plants where id = $1", rid)
	// if err != nil {
	// 	BadRequest(c, "Plant does not exist ("+err.Error()+")")
	// 	c.Abort()
	// 	return
	// }

	// _, loggedIn := c.Get("user_id")
	// if uid := plant.UserID; loggedIn && uid != c.GetInt("user_id") {
	// 	c.JSON(http.StatusUnauthorized, gin.H{
	// 		"status":  "error",
	// 		"message": "you don't own that plant",
	// 	})
	// 	c.Abort()
	// 	return
	// }

	// // Store the plant in the context
	// c.Set("plant", &plant)
}

func (a *API) LogEntryDelete(c *gin.Context) {

}
