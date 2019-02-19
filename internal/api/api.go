package api

import (
	"context"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"github.com/teamxiv/growbot-api/internal/config"
)

// API contains all the dependencies of the API server
type API struct {
	Config *config.Config
	Log    *logrus.Logger
	Gin    *gin.Engine
	DB     *sqlx.DB

	Server *http.Server
}

// Start binds the API and starts listening.
func (a *API) Start() error {
	a.Server = &http.Server{
		Addr:    a.Config.BindAddress,
		Handler: a.Gin,
	}
	return a.Server.ListenAndServe()
}

// Shutdown shuts down the API
func (a *API) Shutdown(ctx context.Context) error {
	if err := a.Server.Shutdown(ctx); err != nil {
		return err
	}

	return nil
}

func BadRequest(c *gin.Context, msg string) {
	c.JSON(http.StatusBadRequest, gin.H{
		"status":  "error",
		"message": msg,
	})
}

func (a *API) NotImplemented(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "not implemented (yet)"})
}

// NewAPI sets up a new API module.
func NewAPI(
	conf *config.Config,
	log *logrus.Logger,
	db *sqlx.DB,
) *API {

	router := gin.Default()
	router.Use(cors.Default())

	a := &API{
		Config: conf,
		Log:    log,
		Gin:    router,
		DB:     db,
	}

	// General websocket
	router.GET("/stream", a.NotImplemented)

	// Authentication
	auth := router.Group("/auth")
	auth.POST("/login", a.AuthLoginPost)
	auth.POST("/register", a.AuthRegisterPost)
	auth.POST("/forgot", a.AuthForgotPost)
	auth.POST("/chgpass", a.AuthChgPassPost)

	// Robots
	robots := router.Group("/robots")
	robots.GET("", a.RobotListGet) // List robots
	robots.POST("/register", a.RobotRegisterPost)

	// A robot
	aRobot := router.Group("/robot/:uuid")
	aRobot.GET("", a.RobotStatusGet) // Get (status) info
	aRobot.DELETE("", a.RobotDelete) // Delete this bot
	aRobot.POST("/move", a.RobotMovePost)
	aRobot.POST("/startDemo", a.RobotStartDemoPost)
	aRobot.PATCH("/settings", a.RobotSettingsPatch)
	aRobot.GET("/stream", a.RobotStream)

	// Legacy
	router.GET("/status", a.StatusGet)
	router.POST("/move", a.MovePost)
	router.POST("/demo/start", a.DemoStartPost)
	router.PATCH("/settings", a.SettingsPatch)
	router.GET("/stream/:uuid", a.StreamRobot)

	return a
}
