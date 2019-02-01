package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

func (i *API) MovePost(c *gin.Context) {
	var result struct {
		ID        uuid.UUID
		Direction string
	}

	if err := c.BindJSON(&result); err != nil {
		BadRequest(c, err.Error())
		return
	}

	payload := struct {
		Type string `json:"type"`
		Data string `json:"data"`
	}{
		Type: "move",
		Data: result.Direction,
	}

	robotCtxsMutex.Lock()
	wctx, ok := robotCtxs[result.ID]
	robotCtxsMutex.Unlock()

	if !ok {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Robot not found",
		})
		return
	}

	wsc := wctx.MustGet("ws").(*websocket.Conn)
	wsc.WriteJSON(payload)

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
	})
}
