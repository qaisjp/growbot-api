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

	corsConf := cors.DefaultConfig()
	corsConf.AddAllowHeaders("Authorization")
	corsConf.AllowAllOrigins = true

	router.Use(cors.New(corsConf))

	a := &API{
		Config: conf,
		Log:    log,
		Gin:    router,
		DB:     db,
	}

	// the jwt middleware
	authMiddleware, err := jwt.New(&jwt.GinJWTMiddleware{
		Realm:           "test zone",
		Key:             []byte("secret key"),
		Timeout:         time.Hour,
		MaxRefresh:      time.Hour,
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
	stream := router.Group("/stream")
	{
		// General websocket
		stream.GET("", authRequired, a.NotImplemented)

		// Robots can connect to the stream without authentication
		stream.GET("/:uuid", a.StreamRobot)
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

	// Robots
	robots := router.Group("/robots", authRequired)
	{
		robots.GET("", a.RobotListGet) // List robots
		robots.POST("/register", a.RobotRegisterPost)
	}

	// A robot
	aRobot := router.Group("/robot/:uuid", authRequired, a.RobotCheck)
	{
		aRobot.GET("", a.RobotStatusGet) // Get (status) info
		aRobot.DELETE("", a.RobotDelete) // Delete this bot
		aRobot.POST("/move", a.RobotMovePost)
		aRobot.POST("/startDemo", a.RobotStartDemoPost)
		aRobot.PATCH("/settings", a.RobotSettingsPatch)
	}

	return a
}
