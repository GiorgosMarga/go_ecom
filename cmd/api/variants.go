package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/GiorgosMarga/ecom_go/internal/validator"
	"github.com/GiorgosMarga/ecom_go/models"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
)

func (app *application) registerVariantsRoutes(router *gin.Engine) {
	v1 := router.Group("/api/v1/variants")
	v1.POST("", app.authenticateUser(), app.createVariantHandler)
	// v1.POST("/login", app.loginUserHandler)
	v1.GET("/:id", app.authenticateUser(), app.getVariantHandler)
	v1.DELETE("/:id", app.deleteVariantHandler)
}
func (app *application) getVariantHandler(c *gin.Context) {
	id := c.Params.ByName("id")
	pv, err := app.models.Variant.GetById(id)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrNotFound):
			app.notFoundError(c)
		default:
			app.internalServerError(c, err)
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"v": pv})

}
func (app *application) createVariantHandler(c *gin.Context) {
	if err := c.Request.ParseMultipartForm(32 << 20); err != nil { // 32 MB max memory
		fmt.Println(err)
		app.badRequestError(c, err)
		return
	}

	imagePaths := make([]string, 0)

	form, err := c.MultipartForm()
	if err != nil {
		app.internalServerError(c, err)
		fmt.Println(err)
		return
	}
	files := form.File["images"]
	for _, file := range files {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		f, err := file.Open()
		if err != nil {
			f.Close()
			app.internalServerError(c, err)
			return
		}
		result, err := app.uploader.Upload(ctx, &s3.PutObjectInput{
			Bucket: &app.cfg.bucket,
			Key:    aws.String(file.Filename),
			Body:   f,
		})
		if err != nil {
			app.internalServerError(c, err)
			return
		}
		imagePaths = append(imagePaths, result.Location)
		f.Close()
	}

	jsonData := c.PostForm("data")
	var variant models.Variant

	if err := json.Unmarshal([]byte(jsonData), &variant); err != nil {
		app.badRequestError(c, err)
		return
	}
	variant.Img = imagePaths

	v := validator.NewValidator()
	if models.ValidateVariant(v, variant); !v.IsValid() {
		app.failedValidationError(c, v.Errors)
		return
	}

	if err := app.models.Variant.Insert(&variant); err != nil {
		app.internalServerError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"v": variant})
}

func (app *application) deleteVariantHandler(c *gin.Context) {
	variantId := c.Params.ByName("id")

	if err := app.models.Variant.Delete(variantId); err != nil {
		switch {
		case errors.Is(err, models.ErrNotFound):
			app.notFoundError(c)
		case errors.Is(err, models.ErrInvalidID):
			app.badRequestError(c, err)
		default:
			app.internalServerError(c, err)
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})

}
