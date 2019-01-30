package api

import (
	"context"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/teamxiv/growbot-api/internal/config"
)

// API contains all the dependencies of the API server
type API struct {
	Config *config.Config
	Log    *logrus.Logger
	Gin    *gin.Engine

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

// NewAPI sets up a new API module.
func NewAPI(
	conf *config.Config,
	log *logrus.Logger,
) *API {

	router := gin.Default()
	router.Use(cors.Default())

	a := &API{
		Config: conf,
		Log:    log,
		Gin:    router,
	}

	router.POST("/move", a.MovePost)
	router.GET("/stream", a.Stream)

	return a
}
