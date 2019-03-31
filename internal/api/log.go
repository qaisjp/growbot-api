package api

import (
	"fmt"
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
	PlantID   *int       `json:"plant_id,omitempty" db:"plant_id"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
}

// LogListGet returns a list of log entries
func (a *API) LogListGet(c *gin.Context) {
	userID := c.GetInt("user_id")

	input := struct {
		RobotID  *string `form:"robot_id"`
		PlantID  *int    `form:"plant_id"`
		Severity *int    `form:"severity"`
		Limit    int     `form:"limit,default=10"`
		Offset   int     `form:"offset,default=0"`
	}{}

	if err := c.BindQuery(&input); err != nil {
		a.error(c, http.StatusBadRequest, err.Error())
		return
	}

	args := []interface{}{userID}
	query := "user_id=$1"
	queryIndex := 1

	if input.RobotID != nil {
		queryIndex++
		args = append(args, *input.RobotID)
		query += fmt.Sprintf(" and robot_id=$%d", queryIndex)
	}

	if input.PlantID != nil {
		queryIndex++
		args = append(args, *input.PlantID)
		query += fmt.Sprintf(" and plant_id=$%d", queryIndex)
	}

	if input.Severity != nil {
		queryIndex++
		args = append(args, *input.Severity)
		query += fmt.Sprintf(" and severity=$%d", queryIndex)
	}

	queryIndex += 2
	args = append(args, input.Limit, input.Offset)
	query += fmt.Sprintf(" order by created_at desc limit $%d offset $%d", queryIndex-1, queryIndex)

	entries := []LogEntry{}

	err := a.DB.Select(&entries, "select * from log where "+query+" ", args...)
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
