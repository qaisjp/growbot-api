package api

import (
	"net/http"
	"strconv"

	"github.com/teamxiv/growbot-api/internal/models"

	"github.com/gin-gonic/gin"
)

// EventCheck is a middleware to check whether the passed event id exists,
// and (if logged in) confirms whether the currently logged in user owns that event
func (a *API) EventCheck(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		BadRequest(c, err.Error())
		c.Abort()
		return
	}

	event := models.Event{}
	err = a.DB.Get(&event, "select * from events where id = $1", id)
	if err != nil {
		BadRequest(c, "Event does not exist ("+err.Error()+")")
		c.Abort()
		return
	}

	_, loggedIn := c.Get("user_id")
	if uid := event.UserID; loggedIn && uid != c.GetInt("user_id") {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "error",
			"message": "You don't own that event",
		})
		c.Abort()
		return
	}

	// Store the event in the context
	c.Set("event", &event)
}
