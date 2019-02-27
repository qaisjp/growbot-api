package api

import (
	"encoding/base64"
	"log"
	"sync"

	"github.com/mattn/go-mjpeg"

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

var robotStreams = make(map[uuid.UUID]*mjpeg.Stream)
var robotStreamsMutex = &sync.Mutex{}

func GetStream(rid uuid.UUID) *mjpeg.Stream {
	robotStreamsMutex.Lock()
	stream, ok := robotStreams[rid]
	if !ok {
		stream = mjpeg.NewStream()
		robotStreams[rid] = stream
	}
	robotStreamsMutex.Unlock()
	return stream
}

func (i *API) StreamRobot(ctx *gin.Context) {
	w, r := ctx.Writer, ctx.Request

	rid := ctx.MustGet("robot").(*models.Robot).ID

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

func (a *API) StreamRobotVideo(ctx *gin.Context) {
	w, r := ctx.Writer, ctx.Request

	rid := ctx.MustGet("robot").(*models.Robot).ID

	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}

	stream := GetStream(rid)

	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}

		image, err := base64.StdEncoding.DecodeString(string(message))
		if err != nil {
			a.Log.WithField("rid", rid).WithError(err).Warnln("Could not decode base64 image sent")
		}

		if err := stream.Update(image); err != nil {
			a.Log.WithField("rid", rid).WithError(err).Warnln("Could not update stream")
			return
		}

		a.Log.WithField("rid", rid).Infoln("Image added to stream")
	}
}
