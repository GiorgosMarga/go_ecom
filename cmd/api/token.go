package main

import (
	"errors"
	"net/http"

	"github.com/GiorgosMarga/ecom_go/models"
	"github.com/gin-gonic/gin"
)

func (app *application) registerTokenRoutes(router *gin.Engine) {
	v1 := router.Group("/api/v1/tokens")
	v1.GET("/refresh", app.refreshTokenHandler)
}

func (app *application) refreshTokenHandler(c *gin.Context) {
	cookies, err := c.Request.Cookie("refresh_token")
	if err != nil {
		app.notAuthenticatedError(c)
		return
	}

	token, err := app.verifyToken(cookies.Value)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidJWT):
			app.notAuthenticatedError(c)
		default:
			app.internalServerError(c, err)
		}
		return
	}

	tokenId, err := token.Claims.GetSubject()
	if err != nil {
		app.notAuthenticatedError(c)
		return
	}

	userId, err := app.models.Token.GetForUser(tokenId)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrNotFound):
			app.notFoundError(c)
		default:
			app.internalServerError(c, err)
		}
		return
	}
	user, err := app.models.User.GetByID(userId.Hex())
	if err != nil {
		switch {
		case errors.Is(err, models.ErrNotFound):
			app.notFoundError(c)
		default:
			app.internalServerError(c, err)
		}
		return
	}

	accessToken, err := app.createAccessToken(*user)
	if err != nil {
		app.internalServerError(c, err)
	}

	c.JSON(http.StatusOK, gin.H{"access_token": accessToken})
}
