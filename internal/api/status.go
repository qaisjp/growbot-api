package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (i *API) StatusGet(c *gin.Context) {
	id := c.Query("id")
	rid, err := uuid.Parse(id)
	if err != nil {
		BadRequest(c, err.Error())
		return
	}

	robotCtxsMutex.Lock()
	_, online := robotCtxs[rid]
	robotCtxsMutex.Unlock()

	c.JSON(http.StatusOK, gin.H{
		"online": online,
	})
}
