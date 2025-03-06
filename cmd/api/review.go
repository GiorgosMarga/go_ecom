package main

import (
	"errors"
	"net/http"

	"github.com/GiorgosMarga/ecom_go/internal/validator"
	"github.com/GiorgosMarga/ecom_go/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (app *application) registerReviewRoutes(router *gin.Engine) {
	v1 := router.Group("/api/v1/review")
	v1.POST("", app.authenticateUser(), app.createReviewHandler)
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

	var review models.Review

	if err := c.BindJSON(&review); err != nil {
		app.badRequestError(c, err)
		return
	}

	review.UserID = user.UserID

	v := validator.NewValidator()

	if models.ValidateReview(v, review); !v.IsValid() {
		app.failedValidationError(c, v.Errors)
		return
	}

	if err := app.models.Review.Insert(&review); err != nil {
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

	id := c.Params.ByName("id")

	review, err := app.models.Review.GetByID(id)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrNotFound):
			app.notFoundError(c)
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
	if payload.Content != nil {
		review.Content = *payload.Content
	}
	// validate payload

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

	id := c.Params.ByName("id")
	// validate payload
	reviews, err := app.models.Review.GetForProduct(id)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrNotFound):
			app.notFoundError(c)
		default:
			app.internalServerError(c, err)
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"reviews": reviews})

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
