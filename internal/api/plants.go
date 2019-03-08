package api

import (
	"net/http"

	"github.com/teamxiv/growbot-api/internal/models"

	"github.com/gin-gonic/gin"
)

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
