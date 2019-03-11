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

// EventListGet gets the plant object
func (a *API) EventListGet(c *gin.Context) {
	userID := c.GetInt("user_id")

	events := []models.Event{}

	err := a.DB.Select(&events, "select * from events where user_id=$1", userID)

	if err != nil {
		a.error(c, http.StatusBadRequest, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"events": events,
	})
}

// EventCreatePost gets the plant object
func (a *API) EventCreatePost(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "not implemented (yet)"})
}

// EventGet gets the Event object
func (a *API) EventGet(c *gin.Context) {
	event := c.MustGet("event").(*models.Event)

	c.JSON(http.StatusOK, event)
}

// EventPut updates the event object
func (a *API) EventPut(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "not implemented (yet)"})
}

// EventDelete gets the plant object
func (a *API) EventDelete(c *gin.Context) {
	event := c.MustGet("event").(*models.Event)

	_, err := a.DB.Exec("delete from events where id=$1", event.ID)
	if err != nil {
		a.error(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
	})
}
