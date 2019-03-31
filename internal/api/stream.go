package api

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"

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
	CheckOrigin: func(r *http.Request) bool { return true },
}

var robotCtxs = make(map[uuid.UUID]*gin.Context)
var robotCtxsMutex = &sync.Mutex{}

var robotStreams = make(map[uuid.UUID]*Stream)
var robotStreamsMutex = &sync.Mutex{}

type Stream struct {
	Inner      *mjpeg.Stream
	LastUpdate time.Time
}

func (s *Stream) Update(b []byte) error {
	s.LastUpdate = time.Now()
	return s.Inner.Update(b)
}

var deadBytes []byte

func init() {
	var err error
	deadBytes, err = ioutil.ReadFile("dead.jpeg")
	if err != nil {
		panic(err)
	}
}

// VideoDeathThreshold is the duration we wait before we show the "dead" image
const VideoDeathThreshold = time.Second * 5

// VideoDeathFrequency is the frequency of which we send the "dead"
const VideoDeathFrequency = time.Second

func (a *API) GetStream(rid uuid.UUID) *Stream {
	robotStreamsMutex.Lock()
	stream, ok := robotStreams[rid]
	if !ok {
		stream = &Stream{
			Inner: mjpeg.NewStream(),
		}
		robotStreams[rid] = stream

		go func(stream *Stream) {
			tick := time.NewTicker(VideoDeathFrequency)

			for {
				select {
				case <-tick.C:
					if time.Now().Sub(stream.LastUpdate) > VideoDeathThreshold {
						// a.Log.WithField("uuid", rid).Infoln("Pushed dead image")
						if err := stream.Inner.Update(deadBytes); err != nil {
							a.Log.WithError(err).WithField("uuid", rid).Warnln("Could not add dead image")
						}
					}
				}
			}
		}(stream)
	}

	robotStreamsMutex.Unlock()

	return stream
}

