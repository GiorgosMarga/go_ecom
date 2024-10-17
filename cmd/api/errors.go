package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (app *application) logError(c *gin.Context, err error) {
	app.logger.Println(c.Request.Method, c.Request.URL, err.Error())
}

func (app *application) sendError(c *gin.Context, status int, err any) {
	c.JSON(status, gin.H{
		"errors": err,
	})
}

func (app *application) internalServerError(c *gin.Context, err error) {
	msg := "the server encountered a problem and could not process your request"
	app.logError(c, err)
	app.sendError(c, http.StatusInternalServerError, msg)
}

func (app *application) badRequestError(c *gin.Context, err error) {
	msg := fmt.Sprintf("bad request error: %s", err.Error())
	app.sendError(c, http.StatusBadRequest, msg)
}

func (app *application) notFoundError(c *gin.Context) {
	msg := "resource could not be found"
	app.sendError(c, http.StatusNotFound, msg)
}

func (app *application) notAuthenticatedError(c *gin.Context) {
	msg := "you need to be authenticated to access this resource"
	app.sendError(c, http.StatusUnauthorized, msg)
}

func (app *application) notAuthorizedError(c *gin.Context) {
	msg := "you are not authorized to access this resource"
	app.sendError(c, http.StatusUnauthorized, msg)
}

func (app *application) failedValidationError(c *gin.Context, errors any) {
	app.sendError(c, http.StatusUnprocessableEntity, errors)
}

func (app *application) wrongCredentialsError(c *gin.Context) {
	msg := "wrong credentials"
	app.sendError(c, http.StatusUnauthorized, msg)
}
