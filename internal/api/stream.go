package api

import (
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/teamxiv/growbot-api/internal/models"
)

var upgrader = websocket.Upgrader{
	// Currently uses default options
	// ReadBufferSize:  1024,
	// WriteBufferSize: 1024,
}

var robotCtxs = make(map[uuid.UUID]*gin.Context)
var robotCtxsMutex = &sync.Mutex{}

func (i *API) StreamRobot(ctx *gin.Context) {
	w, r := ctx.Writer, ctx.Request

	var rid uuid.UUID
	var err error
	if rid, err = uuid.Parse(ctx.Param("uuid")); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	var robot models.Robot

	if err := i.DB.Get(&robot, "select * from robots where id = $1", rid); err != nil {
		ctx.JSON(http.StatusForbidden, gin.H{
			"message": "invalid uuid: " + err.Error(),
		})
		return
	}

	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}

	// Set the websocket connection on the context
	ctx.Set("ws", c)

	// Add this websocket connection to the map (and cancel the existing one)
	{
		robotCtxsMutex.Lock()

		// Close the old one if it exists
		oldCtx, exists := robotCtxs[rid]
		if exists {
			ws := oldCtx.MustGet("ws").(*websocket.Conn)
			ws.Close()
		}

		// Add the new context
		robotCtxs[rid] = ctx

		robotCtxsMutex.Unlock()
	}

	defer func() {
		c.Close()

		robotCtxsMutex.Lock()
		defer robotCtxsMutex.Unlock()

		if robotCtxs[rid] == ctx {
			delete(robotCtxs, rid)
		}
	}()

	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		log.Printf("recv: %s", message)
		err = c.WriteMessage(mt, message)
		if err != nil {
			log.Println("write:", err)
			break
		}
	}
}