func (a *API) StreamRobot(ctx *gin.Context) {
	w, r := ctx.Writer, ctx.Request

	robot := ctx.MustGet("robot").(*models.Robot)
	rid := robot.ID

	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}

	// Set the websocket connection on the context
	ctx.Set("ws", c)

	// Update seen_at
	updateSeenAt := func() {
		now := time.Now()
		_, err = a.DB.Exec("update robot_state set seen_at=$2 where id = $1", rid, now)
		if err != nil {
			a.Log.WithError(err).WithField("rid", rid).Warnln("Could not update seen_at")
		}

		if robot.UserID != nil {
			a.userStreams.transmit(*robot.UserID, "UPDATE_ROBOT_STATE", map[string]interface{}{"id": rid, "seen_at": now})
		}
	}
	updateSeenAt()

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

	// On first load, gather events, and push to client
	a.pingRobotEvents(rid)

	for {
		_, b, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}

		updateSeenAt()

		msg := struct {
			Type string
			Data map[string]interface{}
		}{}

		err = json.Unmarshal(b, &msg)
		if err != nil {
			a.Log.WithError(err).Warnln("Could not read websocket message")
			continue
		}

		switch msg.Type {
		case "PLANT_CAPTURE_PHOTO":
			plantID := int(msg.Data["id"].(float64))
			plantImageB64 := msg.Data["image"].(string)

			a.Log.WithField("plant_id", plantID).Infoln("PLANT_CAPTURE_PHOTO received")

			// todo: verify user owns plantID

			u := uuid.New()
			filename := photoBucketKey(u)
			photo := models.PlantPhoto{
				Filename: u,
				PlantID:  plantID,
			}

			w, err := a.Bucket.NewWriter(ctx, filename, nil)
			if err != nil {
				a.Log.WithError(err).Warnln("could not create bucket for PLANT_CAPTURE_PHOTO")
				continue
			}

			a.Log.WithField("plant_id", plantID).Infoln("PLANT_CAPTURE_PHOTO writer created")

			rb := base64.NewDecoder(base64.StdEncoding, bytes.NewReader([]byte(plantImageB64)))

			a.Log.WithField("plant_id", plantID).Infoln("PLANT_CAPTURE_PHOTO base64 decoder created")

			_, err = io.Copy(w, rb)
			if err != nil {
				a.Log.WithError(err).Warnln("could not decode base64 image into bucket")
				continue
			}

			a.Log.WithField("plant_id", plantID).Infoln("PLANT_CAPTURE_PHOTO copied to bucket")

			err = w.Close()
			if err != nil {
				a.Log.WithError(err).Warnln("could not close bucket writer")
			}

			a.Log.WithField("plant_id", plantID).Infoln("PLANT_CAPTURE_PHOTO inserting into db")

			_, err = a.DB.NamedQuery("insert into plant_photos(filename, plant_id) values (:filename, :plant_id)", photo)
			if err != nil {
				_ = a.Bucket.Delete(ctx, filename)
				a.Log.WithError(err).WithField("plant_id", plantID).Warnln("could not insert file for PLANT_CAPTURE_PHOTO")
				continue
			}

			a.Log.WithField("plant_id", plantID).Infoln("PLANT_CAPTURE_PHOTO done")

		case "CREATE_LOG_ENTRY":
			_, plantExists := msg.Data["plant_id"]
			var plantID *int
			if plantExists {
				id := int(msg.Data["plant_id"].(float64))
				plantID = &id
			}

			var uid *int
			err := a.DB.Get(&uid, "select user_id from robots where id=$1", rid)
			if err != nil {
				a.Log.WithField("data", msg.Data).WithError(err).Warnln("could not get user id for CREATE_LOG_ENTRY")
				continue
			}

			// Forget log entries when the robot is unregistered
			if uid == nil {
				continue
			}

			entry := LogEntry{
				UserID:   *uid,
				Type:     msg.Data["type"].(string),
				Message:  msg.Data["message"].(string),
				Severity: int(msg.Data["severity"].(float64)),
				RobotID:  &rid,
				PlantID:  plantID,
			}

			result, err := a.DB.NamedQuery("insert into log(user_id, type, message, severity, robot_id, plant_id) values (:user_id, :type, :message, :severity, :robot_id, :plant_id) returning id, created_at", entry)
			if err != nil {
				a.Log.WithError(err).WithField("data", msg.Data).Warnln("could not insert log entry for CREATE_LOG_ENTRY")
				continue
			}

			if !result.Next() {
				a.Log.Warnln("Expected result.Next() to return true for CREATE_LOG_ENTRY")
				continue
			}

			if err := result.StructScan(&entry); err != nil {
				a.Log.WithError(err).WithField("data", msg.Data).Warnln("could not scan log entry for CREATE_LOG_ENTRY")
				continue
			}

			a.userStreams.transmit(entry.UserID, "CREATE_LOG_ENTRY", entry)

		case "UPDATE_SOIL_MOISTURE":
			_, plantExists := msg.Data["plant_id"]
			if !plantExists {
				a.Log.WithField("data", msg.Data).Warnln("no plant_id provided for UPDATE_SOIL_MOISTURE")
				continue
			}

			plantID := int(msg.Data["plant_id"].(float64))

			plant := models.Plant{}
			err = a.DB.Get(&plant, "select user_id from plants where id=$1", plantID)
			if err != nil {
				a.Log.WithField("data", msg.Data).WithError(err).Warnln("could not get plant for UPDATE_SOIL_MOISTURE")
				continue
			}

			if plant.UserID != *robot.UserID {
				continue
			}

			moisture := int(msg.Data["moisture"].(float64))
			_, err = a.DB.Exec("update plants set soil_moisture=$2 where id=$1", plantID, moisture)
			if err != nil {
				a.Log.WithError(err).WithField("data", msg.Data).Warnln("could not update soil moisture for UPDATE_SOIL_MOISTURE")
				continue
			}

			a.userStreams.transmit(plant.UserID, "UPDATE_SOIL_MOISTURE", map[string]interface{}{
				"plant_id": plantID,
				"moisture": moisture,
			})

		default:
			a.Log.WithField("Type", msg.Type).Warnln("Received message with unk type from robot stream")
		}

		// log.Printf("recv: %s", message)
		// err = c.WriteMessage(mt, message)
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

	stream := a.GetStream(rid)

	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}

		image, err := base64.StdEncoding.DecodeString(string(message))
		if err != nil {
			a.Log.WithField("rid", rid).WithError(err).Warnln("Could not decode base64 image sent")
			continue
		}

		if err := stream.Update(image); err != nil {
			a.Log.WithField("rid", rid).WithError(err).Warnln("Could not update stream")
			continue
		}

		a.Log.WithField("rid", rid).Infoln("Image added to stream")
	}
}
