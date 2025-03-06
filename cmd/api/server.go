package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

func (app *application) run() error {
	gin.SetMode(app.cfg.ginMode)
	r := gin.Default()
	r.Use(app.corsMiddleware())
	app.registerProductRoutes(r)
	app.registerUserRoutes(r)
	app.registerCartRoutes(r)
	app.registerTokenRoutes(r)
	app.registerOrderRoutes(r)
	app.registerReviewRoutes(r)
	app.registerVariantsRoutes(r)
	app.registerPaymentRoutes(r)
	fmt.Printf("Server is listening on port %s\n", app.cfg.port)
	return r.Run(fmt.Sprintf(":%s", app.cfg.port))
}
