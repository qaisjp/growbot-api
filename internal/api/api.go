package api

import (
	"context"
	"net/http"
	"time"

	jwt "github.com/appleboy/gin-jwt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"github.com/teamxiv/growbot-api/internal/config"
	"gocloud.dev/blob"
)

// API contains all the dependencies of the API server
type API struct {
	Config *config.Config
	Log    *logrus.Logger
	Gin    *gin.Engine
	DB     *sqlx.DB
	Bucket *blob.Bucket

	Server *http.Server

	userStreams *userStreams
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

func (a *API) error(c *gin.Context, code int, msg string) {
	c.JSON(code, gin.H{
		"status":  "error",
		"message": msg,
	})
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
	bucket *blob.Bucket,
) *API {

	router := gin.Default()

	corsConf := cors.DefaultConfig()
	corsConf.AddAllowMethods("DELETE", "PATCH")
	corsConf.AddAllowHeaders("Authorization")
	corsConf.AllowAllOrigins = true

	router.Use(cors.New(corsConf))

	a := &API{
		Config: conf,
		Log:    log,
		Gin:    router,
		DB:     db,
		Bucket: bucket,

		userStreams: newUserStream(),
	}

	// the jwt middleware
	authMiddleware, err := jwt.New(&jwt.GinJWTMiddleware{
		Realm:           "test zone",
		Key:             []byte("secret key"),
		Timeout:         time.Hour * 6,
		MaxRefresh:      time.Hour * 24 * 3,
		IdentityKey:     "user_id",
		PayloadFunc:     a.jwtPayloadFunc,
		IdentityHandler: a.jwtIdentityHandler,
		Authenticator:   a.jwtAuthenticator,
		Authorizator:    a.jwtAuthorizator,
		Unauthorized:    a.jwtUnauthorized,
		// TokenLookup is a string in the form of "<source>:<name>" that is used
		// to extract token from the request.
		// Optional. Default value "header:Authorization".
		// Possible values:
		// - "header:<name>"
		// - "query:<name>"
		// - "cookie:<name>"
		// - "param:<name>"
		TokenLookup: "header: Authorization, query: token, cookie: jwt",
		// TokenLookup: "query:token",
		// TokenLookup: "cookie:token",
	})

	if err != nil {
		log.WithField("error", err).Fatal("jwt error")
	}

	// router.NoRoute(authMiddleware.MiddlewareFunc(), func(c *gin.Context) {
	// 	claims := jwt.ExtractClaims(c)
	// 	log.Printf("NoRoute claims: %#v\n", claims)
	// 	c.JSON(404, gin.H{"code": "PAGE_NOT_FOUND", "message": "Page not found"})
	// })

	// Actual auth middleware
	authRequired := authMiddleware.MiddlewareFunc()

	// Stream
	{
		// General websocket
		router.GET("/stream", authRequired, a.StreamUser)

		// Robots can connect to the stream without authentication
		router.GET("/stream/:uuid", a.RobotCheck, a.StreamRobot)

		// Robots can stream videos without authentication
		router.GET("/stream-video/:uuid", a.RobotCheck, a.StreamRobotVideo)
	}

	// Authentication
	auth := router.Group("/auth")
	{
		auth.POST("/login", authMiddleware.LoginHandler)
		auth.POST("/refresh", authMiddleware.RefreshHandler)
		auth.POST("/register", a.AuthRegisterPost)
		auth.POST("/forgot", a.AuthForgotPost)
		auth.POST("/chgpass", authRequired, a.AuthChgPassPost)
	}

	// Log
	logs := router.Group("/log", authRequired)
	{
		logs.GET("", a.LogListGet)
	}

	// Robots
	robots := router.Group("/robots", authRequired)
	{
		robots.GET("", a.RobotListGet) // List robots
		robots.POST("/register", a.RobotRegisterPost)
	}

	// A robot
	aRobot := router.Group("/robot/:uuid", authRequired, a.RobotCheck)
	{
		aRobot.GET("", a.RobotStatusGet)      // Get (status) info
		aRobot.GET("/video", a.RobotVideoGet) // Get video
		aRobot.DELETE("", a.RobotDelete)      // Delete this bot
		aRobot.POST("/move", a.RobotMovePost)
		aRobot.POST("/startDemo", a.RobotStartDemoPost)
		aRobot.PATCH("/settings", a.RobotSettingsPatch)
	}

	// Photos
	photos := router.Group("/photos", authRequired)
	{
		photos.GET("", a.PhotosListGet)

		photo := photos.Group("/:id", a.PhotoCheck)
		{
			photo.GET("", a.PhotoServeGet)
			photo.DELETE("", a.PhotoDelete)
		}
	}

	// Plants
	plants := router.Group("/plants", authRequired)
	{
		plants.GET("", a.PlantListGet)
		plants.POST("", a.PlantCreatePost) // create a plant, only {name: ""}

		plant := plants.Group("/:uuid", a.PlantCheck)
		{
			plant.GET("", a.PlantGet)
			plant.DELETE("", a.PlantDelete)
			plant.PATCH("", a.PlantRenamePatch)
		}
	}

	// Events
	events := router.Group("/events", authRequired)
	{
		events.GET("", a.EventListGet)
		events.POST("", a.EventCreatePost)

		event := events.Group("/:id", a.EventCheck)
		{
			event.GET("", a.EventGet)
			event.PUT("", a.EventPut)
			event.DELETE("", a.EventDelete)
		}
	}

	return a
}
