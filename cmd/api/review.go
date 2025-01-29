package main

import (
	"errors"
	"net/http"

	"github.com/GiorgosMarga/ecom_go/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (app *application) registerReviewRoutes(router *gin.Engine) {
	v1 := router.Group("/api/v1/review")
	v1.POST("/", app.authenticateUser(), app.createReviewHandler)
	v1.PATCH("/:id", app.authenticateUser(), app.updateReviewHandler)
	v1.GET("/:id", app.getReviewHandler)
	v1.DELETE("/:id", app.authenticateUser(), app.deleteReviewHandler)
}

func (app *application) createReviewHandler(c *gin.Context) {
	user, err := GetUser(c)
	if err != nil {
		app.internalServerError(c, err)
		return
	}

	var payload models.ReviewPayload

	if err := c.BindJSON(&payload); err != nil {
		app.badRequestError(c, err)
		return
	}
	// validate payload

	review := models.NewReview(user.UserID, payload.ProductID, payload.Content)

	if err := app.models.Review.Insert(review); err != nil {
		app.internalServerError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"review": review})

}

func (app *application) updateReviewHandler(c *gin.Context) {
	user, err := GetUser(c)
	if err != nil {
		app.internalServerError(c, err)
		return
	}

	id := ReadIdParam(c)
	if id == primitive.NilObjectID {
		app.badRequestError(c, models.ErrInvalidID)
		return
	}

	review, err := app.models.Review.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrNotFound):
			app.notFoundError(c)
			return
		default:
			app.internalServerError(c, err)
		}
		return
	}

	// only the user who created the review OR an admin can update a review
	if review.UserID != user.UserID && user.Role != models.GetRole(models.AdminRole) {
		app.notAuthorizedError(c)
		return
	}

	var payload models.ReviewUpdatePayload

	if err := c.BindJSON(&payload); err != nil {
		app.badRequestError(c, err)
		return
	}
	// validate payload
	if payload.Content != nil {
		review.Content = *payload.Content
	}

	if err := app.models.Review.Update(review); err != nil {
		switch {
		case errors.Is(err, models.ErrNotFound):
			app.notFoundError(c)
		default:
			app.internalServerError(c, err)
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"review": review})

}

func (app *application) getReviewHandler(c *gin.Context) {

	id := ReadIdParam(c)
	if id == primitive.NilObjectID {
		app.badRequestError(c, ErrInvalidJWT)
		return
	}
	// validate payload
	review, err := app.models.Review.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrNotFound):
			app.notFoundError(c)
		default:
			app.internalServerError(c, err)
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"review": review})

}

func (app *application) deleteReviewHandler(c *gin.Context) {
	user, err := GetUser(c)
	if err != nil {
		app.internalServerError(c, err)
		return
	}

	id := ReadIdParam(c)
	if id == primitive.NilObjectID {
		app.badRequestError(c, models.ErrInvalidID)
		return
	}

	if err := app.models.Review.Delete(id, user); err != nil {
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
