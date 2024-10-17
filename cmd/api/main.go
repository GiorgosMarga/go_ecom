package main

import (
	"context"
	"log"
	"os"

	"github.com/GiorgosMarga/ecom_go/models"
)

type application struct {
	cfg    *config
	logger *log.Logger
	models models.Models
}

func main() {
	cfg := NewConfig()
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)
	db, err := connectDB(*cfg)
	if err != nil {
		logger.Fatal(err)
	}
	defer func() {
		if err := db.Client().Disconnect(context.Background()); err != nil {
			log.Fatal(err)
		}
	}()
	logger.Println("Successfully conneected to the mongo DB")
	app := &application{
		cfg:    cfg,
		logger: logger,
		models: models.NewModels(db),
	}
	if err := app.run(); err != nil {
		log.Fatal(err)
	}
}
