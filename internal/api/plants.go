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

// PlantGet gets the plant object
func (a *API) PlantGet(c *gin.Context) {
	plant := c.MustGet("plant").(*models.Plant)
	c.JSON(http.StatusOK, plant)
}

// PlantCreatePost gets the plant object
func (a *API) PlantCreatePost(c *gin.Context) {
	input := struct {
		ID   uuid.UUID `json:"id,omitempty"`
		Name string    `json:"name"`
	}{}

	err := c.BindJSON(&input)
	if err != nil {
		a.error(c, http.StatusBadRequest, err.Error())
		return
	}

	if input.ID == uuid.Nil {
		input.ID = uuid.New()
	} else {
		// Check the given UUID doesn't already exist
		var count int
		err := a.DB.Get(&count, "select count(id) from plants where id = $1", input.ID)
		if err != nil {
			a.error(c, http.StatusInternalServerError, err.Error())
			return
		}

		if count > 0 {
			a.error(c, http.StatusBadRequest, "This plant is already in the database.")
			return
		}
	}

	_, err = a.DB.NamedQuery("insert into plants(id, name) values (:id, :name)", input)
	if err != nil {
		a.error(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id": input.ID,
	})
}

// PlantRenamePatch gets the plant object
func (a *API) PlantRenamePatch(c *gin.Context) {
	plant := c.MustGet("plant").(*models.Plant)

	input := struct {
		Name string `json:"name"`
	}{}

	err := c.BindJSON(&input)
	if err != nil {
		a.error(c, http.StatusBadRequest, err.Error())
		return
	}

	_, err = a.DB.Exec("update plants set name=$2 where id=$1", plant.ID, input.Name)
	if err != nil {
		a.error(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
	})
}
