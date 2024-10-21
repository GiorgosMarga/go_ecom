package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

func (app *application) run() error {
	gin.SetMode(app.cfg.ginMode)
	r := gin.Default()
	app.registerUserRoutes(r)
	app.registerProductRoutes(r)
	app.registerCartRoutes(r)
	app.registerTokenRoutes(r)
	app.registerOrderRoutes(r)
	fmt.Printf("Server is listening on port %s\n", app.cfg.port)
	return r.Run(fmt.Sprintf(":%s", app.cfg.port))
}
