package api

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/teamxiv/growbot-api/internal/models"

	"github.com/gin-gonic/gin"
)

// PlantCheck is a middleware to check whether the passed plant uuid exists,
// and (if logged in) confirms whether the currently logged in user owns that plant
func (a *API) PlantCheck(c *gin.Context) {
	id := c.Param("uuid")
	rid, err := uuid.Parse(id)
	if err != nil {
		BadRequest(c, err.Error())
		c.Abort()
		return
	}

	plant := models.Plant{}
	err = a.DB.Get(&plant, "select * from plants where id = $1", rid)
	if err != nil {
		BadRequest(c, "Plant does not exist ("+err.Error()+")")
		c.Abort()
		return
	}

	_, loggedIn := c.Get("user_id")
	if uid := plant.UserID; loggedIn && uid != c.GetInt("user_id") {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "error",
			"message": "you don't own that plant",
		})
		c.Abort()
		return
	}

	// Store the plant in the context
	c.Set("plant", &plant)
}

// PlantListGet requires you to be logged in.
// It lists all plants the user owns
func (a *API) PlantListGet(c *gin.Context) {
	user_id := c.GetInt("user_id")

	plants := []models.Plant{}

	err := a.DB.Select(&plants, "select * from plants where user_id=$1", user_id)
	if err != nil {
		BadRequest(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"plants": plants,
	})
}
