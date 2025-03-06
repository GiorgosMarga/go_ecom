package main

import (
	"fmt"

	"github.com/GiorgosMarga/ecom_go/models"
	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/paymentintent"
)

func (app *application) registerPaymentRoutes(router *gin.Engine) {
	v1 := router.Group("/api/v1/payments")
	v1.POST("", app.authenticateUser(), app.createPaymentIntentHandler)
	// v1.GET("/:id", app.getProductHandler)
	// v1.DELETE("/:id", app.authenticateUser(), app.authorizeUser(), app.deleteProductHandler)
	// v1.PATCH("/:id", app.authenticateUser(), app.authorizeUser(), app.updateProductHandler)
	// v1.GET("/:id", app.getUserByIdHandler)
}

func (app *application) createPaymentIntentHandler(c *gin.Context) {

	user, err := GetUser(c)
	if err != nil {
		app.internalServerError(c, err)
		return
	}
	var order models.Order
	if err := c.BindJSON(&order); err != nil {
		app.badRequestError(c, err)
		return
	}
	// TODO: Validate order
	// v := validator.NewValidator()
	// if err := models.ValidateOrder(order); err != nil {
	// 	app.failedValidationError(c, v.Errors)
	// 	return
	// }

	amount, err := app.models.Variant.GetTotalPrice(&order)
	if err != nil {
		app.internalServerError(c, err)
		return
	}

	params := &stripe.PaymentIntentParams{
		Amount:   stripe.Int64(int64(amount)),
		Currency: stripe.String(string(stripe.CurrencyUSD)),
		AutomaticPaymentMethods: &stripe.PaymentIntentAutomaticPaymentMethodsParams{
			Enabled: stripe.Bool(true),
		},
	}

	pi, err := paymentintent.New(params)
	if err != nil {
		if stripeErr, ok := err.(*stripe.Error); ok {
			fmt.Println("Stripe error:", stripeErr)
		}
		app.internalServerError(c, err)
		return
	}
	order.UserId = user.UserID
	order.PaymentIntentId = pi.ID
	order.Total = amount
	if err := app.models.Order.Insert(&order); err != nil {
		app.internalServerError(c, err)
		return
	}
	c.JSON(200, gin.H{"client_secret": pi.ClientSecret, "total": amount})
}
