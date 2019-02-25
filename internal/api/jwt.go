package api

import (
	"errors"

	jwt "github.com/appleboy/gin-jwt"
	"github.com/gin-gonic/gin"
	"github.com/teamxiv/growbot-api/internal/models"
	"golang.org/x/crypto/bcrypt"
)

func (a *API) jwtUnauthorized(c *gin.Context, code int, message string) {
	c.JSON(code, gin.H{
		"code":    code,
		"message": message,
	})
}

func (a *API) jwtAuthorizator(data interface{}, c *gin.Context) bool {
	return true
}

func (a *API) jwtAuthenticator(c *gin.Context) (interface{}, error) {
	var input struct {
		Email    string
		Password string
	}

	if err := c.BindJSON(&input); err != nil {
		return "", jwt.ErrMissingLoginValues
	}

	var user models.User

	err := a.DB.Get(&user, "select id,password,is_activated from users where email = $1 limit 1", input.Email)
	if err != nil {
		return "", err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password))
	if err != nil {
		return "", err
	}

	if !user.Activated {
		return "", errors.New("account not activated")
	}

	return &user, nil
}

func (a *API) jwtIdentityHandler(c *gin.Context) interface{} {
	claims := jwt.ExtractClaims(c)

	return int(claims["id"].(float64))
}

func (a *API) jwtPayloadFunc(data interface{}) jwt.MapClaims {
	if v, ok := data.(*models.User); ok {
		return jwt.MapClaims{
			"id": v.ID,
		}
	}
	return jwt.MapClaims{}
}
