package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/GiorgosMarga/ecom_go/models"
	"github.com/gin-gonic/gin"
)

func (app *application) registerCartRoutes(router *gin.Engine) {
	v1 := router.Group("/api/v1/cart")
	v1.POST("", app.authenticateUser(), app.createCartHandler)
	v1.GET("/:id", app.authenticateUser(), app.getCartHandler)
	v1.DELETE("/:id", app.authenticateUser(), app.deleteCartHandler)
	v1.PATCH("/:id", app.authenticateUser(), app.updateCartHandler)
	// v1.GET("/:id", app.getUserByIdHandler)
}

func (app *application) createCartHandler(c *gin.Context) {
	cart := &models.Cart{}
	if err := c.BindJSON(&cart); err != nil {
		app.badRequestError(c, err)
		return
	}

	user, err := GetUser(c)
	if err != nil {
		app.internalServerError(c, err)
		return
	}
	cart.UserId = user.UserID

	if err := app.models.Cart.Insert(cart); err != nil {
		app.internalServerError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"cart": cart})
}
func (app *application) getCartHandler(c *gin.Context) {
	id := ReadIdParam(c)
	if id.IsZero() {
		app.badRequestError(c, models.ErrInvalidID)
		return
	}

	cart, err := app.models.Cart.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrNotFound):
			app.notFoundError(c)
		default:
			app.internalServerError(c, err)
		}
		return
	}

	user, err := GetUser(c)
	if err != nil {
		app.internalServerError(c, err)
		return
	}
	if cart.UserId != user.UserID {
		app.notAuthorizedError(c)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"cart": cart})
}
func (app *application) updateCartHandler(c *gin.Context) {
	id := ReadIdParam(c)
	if id.IsZero() {
		app.badRequestError(c, fmt.Errorf("bad id"))
		return
	}

	cart, err := app.models.Cart.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrNotFound):
			app.notFoundError(c)
		default:
			app.internalServerError(c, err)
		}
		return
	}

	user, err := GetUser(c)
	if err != nil {
		app.internalServerError(c, err)
		return
	}
	if cart.UserId != user.UserID {
		app.notAuthorizedError(c)
		return
	}

	payload := &models.Cart{}
	if err := c.BindJSON(&payload); err != nil {
		app.badRequestError(c, err)
		return
	}

	cart.Products = payload.Products

	if err := app.models.Cart.Update(cart); err != nil {
		app.internalServerError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"cart": cart})
}

func (app *application) deleteCartHandler(c *gin.Context) {
	id := ReadIdParam(c)
	if id.IsZero() {
		app.badRequestError(c, fmt.Errorf("bad id"))
		return
	}

	user, err := GetUser(c)
	if err != nil {
		app.notAuthenticatedError(c)
		return
	}
	if err := app.models.Cart.Delete(id, user.UserID); err != nil {
		app.internalServerError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"msg": "success"})
}
