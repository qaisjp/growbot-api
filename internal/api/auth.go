package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (a *API) AuthLoginPost(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "not implemented (yet)"})
}

func (a *API) AuthRegisterPost(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "not implemented (yet)"})
}

func (a *API) AuthForgotPost(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "not implemented (yet)"})
}

func (a *API) AuthChgPassPost(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "not implemented (yet)"})
}
