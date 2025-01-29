package main

import (
	"errors"
	"net/http"

	"github.com/GiorgosMarga/ecom_go/internal/validator"
	"github.com/GiorgosMarga/ecom_go/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (app *application) registerOrderRoutes(router *gin.Engine) {
	v1 := router.Group("/api/v1/orders")
	v1.POST("/", app.authenticateUser(), app.createOrderHandler)
	v1.GET("/:id", app.authenticateUser(), app.getOrderHandler)
	v1.DELETE("/:id", app.authenticateUser(), app.authorizeUser(), app.deleteOrderHandler)
	v1.PATCH("/:id", app.authenticateUser(), app.updateOrderHandler)
}

func (app *application) createOrderHandler(c *gin.Context) {
	var order models.Order

	if err := c.BindJSON(&order); err != nil {
		app.badRequestError(c, err)
		return
	}

	user, err := GetUser(c)
	if err != nil {
		app.internalServerError(c, err)
		return
	}

	idToQuantity := make(map[primitive.ObjectID]int)
	for _, p := range order.Products {
		idToQuantity[p.ProductId] += p.Quantity
	}

	total, err := app.models.Product.GetPriceForOrder(idToQuantity)
	if err != nil {
		app.internalServerError(c, err)
		return
	}
	order.UserId = user.UserID
	order.Total = total
	if err := app.models.Order.Insert(&order); err != nil {
		app.internalServerError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"order": order})

}

func (app *application) getOrderHandler(c *gin.Context) {
	orderId := ReadIdParam(c)
	user, err := GetUser(c)
	if err != nil {
		app.internalServerError(c, err)
		return
	}
	order, err := app.models.Order.Get(orderId)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrNotFound):
			app.notFoundError(c)
		default:
			app.internalServerError(c, err)
		}
		return
	}

	// if the order doesnt belong to the user who made the request return notAuthorized
	// if the user is an admin then return the order
	if user.UserID != order.UserId && user.Role != models.GetRole(models.AdminRole) {
		app.notAuthorizedError(c)
		return
	}

	c.JSON(http.StatusOK, gin.H{"order": order})

}

// only admins can delete an order. Deleting an order is different than canceling it
// if you want to cancel and order user updateOrderHandler
func (app *application) deleteOrderHandler(c *gin.Context) {
	orderId := ReadIdParam(c)
	err := app.models.Order.Delete(orderId)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrNotFound):
			app.notFoundError(c)
		default:
			app.internalServerError(c, err)
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "success"})

}

func (app *application) updateOrderHandler(c *gin.Context) {
	orderId := ReadIdParam(c)
	user, err := GetUser(c)
	if err != nil {
		app.internalServerError(c, err)
		return
	}

	order, err := app.models.Order.Get(orderId)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrNotFound):
			app.notFoundError(c)
		default:
			app.internalServerError(c, err)
		}
		return
	}

	// if the order doesnt belong to the user who made the request return notAuthorized
	// but if the user is an admin then continue
	if order.UserId != user.UserID && user.Role != models.GetRole(models.AdminRole) {
		app.notAuthorizedError(c)
		return
	}

	var payload models.OrderUpdatePayload

	if err := c.Bind(&payload); err != nil {
		app.badRequestError(c, err)
		return
	}

	v := validator.NewValidator()

	if models.ValidateOrderUpdatePayload(v, payload); !v.IsValid() {
		app.failedValidationError(c, v.Errors)
		return
	}

	if len(payload.Products) != 0 {
		order.Products = payload.Products
	}

	if payload.Status != nil {
		order.Status = *payload.Status
	}

	err = app.models.Order.Update(order)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrNotFound):
			app.notFoundError(c)
		default:
			app.internalServerError(c, err)
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"order": order})

}
