package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/GiorgosMarga/ecom_go/internal/validator"
	"github.com/GiorgosMarga/ecom_go/models"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
)

func (app *application) registerProductRoutes(router *gin.Engine) {
	v1 := router.Group("/api/v1/products")
	v1.POST("", app.authenticateUser(), app.authorizeUser(), app.createProductHandler)
	v1.GET("/:id", app.getProductHandler)
	v1.DELETE("/:id", app.authenticateUser(), app.authorizeUser(), app.deleteProductHandler)
	v1.PATCH("/:id", app.authenticateUser(), app.authorizeUser(), app.updateProductHandler)
	// v1.GET("/:id", app.getUserByIdHandler)
}

func (app *application) createProductHandler(c *gin.Context) {
	if err := c.Request.ParseMultipartForm(32 << 20); err != nil { // 32 MB max memory
		app.badRequestError(c, err)
		return
	}
	file, header, err := c.Request.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Image is required"})
		return
	}
	defer file.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	result, err := app.uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket: &app.cfg.bucket,
		Key:    aws.String(header.Filename),
		Body:   file,
	})
	if err != nil {
		app.internalServerError(c, err)
		return
	}
	jsonData := c.PostForm("data")
	var product models.Product

	if err := json.Unmarshal([]byte(jsonData), &product); err != nil {
		app.badRequestError(c, err)
		return
	}

	v := validator.NewValidator()
	if models.ValidateProduct(v, product); !v.IsValid() {
		app.failedValidationError(c, v.Errors)
		return
	}
	product.Img = result.Location
	err = app.models.Product.Insert(&product)
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

	// By creating a product payload we allow only specific fields to be updated.
	// For example, if we dont include the price in the product payload, then even
	// if the user sends the price, we will just ignore it
	productPayload := models.ProductUpdatePayload{}
	if err := c.BindJSON(&productPayload); err != nil {
		app.badRequestError(c, err)
		return
	}

	if productPayload.Description != nil {
		product.Description = *productPayload.Description
	}

	if productPayload.Price != nil {
		product.Price = *productPayload.Price
	}

	if productPayload.Rating != nil {
		product.Rating = *productPayload.Rating
	}

	if productPayload.Stock != nil {
		product.Stock = *productPayload.Stock
	}

	if productPayload.Name != nil {
		product.Name = *productPayload.Name
	}

	if productPayload.Tags != nil {
		product.Tags = *productPayload.Tags
	}
	v := validator.NewValidator()

	if models.ValidateProduct(v, *product); !v.IsValid() {
		app.failedValidationError(c, v.Errors)
		return
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
