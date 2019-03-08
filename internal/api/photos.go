package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/teamxiv/growbot-api/internal/models"
)

// PhotosListGet requires you to be logged in.
// It lists all photos of plants the user owns
func (a *API) PhotosListGet(c *gin.Context) {
	user_id := c.GetInt("user_id")

	plant_id_str := c.Query("plant_id")
	var plant_id int
	if plant_id_str != "" {
		var err error
		plant_id, err = strconv.Atoi(plant_id_str)
		if err != nil {
			BadRequest(c, err.Error())
			return
		}
	}

	photos := []models.PlantPhoto{}

	var err error
	if plant_id_str == "" {
		err = a.DB.Select(&photos, "select ph.* from plants as pl, plant_photos as ph where pl.user_id=$1 and ph.plant_id=pl.plant_id", user_id)
	} else {
		err = a.DB.Select(&photos, "select ph.* from plants as pl, plant_photos as ph where pl.user_id=$1 and ph.plant_id=pl.plant_id and pl.plant_id=$2", user_id, plant_id)
	}

	if err != nil {
		BadRequest(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"photos": photos,
	})
}
