package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/GiorgosMarga/ecom_go/internal/validator"
	"github.com/GiorgosMarga/ecom_go/models"
	"github.com/gin-gonic/gin"
)

func (app *application) registerProductRoutes(router *gin.Engine) {
	v1 := router.Group("/api/v1/products")
	v1.POST("/", app.authenticateUser(), app.authorizeUser(), app.createProductHandler)
	v1.GET("/:id", app.getProductHandler)
	v1.DELETE("/:id", app.authenticateUser(), app.authorizeUser(), app.deleteProductHandler)
	v1.PATCH("/:id", app.authenticateUser(), app.authorizeUser(), app.updateProductHandler)
	// v1.GET("/:id", app.getUserByIdHandler)
}

func (app *application) createProductHandler(c *gin.Context) {
	var payload models.Product

	if err := c.BindJSON(&payload); err != nil {
		app.badRequestError(c, err)
		return
	}

	// TODO: validate payload
	product := models.Product{
		Description: payload.Description,
		Price:       payload.Price,
		Stock:       payload.Stock,
	}

	v := validator.NewValidator()
	if models.ValidateProduct(v, product); !v.IsValid() {
		app.failedValidationError(c, v.Errors)
		return
	}

	err := app.models.Product.Insert(&product)
	if err != nil {
		app.internalServerError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"product": product})

}

func (app *application) getProductHandler(c *gin.Context) {
	hexID := c.Param("id")
	product, err := app.models.Product.GetById(hexID)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrInvalidID):
			app.badRequestError(c, err)
		case errors.Is(err, models.ErrNotFound):
			app.notFoundError(c)
		default:
			app.internalServerError(c, err)
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"product": product})
}

func (app *application) deleteProductHandler(c *gin.Context) {
	hexID := c.Param("id")
	err := app.models.Product.Delete(hexID)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrInvalidID):
			app.badRequestError(c, err)
		case errors.Is(err, models.ErrNotFound):
			app.notFoundError(c)
		default:
			app.internalServerError(c, err)
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "success"})
}

func (app *application) updateProductHandler(c *gin.Context) {
	hexID := c.Param("id")
	fmt.Println(hexID)
	product, err := app.models.Product.GetById(hexID)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrNotFound):
			app.notFoundError(c)
		default:
			app.internalServerError(c, err)
		}
		return
	}

	productPayload := models.ProductUpdatePayload{}

	if err := c.BindJSON(&productPayload); err != nil {
		app.badRequestError(c, err)
		return
	}

	v := validator.NewValidator()

	if models.ValidateUpdatePayload(v, productPayload); !v.IsValid() {
		app.failedValidationError(c, v.Errors)
		return
	}

	if productPayload.Description != nil {
		product.Description = *productPayload.Description
	}
	if productPayload.Price != nil {
		product.Price = *productPayload.Price
	}
	if productPayload.Stock != nil {
		product.Stock = *productPayload.Stock
	}

	err = app.models.Product.Update(product)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrNotFound):
			app.notFoundError(c)
		default:
			app.internalServerError(c, err)
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"product": product})
}
