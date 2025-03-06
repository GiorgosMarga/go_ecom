package main

import (
	"errors"
	"net/http"

	"github.com/GiorgosMarga/ecom_go/internal/validator"
	"github.com/GiorgosMarga/ecom_go/models"
	"github.com/GiorgosMarga/ecom_go/utils"
	"github.com/gin-gonic/gin"
)

func (app *application) registerUserRoutes(router *gin.Engine) {
	v1 := router.Group("/api/v1/users")
	v1.POST("/register", app.registerUserHandler)
	v1.POST("/login", app.loginUserHandler)
	v1.PATCH("/:id", app.authenticateUser(), app.updateUserHandler)
	v1.GET("/:id", app.getUserByIdHandler)
}

func (app *application) getUserByIdHandler(c *gin.Context) {
	id := c.Params.ByName("id")

	user, err := app.models.User.GetByID(id)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrNotFound):
			app.notFoundError(c)
		default:
			app.internalServerError(c, err)
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}
func (app *application) registerUserHandler(c *gin.Context) {
	user := models.User{}
	err := c.BindJSON(&user)
	if err != nil {
		app.badRequestError(c, err)
		return
	}

	v := validator.NewValidator()
	if models.ValidateUser(v, user); !v.IsValid() {
		app.failedValidationError(c, v.Errors)
		return
	}

	err = app.models.User.Insert(&user)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrUsedEmail):
			app.badRequestError(c, err)
		default:
			app.internalServerError(c, err)
		}
		return
	}

	accessToken, err := app.createAccessToken(user)
	if err != nil {
		app.internalServerError(c, err)
		return
	}

	refreshToken, err := app.createRefreshToken(user)
	if err != nil {
		app.internalServerError(c, err)
		return
	}
	// keep cookie for a week
	c.SetCookie("refrest_token", refreshToken, 60*60*24*7, "/", "localhost", false, true)
	c.SetCookie("access_token", accessToken, 60*60*24*7, "/", "localhost", false, true)
	c.JSON(http.StatusCreated, gin.H{"user": user})
}

func (app *application) loginUserHandler(c *gin.Context) {
	user := models.User{}

	err := c.BindJSON(&user)
	if err != nil {
		app.badRequestError(c, err)
		return
	}

	v := validator.NewValidator()
	if models.ValidateUserLogin(v, user); !v.IsValid() {
		app.failedValidationError(c, v.Errors)
		return
	}

	u, err := app.models.User.GetByEmail(user.Email)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrNotFound):
			app.wrongCredentialsError(c)
		default:
			app.internalServerError(c, err)
		}
		return
	}
	err = utils.CheckPasswordHash(user.PasswordHash, u.PasswordHash)
	if err != nil {
		switch {
		case errors.Is(err, utils.ErrInvalidPassword):
			app.wrongCredentialsError(c)
		default:
			app.internalServerError(c, err)
		}
		return
	}

	// Change user password to not send them back to the client
	u.PasswordHash = ""

	accessToken, err := app.createAccessToken(*u)
	if err != nil {
		app.internalServerError(c, err)
		return
	}
	refreshToken, err := app.createRefreshToken(*u)
	if err != nil {
		app.internalServerError(c, err)
		return
	}
	c.SetCookie("refresh_token", refreshToken, 60*60*24*7, "/", "localhost", false, false) // 7 days
	c.SetCookie("access_token", accessToken, 60*60*24*7, "/", "localhost", false, false)   // 7 days
	c.JSON(http.StatusOK, gin.H{"user": u})
}

func (app *application) updateUserHandler(c *gin.Context) {
	id := c.Param("id")
	user, err := app.models.User.GetByID(id)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrNotFound):
			app.notFoundError(c)
		default:
			app.internalServerError(c, err)
		}
		return
	}
	contextUser, err := GetUser(c)
	if err != nil {
		app.notAuthenticatedError(c)
		return
	}

	// check if user is updating itself or tryign to update someone else
	if contextUser.UserID.Hex() != id {
		app.notAuthorizedError(c)
		return
	}

	payload := models.UserUpdatePayload{}

	err = c.BindJSON(&payload)
	if err != nil {
		app.badRequestError(c, err)
		return
	}

	if payload.Email != nil {
		user.Email = *payload.Email
	}
	if payload.Role != nil {
		user.Role = *payload.Role
	}
	if payload.Name != nil {
		user.Name = *payload.Name
	}
	if payload.Password != nil {
		pwd, err := utils.HashPassword(*payload.Password)
		if err != nil {
			app.internalServerError(c, err)
			return
		}
		user.PasswordHash = pwd
	}

	v := validator.NewValidator()
	if models.ValidateUser(v, *user); !v.IsValid() {
		app.failedValidationError(c, v.Errors)
		return
	}
	if !user.Role.Validate() {
		v.AddError("role", "invalid role")
		app.failedValidationError(c, v.Errors)
		return
	}

	err = app.models.User.Update(user)
	if err != nil {
		app.internalServerError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"user": user})
}
