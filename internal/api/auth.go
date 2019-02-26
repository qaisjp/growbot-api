package api

import (
	"net/http"

	"golang.org/x/crypto/bcrypt"

	"github.com/gin-gonic/gin"
	"github.com/teamxiv/growbot-api/internal/models"
)

func validatePassword(pw string) (bool, string) {
	if len(pw) < 8 {
		return false, "Password too short"
	}

	return true, ""
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

	success, errMsg := validatePassword(input.Password)
	if !success {
		BadRequest(c, errMsg)
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
	result, err := a.DB.NamedQuery("insert into users(forename, surname, email, password, is_activated) values (:forename, :surname, :email, :password, :is_activated) RETURNING id", row)
	if err != nil {
		BadRequest(c, err.Error())
		return
	}

	if result.Rows.Next() {
		var n int
		result.Rows.Scan(&n)

		c.JSON(http.StatusOK, gin.H{
			"message": "success",
		})
	} else {
		BadRequest(c, result.Rows.Err().Error())
	}
}

func (a *API) AuthForgotPost(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "not implemented (yet)"})
}

func (a *API) AuthChgPassPost(c *gin.Context) {
	userID := c.GetInt("user_id")

	input := struct {
		Old string `json:"old"`
		New string `json:"new"`
	}{}

	err := c.BindJSON(&input)
	if err != nil {
		BadRequest(c, err.Error())
		return
	}

	success, errMsg := validatePassword(input.New)
	if !success {
		BadRequest(c, errMsg)
		return
	}

	var user models.User

	err = a.DB.Get(&user, "select id,password from users where id = $1 limit 1", userID)
	if err != nil {
		BadRequest(c, err.Error())
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Old))
	if err != nil {
		BadRequest(c, err.Error())
		return
	}

	// Bcrypt this password
	password, err := bcrypt.GenerateFromPassword([]byte(input.New), bcrypt.DefaultCost)
	if err != nil {
		BadRequest(c, err.Error())
	}

	_, err = a.DB.Exec("update users set password = $2 where id = $1", userID, password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Successfully updated password!",
	})
}
