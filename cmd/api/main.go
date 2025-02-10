package main

import (
	"context"
	"log"
	"os"

	"github.com/GiorgosMarga/ecom_go/models"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type application struct {
	cfg      *config
	logger   *log.Logger
	models   models.Models
	uploader *manager.Uploader
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

	awsCfg, err := awsConfig.LoadDefaultConfig(context.TODO(), awsConfig.WithRegion("eu-north-1"))
	if err != nil {
		log.Fatal(err)
	}
	logger.Println("Successfully conneected to the S3")

	uploader := manager.NewUploader(s3.NewFromConfig(awsCfg))

	app := &application{
		cfg:      cfg,
		logger:   logger,
		models:   models.NewModels(db),
		uploader: uploader,
	}
	if err := app.run(); err != nil {
		log.Fatal(err)
	}
}
