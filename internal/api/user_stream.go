package api

import (
	"log"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/teamxiv/growbot-api/internal/models"
)

type userStreams struct {
	m   map[int][]*websocket.Conn
	mux sync.RWMutex
}

func newUserStream() *userStreams {
	return &userStreams{
		m: make(map[int][]*websocket.Conn),
	}
}

func (s *userStreams) transmit(uid int, typ string, data interface{}) {
	s.mux.RLock()
	defer s.mux.RUnlock()

	a, ok := s.m[uid]
	if !ok {
		return
	}

	message := struct {
		Type string      `json:"type"`
		Data interface{} `json:"data"`
	}{typ, data}

	for _, c := range a {
		c.WriteJSON(message)
	}
}

func (s *userStreams) add(uid int, conn *websocket.Conn) {
	s.mux.Lock()
	defer s.mux.Unlock()

	a, ok := s.m[uid]
	if !ok {
		s.m[uid] = []*websocket.Conn{conn}
		return
	}

	s.m[uid] = append(a, conn)
}

func (s *userStreams) remove(uid int, conn *websocket.Conn) {
	s.mux.Lock()
	defer s.mux.Unlock()

	a, ok := s.m[uid]
	if !ok {
		return
	}

	slice := a[:0]
	for _, c := range a {
		if c != conn {
			slice = append(slice, c)
		}
	}
}

func (a *API) StreamUser(ctx *gin.Context) {
	w, r := ctx.Writer, ctx.Request

	user := ctx.MustGet("user").(*models.User)
	uid := user.ID

	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}

	// Set the websocket connection on the context
	ctx.Set("ws", c)

	// Add this websocket connection to the map
	a.userStreams.add(uid, c)

	defer func() {
		a.userStreams.remove(uid, c)
		c.Close()
	}()

	for {
		_, b, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}

		a.Log.WithField("msg", string(b)).Warnln("Received user stream message, didn't expect any")
	}
}
