package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/GiorgosMarga/ecom_go/models"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
)

func (app *application) registerVariantsRoutes(router *gin.Engine) {
	v1 := router.Group("/api/v1/variants")
	v1.POST("", app.authenticateUser(), app.createVariantHandler)
	// v1.POST("/login", app.loginUserHandler)
	// v1.PATCH("/:id", app.authenticateUser(), app.updateUserHandler)
	// v1.GET("/:id", app.getUserByIdHandler)
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
	var variant models.ProductVariant

	if err := json.Unmarshal([]byte(jsonData), &variant); err != nil {
		app.badRequestError(c, err)
		return
	}
	variant.Img = imagePaths
	if err := app.models.Variant.Insert(&variant); err != nil {
		app.internalServerError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"v": variant})
}
