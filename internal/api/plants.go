package api

import (
	"net/http"
	"strconv"

	"github.com/teamxiv/growbot-api/internal/models"

	"github.com/gin-gonic/gin"
)

// PlantCheck is a middleware to check whether the passed plant uuid exists,
// and (if logged in) confirms whether the currently logged in user owns that plant
func (a *API) PlantCheck(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		BadRequest(c, err.Error())
		c.Abort()
		return
	}

	plant := models.Plant{}
	err = a.DB.Get(&plant, "select * from plants where id = $1", id)
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
	userID := c.GetInt("user_id")

	plants := []models.Plant{}

	err := a.DB.Select(&plants, "select * from plants where user_id=$1", userID)
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

// PlantDelete deletes the plant object
func (a *API) PlantDelete(c *gin.Context) {
	plant := c.MustGet("plant").(*models.Plant)

	_, err := a.DB.Exec("delete from plants where id = $1", plant.ID)
	if err != nil {
		a.error(c, http.StatusInternalServerError, "could not delete plant: "+err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

// PlantCreatePost gets the plant object
func (a *API) PlantCreatePost(c *gin.Context) {
	input := struct {
		Name string `json:"name"`
	}{}

	err := c.BindJSON(&input)
	if err != nil {
		a.error(c, http.StatusBadRequest, err.Error())
		return
	}

	row := models.Plant{
		Name:   input.Name,
		UserID: c.GetInt("user_id"),
	}

	result, err := a.DB.NamedQuery("insert into plants(name, user_id) values (:name, :user_id) returning id", row)
	if err != nil {
		a.error(c, http.StatusInternalServerError, err.Error())
		return
	}

	if !result.Next() {
		a.error(c, http.StatusInternalServerError, "Expected result.Next() to return true")
		return
	}

	var id int
	if err := result.Scan(&id); err != nil {
		a.error(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id": id,
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
