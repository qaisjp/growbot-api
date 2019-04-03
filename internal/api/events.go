package api

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"github.com/jmoiron/sqlx/types"
	"github.com/teamxiv/growbot-api/internal/models"

	"github.com/gin-gonic/gin"
)

// EventCheck is a middleware to check whether the passed event id exists,
// and (if logged in) confirms whether the currently logged in user owns that event
func (a *API) EventCheck(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		BadRequest(c, err.Error())
		c.Abort()
		return
	}

	event := models.Event{}
	err = a.DB.Get(&event, "select * from events where id = $1", id)
	if err != nil {
		BadRequest(c, "Event does not exist ("+err.Error()+")")
		c.Abort()
		return
	}

	_, loggedIn := c.Get("user_id")
	if uid := event.UserID; loggedIn && uid != c.GetInt("user_id") {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "error",
			"message": "You don't own that event",
		})
		c.Abort()
		return
	}

	// Store the event in the context
	c.Set("event", &event)
}

// EventGet gets the Event object
func (a *API) EventGet(c *gin.Context) {
	event := c.MustGet("event").(*models.Event)

	result := struct {
		models.Event
		Actions []models.EventAction `json:"actions"`
	}{Event: *event}

	err := a.DB.Select(&result.Actions, "select * from event_actions as a where a.event_id=$1", event.ID)
	if err != nil {
		a.error(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, result)
}

type expandedEvent struct {
	models.Event
	Action []models.EventAction `json:"actions"`
}

func (a *API) expandedEventsByUserID(userID int) ([]expandedEvent, error) {
	events := []struct {
		models.Event
		Actions types.JSONText `json:"actions" db:"actions"`
	}{}

	err := a.DB.Select(&events, "select e.*, json_agg(a) as actions from event_actions as a, events as e where e.user_id=$1 and a.event_id=e.id group by e.id", userID)
	if err != nil {
		return nil, err
	}

	result := make([]expandedEvent, len(events))

	for i, event := range events {
		result[i].Event = event.Event

		if err := event.Actions.Unmarshal(&result[i].Action); err != nil {
			return nil, err
		}
	}

	return result, nil
}

func (a *API) pingRobotEvents(rid uuid.UUID, boot bool) {
	robotCtxsMutex.Lock()
	c, ok := robotCtxs[rid]
	robotCtxsMutex.Unlock()

	// If not connected, stop
	if !ok {
		return
	}

	events := []struct {
		models.Event
		Actions types.JSONText `json:"actions" db:"actions"`
	}{}

	query := ""
	if boot {
		query = "and not e.ephemeral"
	}

	err := a.DB.Select(&events, "select e.*, json_agg(a) as actions from event_actions as a, events as e where a.robot_id=$1 and a.event_id=e.id "+query+" group by e.id", rid)
	if err != nil {
		a.Log.WithError(err).WithField("rid", rid).Warnln("could not get events from db")
		return
	}

	result := make([]expandedEvent, len(events))

	for i, event := range events {
		if event.Ephemeral {
			a.DB.MustExec("delete from events where id = $1", event.ID)
		}
		result[i].Event = event.Event

		if err := event.Actions.Unmarshal(&result[i].Action); err != nil {
			a.Log.WithError(err).WithField("rid", rid).Warnln("could not unmarshal actions")
			return
		}
	}

	ws := c.MustGet("ws").(*websocket.Conn)

	msg := struct {
		Type string          `json:"type"`
		Data []expandedEvent `json:"data"`
	}{"events", result}

	err = ws.WriteJSON(msg)
	if err != nil {
		a.Log.WithError(err).WithField("rid", rid).Warnln("could not send events")
	}
}

// func (a *API) pingUserEvents(uid *int, rid *uuid.UUID) {
// 	var events []expandedEvent

// 	rids := []uuid.UUID{}
// 	if uid != nil {
// 		var error
// 		events, err = a.expandedEventsByUserID(uid)
// 		if err != nil {
// 			a.Log.WithError(err).WithField("uid", uid).Warnln("could not gather events")
// 			return
// 		}
// 	} else {
// 		// Get associated IDs
// 		err := a.DB.Get(&rids, "select array_agg(id) from robots where user_id=$1", uid)
// 		if err != nil {
// 			a.Log.WithError(err).WithField("uid", uid).Warnln("could not gather events")
// 			return
// 		}
// 	}

// 	if rid != nil {
// 		fmt.Println(*rid)
// 		rids = append(rids, *rid)
// 		events, err =  a.expandedEventsByRobotID(*rid)
// 		if err != nil {
// 			a.Log.WithError(err).WithField("rid", *rid).Warnln("could not gather events")
// 			return
// 		}
// 	}

// 	eventMap := make(map[uuid.UUID][]expandedEvent)
// 	for _, rid := range rids {
// 		// Initialise to empty list
// 		eventMap[rid] = []expandedEvent{}
// 		fmt.Println("intialised")

// 		for _, event := range events {
// 			actions := []models.EventAction{}
// 			for _, action := range event.Action {
// 				if action.RobotID == rid {
// 					actions = append(actions, action)
// 				}
// 			}

// 			if len(actions) > 0 {
// 				e := expandedEvent{
// 					Event:  event.Event,
// 					Action: actions,
// 				}
// 				eventMap[rid] = append(eventMap[rid], e)
// 			}
// 		}
// 	}

// 	for _, rid := range rids {

// 		robotCtxsMutex.Lock()
// 		c, ok := robotCtxs[rid]
// 		robotCtxsMutex.Unlock()

// 		// If not connected, continue
// 		if !ok {
// 			continue
// 		}

// 		ws := c.MustGet("ws").(*websocket.Conn)

// 		msg := struct {
// 			Type string          `json:"type"`
// 			Data []expandedEvent `json:"data"`
// 		}{"events", eventMap[rid]}

// 		err := ws.WriteJSON(msg)
// 		if err != nil {
// 			a.Log.WithError(err).WithField("uid", uid).Warnln("could not send events")
// 		}
// 	}

// }

// EventListGet gets the plant object
func (a *API) EventListGet(c *gin.Context) {
	userID := c.GetInt("user_id")

	events, err := a.expandedEventsByUserID(userID)
	if err != nil {
		a.error(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"events": events,
	})
}

// EventCreatePost gets the plant object
func (a *API) EventCreatePost(c *gin.Context) {
	userID := c.GetInt("user_id")
	input := struct {
		models.Event
		Actions []models.EventAction
	}{}

	err := c.BindJSON(&input)
	if err != nil {
		a.error(c, http.StatusBadRequest, err.Error())
		return
	}

	hasActions := len(input.Actions) > 0

	query := `insert into events (summary, recurrence, user_id, ephemeral) values ($1, $2, $3, $4) returning id`
	if hasActions {
		query = "with inserted as (" + query + ")"
	}

	args := []interface{}{input.Event.Summary, input.Recurrences, userID, input.Ephemeral}
	rids := make(map[uuid.UUID]struct{})

	for i, action := range input.Actions {
		if i == 0 {
			query += "\ninsert into event_actions (event_id, name, plant_id, robot_id, data) values "
		} else {
			query += ", "
		}

		rids[action.RobotID] = struct{}{}

		argCount := len(args)
		query += fmt.Sprintf("\n( (select id from inserted), $%d, $%d, $%d, $%d )", argCount+1, argCount+2, argCount+3, argCount+4)
		args = append(args, action.Name, action.PlantID, action.RobotID, action.Data)
	}

	if hasActions {
		query += " returning (select id from inserted)"
	}

	var id int
	err = a.DB.Get(&id, query, args...)
	if err != nil {
		a.error(c, http.StatusInternalServerError, err.Error())
		return
	}

	for rid := range rids {
		a.pingRobotEvents(rid, false)
	}

	c.JSON(http.StatusCreated, gin.H{
		"id": id,
	})
}

// EventPut updates the event object
func (a *API) EventPut(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "not implemented (yet)"})
}

// EventDelete gets the plant object
func (a *API) EventDelete(c *gin.Context) {
	event := c.MustGet("event").(*models.Event)

	_, err := a.DB.Exec("delete from events where id=$1", event.ID)
	if err != nil {
		a.error(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
	})
}
