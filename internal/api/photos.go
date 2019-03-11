package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/teamxiv/growbot-api/internal/models"
)

// PhotoCheck is a middleware to check whether the passed photo uuid exists,
// and (if logged in) confirms whether the currently logged in user owns that robot
func (a *API) PhotoCheck(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		BadRequest(c, err.Error())
		c.Abort()
		return
	}

	photo := struct {
		models.PlantPhoto
		UserID int `db:"photo_id"`
	}{}
	err = a.DB.Get(&photo, "select ph.*, pl.user_id as user_id from plant_photos as ph, plants as pl where ph.id = $1 and ph.plant_id = pl.id", id)
	if err != nil {
		BadRequest(c, "Photo does not exist ("+err.Error()+")")
		c.Abort()
		return
	}

	_, loggedIn := c.Get("user_id")
	if uid := photo.UserID; loggedIn && uid != c.GetInt("user_id") {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "error",
			"message": "You don't own that plant",
		})
		c.Abort()
		return
	}

	// Store the robot in the context
	c.Set("photo", &photo.PlantPhoto)
}

// PhotosListGet requires you to be logged in.
// It lists all photos of plants the user owns
func (a *API) PhotosListGet(c *gin.Context) {
	userID := c.GetInt("user_id")

	plantIDstr := c.Query("plant_id")
	var plantID int
	if plantIDstr != "" {
		var err error
		plantID, err = strconv.Atoi(plantIDstr)
		if err != nil {
			BadRequest(c, err.Error())
			return
		}
	}

	photos := []models.PlantPhoto{}

	var err error
	if plantIDstr == "" {
		err = a.DB.Select(&photos, "select ph.* from plants as pl, plant_photos as ph where pl.user_id=$1 and ph.plant_id=pl.plant_id", userID)
	} else {
		err = a.DB.Select(&photos, "select ph.* from plants as pl, plant_photos as ph where pl.user_id=$1 and ph.plant_id=pl.plant_id and pl.plant_id=$2", userID, plantID)
	}

	if err != nil {
		BadRequest(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"photos": photos,
	})
}
