package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (i *API) MovePost(c *gin.Context) {
	var result struct {
		Direction string
	}

	if err := c.BindJSON(&result); err != nil {
		BadRequest(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   result.Direction,
	})
}
