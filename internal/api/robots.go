package api

import (
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/teamxiv/growbot-api/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RobotCheck is a middleware to check whether the passed robot uuid exists,
// and (if logged in) confirms whether the currently logged in user owns that robot
func (a *API) RobotCheck(c *gin.Context) {
	id := c.Param("uuid")
	rid, err := uuid.Parse(id)
	if err != nil {
		BadRequest(c, err.Error())
		c.Abort()
		return
	}

	robot := models.Robot{}
	err = a.DB.Get(&robot, "select * from robots where id = $1", rid)
	if err != nil {
		BadRequest(c, "Robot does not exist ("+err.Error()+")")
		c.Abort()
		return
	}

	_, loggedIn := c.Get("user_id")
	if uid := robot.UserID; loggedIn && (uid == nil || *uid != c.GetInt("user_id")) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "error",
			"message": "you don't own that robot",
		})
		c.Abort()
		return
	}

	// Store the robot in the context
	c.Set("robot", &robot)
}

// RobotVideoGet streams the video
func (a *API) RobotVideoGet(c *gin.Context) {
	robot := c.MustGet("robot").(*models.Robot)

	stream := a.GetStream(robot.ID)
	stream.Inner.ServeHTTP(c.Writer, c.Request)
}

// RobotListGet requires you to be logged in.
// It lists all robots the user owns + the robot state.
func (a *API) RobotListGet(c *gin.Context) {
	user_id := c.GetInt("user_id")

	robots := []struct {
		ID uuid.UUID `json:"id" db:"robot_id"`
		models.Robot
		models.RobotState
	}{}

	err := a.DB.Select(&robots, "select robots.id as robot_id,created_at,updated_at,robot_state.*,title from robots,robot_state where robots.user_id=$1 and robot_state.id=robots.id", user_id)
	if err != nil {
		BadRequest(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"robots": robots,
	})
}

// RobotRegisterPost takes a "robot_id" in the JSON body.
// It registers the robot corresponding to the UUID to the currently logged in user.
//
// Usually returns HTTP Status OK.
// Otherwise complains.
func (a *API) RobotRegisterPost(c *gin.Context) {
	user_id := c.GetInt("user_id")

	input := struct {
		RobotID uuid.UUID `json:"robot_id"`
		Title   string    `json:"title"`
	}{}

	err := c.BindJSON(&input)
	if err != nil {
		BadRequest(c, err.Error())
		return
	}

	if input.Title == "" {
		input.Title = "Unnamed Robot"
	}

	var robot models.Robot

	err = a.DB.Get(&robot, "select * from robots where id = $1", input.RobotID)
	if err != nil {
		BadRequest(c, err.Error())
		return
	}

	if robot.UserID != nil {
		BadRequest(c, "This robot has already been registered. Please email us.")
		return
	}

	_, err = a.DB.Exec("update robots set user_id=$1, title=$3 where id=$2 returning id", user_id, input.RobotID, input.Title)
	if err != nil {
		BadRequest(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
	})
}

// RobotStatusGet returns the state of the robot + whether or not it is currently connected
func (a *API) RobotStatusGet(c *gin.Context) {
	robot := c.MustGet("robot").(*models.Robot)

	var state struct {
		models.RobotState
		Connected bool `json:"is_connected"`
	}

	err := a.DB.Get(&state, "select * from robot_state where id = $1", robot.ID)
	if err != nil {
		BadRequest(c, err.Error())
		return
	}

	robotCtxsMutex.Lock()
	_, connected := robotCtxs[robot.ID]
	robotCtxsMutex.Unlock()

	state.Connected = connected

	c.JSON(http.StatusOK, state)
}

// RobotDelete disassociates the robot from the current user
func (a *API) RobotDelete(c *gin.Context) {
	robot := c.MustGet("robot").(*models.Robot)

	_, err := a.DB.Exec("update robots set user_id=null,title='' where id=$1", robot.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "could not delete row: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
	})
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
		c.JSON(http.StatusFailedDependency, gin.H{
			"message": "Robot not connected",
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
		c.JSON(http.StatusFailedDependency, gin.H{
			"message": "Robot not connected",
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

	var input struct {
		Key   string
		Value interface{}
	}

	type settingsOut struct {
		Key   string
		Value interface{}
	}

	if err := c.BindJSON(&input); err != nil {
		BadRequest(c, err.Error())
		return
	}

	// Key `title` is special, database only.
	if input.Key == "title" {
		_, err := a.DB.Exec("update robots set title = $2 where id = $1", robot.ID, input.Value)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Could not update database: " + err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":  "success",
			"message": "Robot has been renamed",
		})
		return
	}

	payload := struct {
		Type string      `json:"type"`
		Data settingsOut `json:"data"`
	}{
		Type: "settings/patch",
		Data: settingsOut{
			Key:   input.Key,
			Value: input.Value,
		},
	}

	robotCtxsMutex.Lock()
	wctx, ok := robotCtxs[robot.ID]
	robotCtxsMutex.Unlock()

	if !ok {
		c.JSON(http.StatusFailedDependency, gin.H{
			"message": "Robot not connected",
		})
		return
	}

	wsc := wctx.MustGet("ws").(*websocket.Conn)
	wsc.WriteJSON(payload)

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
	})
}
