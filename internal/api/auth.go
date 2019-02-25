package api

import (
	"net/http"

	"golang.org/x/crypto/bcrypt"

	"github.com/gin-gonic/gin"
	"github.com/teamxiv/growbot-api/internal/models"
)

func (a *API) AuthLoginPost(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "not implemented (yet)"})
}

// AuthRegisterPost takes:
// - forename
// - surname
// - email address
// - password (in plaintext)
//
// Usually returns:
// - HTTP Status OK (200)
//   At this point the user should receive a verification token in their inbox.
//   This will be done later.
//
// Otherwise, complains about:
// - Bad email format
// - Email address already existing
// - Shitty password (min 8 characters please)
func (a *API) AuthRegisterPost(c *gin.Context) {
	// TODO(q): add govalidator
	input := struct {
		Forename string
		Surname  string
		Email    string
		Password string
	}{}

	err := c.BindJSON(&input)
	if err != nil {
		BadRequest(c, err.Error())
		return
	}

	if len(input.Password) < 8 {
		BadRequest(c, "Password too short")
		return
	}

	// Bcrypt this password
	password, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		BadRequest(c, err.Error())
	}

	// Create their row
	row := models.User{
		Forename: input.Forename,
		Surname:  input.Surname,
		Email:    input.Email,
		Password: string(password),

		// TODO(q): remove this line and the field in the row insertion
		Activated: true,
	}

	// TODO(q): Send them an email with their verification code
	// 1. Generate verification code
	// 2. Add to verification code table with their user ID
	// 3. Email them the code

	// Check if email already exists
	// ...

	// Create the row
	result, err := a.DB.NamedQuery("insert into users(forename, surname, email, password, activated) values (:forename, :surname, :email, :password, :activated) RETURNING id", row)
	if err != nil {
		BadRequest(c, err.Error())
		return
	}

	if result.Rows.Next() {
		var n int
		result.Rows.Scan(&n)

		c.JSON(http.StatusOK, gin.H{
			"expire": "todo",
			"token":  "asdfsdf",
			"id":     n,
		})
	} else {
		BadRequest(c, result.Rows.Err().Error())
	}
}

func (a *API) AuthForgotPost(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "not implemented (yet)"})
}

func (a *API) AuthChgPassPost(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "not implemented (yet)"})
}
