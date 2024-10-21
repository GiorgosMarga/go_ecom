package main

import (
	"net/http"

	"github.com/GiorgosMarga/ecom_go/models"
	"github.com/gin-gonic/gin"
)

func (app *application) registerOrderRoutes(router *gin.Engine) {
	v1 := router.Group("/api/v1/orders")
	v1.POST("/", app.createOrderHandler)
	// v1.GET("/:id", app.getProductHandler)
	// v1.DELETE("/:id", app.authenticateUser(), app.authorizeUser(), app.deleteProductHandler)
	// v1.PATCH("/:id", app.authenticateUser(), app.authorizeUser(), app.updateProductHandler)
	// v1.GET("/:id", app.getUserByIdHandler)
}

func (app *application) createOrderHandler(c *gin.Context) {
	var order models.Order

	if err := c.BindJSON(&order); err != nil {
		app.badRequestError(c, err)
		return
	}

	if err := app.models.Order.Insert(&order); err != nil {
		app.internalServerError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"order": order})

}
