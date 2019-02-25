package api

import (
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/teamxiv/growbot-api/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (a *API) RobotCheck(c *gin.Context) {
	id := c.Param("uuid")
	rid, err := uuid.Parse(id)
	if err != nil {
		BadRequest(c, err.Error())
		c.Abort()
		return
	}

	// Check if the robot exists
	exists := rid == a.Config.UUID
	if !exists {
		BadRequest(c, "Robot "+rid.String()+" does not exist")
		c.Abort()
		return
	}

	robot := models.Robot{}

	// Store the robot in the context
	c.Set("robot", &robot)
}

func (a *API) RobotListGet(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "not implemented (yet)"})
}

func (a *API) RobotRegisterPost(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "not implemented (yet)"})
}

func (a *API) RobotStatusGet(c *gin.Context) {
	robot := c.MustGet("robot").(*models.Robot)

	robotCtxsMutex.Lock()
	_, online := robotCtxs[robot.ID]
	robotCtxsMutex.Unlock()

	c.JSON(http.StatusOK, gin.H{
		"online": online,
	})
}

func (a *API) RobotDelete(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "not implemented (yet)"})
}

func (a *API) RobotMovePost(c *gin.Context) {
	robot := c.MustGet("robot").(*models.Robot)

	var result struct {
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
	wctx, ok := robotCtxs[robot.ID]
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

func (a *API) RobotStartDemoPost(c *gin.Context) {
	robot := c.MustGet("robot").(*models.Robot)

	var result struct {
		Procedure string
	}

	if err := c.BindJSON(&result); err != nil {
		BadRequest(c, err.Error())
		return
	}

	payload := struct {
		Type string `json:"type"`
		Data string `json:"data"`
	}{
		Type: "demo/start",
		Data: result.Procedure,
	}

	robotCtxsMutex.Lock()
	wctx, ok := robotCtxs[robot.ID]
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

func (a *API) RobotSettingsPatch(c *gin.Context) {
	robot := c.MustGet("robot").(*models.Robot)

	var result struct {
		Key   string
		Value interface{}
	}

	type settingsOut struct {
		Key   string
		Value interface{}
	}

	if err := c.BindJSON(&result); err != nil {
		BadRequest(c, err.Error())
		return
	}

	payload := struct {
		Type string      `json:"type"`
		Data settingsOut `json:"data"`
	}{
		Type: "settings/patch",
		Data: settingsOut{
			Key:   result.Key,
			Value: result.Value,
		},
	}

	robotCtxsMutex.Lock()
	wctx, ok := robotCtxs[robot.ID]
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
